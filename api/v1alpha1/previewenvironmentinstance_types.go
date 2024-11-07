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

// PreviewEnvironmentInstanceSpec defines the desired state of PreviewEnvironmentInstance.
type PreviewEnvironmentInstanceSpec struct {
	// +optional
	// Branch of the repository to be used for the specific preview environment instance
	Branch *string `json:"branch"`

	// +optional
	// PullRequestIdentifier the number of the pull request
	PullRequestNumber int `json:"pullRequestNumber"`

	// +optional
	// GitOrganization the organization of the git repostiory
	GitOrganization string `json:"gitOrganization"`

	// +optional
	// GitRepository the name of the git repository
	GitRepository string `json:"gitRepository"`

	// +optional
	// PreviewEnvironmentRef is the reference to the PreviewEnvironment object
	PreviewEnvironmentRef PreviewEnvironmentRef `json:"previewEnvironmentRef"`

	// +optional
	// CommitHash the last commit hash, this should be the version that the instance is running
	CommitHash string `json:"commitHash"`
}

// PreviewEnvironmentInstanceStatus defines the observed state of PreviewEnvironmentInstance.
type PreviewEnvironmentInstanceStatus struct {
	// +optional
	// RebuildStatus the status of the rebuild
	RebuildStatus string `json:"rebuildStatus"`

	// +optional
	BuiltVersions []BuiltVersion `json:"builtVersions"`

	// +optional
	// PublicFacingUrl the url where the preview environment can be accessed
	PublicFacingUrl string `json:"publicFacingUrl"`
}

type BuiltVersion struct {

	// +optional
	// Tag of the built version
	Tag string `json:"tag"`

	// +optional
	// Timestamp of the upload
	Timestamp metav1.Time `json:"timestamp"`
}

func PreviewEnvironmentInstanceNameFromPullRequest(pe string, owner, repo string, number int) string {
	name := fmt.Sprintf("%s-%s-%s-%d", pe, owner, repo, number)
	name = strings.ReplaceAll(name, "/", "-")
	return strings.ToLower(name)
}

const (
	RebuildStatusBuildingOutdated   = "buildingOutdated"
	RebuildStatusDeploymentOutdated = "deploymentOutdated"
	RebuildStatusBuilding           = "building"
	RebuildStatusDeploying          = "deploying"
	RebuildStatusFailed             = "failed"
	RebuildStatusSuccess            = "success"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PreviewEnvironmentInstance is the Schema for the previewenvironmentinstances API.
type PreviewEnvironmentInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PreviewEnvironmentInstanceSpec   `json:"spec,omitempty"`
	Status PreviewEnvironmentInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PreviewEnvironmentInstanceList contains a list of PreviewEnvironmentInstance.
type PreviewEnvironmentInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PreviewEnvironmentInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PreviewEnvironmentInstance{}, &PreviewEnvironmentInstanceList{})
}
