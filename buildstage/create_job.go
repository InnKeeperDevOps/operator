package buildstage

import (
	"github.com/Synload/build-deploy-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *BuildStage) createBuildJob(deploy *v1alpha1.BuildDeploy) *batchv1.Job {
	var retries int32 = 5
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	envars := []corev1.EnvVar{
		{
			Name:  "GIT_REPO",
			Value: deploy.Spec.Git.URI,
		},
		{
			Name:  "REGISTRY_HOST",
			Value: deploy.Spec.Publish.Host,
		},
		{
			Name:  "DOCKER_TAG",
			Value: deploy.Spec.Publish.Tag,
		},
		{
			Name:  "DOCKER_VERSION",
			Value: deploy.Spec.Publish.Version,
		},
	}

	if deploy.Spec.Publish.Secret != "" {
		envars = append(envars, corev1.EnvVar{
			Name: "REGISTRY_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: deploy.Spec.Publish.Secret,
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
							Name: deploy.Spec.Publish.Secret,
						},
						Key: "password",
					},
				},
			})
	}

	if deploy.Spec.Build.Dockerfile != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "DOCKERFILE",
			Value: deploy.Spec.Build.Dockerfile,
		})
	}
	if deploy.Spec.Build.WorkDir != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "WORKDIR",
			Value: deploy.Spec.Build.WorkDir,
		})
	}
	if deploy.Spec.Git.Commit != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "GIT_COMMIT",
			Value: deploy.Spec.Git.Commit,
		})
	}
	if deploy.Spec.Git.Branch != "" {
		envars = append(envars, corev1.EnvVar{
			Name:  "GIT_BRANCH",
			Value: deploy.Spec.Git.Branch,
		})
	}

	if deploy.Spec.Git.Secret != "" {
		volumes = append(volumes, corev1.Volume{
			Name: deploy.Spec.Git.Secret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: deploy.Spec.Git.Secret,
					Items: []corev1.KeyToPath{{
						Key:  "ssh-key",
						Path: "key",
					}},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      deploy.Spec.Git.Secret,
			MountPath: "/ssh/",
			ReadOnly:  false,
		})
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploy.GetBuilderName(),
			Namespace: "git-builder",
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &retries,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploy.GetBuilderName(),
					Namespace: "git-builder",
					Labels: map[string]string{
						"buildFor": deploy.GetBuilderName(),
					},
				},
				Spec: corev1.PodSpec{
					Volumes: volumes,
					Containers: []corev1.Container{{
						Name:            deploy.GetBuilderName(),
						Image:           latestBuilderImage,
						ImagePullPolicy: corev1.PullAlways,
						VolumeMounts:    volumeMounts,
						Env:             envars,
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
		Status: batchv1.JobStatus{},
	}
}
