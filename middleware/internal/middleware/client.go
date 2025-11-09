package middleware

import (
	"imperm-middleware/internal/models"
)

// Client defines the interface for interacting with the Kubernetes middleware
type Client interface {
	// Environment operations
	ListEnvironments() ([]models.Environment, error)
	CreateEnvironment(name string, withOptions bool) error
	DestroyEnvironment(name string) error

	// Pod operations
	ListPods(namespace string) ([]models.Pod, error)
	GetPodLogs(namespace, podName string) (string, error)

	// Deployment operations
	ListDeployments(namespace string) ([]models.Deployment, error)

	// History
	GetEnvironmentHistory() ([]models.EnvironmentHistory, error)
}
