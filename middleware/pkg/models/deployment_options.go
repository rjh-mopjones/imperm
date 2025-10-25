package models

// DeploymentOptions contains configuration for creating environments
type DeploymentOptions struct {
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"` // Terraform variables as key-value pairs
}

// HasVariables returns true if any variables are configured
func (d *DeploymentOptions) HasVariables() bool {
	return len(d.Variables) > 0
}
