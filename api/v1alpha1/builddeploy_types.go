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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type GitDetails struct {
	URI string `json:"uri"`
	// Secret contains a ssh key to connect to private repos
	Secret string `json:"secret,omitempty"`
}
type DockerBuild struct {
	Dockerfile string `json:"dockerfile"`
	WorkDir    string `json:"work_dir,omitempty"`
}
type DockerPublish struct {
	Secret string `json:"secret"`
	Host    string `json:"host"`
	Tag string `json:"tag"`
	Version string `json:"version"`
}

// BuildDeploySpec defines the desired state of BuildDeploy
type BuildDeploySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Build   DockerBuild   `json:"build,omitempty"`
	Git     GitDetails    `json:"git"`
	Publish DockerPublish `json:"publish"`
}

type Deployed struct {
	Pod string `json:"pod"`
}

// BuildDeployStatus defines the observed state of BuildDeploy
type BuildDeployStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Deployed Deployed `json:"deployed"`
	Built    bool     `json:"built"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BuildDeploy is the Schema for the builddeploys API
type BuildDeploy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuildDeploySpec   `json:"spec,omitempty"`
	Status BuildDeployStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BuildDeployList contains a list of BuildDeploy
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
