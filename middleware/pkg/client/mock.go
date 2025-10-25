package client

import (
	"fmt"
	"imperm-middleware/pkg/models"
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

func (m *MockClient) CreateEnvironment(name string, options *models.DeploymentOptions) error {
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
	hasOptions := options != nil && len(options.Variables) > 0
	historyEntry := models.EnvironmentHistory{
		Name:        name,
		LaunchedAt:  now,
		Status:      "Success",
		WithOptions: hasOptions,
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
	logs := fmt.Sprintf("Logs for pod: %s (namespace: %s)\n\n", podName, namespace)
	logs += "[2025-10-22 20:30:15] INFO: Container started\n"
	logs += "[2025-10-22 20:30:16] INFO: Initializing application\n"
	logs += "[2025-10-22 20:30:17] INFO: Loading configuration from /etc/config\n"
	logs += "[2025-10-22 20:30:18] INFO: Connecting to database at db.default.svc.cluster.local:5432\n"
	logs += "[2025-10-22 20:30:19] INFO: Database connection established\n"
	logs += "[2025-10-22 20:30:20] INFO: Starting HTTP server on port 8080\n"
	logs += "[2025-10-22 20:30:21] INFO: Server listening on 0.0.0.0:8080\n"
	logs += "[2025-10-22 20:30:25] INFO: Health check endpoint registered at /health\n"
	logs += "[2025-10-22 20:31:00] INFO: Processing request GET /api/health\n"
	logs += "[2025-10-22 20:31:30] INFO: Processing request GET /api/status\n"
	logs += "[2025-10-22 20:32:15] INFO: Processing request POST /api/data\n"
	logs += "[2025-10-22 20:32:45] WARN: High memory usage detected: 85%\n"
	logs += "[2025-10-22 20:33:00] INFO: Garbage collection completed\n"
	logs += "[2025-10-22 20:33:30] INFO: Memory usage normalized: 65%\n"
	return logs, nil
}

func (m *MockClient) GetPodEvents(namespace, podName string) ([]models.Event, error) {
	now := time.Now()
	events := []models.Event{
		{
			Type:      "Normal",
			Reason:    "Pulled",
			Message:   "Successfully pulled image \"nginx:latest\"",
			Timestamp: now.Add(-2 * time.Minute),
			Count:     1,
		},
		{
			Type:      "Normal",
			Reason:    "Created",
			Message:   "Created container nginx",
			Timestamp: now.Add(-5 * time.Minute),
			Count:     1,
		},
		{
			Type:      "Normal",
			Reason:    "Started",
			Message:   "Started container nginx",
			Timestamp: now.Add(-5 * time.Minute),
			Count:     1,
		},
		{
			Type:      "Warning",
			Reason:    "BackOff",
			Message:   "Back-off restarting failed container",
			Timestamp: now.Add(-1 * time.Hour),
			Count:     3,
		},
		{
			Type:      "Normal",
			Reason:    "Scheduled",
			Message:   fmt.Sprintf("Successfully assigned %s/%s to node-1", namespace, podName),
			Timestamp: now.Add(-2 * time.Hour),
			Count:     1,
		},
	}
	return events, nil
}

func (m *MockClient) GetDeploymentEvents(namespace, deploymentName string) ([]models.Event, error) {
	now := time.Now()
	events := []models.Event{
		{
			Type:      "Normal",
			Reason:    "ScalingReplicaSet",
			Message:   fmt.Sprintf("Scaled up replica set %s to 2", deploymentName),
			Timestamp: now.Add(-2 * time.Hour),
			Count:     1,
		},
		{
			Type:      "Normal",
			Reason:    "ScalingReplicaSet",
			Message:   fmt.Sprintf("Scaled down replica set %s to 1", deploymentName),
			Timestamp: now.Add(-3 * time.Hour),
			Count:     1,
		},
	}
	return events, nil
}

func (m *MockClient) GetResourceStats(resourceType, namespace string) (*models.ResourceStats, error) {
	stats := &models.ResourceStats{}

	switch resourceType {
	case "environments":
		stats.TotalEnvironments = len(m.environments)
		for _, env := range m.environments {
			stats.TotalPods += len(env.Pods)
			stats.TotalDeployments += len(env.Deployments)
		}
		stats.TotalCount = stats.TotalEnvironments

	case "pods":
		var pods []models.Pod
		for _, env := range m.environments {
			if namespace == "" || env.Namespace == namespace {
				pods = append(pods, env.Pods...)
			}
		}
		stats.TotalCount = len(pods)
		for _, pod := range pods {
			switch pod.Status {
			case "Running":
				stats.RunningPods++
			case "Pending":
				stats.PendingPods++
			case "Failed":
				stats.FailedPods++
			}
		}

	case "deployments":
		var deployments []models.Deployment
		for _, env := range m.environments {
			if namespace == "" || env.Namespace == namespace {
				deployments = append(deployments, env.Deployments...)
			}
		}
		stats.TotalCount = len(deployments)
		for _, dep := range deployments {
			stats.TotalReplicas += dep.UpToDate
			stats.AvailableReplicas += dep.Available
		}
	}

	return stats, nil
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

func (m *MockClient) DeletePod(namespace, podName string) error {
	// Find and remove the pod from environments
	for i := range m.environments {
		env := &m.environments[i]
		if env.Namespace == namespace {
			for j, pod := range env.Pods {
				if pod.Name == podName {
					// Remove pod from slice
					env.Pods = append(env.Pods[:j], env.Pods[j+1:]...)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("pod %s not found in namespace %s", podName, namespace)
}

func (m *MockClient) DeleteDeployment(namespace, deploymentName string) error {
	// Find and remove the deployment from environments
	for i := range m.environments {
		env := &m.environments[i]
		if env.Namespace == namespace {
			for j, dep := range env.Deployments {
				if dep.Name == deploymentName {
					// Remove deployment from slice
					env.Deployments = append(env.Deployments[:j], env.Deployments[j+1:]...)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("deployment %s not found in namespace %s", deploymentName, namespace)
}
