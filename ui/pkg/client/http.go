package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"imperm-ui/pkg/models"
	"net/http"
)

// HTTPClient implements the Client interface for a real middleware API
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client for the middleware API
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// ListEnvironments fetches all environments from the middleware API
func (c *HTTPClient) ListEnvironments() ([]models.Environment, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/environments")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch environments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var envs []models.Environment
	if err := json.NewDecoder(resp.Body).Decode(&envs); err != nil {
		return nil, fmt.Errorf("failed to decode environments: %w", err)
	}

	// Handle null response
	if envs == nil {
		envs = []models.Environment{}
	}

	return envs, nil
}

// CreateEnvironment creates a new environment via the middleware API
func (c *HTTPClient) CreateEnvironment(name string, options *models.DeploymentOptions) error {
	payload := map[string]interface{}{
		"name":    name,
		"options": options,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/environments/create", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// DestroyEnvironment destroys an environment via the middleware API
func (c *HTTPClient) DestroyEnvironment(name string) error {
	payload := map[string]interface{}{
		"name": name,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/environments/destroy", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to destroy environment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// ListPods fetches all pods from the middleware API
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
	// For now, assuming the middleware transforms it to []models.Pod
	var pods []models.Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		return nil, fmt.Errorf("failed to decode pods: %w", err)
	}

	// Handle null response
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
	// Read the response body as plain text
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

// GetResourceStats fetches statistics for a resource type
func (c *HTTPClient) GetResourceStats(resourceType, namespace string) (*models.ResourceStats, error) {
	url := fmt.Sprintf("%s/api/stats?type=%s", c.baseURL, resourceType)
	if namespace != "" {
		url += "&namespace=" + namespace
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats models.ResourceStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return &stats, nil
}

// ListDeployments fetches all deployments from the middleware API
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
	// For now, assuming the middleware transforms it to []models.Deployment
	var deployments []models.Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return nil, fmt.Errorf("failed to decode deployments: %w", err)
	}

	// Handle null response
	if deployments == nil {
		deployments = []models.Deployment{}
	}

	return deployments, nil
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

	// Handle null response
	if metrics == nil {
		metrics = []models.PodMetrics{}
	}

	return metrics, nil
}

// GetEnvironmentHistory fetches the history of environment operations
func (c *HTTPClient) GetEnvironmentHistory() ([]models.EnvironmentHistory, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/environments/history")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var history []models.EnvironmentHistory
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("failed to decode history: %w", err)
	}

	// Handle null response
	if history == nil {
		history = []models.EnvironmentHistory{}
	}

	return history, nil
}

func (c *HTTPClient) DeletePod(namespace, podName string) error {
	url := fmt.Sprintf("%s/api/pods?namespace=%s&pod=%s", c.baseURL, namespace, podName)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *HTTPClient) DeleteDeployment(namespace, deploymentName string) error {
	url := fmt.Sprintf("%s/api/deployments?namespace=%s&deployment=%s", c.baseURL, namespace, deploymentName)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetOperationLogs fetches operation logs for an environment
func (c *HTTPClient) GetOperationLogs(environmentName string) (*models.OperationLogs, error) {
	url := fmt.Sprintf("%s/api/operations/logs?environment=%s", c.baseURL, environmentName)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch operation logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var logs models.OperationLogs
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	return &logs, nil
}
