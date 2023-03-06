package buildstage

import (
	"context"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *BuildStage) UpdateDeploymentToSpec(deployment *v1.Deployment) *v1.Deployment {
	deployment.Spec.Template.Spec.Volumes = r.Deploy.Spec.Deploy.Volumes
	if r.Deploy.Spec.Deploy.Deployment.Spec.Replicas != nil {
		deployment.Spec.Replicas = r.Deploy.Spec.Deploy.Deployment.Spec.Replicas
	}
	//deployment.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].VolumeMounts = r.Deploy.Spec.Deploy.Mounts
	//deployment.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Spec.Publish.Version
	//deployment.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Env = r.Deploy.Spec.Deploy.Env
	return deployment
}

func (r *BuildStage) MergeDeploymentToSpec(deployment *v1.Deployment) error {
	err := mergo.Merge(deployment, r.Deploy.Spec.Deploy.Deployment)
	deployment.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Status.Built.Git.Commit
	return err
}

func (b *BuildStage) CreateDeployment(ctx context.Context) error {
	name := b.Deploy.Namespace + "_" + b.Deploy.Name
	deployment := b.Deploy.Spec.Deploy.Deployment

	exists := false
	for _, ps := range deployment.Spec.Template.Spec.ImagePullSecrets {
		if ps.Name == "docker-pull-"+b.Deploy.Spec.Publish.Host {
			exists = true
		}
	}
	if !exists {
		deployment.Spec.Template.Spec.ImagePullSecrets = append(deployment.Spec.Template.Spec.ImagePullSecrets, v12.LocalObjectReference{
			Name: "docker-pull-" + b.Deploy.Spec.Publish.Host,
		})
	}
	deployment.ObjectMeta.Namespace = b.Deploy.Namespace
	deployment.ObjectMeta.Name = b.Deploy.Name
	if deployment.ObjectMeta.Labels == nil {
		deployment.ObjectMeta.Labels = map[string]string{}
	}
	_, ok := deployment.ObjectMeta.Labels["app-connector"]
	if !ok {
		deployment.ObjectMeta.Labels["app-connector"] = b.Deploy.Name + "_" + b.Deploy.Namespace
	}

	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}
	_, ok = deployment.Spec.Template.ObjectMeta.Labels["app-connector"]
	if !ok {
		deployment.Spec.Template.ObjectMeta.Labels["app-connector"] = b.Deploy.Name + "_" + b.Deploy.Namespace
	}

	deployment.Spec.Template.Spec.Containers[b.Deploy.Spec.Deploy.HandleContainer].Image = b.Deploy.Spec.Publish.Host + "/" + b.Deploy.Spec.Publish.Tag + ":" + b.Deploy.Status.Built.Git.Commit
	log.Info(name + ": deployment created!")
	err := b.Create(ctx, deployment)
	return err
}

func (r *BuildStage) UpdateDeployment(ctx context.Context, deployment *v1.Deployment) error {
	name := r.Deploy.Namespace + "_" + r.Deploy.Name
	if deployment.ObjectMeta.Labels == nil {
		deployment.ObjectMeta.Labels = map[string]string{}
	}
	_, ok := deployment.ObjectMeta.Labels["app-connector"]
	if !ok {
		deployment.ObjectMeta.Labels["app-connector"] = r.Deploy.Name + "_" + r.Deploy.Namespace
	}

	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}
	_, ok = deployment.Spec.Template.ObjectMeta.Labels["app-connector"]
	if !ok {
		deployment.Spec.Template.ObjectMeta.Labels["app-connector"] = r.Deploy.Name + "_" + r.Deploy.Namespace
	}

	r.UpdateDeploymentToSpec(deployment)
	log.Info(name + ": deployment updated!")
	err := r.Update(ctx, deployment)
	return err
}

func (r *BuildStage) DeleteDeployment(ctx context.Context) error {
	deleteOpt := v13.DeletePropagationBackground
	err := r.Delete(ctx, &v1.Deployment{
		ObjectMeta: v13.ObjectMeta{
			Name:      r.Deploy.Name,
			Namespace: r.Deploy.Namespace,
		},
	}, &client.DeleteOptions{
		PropagationPolicy: &deleteOpt,
	})
	return err
}
