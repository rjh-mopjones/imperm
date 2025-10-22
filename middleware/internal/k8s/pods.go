package k8s

import (
	"fmt"
	"imperm-middleware/pkg/models"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPods lists all pods in a namespace (or all namespaces if namespace is empty)
func (c *K8sClient) ListPods(namespace string) ([]models.Pod, error) {
	listOptions := metav1.ListOptions{}

	podList, err := c.clientset.CoreV1().Pods(namespace).List(c.ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var pods []models.Pod

	for _, pod := range podList.Items {
		// Calculate ready status (e.g., "1/2")
		totalContainers := len(pod.Status.ContainerStatuses)
		readyContainers := 0
		restarts := int32(0)

		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				readyContainers++
			}
			restarts += status.RestartCount
		}

		readyStatus := fmt.Sprintf("%d/%d", readyContainers, totalContainers)

		// Get resource usage (these would be from metrics-server in production)
		cpu := "0m"
		memory := "0Mi"

		// Try to get resource requests as a proxy
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				if cpuReq := container.Resources.Requests.Cpu(); cpuReq != nil {
					cpu = cpuReq.String()
				}
				if memReq := container.Resources.Requests.Memory(); memReq != nil {
					memory = memReq.String()
				}
			}
		}

		p := models.Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    string(pod.Status.Phase),
			Ready:     readyStatus,
			Restarts:  int(restarts),
			Age:       pod.CreationTimestamp.Time,
			CPU:       cpu,
			Memory:    memory,
		}

		pods = append(pods, p)
	}

	return pods, nil
}

// GetPodLogs retrieves logs for a specific pod
func (c *K8sClient) GetPodLogs(namespace, podName string) (string, error) {
	// Get the pod to find a container name
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(c.ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod: %w", err)
	}

	if len(pod.Spec.Containers) == 0 {
		return "", fmt.Errorf("pod has no containers")
	}

	// Use the first container
	containerName := pod.Spec.Containers[0].Name

	// Get logs
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: int64Ptr(100), // Last 100 lines
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	logs, err := req.DoRaw(c.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return string(logs), nil
}

// GetPodEvents retrieves events for a specific pod
func (c *K8sClient) GetPodEvents(namespace, podName string) ([]models.Event, error) {
	// Get events related to the pod
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.namespace=%s", podName, namespace)
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

// int64Ptr is a helper to get pointer to int64
func int64Ptr(i int64) *int64 {
	return &i
}
