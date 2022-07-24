package buildstage

import (
	"context"
	cicdv1alpha1 "github.com/Synload/build-deploy-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
)

func (r *BuildStage) UpdateDeploymentToSpec(deploy *cicdv1alpha1.BuildDeploy, simple *cicdv1alpha1.DeployParams, deployment *v1.Deployment) *v1.Deployment {

	deployment.Spec.Template.Spec.Volumes = simple.Volumes
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = simple.Mounts
	deployment.Spec.Template.Spec.Containers[0].Image = deploy.Spec.Publish.Host + "/" + deploy.Spec.Publish.Tag + ":" + deploy.Spec.Publish.Version
	deployment.Spec.Template.Spec.Containers[0].Env = simple.Env
	return deployment
}

func (r *BuildStage) UpdateDeployment(ctx context.Context, deployment *v1.Deployment) {
	err := r.Update(ctx, deployment)

	if err != nil {
		println(err.Error())
	}

}
