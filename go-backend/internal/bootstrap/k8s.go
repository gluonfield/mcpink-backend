package bootstrap

import (
	"log/slog"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func NewK8sClient() (kubernetes.Interface, error) {
	// Try in-cluster config first (when running inside k8s)
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig (for local development)
		slog.Info("Not running in-cluster, falling back to kubeconfig")
		kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}
