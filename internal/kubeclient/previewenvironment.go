package kubeclient

import (
	"context"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: figure out how to handle namespaces
const namespace = "default"

func (k *KubeClient) ListPreviewEnvironments(ctx context.Context) (*coflnetv1alpha1.PreviewEnvironmentList, error) {
	k.log.Info("Listing PreviewEnvironments from the cluster")

	var peiList coflnetv1alpha1.PreviewEnvironmentList
	err := k.kClient.List(ctx, &peiList, &client.ListOptions{})
	if err != nil {
		return nil, err
	}
	return &peiList, nil
}

func (k *KubeClient) PreviewEnvironmentByName(ctx context.Context, name string) (*coflnetv1alpha1.PreviewEnvironment, error) {
	k.log.Info("Getting PreviewEnvironment from the cluster", "name", name)

	var pe coflnetv1alpha1.PreviewEnvironment

	err := k.kClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &pe)
	if err != nil {
		return nil, err
	}
	return &pe, nil
}

func (k *KubeClient) CreatePreviewEnvironment(ctx context.Context, pe *coflnetv1alpha1.PreviewEnvironment) error {
	k.log.Info("Creating PreviewEnvironment in the cluster")
	pe.SetNamespace(namespace)

	err := k.kClient.Create(ctx, pe)
	if err != nil {
		return err
	}
	return nil
}

func (k *KubeClient) DeletePreviewEnvironment(ctx context.Context, name string) (*coflnetv1alpha1.PreviewEnvironment, error) {
	k.log.Info("Deleting PreviewEnvironment from the cluster", "name", name)
	pe, err := k.PreviewEnvironmentByName(ctx, name)
	if err != nil {
		return nil, err
	}

	err = k.kClient.Delete(ctx, pe)
	if err != nil {
		return nil, err
	}
	return pe, nil
}
