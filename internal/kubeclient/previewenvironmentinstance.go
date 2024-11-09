package kubeclient

import (
	"context"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *KubeClient) ListPreviewEnvironmentInstances(ctx context.Context, name string) (*coflnetv1alpha1.PreviewEnvironmentInstanceList, error) {
	k.log.Info("Listing PreviewEnvironmentInstances from the cluster", "name", name)

	// TODO: use the field selector to filter by name
	var peiList coflnetv1alpha1.PreviewEnvironmentInstanceList
	err := k.kClient.List(ctx, &peiList, &client.ListOptions{
		Namespace: namespace,
	})

	if err != nil {
		return nil, err
	}

	items := []coflnetv1alpha1.PreviewEnvironmentInstance{}
	for _, pei := range peiList.Items {
		if pei.Spec.PreviewEnvironmentRef.Name == name {
			k.log.Info("Found matching PreviewEnvironmentInstance", "name", pei.GetName(), "namespace", pei.GetNamespace())
			items = append(items, pei)
		}
	}

	peiList.Items = items
	return &peiList, nil
}
