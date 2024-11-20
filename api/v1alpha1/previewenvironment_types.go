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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PreviewEnvironment is the Schema for the previewenvironments API.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type PreviewEnvironment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PreviewEnvironmentSpec   `json:"spec,omitempty"`
	Status PreviewEnvironmentStatus `json:"status,omitempty"`
}

// PreviewEnvironmentSpec defines the desired state of PreviewEnvironment.
type PreviewEnvironmentSpec struct {
	// +kubebuilder:validation:Required
	// GitSettings configuration of the git repository that should be used for the preview environments
	GitSettings GitSettings `json:"gitSettings"`

	// +kubebuilder:validation:Required
	// ContainerRegistry configuration of the container registry that should be used for the preview environments
	ContainerRegistry *ContainerRegistry `json:"containerRegistry"`

	// +kubebuilder:validation:Required
	// ApplicationSettings configuration for the running application
	ApplicationSettings ApplicationSettings `json:"applicationSettings"`

	// +kubebuilder:validation:Required
	// BuildSettings configuration for the build process
	BuildSettings BuildSettings `json:"buildSettings"`

	// +kubebuilder:validation:Required
	// DisplayName is the name that can be displayed to the user
	DisplayName string `json:"displayName"`

	// +kubebuilder:validation:Required
	// AccessSettings configuration for the access control
	AccessSettings AccessSettings `json:"accessSettings"`
}

type BuildSettings struct {
	// +kubebuilder:validation:Required
	// BuildAllPullRequests is a flag that can be used to build all pull requests
	BuildAllPullRequests bool `json:"buildAllPullRequests"`

	// +kubebuilder:validation:Required
	// BuildAllBranches is a flag that can be used to build all branches
	BuildAllBranches bool `json:"buildAllBranches"`

	// +optional
	// BranchWildcard is optional and can be used to specify a wildcard for branches that should be built
	BranchWildcard *string `json:"branchWildcard"`

	// +optional
	// DockerfilePath is optional and can be used to override the default Dockerfile that is used to build the application
	DockerfilePath *string `json:"dockerfile"`
}

type GitSettings struct {
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	Organization string `json:"organization"`

	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	Repository string `json:"repository"`
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

	// +optional
	// EnvironmentVariables is a list of environment variables that should be set for the application
	EnvironmentVariables *[]EnvironmentVariable `json:"environmentVariables"`

	// +optional
	// Command is optional and can be used to override the default command that is used to start the application
	Command *string `json:"command"`
}

type AccessSettings struct {
	// +kubebuilder:validation:Required
	// Users is a list of users that should have access to the preview environment
	Users []UserAccess `json:"users"`

	// +kubebuilder:validation:Required
	// PublicAccess is a flag that can be used to allow public access to the preview environment
	PublicAccess bool `json:"publicAccess"`
}

type UserAccess struct {
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	// Username is the username of the user
	Username string `json:"username"`

	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	// UserId is the keycloak id of the user
	UserId string `json:"userId"`
}

// PreviewEnvironmentStatus defines the observed state of PreviewEnvironment.
type PreviewEnvironmentStatus struct {
	// +optional
	// PullRequestsDetected is a list of pullRequests that were detected
	PullRequestsDetected []int `json:"pullRequests"`

	// +optional
	Phase string `json:"phase"`
}

const (
	PreviewEnvironmentPhasePending    = "pending"
	PreviewEnvironmentPhaseProcessing = "procesing"
	PreviewEnvironmentPhaseReady      = "ready"
	PreviewEnvironmentPhaseError      = "error"
)

type EnvironmentVariable struct {

	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	// Key is the key of the environment variable
	Key string `json:"key"`

	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=63
	// Value is the value of the environment variable
	Value string `json:"value"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PreviewEnvironmentList contains a list of PreviewEnvironment.
type PreviewEnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PreviewEnvironment `json:"items"`
}

func PreviewEnvironmentName(organization, repo string) string {
	str := strings.ToLower(fmt.Sprintf("%s-%s", organization, repo))
	if len(str) > 50 {
		return str[len(str)-50:]
	}
	return str
}

func init() {
	SchemeBuilder.Register(&PreviewEnvironment{}, &PreviewEnvironmentList{})
}

func (pe *PreviewEnvironment) GetOwner() string {
	return pe.GetLabels()["owner"]
}
