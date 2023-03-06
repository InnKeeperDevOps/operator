package buildstage

import (
	"context"
	"errors"
	"github.com/InnKeeperDevOps/operator/api/v1alpha1"
	"github.com/InnKeeperDevOps/operator/buildstage/changelog"
	log "github.com/sirupsen/logrus"
	v12 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type BuildStage struct {
	client.Client
	Deploy  *v1alpha1.BuildDeploy
	Changes changelog.Changes
	Deletes changelog.Changes
}

const LatestBuilderImage = "ghcr.io/innkeeperdevops/git-builder:28"
const LatestGitImage = "ghcr.io/innkeeperdevops/git-latest:2"

func (b *BuildStage) LatestBuilt() bool {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	if b.Deploy.Status.Built != nil {
		if (b.Deploy.Spec.Git.Commit != "" && b.Deploy.Status.Built.Git.Commit != b.Deploy.Spec.Git.Commit) || b.Deploy.Status.Built.Git.Commit == "" {
			log.Debug(name + ": COMMIT DIFF")
			return false
		}
		if b.Deploy.Spec.Git.Branch != "" && b.Deploy.Status.Built.Git.Branch != b.Deploy.Spec.Git.Branch {
			log.Debug(name + ": BRANCH DIFF")
			return false
		}
	} else {
		return false
	}
	return true
}

func remove(arr []string, val string) []string {
	newArr := []string{}
	for _, s := range arr {
		if s != val {
			newArr = append(newArr, s)
		}
	}
	return newArr
}

func (b *BuildStage) LatestDeployed(ctx context.Context) bool {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	deploymentExists := &v12.Deployment{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.Name, Namespace: b.Deploy.Spec.Deploy.Namespace}, deploymentExists)
	if err != nil {
		return false
	}
	if b.Deploy.Status.Deployed != nil {
		if b.Deploy.Spec.Deploy.Deployment.Spec.Replicas != nil && *b.Deploy.Spec.Deploy.Deployment.Spec.Replicas != *deploymentExists.Spec.Replicas {
			log.Debug(name + ": REPLICA DIFF")
			return false
		}
		if b.Deploy.Spec.Git.Commit != "" && b.Deploy.Status.Deployed.Git.Commit != b.Deploy.Spec.Git.Commit || b.Deploy.Status.Deployed.Git.Commit == "" {
			log.Debug(name + ": COMMIT DIFF")
			return false
		}
		if b.Deploy.Spec.Git.Branch != "" && b.Deploy.Status.Deployed.Git.Branch != b.Deploy.Spec.Git.Branch {
			log.Debug(name + ": BRANCH DIFF")
			return false
		}
	} else {
		return false
	}
	return true
}

const BUILD_STAGE = "BUILD"
const DEPLOY_STAGE = "DEPLOY"
const INGRESS_STAGE = "INGRESS"
const CHECK_LATEST = "LATEST"
const DESTROY_STAGE = "DESTROY"
const SLEEP = "SLEEP"

func match(string string, regex string) string {
	r, _ := regexp.Compile(regex)
	return r.FindStringSubmatch(string)[1]
}

func (b *BuildStage) PullSecret(ctx context.Context, namespace string) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	// check docker pull secret exists for this app
	pullSecret := &v1.Secret{}
	err := b.Get(ctx, types.NamespacedName{Name: "docker-pull-" + b.Deploy.Spec.Publish.Host, Namespace: namespace}, pullSecret)
	if err != nil {
		publishSecret := &v1.Secret{}
		err = b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Publish.Secret, Namespace: "git-builder"}, publishSecret)
		if err != nil {
			log.Error(name + ": " + err.Error())
			return errors.New(name + ": Could not get data from git-builder secret " + b.Deploy.Spec.Publish.Secret)
		}
		username := string(publishSecret.Data["username"])
		password := string(publishSecret.Data["password"])
		pullSecret = b.CreatePullSecret(b.Deploy, username, password, namespace)
		err = b.Create(ctx, pullSecret)
		if err != nil {
			return err
		}
		log.Debug(name + ": Created pull secret in the namespace " + namespace)
	}
	log.Debug(name + ": Pull secret exists in the namespace " + namespace)
	return nil
}

func (b *BuildStage) HandleDaemonSet(ctx context.Context) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	daemonSetExists := &v12.DaemonSet{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.CronJob.Name, Namespace: b.Deploy.Spec.Deploy.CronJob.Namespace}, daemonSetExists)
	if err == nil {
		log.Debug(name + ": Updating daemonSet, " + daemonSetExists.Name)
		err = b.MergeDaemonSetToSpec(daemonSetExists)
		if err != nil {
			return err
		}
		err = b.UpdateDaemonSet(ctx, daemonSetExists)
	} else {
		log.Debug(name + ": Creating new daemonSet,  " + b.Deploy.Name)
		err = b.CreateDaemonSet(ctx)
	}
	return err
}

func (b *BuildStage) HandleCronJob(ctx context.Context) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	cronJobExists := &batchv1.CronJob{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.CronJob.Name, Namespace: b.Deploy.Spec.Deploy.CronJob.Namespace}, cronJobExists)
	if err == nil {
		log.Debug(name + ": Updating cronjob, " + cronJobExists.Name)
		err = b.MergeCronJobToSpec(cronJobExists)
		if err != nil {
			return err
		}
		err = b.UpdateCronJob(ctx, cronJobExists)
	} else {
		log.Debug(name + ": Creating new cronjob,  " + b.Deploy.Name)
		err = b.CreateCronJob(ctx)
	}
	return err
}

func (b *BuildStage) HandleDeployment(ctx context.Context) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	deploymentExists := &v12.Deployment{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Name, Namespace: b.Deploy.Namespace}, deploymentExists)
	if err == nil {
		log.Debug(name + ": Updating deployment, " + deploymentExists.Name)
		err = b.MergeDeploymentToSpec(deploymentExists)
		if err != nil {
			return err
		}
		err = b.UpdateDeployment(ctx, deploymentExists)
	} else {
		log.Debug(name + ": Creating new deployment,  " + b.Deploy.Name)
		err = b.CreateDeployment(ctx)
	}
	return err
}

func (b *BuildStage) HandleSimpleDeployment(ctx context.Context) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	deploymentExists := &v12.Deployment{}
	err := b.Get(ctx, types.NamespacedName{Name: b.Deploy.Spec.Deploy.Name, Namespace: b.Deploy.Spec.Deploy.Namespace}, deploymentExists)
	if err == nil {
		log.Debug(name + ": Updating simple deployment, " + deploymentExists.Name)
		b.UpdateDeploymentToSpec(deploymentExists)
		err = b.UpdateDeployment(ctx, deploymentExists)
		if err != nil {
			return err
		}
	} else {
		name := b.Deploy.Namespace + "_" + b.Deploy.Name
		log.Debug(name + ": Creating new simple deployment,  " + b.Deploy.Name)
		deployment := b.CreateSimpleDeployment()
		err = b.Create(ctx, deployment)

	}
	return err
}

func (b *BuildStage) Ingress(ctx context.Context) (ctrl.Result, error) {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	log.Debug(name + ": [" + b.Deploy.Name + "] In Ingress Stage")
	ingressEntries := map[string]*v1alpha1.Ingress{}
	for _, ingressEntry := range b.Deploy.Spec.Ingress {
		ingressEntries[ingressEntry.Name] = ingressEntry
	}
	for _, ingressEntry := range b.Deploy.Status.Ingress {
		if ingressEntries[ingressEntry.Name] == nil {
			ingressEntries[ingressEntry.Name] = ingressEntry
		}
	}
	for ingressName, _ := range b.Deletes.Get() {
		log.Debug(name + ": Deleting ingress in k8s")
		log.Debug(name + ": " + ingressName + ": " + strings.Join(*b.Deletes.GetList(ingressName), ","))
		co := ingressEntries[ingressName]
		b.DeleteIngress(ctx, co)
	}
	for ingressName, _ := range b.Changes.Get() {
		log.Debug(name + ": Saving ingress in k8s")
		log.Debug(name + ": " + ingressName + ": " + strings.Join(*b.Changes.GetList(ingressName), ","))
		co := ingressEntries[ingressName]
		err := b.HandleIngress(ctx, co)
		if err != nil {
			log.Debug(ingressName + ": " + err.Error())
		}
	}
	log.Debug(name + ": Saving status")
	b.Deploy.Status.Ingress = b.Deploy.Spec.Ingress
	err := b.Status().Update(ctx, b.Deploy)
	if err != nil {
		log.Error(name + ": " + err.Error())
	}
	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}

func (b *BuildStage) Destroyer(ctx context.Context) (ctrl.Result, error) {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	log.Debug(name + ": [" + b.Deploy.Name + "] In Destroyer Stage")

	if b.Deploy.Spec.Deploy != nil {
		b.DeleteDeployment(ctx)
		//b.DeleteCronJob(ctx)
		//b.DeleteDaemonSet(ctx)
		//b.DeleteSimpleDeployment(ctx)

	}

	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}

func (b *BuildStage) Route(ctx context.Context) (ctrl.Result, error) {
	switch b.GetStage(ctx) {
	case BUILD_STAGE:
		return b.Builder(ctx)
	case DEPLOY_STAGE:
		return b.Deployer(ctx)
	case DESTROY_STAGE:
		return b.Destroyer(ctx)
	case INGRESS_STAGE:
		return b.Ingress(ctx)
	case CHECK_LATEST:
		return b.CheckLatest(ctx)
	}
	return ctrl.Result{RequeueAfter: time.Second * 20}, nil
}

func (b *BuildStage) GetStage(ctx context.Context) string {
	if !b.LatestBuilt() {
		return BUILD_STAGE
	} else if !b.LatestDeployed(ctx) {
		return DEPLOY_STAGE
	} else if !b.LatestIngress(ctx) {
		return INGRESS_STAGE
	} else if b.Deploy.Spec.Deploy == nil {
		return DESTROY_STAGE
	} else if b.Deploy.Status.Built != nil && b.Deploy.Status.Deployed != nil {
		return CHECK_LATEST
	}
	return SLEEP
}
