package client

import (
	"imperm-middleware/pkg/models"
)

// Client defines the interface for interacting with the Kubernetes middleware
type Client interface {
	// Environment operations
	ListEnvironments() ([]models.Environment, error)
	CreateEnvironment(name string, options *models.DeploymentOptions) error
	DestroyEnvironment(name string) error

	// Pod operations
	ListPods(namespace string) ([]models.Pod, error)
	GetPodLogs(namespace, podName string) (string, error)
	GetPodEvents(namespace, podName string) ([]models.Event, error)
	DeletePod(namespace, podName string) error

	// Deployment operations
	ListDeployments(namespace string) ([]models.Deployment, error)
	GetDeploymentEvents(namespace, deploymentName string) ([]models.Event, error)
	DeleteDeployment(namespace, deploymentName string) error

	// Metrics operations
	GetPodMetrics(namespace string) ([]models.PodMetrics, error)

	// Stats operations
	GetResourceStats(resourceType, namespace string) (*models.ResourceStats, error)

	// History
	GetEnvironmentHistory() ([]models.EnvironmentHistory, error)
}
