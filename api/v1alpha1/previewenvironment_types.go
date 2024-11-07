/*
Copyright 2024.

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

// PreviewEnvironmentSpec defines the desired state of PreviewEnvironment.
type PreviewEnvironmentSpec struct {
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	GitOrganization string `json:"gitOrganization"`

	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	GitRepository string `json:"gitRepository"`

	// +optional
	// ContainerRegistry configuration of the container registry that should be used for the preview environments
	ContainerRegistry ContainerRegistry `json:"containerRegistry"`

	// +optional
	// ApplicationSettings configuration for the running application
	ApplicationSettings ApplicationSettings `json:"applicationSettings"`
}

type PreviewEnvironmentRef struct {
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`

	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	Namespace string `json:"namespace"`
}

type ContainerRegistry struct {
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	Registry string `json:"registry"`

	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	Repository string `json:"repository"`
}

type ApplicationSettings struct {

	// +kubebuilder:validation:MinLength=0
	// IngressHostname the hostname the application should get exposed on
	IngressHostname string `json:"ingressHostname"`

	// +optional
	// Port is the port the application is listening on
	Port int `json:"port"`
}

// PreviewEnvironmentStatus defines the observed state of PreviewEnvironment.
type PreviewEnvironmentStatus struct {
	// PullRequestsDetected is a list of pullRequests that were detected
	PullRequestsDetected []int `json:"pullRequests"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PreviewEnvironment is the Schema for the previewenvironments API.
type PreviewEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PreviewEnvironmentSpec   `json:"spec,omitempty"`
	Status PreviewEnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PreviewEnvironmentList contains a list of PreviewEnvironment.
type PreviewEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PreviewEnvironment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PreviewEnvironment{}, &PreviewEnvironmentList{})
}
