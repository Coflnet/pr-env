package kubeclient

import (
	"log"
	"os"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	coflnetv1alpha1 "github.com/coflnet/pr-env/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	previewEnvironmentResource         = schema.GroupVersionResource{Group: "coflnet.coflnet.com", Version: "v1alpha1", Resource: "previewenvironments"}
	previewEnvironmentInstanceResource = schema.GroupVersionResource{Group: "coflnet.coflnet.com", Version: "v1alpha1", Resource: "previewenvironmentinstances"}
)

type KubeClient struct {
	log          logr.Logger
	kClient      client.Client
	ownNamespace string
}

func NewKubeClient(logger logr.Logger) *KubeClient {
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
		log:     logger,
		kClient: controllerClient,
	}
}

func namespace() string {
	v := os.Getenv("NAMESPACE")
	if v == "" {
		panic("NAMESPACE environment variable not set")
	}
	return v
}
