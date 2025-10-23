package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"imperm-middleware/internal/k8s"
	"imperm-middleware/internal/terraform"
	"imperm-middleware/pkg/client"
	"imperm-middleware/pkg/models"
)

type Handler struct {
	client client.Client
}

type HandlerMode string

const (
	ModeMock      HandlerMode = "mock"
	ModeK8s       HandlerMode = "k8s"
	ModeTerraform HandlerMode = "terraform"
)

func NewHandler(mode HandlerMode) *Handler {
	var c client.Client

	switch mode {
	case ModeMock:
		log.Println("Initializing mock client...")
		c = client.NewMockClient()

	case ModeTerraform:
		log.Println("Initializing Terraform client...")

		// Get paths
		projectRoot := getProjectRoot()
		terraformDir := filepath.Join(projectRoot, "terraform", "environments")
		modulePath := filepath.Join(projectRoot, "terraform", "modules", "k8s-namespace")
		kubeconfig := getKubeconfig()

		tfClient, err := terraform.NewClient(terraformDir, modulePath, kubeconfig)
		if err != nil {
			log.Fatalf("Failed to create Terraform client: %v", err)
		}
		c = tfClient
		log.Println("Successfully initialized Terraform client")

	default: // ModeK8s
		log.Println("Initializing Kubernetes client...")
		k8sClient, err := k8s.NewClient()
		if err != nil {
			log.Fatalf("Failed to create Kubernetes client: %v", err)
		}
		c = k8sClient
		log.Println("Successfully connected to Kubernetes cluster")
	}

	return &Handler{
		client: c,
	}
}

// getProjectRoot returns the project root directory
func getProjectRoot() string {
	// Try to get from environment variable first
	if root := os.Getenv("IMPERM_PROJECT_ROOT"); root != "" {
		return root
	}

	// Otherwise use current working directory and go up until we find terraform/
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: failed to get working directory: %v", err)
		return "."
	}

	// Walk up the directory tree looking for the terraform directory
	dir := cwd
	for {
		tfDir := filepath.Join(dir, "terraform")
		if _, err := os.Stat(tfDir); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, use cwd
			return cwd
		}
		dir = parent
	}
}

// getKubeconfig returns the path to the kubeconfig file
func getKubeconfig() string {
	// Try KUBECONFIG environment variable first
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Default to ~/.kube/config
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: failed to get home directory: %v", err)
		return ""
	}

	return filepath.Join(home, ".kube", "config")
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Environment endpoints
	mux.HandleFunc("/api/environments", h.handleEnvironments)
	mux.HandleFunc("/api/environments/create", h.handleCreateEnvironment)
	mux.HandleFunc("/api/environments/destroy", h.handleDestroyEnvironment)
	mux.HandleFunc("/api/environments/history", h.handleEnvironmentHistory)

	// Pod endpoints
	mux.HandleFunc("/api/pods", h.handlePods)
	mux.HandleFunc("/api/pods/logs", h.handlePodLogs)
	mux.HandleFunc("/api/pods/events", h.handlePodEvents)

	// Deployment endpoints
	mux.HandleFunc("/api/deployments", h.handleDeployments)
	mux.HandleFunc("/api/deployments/events", h.handleDeploymentEvents)

	// Stats endpoints
	mux.HandleFunc("/api/stats", h.handleStats)

	// Terraform operation logs
	mux.HandleFunc("/api/operations/logs", h.handleOperationLogs)

	// Health check
	mux.HandleFunc("/health", h.handleHealth)
}

func (h *Handler) handleEnvironments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	envs, err := h.client.ListEnvironments()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, envs)
}

func (h *Handler) handleCreateEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name    string                  `json:"name"`
		Options *models.DeploymentOptions `json:"options"`
		// Keep backward compatibility
		WithOptions bool `json:"with_options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// If no options provided but withOptions is true, create empty options
	if req.Options == nil && req.WithOptions {
		req.Options = &models.DeploymentOptions{
			Name: req.Name,
		}
	}

	err := h.client.CreateEnvironment(req.Name, req.Options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, map[string]string{"status": "created"})
}

func (h *Handler) handleDestroyEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.client.DestroyEnvironment(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"status": "destroyed"})
}

func (h *Handler) handleEnvironmentHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history, err := h.client.GetEnvironmentHistory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, history)
}

func (h *Handler) handlePods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		namespace := r.URL.Query().Get("namespace")
		pods, err := h.client.ListPods(namespace)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, pods)

	case http.MethodDelete:
		namespace := r.URL.Query().Get("namespace")
		podName := r.URL.Query().Get("pod")

		if namespace == "" || podName == "" {
			http.Error(w, "namespace and pod parameters are required", http.StatusBadRequest)
			return
		}

		err := h.client.DeletePod(namespace, podName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleDeployments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		namespace := r.URL.Query().Get("namespace")
		deployments, err := h.client.ListDeployments(namespace)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, deployments)

	case http.MethodDelete:
		namespace := r.URL.Query().Get("namespace")
		deploymentName := r.URL.Query().Get("deployment")

		if namespace == "" || deploymentName == "" {
			http.Error(w, "namespace and deployment parameters are required", http.StatusBadRequest)
			return
		}

		err := h.client.DeleteDeployment(namespace, deploymentName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondJSON(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handlePodLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")

	if namespace == "" || podName == "" {
		http.Error(w, "namespace and pod parameters are required", http.StatusBadRequest)
		return
	}

	logs, err := h.client.GetPodLogs(namespace, podName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"logs": logs})
}

func (h *Handler) handlePodEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("pod")

	if namespace == "" || podName == "" {
		http.Error(w, "namespace and pod parameters are required", http.StatusBadRequest)
		return
	}

	events, err := h.client.GetPodEvents(namespace, podName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, events)
}

func (h *Handler) handleDeploymentEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	namespace := r.URL.Query().Get("namespace")
	deploymentName := r.URL.Query().Get("deployment")

	if namespace == "" || deploymentName == "" {
		http.Error(w, "namespace and deployment parameters are required", http.StatusBadRequest)
		return
	}

	events, err := h.client.GetDeploymentEvents(namespace, deploymentName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, events)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resourceType := r.URL.Query().Get("type")
	namespace := r.URL.Query().Get("namespace")

	if resourceType == "" {
		http.Error(w, "type parameter is required", http.StatusBadRequest)
		return
	}

	stats, err := h.client.GetResourceStats(resourceType, namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, stats)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, map[string]string{
		"status": "healthy",
	})
}

func (h *Handler) handleOperationLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	envName := r.URL.Query().Get("environment")
	if envName == "" {
		http.Error(w, "environment parameter is required", http.StatusBadRequest)
		return
	}

	logStore := terraform.GetLogStore()
	opLog := logStore.GetOperation(envName)

	if opLog == nil {
		respondJSON(w, map[string]interface{}{
			"environment": envName,
			"status":      "not_found",
			"logs":        []string{},
		})
		return
	}

	lines := opLog.GetLines()
	logStrings := make([]string, len(lines))
	for i, line := range lines {
		logStrings[i] = line.Content
	}

	respondJSON(w, map[string]interface{}{
		"environment": opLog.EnvironmentName,
		"operation":   opLog.Operation,
		"status":      opLog.GetStatus(),
		"start_time":  opLog.StartTime,
		"end_time":    opLog.EndTime,
		"error":       opLog.Error,
		"logs":        logStrings,
	})
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
