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
	"github.com/Synload/build-deploy-operator/buildstage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

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
		println("Handling " + buildDeploy.Name)
		stage := buildstage.BuildStage{Deploy: buildDeploy, Client: r.Client}
		stage.Route(ctx)
	} else {
		// cleanup
		buildDeploy.Name = req.Name
		buildDeploy.Namespace = req.Namespace
		println("deleting jobs for " + req.Name)
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      buildDeploy.GetBuilderName(),
				Namespace: "git-builder",
			},
		}
		deleteOpt := metav1.DeletePropagationBackground
		err = r.Delete(ctx, job, &client.DeleteOptions{
			PropagationPolicy: &deleteOpt,
		})
		if err != nil {
			println(err.Error())
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: time.Second * 25}, nil
}

func (r *BuildDeployReconciler) updateStatus(deploy *cicdv1alpha1.BuildDeploy) {

}

// SetupWithManager sets up the controller with the Manager.
func (r *BuildDeployReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1alpha1.BuildDeploy{}).
		Complete(r)
}
