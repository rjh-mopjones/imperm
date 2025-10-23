package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"imperm-middleware/internal/k8s"
	"imperm-middleware/pkg/models"
)

// TerraformClient implements the client.Client interface using Terraform for provisioning
// and Kubernetes API for querying
type TerraformClient struct {
	baseDir    string // Base directory for terraform environments
	modulePath string // Path to the k8s-namespace module
	kubeconfig string // Path to kubeconfig file
	k8sClient  *k8s.K8sClient // Embedded K8s client for read operations
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

	// Create K8s client for read operations
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &TerraformClient{
		baseDir:    baseDir,
		modulePath: modulePath,
		kubeconfig: kubeconfig,
		k8sClient:  k8sClient,
	}, nil
}

// ListEnvironments lists all environments using Kubernetes API
func (c *TerraformClient) ListEnvironments() ([]models.Environment, error) {
	// Delegate to K8s client for listing
	return c.k8sClient.ListEnvironments()
}

// CreateEnvironment creates a new environment using Terraform
func (c *TerraformClient) CreateEnvironment(name string, options *models.DeploymentOptions) error {
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
	if err := c.generateConfig(envDir, name, options); err != nil {
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

// ListPods lists pods in a namespace using Kubernetes API
func (c *TerraformClient) ListPods(namespace string) ([]models.Pod, error) {
	return c.k8sClient.ListPods(namespace)
}

// GetPodLogs gets logs for a pod using Kubernetes API
func (c *TerraformClient) GetPodLogs(namespace, podName string) (string, error) {
	return c.k8sClient.GetPodLogs(namespace, podName)
}

// GetPodEvents gets events for a pod using Kubernetes API
func (c *TerraformClient) GetPodEvents(namespace, podName string) ([]models.Event, error) {
	return c.k8sClient.GetPodEvents(namespace, podName)
}

// DeletePod deletes a pod using Kubernetes API
func (c *TerraformClient) DeletePod(namespace, podName string) error {
	return c.k8sClient.DeletePod(namespace, podName)
}

// ListDeployments lists deployments in a namespace using Kubernetes API
func (c *TerraformClient) ListDeployments(namespace string) ([]models.Deployment, error) {
	return c.k8sClient.ListDeployments(namespace)
}

// GetDeploymentEvents gets events for a deployment using Kubernetes API
func (c *TerraformClient) GetDeploymentEvents(namespace, deploymentName string) ([]models.Event, error) {
	return c.k8sClient.GetDeploymentEvents(namespace, deploymentName)
}

// DeleteDeployment deletes a deployment using Kubernetes API
func (c *TerraformClient) DeleteDeployment(namespace, deploymentName string) error {
	return c.k8sClient.DeleteDeployment(namespace, deploymentName)
}

// GetResourceStats gets resource statistics using Kubernetes API
func (c *TerraformClient) GetResourceStats(resourceType, namespace string) (*models.ResourceStats, error) {
	return c.k8sClient.GetResourceStats(resourceType, namespace)
}

// GetEnvironmentHistory gets environment history using Kubernetes API
func (c *TerraformClient) GetEnvironmentHistory() ([]models.EnvironmentHistory, error) {
	return c.k8sClient.GetEnvironmentHistory()
}

// generateConfig generates Terraform configuration files for an environment
func (c *TerraformClient) generateConfig(envDir, name string, options *models.DeploymentOptions) error {
	// Default values
	constantLogger := 0
	fastLogger := 0
	errorLogger := 0
	jsonLogger := 0

	if options != nil {
		constantLogger = options.ConstantLogger
		fastLogger = options.FastLogger
		errorLogger = options.ErrorLogger
		jsonLogger = options.JsonLogger
	}

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

  namespace_name    = "%s"
  constant_logger   = %d
  fast_logger       = %d
  error_logger      = %d
  json_logger       = %d
}

output "namespace_name" {
  value = module.environment.namespace_name
}
`, c.kubeconfig, c.modulePath, name, constantLogger, fastLogger, errorLogger, jsonLogger)

	mainTfPath := filepath.Join(envDir, "main.tf")
	if err := os.WriteFile(mainTfPath, []byte(mainTf), 0644); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	return nil
}
