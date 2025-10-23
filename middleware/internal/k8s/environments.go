package k8s

import (
	"fmt"
	"imperm-middleware/pkg/models"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListEnvironments lists all environments (namespaces with their resources)
func (c *K8sClient) ListEnvironments() ([]models.Environment, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var environments []models.Environment

	for _, ns := range namespaces.Items {
		// Skip system namespaces
		if isSystemNamespace(ns.Name) {
			continue
		}

		// Get pods for this namespace
		pods, err := c.ListPods(ns.Name)
		if err != nil {
			// Log error but continue
			pods = []models.Pod{}
		}

		// Get deployments for this namespace
		deployments, err := c.ListDeployments(ns.Name)
		if err != nil {
			// Log error but continue
			deployments = []models.Deployment{}
		}

		env := models.Environment{
			Name:        ns.Name,
			Namespace:   ns.Name,
			Status:      string(ns.Status.Phase),
			Age:         ns.CreationTimestamp.Time,
			Pods:        pods,
			Deployments: deployments,
		}

		environments = append(environments, env)
	}

	return environments, nil
}

// CreateEnvironment creates a new environment (namespace + optional starter resources)
func (c *K8sClient) CreateEnvironment(name string, options *models.DeploymentOptions) error {
	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"managed-by": "imperm",
			},
			Annotations: map[string]string{
				"created-at": time.Now().Format(time.RFC3339),
			},
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(c.ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// If options provided, create resources
	if options != nil && options.HasLoggers() {
		if err := c.createSampleDeployment(name); err != nil {
			// Namespace was created, so don't fail completely
			// Just log the error (in production, you'd want proper logging)
			fmt.Printf("Warning: failed to create sample deployment: %v\n", err)
		}
	}

	return nil
}

// DestroyEnvironment deletes an environment (namespace and all its resources)
func (c *K8sClient) DestroyEnvironment(name string) error {
	err := c.clientset.CoreV1().Namespaces().Delete(c.ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// GetEnvironmentHistory returns history of environment operations
// TODO: Implement persistent storage for history
func (c *K8sClient) GetEnvironmentHistory() ([]models.EnvironmentHistory, error) {
	// For now, return empty - you'll want to implement storage later
	return []models.EnvironmentHistory{}, nil
}

// createSampleDeployment creates a simple nginx deployment as a starter
func (c *K8sClient) createSampleDeployment(namespace string) error {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-app",
			Namespace: namespace,
			Labels: map[string]string{
				"app":        "sample",
				"created-by": "imperm",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "sample",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "sample",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := c.clientset.AppsV1().Deployments(namespace).Create(c.ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	return nil
}

// GetResourceStats retrieves statistics for a resource type
func (c *K8sClient) GetResourceStats(resourceType, namespace string) (*models.ResourceStats, error) {
	stats := &models.ResourceStats{}

	switch resourceType {
	case "environments":
		envs, err := c.ListEnvironments()
		if err != nil {
			return nil, err
		}
		stats.TotalEnvironments = len(envs)
		for _, env := range envs {
			stats.TotalPods += len(env.Pods)
			stats.TotalDeployments += len(env.Deployments)
		}
		stats.TotalCount = stats.TotalEnvironments

	case "pods":
		pods, err := c.ListPods(namespace)
		if err != nil {
			return nil, err
		}
		stats.TotalCount = len(pods)
		for _, pod := range pods {
			switch pod.Status {
			case "Running":
				stats.RunningPods++
			case "Pending":
				stats.PendingPods++
			case "Failed":
				stats.FailedPods++
			}
		}

	case "deployments":
		deployments, err := c.ListDeployments(namespace)
		if err != nil {
			return nil, err
		}
		stats.TotalCount = len(deployments)
		for _, dep := range deployments {
			stats.TotalReplicas += dep.UpToDate
			stats.AvailableReplicas += dep.Available
		}
	}

	return stats, nil
}

// isSystemNamespace checks if a namespace is a system namespace to skip
func isSystemNamespace(name string) bool {
	systemNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"default", // Optional: uncomment if you want to hide default namespace
	}

	for _, sys := range systemNamespaces {
		if name == sys {
			return true
		}
	}

	return false
}
