package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClient struct {
	ClientSet *kubernetes.Clientset
	Context   context.Context
}

func NewK8sClient() (*K8sClient, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// If not in cluster, use kubeconfig
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		if envKubeconfig := os.Getenv("KUBECONFIG"); envKubeconfig != "" {
			kubeconfigPath = envKubeconfig
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return &K8sClient{
		ClientSet: clientset,
		Context:   context.Background(),
	}, nil
}
