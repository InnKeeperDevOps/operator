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

package v1alpha1

import (
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GitDetails struct {
	URI string `json:"uri"`
	// Secret contains a ssh key to connect to private repos
	Secret string `json:"secret,omitempty"`
	Commit string `json:"commit,omitempty"`
	Branch string `json:"branch,omitempty"`
}
type DockerBuild struct {
	Dockerfile string `json:"dockerfile,omitempty"`
	WorkDir    string `json:"workdir,omitempty"`
}
type DockerPublish struct {
	Secret string `json:"secret"`
	Host   string `json:"host"`
	Tag    string `json:"tag"`
}

type DeployParams struct {
	Env       []v12.EnvVar      `json:"env,omitempty"`
	Mounts    []v12.VolumeMount `json:"mounts,omitempty"`
	Volumes   []v12.Volume      `json:"volumes,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Name      string            `json:"name"`

	HandleContainer int            `json:"handleContainer,omitempty"`
	Deployment      *v1.Deployment `json:"deployment,omitempty"`
	CronJob         *v13.CronJob   `json:"cron_job,omitempty"`
	Job             *v13.Job       `json:"job,omitempty"`
	DaemonSet       *v1.DaemonSet  `json:"daemon_set,omitempty"`
}

type Ingress struct {
	Name    string            `json:"name"`
	Domain  string            `json:"domain"`
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Gateway []string          `json:"gateway"`
	Port    uint32            `json:"port"`
}

type BuildDeploySpec struct {
	Build   DockerBuild   `json:"build,omitempty"`
	Git     GitDetails    `json:"git"`
	Publish DockerPublish `json:"publish"`
	Deploy  *DeployParams `json:"deploy"`
	Ingress []*Ingress    `json:"ingress,omitempty"`
}

type RegistryStatus struct {
	Host string `json:"host,omitempty"`
	Tag  string `json:"tag,omitempty"`
}

type GitStatus struct {
	Commit string `json:"commit,omitempty"`
	Branch string `json:"branch,omitempty"`
	Date   int    `json:"date,omitempty"`
	Author string `json:"author,omitempty"`
}

type Deployed struct {
	Pod      string         `json:"pod,omitempty"`
	Git      GitStatus      `json:"git,omitempty"`
	Complete bool           `json:"complete,omitempty"`
	Registry RegistryStatus `json:"registry,omitempty"`
}
type Built struct {
	Git      GitStatus      `json:"git,omitempty"`
	Complete bool           `json:"complete,omitempty"`
	Registry RegistryStatus `json:"registry,omitempty"`
}

type BuildDeployStatus struct {
	Deployed *Deployed  `json:"deployed,omitempty"`
	Built    *Built     `json:"built,omitempty"`
	Ingress  []*Ingress `json:"ingress,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type BuildDeploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuildDeploySpec   `json:"spec,omitempty"`
	Status BuildDeployStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type BuildDeployList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BuildDeploy `json:"items"`
}

func (d *BuildDeploy) GetBuilderName() string {
	return d.Name + "-builder"
}

func init() {
	SchemeBuilder.Register(&BuildDeploy{}, &BuildDeployList{})
}
