package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"imperm-middleware/pkg/models"
)

// TerraformClient implements the client.Client interface using Terraform
type TerraformClient struct {
	baseDir    string // Base directory for terraform environments
	modulePath string // Path to the k8s-namespace module
	kubeconfig string // Path to kubeconfig file
}

// NewClient creates a new Terraform client
func NewClient(baseDir, modulePath, kubeconfig string) (*TerraformClient, error) {
	// Validate terraform is installed
	executor := NewExecutor(baseDir)
	if err := executor.Validate(); err != nil {
		return nil, fmt.Errorf("terraform validation failed: %w", err)
	}

	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &TerraformClient{
		baseDir:    baseDir,
		modulePath: modulePath,
		kubeconfig: kubeconfig,
	}, nil
}

// ListEnvironments lists all environments managed by Terraform
func (c *TerraformClient) ListEnvironments() ([]models.Environment, error) {
	entries, err := os.ReadDir(c.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var environments []models.Environment

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		envName := entry.Name()
		envDir := filepath.Join(c.baseDir, envName)

		// Check if this is a terraform environment (has .terraform directory)
		if _, err := os.Stat(filepath.Join(envDir, ".terraform")); os.IsNotExist(err) {
			continue
		}

		// Get environment info from state
		env, err := c.getEnvironmentInfo(envName)
		if err != nil {
			// Environment exists but state might be corrupted, add it anyway
			environments = append(environments, models.Environment{
				Name:      envName,
				Namespace: envName,
				Status:    "Unknown",
				Age:       time.Now(), // Default to now
			})
			continue
		}

		environments = append(environments, env)
	}

	return environments, nil
}

// CreateEnvironment creates a new environment using Terraform
func (c *TerraformClient) CreateEnvironment(name string, withOptions bool) error {
	// Create operation log
	logStore := GetLogStore()
	opLog := logStore.CreateOperation(name, "create")

	// Create working directory
	opLog.AddLine("Creating working directory...")
	envDir, err := CreateWorkingDir(c.baseDir, name)
	if err != nil {
		opLog.SetFailed(err)
		return err
	}

	// Generate Terraform configuration
	opLog.AddLine("Generating Terraform configuration...")
	if err := c.generateConfig(envDir, name, withOptions); err != nil {
		opLog.SetFailed(err)
		return err
	}

	// Initialize Terraform
	executor := NewExecutor(envDir)
	executor.SetLogCallback(func(line string) {
		opLog.AddLine(line)
	})

	if err := executor.Init(); err != nil {
		opLog.SetFailed(err)
		return err
	}

	// Apply Terraform configuration
	if err := executor.Apply(); err != nil {
		opLog.SetFailed(err)
		return err
	}

	opLog.SetCompleted()
	opLog.AddLine("Environment created successfully!")
	return nil
}

// DestroyEnvironment destroys an environment using Terraform
func (c *TerraformClient) DestroyEnvironment(name string) error {
	// Create operation log
	logStore := GetLogStore()
	opLog := logStore.CreateOperation(name, "destroy")

	envDir := filepath.Join(c.baseDir, name)

	// Check if environment exists
	opLog.AddLine("Checking if environment exists...")
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		err := fmt.Errorf("environment %s does not exist", name)
		opLog.SetFailed(err)
		return err
	}

	// Destroy Terraform resources
	executor := NewExecutor(envDir)
	executor.SetLogCallback(func(line string) {
		opLog.AddLine(line)
	})

	if err := executor.Destroy(); err != nil {
		opLog.SetFailed(err)
		return err
	}

	// Remove working directory
	opLog.AddLine("Cleaning up working directory...")
	if err := RemoveWorkingDir(c.baseDir, name); err != nil {
		opLog.SetFailed(err)
		return err
	}

	opLog.SetCompleted()
	opLog.AddLine("Environment destroyed successfully!")

	// Clean up log after a short delay
	go func() {
		time.Sleep(30 * time.Second)
		logStore.DeleteOperation(name)
	}()

	return nil
}

// ListPods lists pods in a namespace (not fully implemented for Terraform)
func (c *TerraformClient) ListPods(namespace string) ([]models.Pod, error) {
	// For Terraform-managed environments, we'd need to query K8s directly
	// or use terraform outputs. For now, return empty.
	return []models.Pod{}, nil
}

// GetPodLogs gets logs for a pod (not implemented for Terraform)
func (c *TerraformClient) GetPodLogs(namespace, podName string) (string, error) {
	return "", fmt.Errorf("pod logs not supported in terraform mode")
}

// GetPodEvents gets events for a pod (not implemented for Terraform)
func (c *TerraformClient) GetPodEvents(namespace, podName string) ([]models.Event, error) {
	return []models.Event{}, nil
}

// DeletePod deletes a pod (not implemented for Terraform)
func (c *TerraformClient) DeletePod(namespace, podName string) error {
	return fmt.Errorf("delete pod not supported in terraform mode")
}

// ListDeployments lists deployments in a namespace (not implemented for Terraform)
func (c *TerraformClient) ListDeployments(namespace string) ([]models.Deployment, error) {
	return []models.Deployment{}, nil
}

// GetDeploymentEvents gets events for a deployment (not implemented for Terraform)
func (c *TerraformClient) GetDeploymentEvents(namespace, deploymentName string) ([]models.Event, error) {
	return []models.Event{}, nil
}

// DeleteDeployment deletes a deployment (not implemented for Terraform)
func (c *TerraformClient) DeleteDeployment(namespace, deploymentName string) error {
	return fmt.Errorf("delete deployment not supported in terraform mode")
}

// GetResourceStats gets resource statistics (not fully implemented for Terraform)
func (c *TerraformClient) GetResourceStats(resourceType, namespace string) (*models.ResourceStats, error) {
	stats := &models.ResourceStats{}

	if resourceType == "environments" {
		envs, err := c.ListEnvironments()
		if err != nil {
			return nil, err
		}
		stats.TotalEnvironments = len(envs)
		stats.TotalCount = len(envs)
	}

	return stats, nil
}

// GetEnvironmentHistory gets environment history (not implemented for Terraform)
func (c *TerraformClient) GetEnvironmentHistory() ([]models.EnvironmentHistory, error) {
	return []models.EnvironmentHistory{}, nil
}

// generateConfig generates Terraform configuration files for an environment
func (c *TerraformClient) generateConfig(envDir, name string, withOptions bool) error {
	mainTf := fmt.Sprintf(`terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
  }
}

provider "kubernetes" {
  config_path = "%s"
}

module "environment" {
  source = "%s"

  namespace_name = "%s"
  with_options   = %t
}

output "namespace_name" {
  value = module.environment.namespace_name
}

output "deployment_created" {
  value = module.environment.deployment_created
}
`, c.kubeconfig, c.modulePath, name, withOptions)

	mainTfPath := filepath.Join(envDir, "main.tf")
	if err := os.WriteFile(mainTfPath, []byte(mainTf), 0644); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	return nil
}

// getEnvironmentInfo retrieves environment information from Terraform state
func (c *TerraformClient) getEnvironmentInfo(name string) (models.Environment, error) {
	envDir := filepath.Join(c.baseDir, name)
	executor := NewExecutor(envDir)

	// Get terraform state
	stateJSON, err := executor.Show()
	if err != nil {
		return models.Environment{}, err
	}

	// Parse state to extract info
	var state TerraformState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return models.Environment{}, fmt.Errorf("failed to parse terraform state: %w", err)
	}

	env := models.Environment{
		Name:      name,
		Namespace: name,
		Status:    "Active",
		Age:       time.Now(), // You could extract this from state metadata
	}

	return env, nil
}

// TerraformState represents the structure of terraform state (simplified)
type TerraformState struct {
	Values struct {
		RootModule struct {
			Resources []struct {
				Type   string `json:"type"`
				Name   string `json:"name"`
				Values map[string]interface{} `json:"values"`
			} `json:"resources"`
		} `json:"root_module"`
	} `json:"values"`
}
