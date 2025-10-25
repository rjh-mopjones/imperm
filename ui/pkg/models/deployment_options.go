package models

// DeploymentOptions contains configuration for creating environments
type DeploymentOptions struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"` // Terraform variables as key-value pairs
}
