package buildstage

import (
	"context"
	"github.com/imdario/mergo"
	v1 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
)

func (r *BuildStage) MergeCronJobToSpec(cronjob *v1.CronJob) error {
	err := mergo.Merge(cronjob, r.Deploy.Spec.Deploy.CronJob)
	cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Status.Built.Git.Commit
	exists := false
	for _, ps := range cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets {
		if ps.Name == "docker-pull-"+r.Deploy.Spec.Publish.Host {
			exists = true
		}
	}
	if !exists {
		cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets, v12.LocalObjectReference{
			Name: "docker-pull-" + r.Deploy.Spec.Publish.Host,
		})
	}
	return err
}

func (r *BuildStage) CreateCronJob(ctx context.Context) error {
	cronjob := r.Deploy.Spec.Deploy.CronJob
	cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Status.Built.Git.Commit
	exists := false
	for _, ps := range cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets {
		if ps.Name == "docker-pull-"+r.Deploy.Spec.Publish.Host {
			exists = true
		}
	}
	if !exists {
		cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets, v12.LocalObjectReference{
			Name: "docker-pull-" + r.Deploy.Spec.Publish.Host,
		})
	}
	err := r.Create(ctx, cronjob)
	return err
}

func (r *BuildStage) UpdateCronJob(ctx context.Context, cronjob *v1.CronJob) error {
	err := r.Update(ctx, cronjob)
	return err
}
