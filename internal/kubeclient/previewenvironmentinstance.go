package kubeclient

import (
	"context"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *KubeClient) ListPreviewEnvironmentInstancesByPreviewEnvironmentId(ctx context.Context, owner string, id types.UID) (*coflnetv1alpha1.PreviewEnvironmentInstanceList, error) {
	k.log.Info("Listing PreviewEnvironmentInstances from the cluster", "owner", owner)

	var peiList coflnetv1alpha1.PreviewEnvironmentInstanceList
	err := k.kClient.List(ctx, &peiList, &client.ListOptions{
		Namespace: namespace,
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
