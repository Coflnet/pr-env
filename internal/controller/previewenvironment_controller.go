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

package controller

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/coflnet/pr-env/internal/git"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const finalizerName = "coflnet.com.pr.env/finalizer"

// PreviewEnvironmentReconciler reconciles a PreviewEnvironment object
type PreviewEnvironmentReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	githubClient *git.GithubClient
	log          logr.Logger
}

// +kubebuilder:rbac:groups=coflnet.coflnet.com,resources=previewenvironments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coflnet.coflnet.com,resources=previewenvironments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=coflnet.coflnet.com,resources=previewenvironments/finalizers,verbs=update
func (r *PreviewEnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log.Info("Reconciling PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
	// TODO: mark environment as updating

	// load the preview environment
	var pe coflnetv1alpha1.PreviewEnvironment
	if err := r.Get(ctx, req.NamespacedName, &pe); err != nil {
		r.log.Info("Unable to load PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// check if the preview environment is being deleted
	if !pe.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&pe, finalizerName) {
			// do some deletion work
			err := r.deletePreviewEnvironmentInstancesForPreviewEnvironment(ctx, pe)
			if err != nil {
				r.log.Error(err, "Unable to delete PreviewEnvironmentInstances", "namespace", req.Namespace, "name", req.Name)
				return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
			}

			// remove the finalizer
			controllerutil.RemoveFinalizer(&pe, finalizerName)
			if err := r.Update(ctx, &pe); err != nil {
				r.log.Error(err, "Unable to remove finalizer from PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// add the finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&pe, finalizerName) {
		controllerutil.AddFinalizer(&pe, finalizerName)
		if err := r.Update(ctx, &pe); err != nil {
			r.log.Error(err, "Unable to add finalizer to PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, err
		}
	}

	// list all the instances that should be created
	peis, err := r.detectInstancesThatShouldBeCreated(ctx, pe)
	if err != nil {
		// TODO: mark the pe as failed
		r.log.Error(err, "Unable to detect instances that should be created", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	// create the preview environment instances that should be created for each branch
	err = r.savePreviewEnvironmentInstances(ctx, &pe, peis)
	if err != nil {
		r.log.Error(err, "Unable to create PreviewEnvironment instances", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{
			RequeueAfter: time.Minute * 1,
		}, nil
	}

	// TODO: update status

	return ctrl.Result{}, nil
}

func (r *PreviewEnvironmentReconciler) detectInstancesThatShouldBeCreated(ctx context.Context, pe coflnetv1alpha1.PreviewEnvironment) ([]*coflnetv1alpha1.PreviewEnvironmentInstance, error) {
	result := []*coflnetv1alpha1.PreviewEnvironmentInstance{}

	prs, err := r.detectOpenPullRequests(ctx, pe)
	if err != nil {
		return nil, err
	}

	r.log.Info("detected open pull requests", "count", len(prs), "namespace", pe.Namespace, "name", pe.Name)
	for _, pullRequest := range prs {
		result = append(result,
			r.buildPreviewEnvironmentInstanceForPr(pe, pullRequest))
	}

	branches, err := r.detectOpenBranches(ctx, pe)
	if err != nil {
		return nil, err
	}

	r.log.Info("detected open branches", "count", len(branches), "namespace", pe.Namespace, "name", pe.Name)
	for _, branch := range branches {
		result = append(result, r.buildPreviewEnvironmentInstanceForBranch(pe, branch))
	}

	return result, nil
}

func (r *PreviewEnvironmentReconciler) detectOpenPullRequests(ctx context.Context, pr coflnetv1alpha1.PreviewEnvironment) ([]*github.PullRequest, error) {
	prs, err := r.githubClient.PullRequestsOfRepository(ctx, pr.Spec.GitSettings.Organization, pr.Spec.GitSettings.Repository)
	return prs, err
}

// detectOpenBranches detects the open branches of the repository
// returns the name of the branches as a string slice
func (r *PreviewEnvironmentReconciler) detectOpenBranches(ctx context.Context, pr coflnetv1alpha1.PreviewEnvironment) ([]string, error) {
	if pr.Spec.BuildSettings.BuildAllBranches == false {
		return []string{}, nil
	}

	branches, err := r.githubClient.BranchesOfRepository(ctx, pr.Spec.GitSettings.Organization, pr.Spec.GitSettings.Repository)

	if pr.Spec.BuildSettings.BuildAllBranches || pr.Spec.BuildSettings.BranchWildcard == nil {
		return branches, err
	}

	var filteredBranches []string
	for _, branch := range branches {
		if strings.Contains(branch, *pr.Spec.BuildSettings.BranchWildcard) {
			filteredBranches = append(filteredBranches, branch)
		}
	}
	return filteredBranches, nil
}

func (r *PreviewEnvironmentReconciler) buildPreviewEnvironmentInstanceForPr(pe coflnetv1alpha1.PreviewEnvironment, pullRequest *github.PullRequest) *coflnetv1alpha1.PreviewEnvironmentInstance {

	name := coflnetv1alpha1.PreviewEnvironmentInstanceNameFromPullRequest(
		pe.GetName(),
		pe.GetOwner(),
		pe.Spec.GitSettings.Organization,
		pe.Spec.GitSettings.Repository,
		int(pullRequest.GetNumber()),
	)

	gitSettings := coflnetv1alpha1.InstanceGitSettings{
		PullRequestNumber: intPtr(int(pullRequest.GetNumber())),
		Branch:            strPtr(pullRequest.GetHead().GetRef()),
		CommitHash:        "",
	}

	return createInstanceFromEnvironment(pe, name, gitSettings)
}

func (r *PreviewEnvironmentReconciler) buildPreviewEnvironmentInstanceForBranch(pe coflnetv1alpha1.PreviewEnvironment, branch string) *coflnetv1alpha1.PreviewEnvironmentInstance {
	name := coflnetv1alpha1.PreviewEnvironmentInstanceNameFromBranch(
		pe.Name,
		pe.GetLabels()["owner"],
		pe.Spec.GitSettings.Organization,
		pe.Spec.GitSettings.Repository,
		branch,
	)

	gitSettings := coflnetv1alpha1.InstanceGitSettings{
		Branch: strPtr(branch),
	}

	return createInstanceFromEnvironment(pe, name, gitSettings)
}

func createInstanceFromEnvironment(pe coflnetv1alpha1.PreviewEnvironment, name string, gitSettings coflnetv1alpha1.InstanceGitSettings) *coflnetv1alpha1.PreviewEnvironmentInstance {
	return &coflnetv1alpha1.PreviewEnvironmentInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: pe.Namespace,
			Labels: map[string]string{
				"owner":               pe.GetOwner(),
				"previewenvironment":  string(pe.GetUID()),
				"github-organization": pe.Spec.GitSettings.Organization,
				"github-repository":   pe.Spec.GitSettings.Repository,
				"github-identifier":   gitSettings.BranchOrPullRequestIdentifier(),
			},
		},
		Spec: coflnetv1alpha1.PreviewEnvironmentInstanceSpec{
			InstanceGitSettings: gitSettings,
			DesiredPhase:        coflnetv1alpha1.InstancePhaseRunning,
		},
		Status: coflnetv1alpha1.PreviewEnvironmentInstanceStatus{
			Phase: coflnetv1alpha1.InstancePhasePending,
		},
	}
}

func (r *PreviewEnvironmentReconciler) savePreviewEnvironmentInstances(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, peis []*coflnetv1alpha1.PreviewEnvironmentInstance) error {
	for _, pei := range peis {
		if err := r.savePreviewEnvironmentInstance(ctx, pei); err != nil {
			return err
		}
	}
	return nil
}

func (r *PreviewEnvironmentReconciler) savePreviewEnvironmentInstance(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	r.log.Info("saving preview environment instance", "namespace", pei.Namespace, "name", pei.Name)

	existingPei := &coflnetv1alpha1.PreviewEnvironmentInstance{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: pei.Namespace,
			Name:      pei.Name,
		},
	}

	err := r.Get(ctx, client.ObjectKey{Namespace: pei.Namespace, Name: pei.Name}, existingPei)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	// update the preview environment instance
	if existingPei.Spec.DesiredPhase != "" {
		pei.ObjectMeta = existingPei.ObjectMeta
		err = r.Update(ctx, pei)
		if err != nil {
			return err
		}

		return nil
	}

	// create the preview environment instance
	return r.Create(ctx, pei)
}

func (r *PreviewEnvironmentReconciler) deletePreviewEnvironmentInstancesForPreviewEnvironment(ctx context.Context, pr coflnetv1alpha1.PreviewEnvironment) error {
	r.log.Info("deleting preview environment instances for preview environment", "namespace", pr.Namespace, "name", pr.Name)

	var peis coflnetv1alpha1.PreviewEnvironmentInstanceList
	if err := r.List(ctx, &peis, client.InNamespace(pr.Namespace), &client.ListOptions{}); err != nil {
		return err
	}

	r.log.Info("loaded preview environment instances", "count", len(peis.Items), "namespace", pr.Namespace, "name", pr.Name)
	for _, pei := range peis.Items {
		r.log.Info("deleting preview environment instance", "pei", pei.Name, "namespace", pei.Namespace)
		if err := r.Delete(ctx, &pei); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PreviewEnvironmentReconciler) SetupWithManager(mgr ctrl.Manager, gc *git.GithubClient) error {
	r.log = log.FromContext(context.TODO())

	r.githubClient = gc

	// setup the indexer stuff
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &coflnetv1alpha1.PreviewEnvironmentInstance{}, "spec.previewEnvironmentRef.name", func(o client.Object) []string {
		pei := o.(*coflnetv1alpha1.PreviewEnvironmentInstance)
		owner := metav1.GetControllerOf(pei)
		if owner == nil {
			return nil
		}

		r.log.Info("indexing preview environment instance", "pei", pei.Name, "namespace", pei.Namespace, "owner", owner.Name, "ownerKind", owner.Kind)
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&coflnetv1alpha1.PreviewEnvironment{}).
		Named("previewenvironment").
		Complete(r)
}

func intPtr(i int) *int {
	return &i
}
