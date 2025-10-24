package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// K8sClient implements the client.Client interface for real Kubernetes
type K8sClient struct {
	clientset       *kubernetes.Clientset
	metricsClient   *metricsv.Clientset
	ctx             context.Context
}

// NewClient creates a new Kubernetes client
// It tries in-cluster config first, then falls back to kubeconfig
func NewClient() (*K8sClient, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	// Create metrics client (optional - won't fail if metrics-server isn't available)
	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		// Metrics client creation failed, but we can continue without it
		metricsClient = nil
	}

	return &K8sClient{
		clientset:     clientset,
		metricsClient: metricsClient,
		ctx:           context.Background(),
	}, nil
}

// getKubeConfig attempts to get Kubernetes config from various sources
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first (for when running inside K8s)
	config, err := rest.InClusterConfig()
	if err == nil {
		// Increase rate limits to avoid throttling
		config.QPS = 100    // Default is 5
		config.Burst = 200  // Default is 10
		return config, nil
	}

	// Fall back to kubeconfig file
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
	}

	// Increase rate limits to avoid client-side throttling
	// Default QPS is 5 and Burst is 10, which is way too low for our use case
	config.QPS = 100    // Allow up to 100 queries per second
	config.Burst = 200  // Allow bursts of up to 200 requests

	return config, nil
}
