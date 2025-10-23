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

// Event represents a Kubernetes event
type Event struct {
	Type      string    // Normal, Warning
	Reason    string    // e.g., "Pulled", "Created", "Started"
	Message   string    // Full event message
	Timestamp time.Time // When the event occurred
	Count     int       // Number of times this event occurred
}

// ResourceStats represents statistics for a resource
type ResourceStats struct {
	// Common stats
	TotalCount int

	// Pod-specific stats
	RunningPods int
	PendingPods int
	FailedPods  int

	// Deployment-specific stats
	TotalReplicas     int
	AvailableReplicas int

	// Environment-specific stats
	TotalEnvironments int
	TotalPods         int
	TotalDeployments  int
}

// OperationLogs represents logs from a Terraform operation
type OperationLogs struct {
	Environment string     `json:"environment"`
	Operation   string     `json:"operation"`
	Status      string     `json:"status"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Error       string     `json:"error"`
	Logs        []string   `json:"logs"`
}
