package models

// DeploymentOptions contains configuration for creating environments
type DeploymentOptions struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	ConstantLogger  int    `json:"constant_logger"`  // Number of replicas
	FastLogger      int    `json:"fast_logger"`      // Number of replicas
	ErrorLogger     int    `json:"error_logger"`     // Number of replicas
	JsonLogger      int    `json:"json_logger"`      // Number of replicas
}
