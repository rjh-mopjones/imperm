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

		// Get actual resource usage from metrics-server
		cpu := "N/A"
		memory := "N/A"

		// Try to get metrics from metrics-server if available
		if c.metricsClient != nil {
			podMetrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(c.ctx, pod.Name, metav1.GetOptions{})
			if err == nil && len(podMetrics.Containers) > 0 {
				// Sum up all container metrics
				var totalCPU, totalMemory int64
				for _, container := range podMetrics.Containers {
					totalCPU += container.Usage.Cpu().MilliValue()
					totalMemory += container.Usage.Memory().Value()
				}
				cpu = fmt.Sprintf("%dm", totalCPU)
				memory = fmt.Sprintf("%dMi", totalMemory/(1024*1024))
			} else {
				// Fall back to resource requests if metrics call failed
				for _, container := range pod.Spec.Containers {
					if container.Resources.Requests != nil {
						if cpuReq := container.Resources.Requests.Cpu(); cpuReq != nil {
							cpu = cpuReq.String()
						}
						if memReq := container.Resources.Requests.Memory(); memReq != nil {
							memory = memReq.String()
						}
						break // Use first container's requests
					}
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

// DeletePod deletes a pod in the specified namespace
func (c *K8sClient) DeletePod(namespace, podName string) error {
	deletePolicy := metav1.DeletePropagationForeground
	return c.clientset.CoreV1().Pods(namespace).Delete(c.ctx, podName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// GetPodMetrics gets resource usage metrics for all pods in a namespace
func (c *K8sClient) GetPodMetrics(namespace string) ([]models.PodMetrics, error) {
	var podMetrics []models.PodMetrics

	if c.metricsClient == nil {
		return podMetrics, fmt.Errorf("metrics client not available")
	}

	// Get pod list to get resource limits
	podList, err := c.clientset.CoreV1().Pods(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Get metrics from metrics-server
	metricsList, err := c.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	// Create a map of pod limits for easy lookup
	podLimits := make(map[string]struct {
		cpuLimit    int64
		memoryLimit int64
	})

	for _, pod := range podList.Items {
		var totalCPULimit, totalMemoryLimit int64
		for _, container := range pod.Spec.Containers {
			if limit := container.Resources.Limits; limit != nil {
				totalCPULimit += limit.Cpu().MilliValue()
				totalMemoryLimit += limit.Memory().Value()
			}
		}
		podLimits[pod.Name] = struct {
			cpuLimit    int64
			memoryLimit int64
		}{cpuLimit: totalCPULimit, memoryLimit: totalMemoryLimit}
	}

	// Build metrics response
	for _, podMetric := range metricsList.Items {
		var totalCPU, totalMemory int64
		for _, container := range podMetric.Containers {
			totalCPU += container.Usage.Cpu().MilliValue()
			totalMemory += container.Usage.Memory().Value()
		}

		limits, ok := podLimits[podMetric.Name]
		if !ok {
			// If we can't find limits, use 0
			limits.cpuLimit = 0
			limits.memoryLimit = 0
		}

		cpuUsedPercentage := 0.0
		if limits.cpuLimit > 0 {
			cpuUsedPercentage = float64(totalCPU) / float64(limits.cpuLimit) * 100
		}

		memoryUsedPercentage := 0.0
		if limits.memoryLimit > 0 {
			memoryUsedPercentage = float64(totalMemory) / float64(limits.memoryLimit) * 100
		}

		podMetrics = append(podMetrics, models.PodMetrics{
			Name:                 podMetric.Name,
			CPULimit:             fmt.Sprintf("%dm", limits.cpuLimit),
			CPUUsed:              fmt.Sprintf("%dm", totalCPU),
			CPUUsedPercentage:    cpuUsedPercentage,
			MemoryLimit:          fmt.Sprintf("%dMi", limits.memoryLimit/(1024*1024)),
			MemoryUsed:           fmt.Sprintf("%dMi", totalMemory/(1024*1024)),
			MemoryUsedPercentage: memoryUsedPercentage,
		})
	}

	return podMetrics, nil
}

// int64Ptr is a helper to get pointer to int64
func int64Ptr(i int64) *int64 {
	return &i
}
