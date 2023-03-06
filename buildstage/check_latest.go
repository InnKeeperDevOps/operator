package buildstage

import (
	"context"
	"errors"
	"github.com/InnKeeperDevOps/operator/buildstage/job"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
	"time"
)

func (r *BuildStage) CheckLatest(ctx context.Context) (ctrl.Result, error) {
	name := r.Deploy.Namespace + "_" + r.Deploy.Name
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	envars := []corev1.EnvVar{
		{
			Name:  "GIT_REPO",
			Value: r.Deploy.Spec.Git.URI,
		},
		{
			Name:  "REGISTRY_HOST",
			Value: r.Deploy.Spec.Publish.Host,
		},
		{
			Name:  "DOCKER_TAG",
			Value: r.Deploy.Spec.Publish.Tag,
		},
	}

	if r.Deploy.Spec.Publish.Secret != "" {
		envars = append(envars, corev1.EnvVar{
			Name: "REGISTRY_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.Deploy.Spec.Publish.Secret,
					},
					Key: "username",
				},
			},
		},
			corev1.EnvVar{
				Name: "REGISTRY_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: r.Deploy.Spec.Publish.Secret,
						},
						Key: "password",
					},
				},
			})
	}

	if r.Deploy.Spec.Build.Dockerfile != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "DOCKERFILE",
			Value: r.Deploy.Spec.Build.Dockerfile,
		})
	}
	if r.Deploy.Spec.Build.WorkDir != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "WORKDIR",
			Value: r.Deploy.Spec.Build.WorkDir,
		})
	}
	if r.Deploy.Spec.Git.Commit != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "GIT_COMMIT",
			Value: r.Deploy.Spec.Git.Commit,
		})
	}
	if r.Deploy.Spec.Git.Branch != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "GIT_BRANCH",
			Value: r.Deploy.Spec.Git.Branch,
		})
	}

	if r.Deploy.Spec.Git.Secret != "" {
		volumes = append(volumes, corev1.Volume{
			Name: r.Deploy.Spec.Git.Secret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: r.Deploy.Spec.Git.Secret,
					Items: []corev1.KeyToPath{{
						Key:  "ssh-key",
						Path: "key",
					}},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      r.Deploy.Spec.Git.Secret,
			MountPath: "/ssh/",
			ReadOnly:  false,
		})
	}
	batchJob := job.NewJob(r.Client, name, LatestGitImage, envars, volumes, volumeMounts, 1)
	batchJob.StartJob(ctx)
	monitor := batchJob.Monitor(ctx)
	for !monitor.Failed() && !monitor.Succeeded() {
		if monitor.Started() {
			log.Debug(name + ": Job in progress!")
			time.Sleep(2 * time.Second)
		}
		if monitor.Finished() {
			break
		}
		monitor = batchJob.Monitor(ctx)
	}
	result := ctrl.Result{RequeueAfter: time.Second * 20}
	if monitor.Failed() && !monitor.Succeeded() {
		log.Debug(name + ": Git error, execution failed")
		return result, errors.New("execution failed")
	}
	jobPod := batchJob.Pod(ctx)
	if jobPod == nil {
		log.Debug(name + ": Git error, could not get pod.")
		batchJob.Delete(ctx)
		return result, errors.New("could not find pods.")
	}
	err, str := jobPod.Log(ctx)
	if err != nil {
		log.Debug(name + ": Git error, could not get logs.")
		batchJob.Delete(ctx)
		return result, err
	}
	log.Debug(name + ": Git check latest success")
	hash := match(str, "hash:(.*?),")
	date, _ := strconv.Atoi(match(str, "date:([0-9]+)"))

	if hash != r.Deploy.Status.Built.Git.Commit && r.Deploy.Status.Built.Git.Date < date {
		log.Info(name + ": Git change detected! Changing status to blank to force deployment.")
		r.Deploy.Status.Built = nil
		r.Deploy.Status.Deployed = nil
		err := r.Status().Update(ctx, r.Deploy)
		if err != nil {
			batchJob.Delete(ctx)
			return result, err
		}
	} else {
		log.Debug(name + ": Git no change detected.")
	}
	batchJob.Delete(ctx)
	return result, nil
}
