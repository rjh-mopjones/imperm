package models

import "time"

// Environment represents a Kubernetes environment
type Environment struct {
	Name        string
	Namespace   string
	Status      string
	Age         time.Time
	Pods        []Pod
	Deployments []Deployment
}

// Pod represents a Kubernetes pod
type Pod struct {
	Name      string
	Namespace string
	Status    string
	Ready     string
	Restarts  int
	Age       time.Time
	CPU       string // e.g., "100m", "1.5"
	Memory    string // e.g., "256Mi", "1.5Gi"
}

// Deployment represents a Kubernetes deployment
type Deployment struct {
	Name      string
	Namespace string
	Ready     string
	UpToDate  int
	Available int
	Age       time.Time
}

// EnvironmentHistory represents a historical environment launch
type EnvironmentHistory struct {
	Name        string
	LaunchedAt  time.Time
	Status      string
	WithOptions bool
}
