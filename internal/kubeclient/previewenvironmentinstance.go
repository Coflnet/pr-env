package kubeclient

import (
	"context"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *KubeClient) ListPreviewEnvironmentInstancesByPreviewEnvironmentId(ctx context.Context, owner string, id types.UID) (*coflnetv1alpha1.PreviewEnvironmentInstanceList, error) {
	k.log.Info("Listing PreviewEnvironmentInstances from the cluster", "owner", owner)

	var peiList coflnetv1alpha1.PreviewEnvironmentInstanceList
	err := k.kClient.List(ctx, &peiList, &client.ListOptions{
		Namespace: namespace(),
		LabelSelector: labels.Set(map[string]string{
			"owner":              owner,
			"previewenvironment": string(id),
		}).AsSelector(),
	})

	if err != nil {
		return nil, err
	}

	return &peiList, nil
}

func (k *KubeClient) PreviewEnvironmentByOrganizationRepoAndIdentifier(ctx context.Context, organization, repo, identifier string) (*coflnetv1alpha1.PreviewEnvironmentInstance, error) {

	labelSelector := labels.Set(map[string]string{
		"github-organization": organization,
		"github-repository":   repo,
		"github-identifier":   identifier,
	}).AsSelector()
	k.log.Info("Getting PreviewEnvironmentInstance from the cluster", "organization", organization, "repository", repo, "identifier", identifier)

	var peiList coflnetv1alpha1.PreviewEnvironmentInstanceList
	err := k.kClient.List(ctx, &peiList, &client.ListOptions{
		Namespace:     namespace(),
		LabelSelector: labelSelector,
	})

	if err != nil {
		return nil, err
	}

	if len(peiList.Items) == 0 {
		return nil, errors.NewNotFound(coflnetv1alpha1.PreviewEnvironmentInstanceGVR.GroupResource(), labelSelector.String())
	}

	return &peiList.Items[0], nil
}

func (k *KubeClient) TriggerUpdateForPreviewEnvironmentInstance(ctx context.Context, owner string, peId types.UID, branchOrPullRequestIdentifier string) error {
	peiList, err := k.ListPreviewEnvironmentInstancesByPreviewEnvironmentId(ctx, owner, peId)
	if err != nil {
		return err
	}

	for _, pei := range peiList.Items {
		if pei.BranchOrPullRequestIdentifier() == branchOrPullRequestIdentifier {
			k.log.Info("Found matching PreviewEnvironmentInstance", "name", pei.GetName(), "namespace", pei.GetNamespace())

			// update the status field
			pei.Status.Phase = coflnetv1alpha1.InstancePhasePending
			err := k.kClient.Status().Update(ctx, &pei)
			if err != nil {
				return err
			}

			k.log.Info("Updated PreviewEnvironmentInstance", "name", pei.GetName(), "phase", coflnetv1alpha1.InstancePhasePending, "namespace", pei.GetNamespace())
		}
	}

	return nil
}
