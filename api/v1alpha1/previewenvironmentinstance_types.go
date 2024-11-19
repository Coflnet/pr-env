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
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PreviewEnvironmentInstanceSpec defines the desired state of PreviewEnvironmentInstance.
type PreviewEnvironmentInstanceSpec struct {
	// +kubebuilder:validation:Required
	// InstanceGitSettings configuration of the git repository that should be used for the preview environment instance
	InstanceGitSettings InstanceGitSettings `json:"instanceGitSettings"`

	// +kubebuilder:validation:Required
	// DesiredPhase the desired phase of the preview environment instance
	DesiredPhase string `json:"desiredPhase"`
}

type InstanceGitSettings struct {
	// +kubebuilder:validation:Required
	// Branch the branch that should be used for the preview environment instance
	Branch *string `json:"branch"`

	// +optional
	// PullRequestNumber the pull request number for the preview environment instance
	PullRequestNumber *int `json:"pullRequestNumber"`

	// +optional
	// CommitHash the last commit hash, this should be the version that the instance is running
	CommitHash string `json:"commitHash"`
}

// PreviewEnvironmentInstanceStatus defines the observed state of PreviewEnvironmentInstance.
type PreviewEnvironmentInstanceStatus struct {

	// +optional
	// Phase is the current phase of the preview environment instance
	Phase string `json:"phase"`

	// +optional
	// BuiltVersions a list of already built versions, these include the commit hash and the timestamp
	BuiltVersions []BuiltVersion `json:"builtVersions"`

	// +optional
	// PublicFacingUrl the url where the preview environment can be accessed
	PublicFacingUrl string `json:"publicFacingUrl"`
}

type BuiltVersion struct {
	// +kubebuilder:validation:Required
	// Tag of the built version
	Tag string `json:"tag"`

	// +kubebuilder:validation:Required
	// Timestamp of the upload
	Timestamp metav1.Time `json:"timestamp"`
}

const (
	InstancePhasePending   = "pending"
	InstancePhaseBuilding  = "building"
	InstancePhaseDeploying = "deploying"
	InstancePhaseRunning   = "running"
	InstancePhaseFailed    = "failed"
	InstancePhaseStopped   = "stopped"
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

func PreviewEnvironmentInstanceNameFromPullRequest(pe string, owner, gitOrganization, gitRepository string, pullRequestIdentifier int) string {
	return previewEnvironmentInstanceName(pe, owner, gitOrganization, gitRepository, fmt.Sprintf("pr-%d", pullRequestIdentifier))
}

func PreviewEnvironmentInstanceNameFromBranch(pe string, owner, gitOrganization, gitRepository, branch string) string {
	return previewEnvironmentInstanceName(pe, owner, gitOrganization, gitRepository, branch)
}

func previewEnvironmentInstanceName(pe string, owner, gitOrganization, gitRepository, identifier string) string {
	str := strings.ToLower(fmt.Sprintf("pei-%s-%s-%s-%s-%s", owner, pe, gitOrganization, gitRepository, identifier))
	str = strings.ReplaceAll(str, "/", "-")
	if len(str) > 45 {
		return fmt.Sprintf("pei-%s", str[len(str)-45:])
	}
	return str
}

func PreviewEnvironmentInstanceContainerName(pe *PreviewEnvironment, identifier, commitHash string) string {
	return fmt.Sprintf("%s/%s/tmpenv:%s-%s-%s-%s-%s", pe.Spec.ContainerRegistry.Registry, pe.Spec.ContainerRegistry.Repository, pe.GetOwner(), pe.Spec.GitSettings.Organization, pe.Spec.GitSettings.Repository, identifier, commitHash)
}

func PreviewEnvironmentHttpPath(pe *PreviewEnvironment, pei *PreviewEnvironmentInstance) string {
	return fmt.Sprintf("/%s/%s/%s/%s", pe.Spec.GitSettings.Organization, pe.Spec.GitSettings.Repository, pei.BranchOrPullRequestIdentifier(), pei.Spec.InstanceGitSettings.CommitHash)
}

func (g *InstanceGitSettings) BranchOrPullRequestIdentifier() string {
	if g.PullRequestNumber != nil {
		return strconv.Itoa(*g.PullRequestNumber)
	}
	if g.Branch != nil {
		return *g.Branch
	}
	panic("neither branch nor pull request number is set")
}

func (pei *PreviewEnvironmentInstance) BranchOrPullRequestIdentifier() string {
	return pei.Spec.InstanceGitSettings.BranchOrPullRequestIdentifier()
}

func (pei *PreviewEnvironmentInstance) GetOwner() string {
	return pei.GetLabels()["owner"]
}

func (pei *PreviewEnvironmentInstance) GetPreviewEnvironmentId() string {
	return pei.GetLabels()["previewenvironment"]
}

func (pei *PreviewEnvironmentInstance) NameForAuthProxy() string {
	return fmt.Sprintf("%s-auth-proxy", pei.GetName())
}
