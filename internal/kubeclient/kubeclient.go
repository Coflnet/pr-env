package kubeclient

import (
	"context"
	"encoding/json"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var (
	previewEnvironmentResource         = schema.GroupVersionResource{Group: "coflnet.coflnet.com", Version: "v1alpha1", Resource: "previewenvironments"}
	previewEnvironmentInstanceResource = schema.GroupVersionResource{Group: "coflnet.coflnet.com", Version: "v1alpha1", Resource: "previewenvironmentinstances"}
)

type KubeClient struct {
	dynamicClient *dynamic.DynamicClient
	log           logr.Logger
}

func NewKubeClient(logger logr.Logger, dynamicClient *dynamic.DynamicClient) *KubeClient {
	return &KubeClient{
		dynamicClient: dynamicClient,
		log:           logger,
	}
}

func (k *KubeClient) TriggerUpdateForPreviewEnvironmentInstance(ctx context.Context, owner, repo string, prNumber int) error {
	k.log.Info("Triggering update for PreviewEnvironmentInstance", "owner", owner, "repo", repo, "prNumber", prNumber)

	list, err := k.dynamicClient.Resource(previewEnvironmentInstanceResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		ownerVal, _, _ := unstructured.NestedString(item.Object, "spec", "gitOrganization")
		repoVal, _, _ := unstructured.NestedString(item.Object, "spec", "gitRepository")
		prNumberVal, _, _ := unstructured.NestedInt64(item.Object, "spec", "pullRequestNumber")

		if ownerVal == owner && repoVal == repo && int(prNumberVal) == prNumber {
			k.log.Info("Found matching PreviewEnvironmentInstance", "name", item.GetName())

			patch := []map[string]interface{}{
				{
					"op":    "replace",
					"path":  "/status/rebuildStatus",
					"value": coflnetv1alpha1.RebuildStatusBuilding,
				},
				{
					"op":    "replace",
					"path":  "/spec/commitHash",
					"value": "newcommit",
				},
			}

			payload, err := json.Marshal(patch)
			if err != nil {
				return err
			}

			// Apply the JSON Patch to the resource
			_, err = k.dynamicClient.Resource(previewEnvironmentInstanceResource).Namespace(item.GetNamespace()).Patch(context.TODO(), item.GetName(), types.JSONPatchType, payload, metav1.PatchOptions{})
			if err != nil {
				return err
			}

			k.log.Info("Updated PreviewEnvironmentInstance", "name", item.GetName(), "status", coflnetv1alpha1.RebuildStatusBuildingOutdated, "namespace", item.GetNamespace())
		}
	}

	return nil
}
