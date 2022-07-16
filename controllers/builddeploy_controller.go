/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cicdv1alpha1 "github.com/Synload/build-deploy-operator/api/v1alpha1"
)

// BuildDeployReconciler reconciles a BuildDeploy object
type BuildDeployReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cicd.synload.com,resources=builddeploys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cicd.synload.com,resources=builddeploys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cicd.synload.com,resources=builddeploys/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BuildDeploy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *BuildDeployReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// check namespace for builder
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: "git-builder"}, namespace)
	if err != nil {
		println(err.Error())
		println("Creating namespace")
		namespace.Name = "git-builder"
		r.Create(ctx, namespace)
	}

	buildDeploy := &cicdv1alpha1.BuildDeploy{}
	err = r.Get(ctx, req.NamespacedName, buildDeploy)
	if err == nil {
		println(buildDeploy.Name)
		jobExists := &batchv1.Job{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: buildDeploy.GetBuilderName(), Namespace: "git-builder"}, jobExists)
		if err != nil {
			println(err.Error())
			err = r.Client.Create(ctx, r.createBuildJob(buildDeploy))
			if err != nil {
				println(err.Error())
			}
		} else {
			println(jobExists.Status.Succeeded)
		}
	}
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *BuildDeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1alpha1.BuildDeploy{}).
		Complete(r)
}

func (r *BuildDeployReconciler) createBuildJob(deploy *cicdv1alpha1.BuildDeploy) *batchv1.Job {
	var retries int32 = 0
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
	}

	if deploy.Spec.Publish.Secret !="" {
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

	if deploy.Spec.Git.Secret!="" {
		volumes = append(volumes, corev1.Volume{
			Name: deploy.Spec.Git.Secret,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: deploy.Spec.Git.Secret,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      deploy.Spec.Git.Secret,
			MountPath: "/ssh/key",
			ReadOnly:  true,
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
				},
				Spec: corev1.PodSpec{
					Volumes: volumes,
					Containers: []corev1.Container{{
						Name:  deploy.GetBuilderName(),
						Image: "ghcr.io/synload/git-buildah:main",
						VolumeMounts: volumeMounts,
						Env: envars,
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
		Status: batchv1.JobStatus{},
	}
}
