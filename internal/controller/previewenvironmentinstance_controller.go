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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/coflnet/pr-env/internal/git"
	"github.com/go-logr/logr"
)

// PreviewEnvironmentInstanceReconciler reconciles a PreviewEnvironmentInstance object
type PreviewEnvironmentInstanceReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	log          logr.Logger
	githubClient *git.GithubClient
}

// +kubebuilder:rbac:groups=coflnet.coflnet.com,resources=previewenvironmentinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coflnet.coflnet.com,resources=previewenvironmentinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=coflnet.coflnet.com,resources=previewenvironmentinstances/finalizers,verbs=update
func (r *PreviewEnvironmentInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log = log.FromContext(ctx)

	// load the pei
	var pei coflnetv1alpha1.PreviewEnvironmentInstance
	if err := r.Get(ctx, req.NamespacedName, &pei); err != nil {
		r.log.Error(err, "unable to load the PreviewEnvironmentInstance", "namespace", req.Namespace, "name", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// check if the preview environment instance is being deleted
	if !pei.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&pei, finalizerName) {
			// do some deletion work
			err := r.deleteResourcesForPreviewEnvironmentInstance(ctx, &pei)
			if err != nil {
				r.log.Error(err, "Unable to delete resources dependent on preview environment instance", "namespace", req.Namespace, "name", req.Name)
				return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
			}

			// remove the finalizer
			controllerutil.RemoveFinalizer(&pei, finalizerName)
			if err := r.Update(ctx, &pei); err != nil {
				r.log.Error(err, "Unable to remove finalizer from PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// add the finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&pei, finalizerName) {
		controllerutil.AddFinalizer(&pei, finalizerName)
		if err := r.Update(ctx, &pei); err != nil {
			r.log.Error(err, "Unable to add finalizer to PreviewEnvironment", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, err
		}
	}

	pe, err := r.loadPreviewEnvironmentForInstance(ctx, &pei)
	if err != nil {
		return ctrl.Result{}, err
	}

	if pei.Status.RebuildStatus == coflnetv1alpha1.RebuildStatusBuilding || pei.Status.RebuildStatus == coflnetv1alpha1.RebuildStatusDeploying {
		r.log.Info("instance is already being rebuilt or deployed", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	if pei.Status.RebuildStatus == coflnetv1alpha1.RebuildStatusBuildingOutdated || pei.Status.RebuildStatus == coflnetv1alpha1.RebuildStatusFailed || pei.Status.RebuildStatus == "" {
		err := r.rebuildInstance(ctx, pe, &pei)
		if err != nil {
			r.log.Error(err, "unable to rebuild the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
			err = r.markPreviewEnvironmentInstanceAsFailed(ctx, &pei)
			if err != nil {
				r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as failed", "namespace", pei.Namespace, "name", pei.Name)
			}
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}

		r.log.Info("instance is being rebuilt", "namespace", pei.Namespace, "name", pei.Name)
		pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusDeploymentOutdated
		if err := r.Status().Update(ctx, &pei); err != nil {
			r.log.Error(err, "unable to update the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
			err = r.markPreviewEnvironmentInstanceAsFailed(ctx, &pei)
			if err != nil {
				r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as failed", "namespace", pei.Namespace, "name", pei.Name)
			}
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}

		r.log.Info("marked the instance as deployment outdated", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{}, nil
	}

	if pei.Status.RebuildStatus == coflnetv1alpha1.RebuildStatusDeploymentOutdated {
		r.log.Info("instance is being redeployed", "namespace", pei.Namespace, "name", pei.Name)
		err := r.redeployInstance(ctx, pe, &pei)
		if err != nil {
			r.log.Error(err, "unable to redeploy the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
			err = r.markPreviewEnvironmentInstanceAsFailed(ctx, &pei)
			if err != nil {
				r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as failed", "namespace", pei.Namespace, "name", pei.Name)
			}
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}

		r.log.Info("instance is being redeployed", "namespace", pei.Namespace, "name", pei.Name)
		pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusSuccess
		if err := r.Status().Update(ctx, &pei); err != nil {
			r.log.Error(err, "unable to update the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
			return ctrl.Result{}, err
		}

		r.log.Info("updating github pull request", "namespace", pei.Namespace, "name", pei.Name)
		if err := r.githubClient.UpdatePullRequestAnswer(ctx, pe, &pei); err != nil {
			r.log.Error(err, "unable to update the pull request", "namespace", pei.Namespace, "name", pei.Name)
			return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
		}

	}

	// fetching the latest information about the pull request
	pr, err := r.githubClient.PullRequest(ctx, pei.Spec.GitOrganization, pei.Spec.GitRepository, pei.Spec.PullRequestNumber)
	if err != nil {
		r.log.Error(err, "unable to fetch pull request")
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	// check if the pull request is closed
	if pr.GetState() == "closed" {
		r.log.Info("pull request is closed, we can ignore that one", "pull request number", pei.Spec.PullRequestNumber)
		return ctrl.Result{}, nil
	}

	// check if the pull request is merged
	if pr.GetMerged() {
		r.log.Info("pull request is merged, we can ignore that one", "pull request number", pei.Spec.PullRequestNumber)
		return ctrl.Result{}, nil
	}

	commitHash := pr.GetHead().GetSHA()
	if pei.Spec.CommitHash == commitHash {
		r.log.Info("commit hash is the same, don't need to update the instance", "commit hash", commitHash)
		return ctrl.Result{}, nil
	}

	pei.Spec.CommitHash = commitHash
	if err := r.Update(ctx, &pei); err != nil {
		r.log.Error(err, "unable to update the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{}, err
	}

	pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusBuildingOutdated
	if err := r.Status().Update(ctx, &pei); err != nil {
		r.log.Error(err, "unable to update the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PreviewEnvironmentInstanceReconciler) markPreviewEnvironmentInstanceAsFailed(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusFailed
	if err := r.Status().Update(ctx, pei); err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PreviewEnvironmentInstanceReconciler) SetupWithManager(mgr ctrl.Manager, gh *git.GithubClient) error {
	r.githubClient = gh

	return ctrl.NewControllerManagedBy(mgr).
		For(&coflnetv1alpha1.PreviewEnvironmentInstance{}).
		Named("previewenvironmentinstance").
		Complete(r)
}

func (r *PreviewEnvironmentInstanceReconciler) loadPreviewEnvironmentForInstance(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) (*coflnetv1alpha1.PreviewEnvironment, error) {
	var peList coflnetv1alpha1.PreviewEnvironmentList
	err := r.List(ctx, &peList, &client.ListOptions{
		Namespace: pei.GetNamespace(),
	})

	if err != nil {
		return nil, err
	}

	for _, pe := range peList.Items {
		if string(pe.GetUID()) == pei.GetLabels()["previewenvironment"] {
			return &pe, nil
		}
	}

	return nil, errors.NewNotFound(coflnetv1alpha1.PreviewEnvironmentGVR.GroupResource(), pei.GetLabels()["previewenvironment"])
}
