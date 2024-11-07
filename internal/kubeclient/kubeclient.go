package kubeclient

import (
	"context"
	"log"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	previewEnvironmentResource         = schema.GroupVersionResource{Group: "coflnet.coflnet.com", Version: "v1alpha1", Resource: "previewenvironments"}
	previewEnvironmentInstanceResource = schema.GroupVersionResource{Group: "coflnet.coflnet.com", Version: "v1alpha1", Resource: "previewenvironmentinstances"}
)

type KubeClient struct {
	dynamicClient *dynamic.DynamicClient
	kClient       client.Client
	log           logr.Logger
}

func NewKubeClient(logger logr.Logger, dynamicClient *dynamic.DynamicClient) *KubeClient {

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = coflnetv1alpha1.AddToScheme(scheme)

	kubeconfig := ctrl.GetConfigOrDie()
	controllerClient, err := client.New(kubeconfig, client.Options{Scheme: scheme})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &KubeClient{
		dynamicClient: dynamicClient,
		kClient:       controllerClient,
		log:           logger,
	}
}

func (k *KubeClient) TriggerUpdateForPreviewEnvironmentInstance(ctx context.Context, owner, repo string, prNumber int) error {
	k.log.Info("Triggering update for PreviewEnvironmentInstance", "owner", owner, "repo", repo, "prNumber", prNumber)

	var peiList coflnetv1alpha1.PreviewEnvironmentInstanceList
	err := k.kClient.List(ctx, &peiList, &client.ListOptions{})
	if err != nil {
		return err
	}

	for _, pei := range peiList.Items {
		if pei.Spec.GitOrganization == owner && pei.Spec.GitRepository == repo && pei.Spec.PullRequestNumber == prNumber {
			k.log.Info("Found matching PreviewEnvironmentInstance", "name", pei.GetName(), "namespace", pei.GetNamespace())

			// update the status field
			pei.Status.RebuildStatus = coflnetv1alpha1.RebuildStatusBuildingOutdated

			err := k.kClient.Status().Update(ctx, &pei)
			if err != nil {
				return err
			}

			k.log.Info("Updated PreviewEnvironmentInstance", "name", pei.GetName(), "status", coflnetv1alpha1.RebuildStatusBuildingOutdated, "namespace", pei.GetNamespace())
		}
	}

	return nil
}
