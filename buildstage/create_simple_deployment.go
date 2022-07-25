package buildstage

import (
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *BuildStage) CreateSimpleDeployment() *v1.Deployment {
	namespace := "default"
	if r.Deploy.Spec.Deploy.Namespace != "" {
		namespace = r.Deploy.Spec.Deploy.Namespace
	}
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      r.Deploy.Spec.Deploy.Name,
			Labels:    map[string]string{"app-selector": r.Deploy.Spec.Deploy.Name},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app-selector": r.Deploy.Spec.Deploy.Name},
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      r.Deploy.Spec.Deploy.Name,
					Labels:    map[string]string{"app-selector": r.Deploy.Spec.Deploy.Name},
				},
				Spec: v12.PodSpec{
					Volumes: r.Deploy.Spec.Deploy.Volumes,
					ImagePullSecrets: []v12.LocalObjectReference{{
						Name: "docker-pull-" + r.Deploy.Spec.Publish.Host,
					}},
					Containers: []v12.Container{{
						Name:            r.Deploy.Spec.Deploy.Name,
						Env:             r.Deploy.Spec.Deploy.Env,
						Image:           r.Deploy.Spec.Publish.Host + "/" + r.Deploy.Spec.Publish.Tag + ":" + r.Deploy.Spec.Publish.Version,
						VolumeMounts:    r.Deploy.Spec.Deploy.Mounts,
						ImagePullPolicy: v12.PullAlways,
					}},
				},
			},
		},
	}
	return deployment
}
