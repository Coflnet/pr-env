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
	"fmt"
	"time"

	"github.com/google/go-github/v66/github"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/coflnet/pr-env/pkg/git"
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

	// load the preview environment
	var pr coflnetv1alpha1.PreviewEnvironment
	if err := r.Get(ctx, req.NamespacedName, &pr); err != nil {
		r.log.Info("Unable to load PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// check if the preview environment is being deleted
	if !pr.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&pr, finalizerName) {
			// do some deletion work
			err := r.deletePreviewEnvironmentInstancesForPreviewEnvironment(ctx, pr)
			if err != nil {
				r.log.Error(err, "Unable to delete PreviewEnvironmentInstances", "namespace", req.Namespace, "name", req.Name)
				return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
			}

			// remove the finalizer
			controllerutil.RemoveFinalizer(&pr, finalizerName)
			if err := r.Update(ctx, &pr); err != nil {
				r.log.Error(err, "Unable to remove finalizer from PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// add the finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&pr, finalizerName) {
		controllerutil.AddFinalizer(&pr, finalizerName)
		if err := r.Update(ctx, &pr); err != nil {
			r.log.Error(err, "Unable to add finalizer to PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, err
		}
	}

	// list all the available branches of the specific preview environment
	prs, err := r.githubClient.PullRequestsOfRepository(ctx, pr.Spec.GitOrganization, pr.Spec.GitRepository)
	if err != nil {
		r.log.Error(err, "Unable to list branches of repository", "namespace", req.Namespace, "name", req.Name, "owner", pr.Spec.GitOrganization, "repo", pr.Spec.GitRepository)
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	prIdentifiers := make([]int, len(prs))
	for i, pr := range prs {
		prIdentifiers[i] = int(*pr.ID)
	}

	// update the status of the preview environment
	pr.Status.PullRequestsDetected = prIdentifiers
	r.log.Info("detected pullrequests for pe", "branches", pr.Status.PullRequestsDetected, "namespace", req.Namespace, "name", req.Name)

	// create the preview environment instances that should be created for each branch
	err = r.createPreviewEnvironmentInstancesForDetectedPullRequests(ctx, pr, prs)
	if err != nil {
		r.log.Error(err, "Unable to create PreviewEnvironment instances", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{
			RequeueAfter: time.Minute * 1,
		}, nil
	}

	if err := r.Status().Update(ctx, &pr); err != nil {
		r.log.Error(err, "Unable to update status of PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PreviewEnvironmentReconciler) createPreviewEnvironmentInstancesForDetectedPullRequests(ctx context.Context, pr coflnetv1alpha1.PreviewEnvironment, prs []*github.PullRequest) error {
	for _, githubPr := range prs {
		peiName := coflnetv1alpha1.PreviewEnvironmentInstanceNameFromPullRequest(pr.Name, pr.Spec.GitOrganization, pr.Spec.GitRepository, int(*githubPr.Number))
		r.log.Info("check if a preview environment instance already exists", "pei", peiName, "namespace", pr.Namespace)

		pei := coflnetv1alpha1.PreviewEnvironmentInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      peiName,
				Namespace: pr.Namespace,
			},
		}

		// check if the preview environment instance already exists
		err := r.Get(ctx, client.ObjectKey{Namespace: pei.Namespace, Name: pei.Name}, &pei)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			}
		}
		if pei.Spec.PullRequestNumber == *githubPr.Number {
			r.log.Info("pei exists with the same branch, skip creating it", "pei", pei.Name, "namespace", pei.Namespace)
			continue
		}
		fmt.Printf("%v\n", pei)

		pei = coflnetv1alpha1.PreviewEnvironmentInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      peiName,
				Namespace: pr.Namespace,
			},
			Spec: coflnetv1alpha1.PreviewEnvironmentInstanceSpec{
				PullRequestNumber: *githubPr.Number,
				Branch:            githubPr.Head.Ref,
				GitOrganization:   pr.Spec.GitOrganization,
				GitRepository:     pr.Spec.GitRepository,
				PreviewEnvironmentRef: coflnetv1alpha1.PreviewEnvironmentRef{
					Name:      pr.Name,
					Namespace: pr.Namespace,
				},
			},
			Status: coflnetv1alpha1.PreviewEnvironmentInstanceStatus{
				RebuildStatus: coflnetv1alpha1.RebuildStatusBuildingOutdated,
			},
		}

		// create the preview environment instance
		r.log.Info("pei does not exist yet, creating it", "pei", pei.Name, "namespace", pei.Namespace)
		if err := r.Create(ctx, &pei); err != nil {
			return err
		}
		r.log.Info("created preview environment instance", "pei", pei.Name, "namespace", pei.Namespace)
	}
	return nil
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
