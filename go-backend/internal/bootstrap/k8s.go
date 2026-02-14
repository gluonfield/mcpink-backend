package bootstrap

import (
	"log/slog"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func k8sRestConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Info("Not running in-cluster, falling back to kubeconfig")
		kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func NewK8sClient() (kubernetes.Interface, error) {
	config, err := k8sRestConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func NewK8sDynamicClient() (dynamic.Interface, error) {
	config, err := k8sRestConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}
