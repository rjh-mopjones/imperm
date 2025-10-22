package middleware

import (
	"fmt"
	"imperm/internal/models"
	"time"
)

// MockClient is a mock implementation of the Client interface
type MockClient struct {
	environments []models.Environment
	history      []models.EnvironmentHistory
}

// NewMockClient creates a new mock client with sample data
func NewMockClient() *MockClient {
	now := time.Now()
	return &MockClient{
		environments: []models.Environment{
			{
				Name:      "dev-env-1",
				Namespace: "default",
				Status:    "Running",
				Age:       now.Add(-2 * time.Hour),
				Pods: []models.Pod{
					{
						Name:      "app-deployment-abc123",
						Namespace: "default",
						Status:    "Running",
						Ready:     "1/1",
						Restarts:  0,
						Age:       now.Add(-2 * time.Hour),
						CPU:       "150m",
						Memory:    "256Mi",
					},
					{
						Name:      "app-deployment-def456",
						Namespace: "default",
						Status:    "Running",
						Ready:     "1/1",
						Restarts:  2,
						Age:       now.Add(-1 * time.Hour),
						CPU:       "75m",
						Memory:    "128Mi",
					},
				},
				Deployments: []models.Deployment{
					{
						Name:      "app-deployment",
						Namespace: "default",
						Ready:     "2/2",
						UpToDate:  2,
						Available: 2,
						Age:       now.Add(-2 * time.Hour),
					},
				},
			},
			{
				Name:      "staging-env-1",
				Namespace: "staging",
				Status:    "Running",
				Age:       now.Add(-24 * time.Hour),
				Pods: []models.Pod{
					{
						Name:      "nginx-pod-xyz789",
						Namespace: "staging",
						Status:    "Running",
						Ready:     "1/1",
						Restarts:  0,
						Age:       now.Add(-24 * time.Hour),
						CPU:       "50m",
						Memory:    "64Mi",
					},
				},
				Deployments: []models.Deployment{
					{
						Name:      "nginx-deployment",
						Namespace: "staging",
						Ready:     "1/1",
						UpToDate:  1,
						Available: 1,
						Age:       now.Add(-24 * time.Hour),
					},
				},
			},
		},
		history: []models.EnvironmentHistory{
			{
				Name:        "dev-env-1",
				LaunchedAt:  now.Add(-2 * time.Hour),
				Status:      "Success",
				WithOptions: false,
			},
			{
				Name:        "staging-env-1",
				LaunchedAt:  now.Add(-24 * time.Hour),
				Status:      "Success",
				WithOptions: true,
			},
		},
	}
}

func (m *MockClient) ListEnvironments() ([]models.Environment, error) {
	return m.environments, nil
}

func (m *MockClient) CreateEnvironment(name string, withOptions bool) error {
	// Simulate environment creation
	now := time.Now()
	newEnv := models.Environment{
		Name:      name,
		Namespace: "default",
		Status:    "Creating",
		Age:       now,
		Pods:      []models.Pod{},
	}
	m.environments = append(m.environments, newEnv)

	// Add to history
	historyEntry := models.EnvironmentHistory{
		Name:        name,
		LaunchedAt:  now,
		Status:      "Success",
		WithOptions: withOptions,
	}
	m.history = append(m.history, historyEntry)

	return nil
}

func (m *MockClient) DestroyEnvironment(name string) error {
	// Simulate environment destruction
	for i, env := range m.environments {
		if env.Name == name {
			m.environments = append(m.environments[:i], m.environments[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("environment %s not found", name)
}

func (m *MockClient) ListPods(namespace string) ([]models.Pod, error) {
	var pods []models.Pod
	for _, env := range m.environments {
		if namespace == "" || env.Namespace == namespace {
			pods = append(pods, env.Pods...)
		}
	}
	return pods, nil
}

func (m *MockClient) GetPodLogs(namespace, podName string) (string, error) {
	return fmt.Sprintf("Mock logs for pod %s in namespace %s", podName, namespace), nil
}

func (m *MockClient) ListDeployments(namespace string) ([]models.Deployment, error) {
	var deployments []models.Deployment
	for _, env := range m.environments {
		if namespace == "" || env.Namespace == namespace {
			deployments = append(deployments, env.Deployments...)
		}
	}
	return deployments, nil
}

func (m *MockClient) GetEnvironmentHistory() ([]models.EnvironmentHistory, error) {
	return m.history, nil
}
