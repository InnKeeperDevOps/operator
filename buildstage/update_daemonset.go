package buildstage

import (
	"context"
	"github.com/imdario/mergo"
	v1 "k8s.io/api/apps/v1"
)

func (r *BuildStage) MergeDaemonSetToSpec(daemonset *v1.DaemonSet) error {
	err := mergo.Merge(daemonset, r.Deploy.Spec.Deploy.DaemonSet)
	daemonset.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Spec.Publish.Version
	return err
}
func (r *BuildStage) CreateDaemonSet(ctx context.Context) error {
	daemonset := r.Deploy.Spec.Deploy.DaemonSet
	daemonset.Spec.Template.Spec.Containers[r.Deploy.Spec.Deploy.HandleContainer].Image = r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Spec.Publish.Version
	err := r.Create(ctx, daemonset)
	return err
}

func (r *BuildStage) UpdateDaemonSet(ctx context.Context, daemonset *v1.DaemonSet) error {
	err := r.Update(ctx, daemonset)
	return err
}
