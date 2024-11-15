package kubeclient

import (
	"context"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *KubeClient) ListPreviewEnvironments(ctx context.Context, owner string) (*coflnetv1alpha1.PreviewEnvironmentList, error) {
	k.log.Info("Listing PreviewEnvironments from the cluster")

	var peiList coflnetv1alpha1.PreviewEnvironmentList
	err := k.kClient.List(ctx, &peiList, &client.ListOptions{
		Namespace:     namespace(),
		LabelSelector: labels.Set(map[string]string{"owner": owner}).AsSelector(),
	})

	if err != nil {
		return nil, err
	}
	return &peiList, nil
}

func (k *KubeClient) PreviewEnvironmentOfUser(ctx context.Context, owner string) (*coflnetv1alpha1.PreviewEnvironmentList, error) {
	k.log.Info("Getting PreviewEnvironment from the cluster", "user", owner)

	var peList coflnetv1alpha1.PreviewEnvironmentList

	err := k.kClient.List(ctx, &peList, &client.ListOptions{
		Namespace:     namespace(),
		LabelSelector: labels.Set(map[string]string{"owner": owner}).AsSelector(),
	})

	if err != nil {
		return nil, err
	}

	return &peList, nil
}

func (k *KubeClient) PreviewEnvironmentById(ctx context.Context, owner string, id types.UID) (*coflnetv1alpha1.PreviewEnvironment, error) {
	k.log.Info("Getting PreviewEnvironment from the cluster", "owner", owner, "id", id)

	peList, err := k.PreviewEnvironmentOfUser(ctx, owner)
	if err != nil {
		return nil, err
	}

	for _, pe := range peList.Items {
		if pe.UID == id {
			return &pe, nil
		}
	}

	return nil, errors.NewNotFound(coflnetv1alpha1.PreviewEnvironmentGVR.GroupResource(), string(id))
}

func (k *KubeClient) PreviewEnvironmentByName(ctx context.Context, owner, name string) (*coflnetv1alpha1.PreviewEnvironment, error) {
	k.log.Info("Getting PreviewEnvironment from the cluster", "owner", owner, "name", name)

	peList, err := k.PreviewEnvironmentOfUser(ctx, owner)
	if err != nil {
		return nil, err
	}

	for _, pe := range peList.Items {
		if pe.Name == name {
			return &pe, nil
		}
	}

	return nil, errors.NewNotFound(coflnetv1alpha1.PreviewEnvironmentGVR.GroupResource(), name)
}

func (k *KubeClient) PreviewEnvironmentByDisplayName(ctx context.Context, owner, displayName string) (*coflnetv1alpha1.PreviewEnvironment, error) {
	k.log.Info("Getting PreviewEnvironment from the cluster", "owner", owner, "displayName", displayName)

	peList, err := k.PreviewEnvironmentOfUser(ctx, owner)
	if err != nil {
		return nil, err
	}

	for _, pe := range peList.Items {
		if pe.Spec.DisplayName == displayName {
			return &pe, nil
		}
	}

	return nil, errors.NewNotFound(coflnetv1alpha1.PreviewEnvironmentGVR.GroupResource(), displayName)
}

func (k *KubeClient) CreatePreviewEnvironment(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment) error {
	k.log.Info("Creating PreviewEnvironment in the cluster")
	pe.SetNamespace(namespace())

	err := k.kClient.Create(ctx, pe)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubeClient) DeletePreviewEnvironment(ctx context.Context, owner string, id types.UID) (*coflnetv1alpha1.PreviewEnvironment, error) {
	k.log.Info("Deleting PreviewEnvironment from the cluster", "id", id)
	pe, err := k.PreviewEnvironmentById(ctx, owner, id)
	if err != nil {
		return nil, err
	}

	err = k.kClient.Delete(ctx, pe)
	if err != nil {
		return nil, err
	}
	return pe, nil
}
