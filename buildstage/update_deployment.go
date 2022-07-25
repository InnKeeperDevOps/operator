package buildstage

import (
	"context"
	"github.com/imdario/mergo"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
)

func (r *BuildStage) UpdateDeploymentToSpec(deployment *v1.Deployment) *v1.Deployment {
	deployment.Spec.Template.Spec.Volumes = r.Deploy.Spec.Deploy.Volumes
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = r.Deploy.Spec.Deploy.Mounts
	deployment.Spec.Template.Spec.Containers[0].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Spec.Publish.Version
	deployment.Spec.Template.Spec.Containers[0].Env = r.Deploy.Spec.Deploy.Env
	return deployment
}

func (r *BuildStage) MergeDeploymentToSpec(deployment *v1.Deployment) error {
	err := mergo.Merge(deployment, r.Deploy.Spec.Deploy.Deployment)
	deployment.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Spec.Publish.Version
	return err
}

func (r *BuildStage) CreateDeployment(ctx context.Context) error {
	deployment := r.Deploy.Spec.Deploy.Deployment

	exists := false
	for _, ps := range deployment.Spec.Template.Spec.ImagePullSecrets {
		if ps.Name == "docker-pull-"+r.Deploy.Spec.Publish.Host {
			exists = true
		}
	}
	if !exists {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(deployment.Spec.Template.Spec.ImagePullSecrets, v12.LocalObjectReference{
			Name: "docker-pull-" + r.Deploy.Spec.Publish.Host,
		})
	}
	deployment.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Spec.Publish.Version
	err := r.Create(ctx, deployment)
	return err
}

func (r *BuildStage) UpdateDeployment(ctx context.Context, deployment *v1.Deployment) error {
	err := r.Update(ctx, deployment)
	return err
}
