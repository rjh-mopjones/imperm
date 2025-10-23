package k8s

import (
	"fmt"
	"imperm-middleware/pkg/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDeployments lists all deployments in a namespace (or all namespaces if namespace is empty)
func (c *K8sClient) ListDeployments(namespace string) ([]models.Deployment, error) {
	listOptions := metav1.ListOptions{}

	deploymentList, err := c.clientset.AppsV1().Deployments(namespace).List(c.ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var deployments []models.Deployment

	for _, deploy := range deploymentList.Items {
		// Calculate ready replicas string (e.g., "2/3")
		ready := fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, *deploy.Spec.Replicas)

		d := models.Deployment{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			Ready:     ready,
			UpToDate:  int(deploy.Status.UpdatedReplicas),
			Available: int(deploy.Status.AvailableReplicas),
			Age:       deploy.CreationTimestamp.Time,
		}

		deployments = append(deployments, d)
	}

	return deployments, nil
}

// GetDeploymentEvents retrieves events for a specific deployment
func (c *K8sClient) GetDeploymentEvents(namespace, deploymentName string) ([]models.Event, error) {
	// Get events related to the deployment
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", deploymentName, namespace)
	eventList, err := c.clientset.CoreV1().Events(namespace).List(c.ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var events []models.Event
	for _, event := range eventList.Items {
		e := models.Event{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Timestamp: event.LastTimestamp.Time,
			Count:     int(event.Count),
		}
		events = append(events, e)
	}

	return events, nil
}

// DeleteDeployment deletes a deployment in the specified namespace
func (c *K8sClient) DeleteDeployment(namespace, deploymentName string) error {
	deletePolicy := metav1.DeletePropagationForeground
	return c.clientset.AppsV1().Deployments(namespace).Delete(c.ctx, deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}
