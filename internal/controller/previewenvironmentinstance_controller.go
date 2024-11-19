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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/coflnet/pr-env/internal/git"
	"github.com/coflnet/pr-env/internal/keycloak"
	"github.com/go-logr/logr"
)

// PreviewEnvironmentInstanceReconciler reconciles a PreviewEnvironmentInstance object
type PreviewEnvironmentInstanceReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	log            logr.Logger
	githubClient   *git.GithubClient
	keycloakClient *keycloak.KeycloakClient
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

	// check if the instance has to be rebuild
	if pei.Status.Phase == coflnetv1alpha1.InstancePhasePending {
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

		err = r.markPreviewEnvironmentInstanceAsDeploying(ctx, &pei)
		if err != nil {
			r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as deploying", "namespace", pei.Namespace, "name", pei.Name)
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}

		return ctrl.Result{}, nil
	}

	// check if the instance has to be deployed
	if pei.Status.Phase == coflnetv1alpha1.InstancePhaseDeploying {
		err := r.redeployInstance(ctx, pe, &pei)
		if err != nil {
			r.log.Error(err, "unable to redeploy the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
			err = r.markPreviewEnvironmentInstanceAsFailed(ctx, &pei)
			if err != nil {
				r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as failed", "namespace", pei.Namespace, "name", pei.Name)
			}
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}
		r.log.Info("instance was deployed", "namespace", pei.Namespace, "name", pei.Name)

		r.log.Info("updating github pull request", "namespace", pei.Namespace, "name", pei.Name)
		if err := r.githubClient.UpdatePullRequestAnswer(ctx, pe, &pei); err != nil {
			r.log.Error(err, "unable to update the pull request", "namespace", pei.Namespace, "name", pei.Name)
			err = r.markPreviewEnvironmentInstanceAsFailed(ctx, &pei)
			if err != nil {
				r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as failed", "namespace", pei.Namespace, "name", pei.Name)
			}
			return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
		}

		err = r.markPreviewEnvironmentInstanceAsRunning(ctx, &pei)
		if err != nil {
			r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as running", "namespace", pei.Namespace, "name", pei.Name)
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}

		return ctrl.Result{}, nil
	}

	// refresh the latest commit hash to check if the instance is outdated
	latestCommitHash, err := r.latestCommitHashForPei(ctx, pe, &pei)
	if err != nil {
		r.log.Error(err, "unable to get the latest commit hash", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	if latestCommitHash == pei.Spec.InstanceGitSettings.CommitHash {
		r.log.Info("instance is up to date", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{}, nil
	}

	pei.Spec.InstanceGitSettings.CommitHash = latestCommitHash
	if err := r.Update(ctx, &pei); err != nil {
		r.log.Error(err, "unable to update the PreviewEnvironmentInstance", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{}, err
	}

	err = r.markPreviewEnvironmentInstanceAsPending(ctx, &pei)
	if err != nil {
		r.log.Error(err, "unable to mark the PreviewEnvironmentInstance as pending", "namespace", pei.Namespace, "name", pei.Name)
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	return ctrl.Result{}, nil
}

func (r *PreviewEnvironmentInstanceReconciler) markPreviewEnvironmentInstanceAsFailed(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	return r.markPreviewEnvironmentInstanceWithStatus(ctx, pei, coflnetv1alpha1.InstancePhaseFailed)
}

func (r *PreviewEnvironmentInstanceReconciler) markPreviewEnvironmentInstanceAsRunning(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	return r.markPreviewEnvironmentInstanceWithStatus(ctx, pei, coflnetv1alpha1.InstancePhaseRunning)
}

func (r *PreviewEnvironmentInstanceReconciler) markPreviewEnvironmentInstanceAsDeploying(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	return r.markPreviewEnvironmentInstanceWithStatus(ctx, pei, coflnetv1alpha1.InstancePhaseDeploying)
}

func (r *PreviewEnvironmentInstanceReconciler) markPreviewEnvironmentInstanceAsBuilding(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	return r.markPreviewEnvironmentInstanceWithStatus(ctx, pei, coflnetv1alpha1.InstancePhaseBuilding)
}

func (r *PreviewEnvironmentInstanceReconciler) markPreviewEnvironmentInstanceAsPending(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance) error {
	return r.markPreviewEnvironmentInstanceWithStatus(ctx, pei, coflnetv1alpha1.InstancePhasePending)
}

func (r *PreviewEnvironmentInstanceReconciler) markPreviewEnvironmentInstanceWithStatus(ctx context.Context, pei *coflnetv1alpha1.PreviewEnvironmentInstance, status string) error {
	pei.Status.Phase = status
	return r.Status().Update(ctx, pei)
}

func (r *PreviewEnvironmentInstanceReconciler) latestCommitHashForPei(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment, pei *coflnetv1alpha1.PreviewEnvironmentInstance) (string, error) {

	// TODO: check if the pei is a branch pei
	// and update the pull request based on that
	if pei.Spec.InstanceGitSettings.PullRequestNumber == nil {
		return "", fmt.Errorf("not implemented")
	}

	pr, err := r.githubClient.PullRequestOfPei(ctx, pe, pei)
	if err != nil {
		return "", err
	}
	return pr.GetHead().GetSHA(), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PreviewEnvironmentInstanceReconciler) SetupWithManager(mgr ctrl.Manager, gh *git.GithubClient, kClient *keycloak.KeycloakClient) error {
	r.githubClient = gh
	r.keycloakClient = kClient

	return ctrl.NewControllerManagedBy(mgr).
		For(&coflnetv1alpha1.PreviewEnvironmentInstance{}).
		Named("previewenvironmentinstance").
		Complete(r)
}

// PERF: loadPreviewEnvironmentForInstance loads the preview environment for the given instance
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
