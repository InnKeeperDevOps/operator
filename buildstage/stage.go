package buildstage

import (
	"context"
	"errors"
	"github.com/InnKeeperDevOps/operator/api/v1alpha1"
	v12 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

type BuildStage struct {
	client.Client
	Deploy *v1alpha1.BuildDeploy
}

const latestBuilderImage = "ghcr.io/innkeeperdevops/git-buildah:main"

func (b *BuildStage) LatestBuilt() bool {
	if b.Deploy.Status.Built != nil {
		if b.Deploy.Spec.Git.Commit != "" && b.Deploy.Status.Built.Git.Commit != b.Deploy.Spec.Git.Commit {
			println("COMMIT DIFF")
			return false
		}
		if b.Deploy.Spec.Git.Branch != "" && b.Deploy.Status.Built.Git.Branch != b.Deploy.Spec.Git.Branch {
			println("BRANCH DIFF")
			return false
		}
		if b.Deploy.Spec.Publish.Version != "" && b.Deploy.Status.Built.Version != b.Deploy.Spec.Publish.Version {
			println("VERSION DIFF")
			return false
		}
	} else {
		return false
	}
	return true
}

func (b *BuildStage) LatestDeployed(ctx context.Context) bool {
	deploymentExists := &v12.Deployment{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.Name, Namespace: b.Deploy.Spec.Deploy.Namespace}, deploymentExists)
	if err != nil {
		return false
	}
	if b.Deploy.Status.Deployed != nil {
		if b.Deploy.Spec.Git.Commit != "" && b.Deploy.Status.Deployed.Git.Commit != b.Deploy.Spec.Git.Commit {
			return false
		}
		if b.Deploy.Spec.Git.Branch != "" && b.Deploy.Status.Deployed.Git.Branch != b.Deploy.Spec.Git.Branch {
			return false
		}
		if b.Deploy.Spec.Publish.Version != "" && b.Deploy.Status.Deployed.Version != b.Deploy.Spec.Publish.Version {
			return false
		}
	} else {
		return false
	}
	return true
}

const BUILD_STAGE = "BUILD"
const DEPLOY_STAGE = "DEPLOY"
const SLEEP_STAGE = "SLEEP"

func match(string string, regex string) string {
	r, _ := regexp.Compile(regex)
	return r.FindStringSubmatch(string)[1]
}
func (b *BuildStage) Builder(ctx context.Context) (ctrl.Result, error) {
	jobExists := &batchv1.Job{}
	err := b.Client.Get(ctx, types.NamespacedName{Name: b.Deploy.GetBuilderName(), Namespace: "git-builder"}, jobExists)
	if err != nil {
		println(err.Error())
		err = b.Client.Create(ctx, b.createBuildJob(b.Deploy))

		if err != nil {
			println(err.Error())
		}
		if b.Deploy.Status.Built != nil {
			b.Deploy.Status.Built.Complete = false
			b.Status().Update(ctx, b.Deploy)
		}
		return ctrl.Result{RequeueAfter: time.Second * 25}, nil
	} else {
		if jobExists.Status.Succeeded == 0 {
			println("Waiting for image to build. [" + b.Deploy.Name + "]")
			return ctrl.Result{RequeueAfter: time.Second * 25}, nil
		}

		// continue to next stage.
		println("Builder completed making image")
		podList := &v1.PodList{}
		selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
			MatchLabels: map[string]string{"buildFor": b.Deploy.GetBuilderName()},
		})
		err = b.Client.List(ctx, podList, client.MatchingLabelsSelector{
			Selector: selector,
		})
		if err == nil {
			if len(podList.Items) > 0 {
				pod := podList.Items[0]
				println(pod.Name)
				log := b.getPodLogs(ctx, pod)
				branch := match(log, "branch:([a-zA-Z0-9/]+)")
				hash := match(log, "hash:(.*?),")
				author := match(log, "author:(.*?),")
				date, _ := strconv.Atoi(match(log, "date:([0-9]+)"))
				b.Deploy.Status = v1alpha1.BuildDeployStatus{
					Built: &v1alpha1.Built{
						Registry: v1alpha1.RegistryStatus{
							Tag:  b.Deploy.Spec.Publish.Tag,
							Host: b.Deploy.Spec.Publish.Host,
						},
						Git: v1alpha1.GitStatus{
							Commit: hash,
							Branch: branch,
							Date:   date,
							Author: author,
						},
						Version:  b.Deploy.Spec.Publish.Version,
						Complete: true,
					},
				}
				policy := metav1.DeletePropagationForeground
				err = b.Delete(ctx, jobExists, &client.DeleteOptions{PropagationPolicy: &policy})
				if err != nil {
					println(err.Error())
				}
				err := b.Status().Update(context.Background(), b.Deploy)
				if err != nil {
					println(err.Error())
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}
		} else {
			print(err.Error())
		}

	}

	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}

func (b *BuildStage) PullSecret(ctx context.Context, namespace string) error {
	// check docker pull secret exists for this app
	pullSecret := &v1.Secret{}
	err := b.Get(ctx, types.NamespacedName{Name: "docker-pull-" + b.Deploy.Spec.Publish.Host, Namespace: namespace}, pullSecret)
	if err != nil {
		publishSecret := &v1.Secret{}
		err = b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Publish.Secret, Namespace: "git-builder"}, publishSecret)
		if err != nil {
			println(err.Error())
			return errors.New("Could not get data from git-builder secret")
		}
		username := string(publishSecret.Data["username"])
		password := string(publishSecret.Data["password"])
		pullSecret = b.CreatePullSecret(b.Deploy, username, password, namespace)
		err = b.Create(ctx, pullSecret)
		if err != nil {
			return err
		}
		println("Created pull secret in the namespace " + namespace)
	}
	return nil
}

func (b *BuildStage) HandleDaemonSet(ctx context.Context) error {
	daemonSetExists := &v12.DaemonSet{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.CronJob.Name, Namespace: b.Deploy.Spec.Deploy.CronJob.Namespace}, daemonSetExists)
	if err == nil {
		println("Updating daemonSet, " + daemonSetExists.Name)
		err = b.MergeDaemonSetToSpec(daemonSetExists)
		if err != nil {
			return err
		}
		err = b.UpdateDaemonSet(ctx, daemonSetExists)
	} else {
		println("Creating new daemonSet,  " + b.Deploy.Name)
		err = b.CreateDaemonSet(ctx)
	}
	return err
}

func (b *BuildStage) HandleCronJob(ctx context.Context) error {
	cronJobExists := &batchv1.CronJob{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.CronJob.Name, Namespace: b.Deploy.Spec.Deploy.CronJob.Namespace}, cronJobExists)
	if err == nil {
		println("Updating cronjob, " + cronJobExists.Name)
		err = b.MergeCronJobToSpec(cronJobExists)
		if err != nil {
			return err
		}
		err = b.UpdateCronJob(ctx, cronJobExists)
	} else {
		println("Creating new cronjob,  " + b.Deploy.Name)
		err = b.CreateCronJob(ctx)
	}
	return err
}

func (b *BuildStage) HandleDeployment(ctx context.Context) error {
	deploymentExists := &v12.Deployment{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.Deployment.Name, Namespace: b.Deploy.Spec.Deploy.Deployment.Namespace}, deploymentExists)
	if err == nil {
		println("Updating deployment, " + deploymentExists.Name)
		err = b.MergeDeploymentToSpec(deploymentExists)
		if err != nil {
			return err
		}
		err = b.UpdateDeployment(ctx, deploymentExists)
	} else {
		println("Creating new deployment,  " + b.Deploy.Name)
		err = b.CreateDeployment(ctx)
	}
	return err
}

func (b *BuildStage) HandleSimpleDeployment(ctx context.Context) error {
	deploymentExists := &v12.Deployment{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.Name, Namespace: b.Deploy.Spec.Deploy.Namespace}, deploymentExists)
	if err == nil {
		println("Updating simple deployment, " + deploymentExists.Name)
		b.UpdateDeploymentToSpec(deploymentExists)
		err = b.UpdateDeployment(ctx, deploymentExists)
		if err != nil {
			return err
		}
	} else {
		println("Creating new simple deployment,  " + b.Deploy.Name)
		deployment := b.CreateSimpleDeployment()
		err = b.Create(ctx, deployment)

	}
	return err
}

func (b *BuildStage) Deployer(ctx context.Context) (ctrl.Result, error) {
	println("In Deployer Stage")

	namespace := "default"
	if b.Deploy.Spec.Deploy != nil {
		if b.Deploy.Spec.Deploy.Namespace != "" {
			namespace = b.Deploy.Spec.Deploy.Namespace
		}
		println("Deploying to namespace: " + namespace)
		err := b.Create(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})
		if err != nil {
			println(err.Error())
		}
		if b.Deploy.Spec.Publish.Secret != "" {
			err = b.PullSecret(ctx, namespace)
			if err != nil {
				println(err.Error())
				return ctrl.Result{RequeueAfter: time.Second * 20}, err
			}
		}

		switch {
		case b.Deploy.Spec.Deploy.Deployment != nil:
			println("Handling " + b.Deploy.Name + " as deployment")
			err = b.HandleDeployment(ctx)
			break
		case b.Deploy.Spec.Deploy.CronJob != nil:
			println("Handling " + b.Deploy.Name + " as cronjob")
			err = b.HandleCronJob(ctx)
			break
		case b.Deploy.Spec.Deploy.DaemonSet != nil:
			println("Handling " + b.Deploy.Name + " as daemonset")
			err = b.HandleDaemonSet(ctx)
			break
		default:
			println("Handling " + b.Deploy.Name + " as simple deployment")
			err = b.HandleSimpleDeployment(ctx)
			break
		}

		if err != nil {
			println(err.Error())
			return ctrl.Result{RequeueAfter: time.Second * 20}, nil
		} else {
			b.Deploy.Status.Deployed = &v1alpha1.Deployed{
				Pod:      "",
				Git:      b.Deploy.Status.Built.Git,
				Complete: true,
				Registry: b.Deploy.Status.Built.Registry,
				Version:  b.Deploy.Status.Built.Version,
			}
			err = b.Status().Update(ctx, b.Deploy)
			if err != nil {
				println(err.Error())
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}

func (b *BuildStage) Route(ctx context.Context) (ctrl.Result, error) {
	switch b.GetStage(ctx) {
	case BUILD_STAGE:
		return b.Builder(ctx)
	case DEPLOY_STAGE:
		return b.Deployer(ctx)
	case SLEEP_STAGE:
		println("Waiting for buildDeploy changes on " + b.Deploy.Name)
		return ctrl.Result{RequeueAfter: time.Minute * 60}, nil
	}
	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}

func (b *BuildStage) GetStage(ctx context.Context) string {
	if !b.LatestBuilt() {
		return BUILD_STAGE
	}
	if !b.LatestDeployed(ctx) {
		return DEPLOY_STAGE
	}
	return SLEEP_STAGE
}
