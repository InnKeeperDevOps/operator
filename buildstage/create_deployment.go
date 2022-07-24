package buildstage

import (
	cicdv1alpha1 "github.com/Synload/build-deploy-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *BuildStage) CreateDeployment(deploy *cicdv1alpha1.BuildDeploy, simple *cicdv1alpha1.DeployParams) *v1.Deployment {
	namespace := "default"
	if simple.Namespace != "" {
		namespace = simple.Namespace
	}
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      simple.Name,
			Labels:    map[string]string{"app-selector": simple.Name},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app-selector": simple.Name},
			},
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      simple.Name,
					Labels:    map[string]string{"app-selector": simple.Name},
				},
				Spec: v12.PodSpec{
					Volumes: simple.Volumes,
					ImagePullSecrets: []v12.LocalObjectReference{{
						Name: "docker-pull-" + deploy.Spec.Publish.Host,
					}},
					Containers: []v12.Container{{
						Name:            simple.Name,
						Env:             simple.Env,
						Image:           deploy.Spec.Publish.Host + "/" + deploy.Spec.Publish.Tag + ":" + deploy.Spec.Publish.Version,
						VolumeMounts:    simple.Mounts,
						ImagePullPolicy: v12.PullAlways,
					}},
				},
			},
		},
	}
	return deployment
}
