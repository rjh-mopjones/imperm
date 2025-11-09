package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"imperm-middleware/pkg/models"
	"net/http"
)

// HTTPClient implements the Client interface for a real upstream API
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client for the upstream API
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// ListEnvironments fetches all environments from the upstream API
func (c *HTTPClient) ListEnvironments() ([]models.Environment, error) {
	// This would need to be implemented based on how environments are managed
	// For now, return empty list as environments are managed separately
	return []models.Environment{}, nil
}

// CreateEnvironment creates a new environment
func (c *HTTPClient) CreateEnvironment(name string, options *models.DeploymentOptions) error {
	// Environment creation would be managed outside the upstream API
	return fmt.Errorf("not implemented via upstream API")
}

// DestroyEnvironment destroys an environment
func (c *HTTPClient) DestroyEnvironment(name string) error {
	// Environment destruction would be managed outside the upstream API
	return fmt.Errorf("not implemented via upstream API")
}

// ListPods fetches all pods from the upstream API
func (c *HTTPClient) ListPods(namespace string) ([]models.Pod, error) {
	url := fmt.Sprintf("%s/api/k8s/%s/pods", c.baseURL, namespace)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pods: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// The upstream returns v1.PodList, we need to handle this
	// For now, assuming it's transformed to []models.Pod
	var pods []models.Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		return nil, fmt.Errorf("failed to decode pods: %w", err)
	}

	if pods == nil {
		pods = []models.Pod{}
	}

	return pods, nil
}

// GetPodLogs fetches logs for a specific pod
func (c *HTTPClient) GetPodLogs(namespace, podName string) (string, error) {
	url := fmt.Sprintf("%s/api/k8s/%s/logs?podName=%s&follow=false", c.baseURL, namespace, podName)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch pod logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// The upstream is a flushwriter that writes logs directly
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return buf.String(), nil
}

// GetPodEvents fetches events for a specific pod
func (c *HTTPClient) GetPodEvents(namespace, podName string) ([]models.Event, error) {
	url := fmt.Sprintf("%s/api/k8s/%s/events", c.baseURL, namespace)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pod events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Upstream returns []EventTimeEntry, convert to []Event
	var eventEntries []models.EventTimeEntry
	if err := json.NewDecoder(resp.Body).Decode(&eventEntries); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	// Convert EventTimeEntry to Event
	events := make([]models.Event, 0, len(eventEntries))
	for _, entry := range eventEntries {
		events = append(events, models.Event{
			Type:      entry.Type,
			Reason:    entry.Reason,
			Message:   entry.Message,
			Timestamp: entry.LastObservedTime,
			Count:     entry.Count,
		})
	}

	return events, nil
}

// DeletePod deletes a specific pod
func (c *HTTPClient) DeletePod(namespace, podName string) error {
	// This would need to call the upstream API's delete endpoint
	return fmt.Errorf("not implemented via upstream API")
}

// ListDeployments fetches all deployments from the upstream API
func (c *HTTPClient) ListDeployments(namespace string) ([]models.Deployment, error) {
	url := fmt.Sprintf("%s/api/k8s/%s/deployments", c.baseURL, namespace)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// The upstream returns appsv1.DeploymentList
	// For now, assuming it's transformed to []models.Deployment
	var deployments []models.Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return nil, fmt.Errorf("failed to decode deployments: %w", err)
	}

	if deployments == nil {
		deployments = []models.Deployment{}
	}

	return deployments, nil
}

// GetDeploymentEvents fetches events for a specific deployment
func (c *HTTPClient) GetDeploymentEvents(namespace, deploymentName string) ([]models.Event, error) {
	url := fmt.Sprintf("%s/api/k8s/%s/events", c.baseURL, namespace)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deployment events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Upstream returns []EventTimeEntry, convert to []Event
	var eventEntries []models.EventTimeEntry
	if err := json.NewDecoder(resp.Body).Decode(&eventEntries); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	// Convert EventTimeEntry to Event
	events := make([]models.Event, 0, len(eventEntries))
	for _, entry := range eventEntries {
		events = append(events, models.Event{
			Type:      entry.Type,
			Reason:    entry.Reason,
			Message:   entry.Message,
			Timestamp: entry.LastObservedTime,
			Count:     entry.Count,
		})
	}

	return events, nil
}

// DeleteDeployment deletes a specific deployment
func (c *HTTPClient) DeleteDeployment(namespace, deploymentName string) error {
	// This would need to call the upstream API's delete endpoint
	return fmt.Errorf("not implemented via upstream API")
}

// GetPodMetrics fetches resource metrics for all pods in a namespace
func (c *HTTPClient) GetPodMetrics(namespace string) ([]models.PodMetrics, error) {
	url := fmt.Sprintf("%s/api/k8s/%s/metrics", c.baseURL, namespace)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pod metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var metrics []models.PodMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode metrics: %w", err)
	}

	if metrics == nil {
		metrics = []models.PodMetrics{}
	}

	return metrics, nil
}

// GetResourceStats fetches statistics for a resource type
func (c *HTTPClient) GetResourceStats(resourceType, namespace string) (*models.ResourceStats, error) {
	// This would aggregate data from various endpoints
	// For now, return empty stats
	return &models.ResourceStats{}, nil
}

// GetEnvironmentHistory fetches the history of environment operations
func (c *HTTPClient) GetEnvironmentHistory() ([]models.EnvironmentHistory, error) {
	// Environment history would be managed outside the upstream API
	return []models.EnvironmentHistory{}, nil
}
