package observe

import (
	"fmt"
	"imperm/internal/models"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (t *Tab) renderDetailsView() string {
	resource := t.getSelectedResource()
	if resource == nil {
		return "Select a resource to view details"
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Width(15)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	var details strings.Builder

	switch r := resource.(type) {
	case models.Environment:
		details.WriteString(labelStyle.Render("Name:") + " " + valueStyle.Render(r.Name) + "\n")
		details.WriteString(labelStyle.Render("Namespace:") + " " + valueStyle.Render(r.Namespace) + "\n")
		details.WriteString(labelStyle.Render("Status:") + " " + valueStyle.Render(r.Status) + "\n")
		details.WriteString(labelStyle.Render("Age:") + " " + valueStyle.Render(formatAge(r.Age)) + "\n")
		details.WriteString(labelStyle.Render("Pods:") + " " + valueStyle.Render(fmt.Sprintf("%d", len(r.Pods))) + "\n")
		details.WriteString(labelStyle.Render("Deployments:") + " " + valueStyle.Render(fmt.Sprintf("%d", len(r.Deployments))) + "\n")

	case models.Pod:
		details.WriteString(labelStyle.Render("Name:") + " " + valueStyle.Render(r.Name) + "\n")
		details.WriteString(labelStyle.Render("Namespace:") + " " + valueStyle.Render(r.Namespace) + "\n")
		details.WriteString(labelStyle.Render("Status:") + " " + valueStyle.Render(r.Status) + "\n")
		details.WriteString(labelStyle.Render("Ready:") + " " + valueStyle.Render(r.Ready) + "\n")
		details.WriteString(labelStyle.Render("Restarts:") + " " + valueStyle.Render(fmt.Sprintf("%d", r.Restarts)) + "\n")
		details.WriteString(labelStyle.Render("CPU:") + " " + valueStyle.Render(r.CPU) + "\n")
		details.WriteString(labelStyle.Render("Memory:") + " " + valueStyle.Render(r.Memory) + "\n")
		details.WriteString(labelStyle.Render("Age:") + " " + valueStyle.Render(formatAge(r.Age)) + "\n")

	case models.Deployment:
		details.WriteString(labelStyle.Render("Name:") + " " + valueStyle.Render(r.Name) + "\n")
		details.WriteString(labelStyle.Render("Namespace:") + " " + valueStyle.Render(r.Namespace) + "\n")
		details.WriteString(labelStyle.Render("Ready:") + " " + valueStyle.Render(r.Ready) + "\n")
		details.WriteString(labelStyle.Render("Up-to-Date:") + " " + valueStyle.Render(fmt.Sprintf("%d", r.UpToDate)) + "\n")
		details.WriteString(labelStyle.Render("Available:") + " " + valueStyle.Render(fmt.Sprintf("%d", r.Available)) + "\n")
		details.WriteString(labelStyle.Render("Age:") + " " + valueStyle.Render(formatAge(r.Age)) + "\n")
	}

	return details.String()
}

func (t *Tab) renderLogsView() string {
	resource := t.getSelectedResource()
	if resource == nil {
		return "Select a pod to view logs"
	}

	pod, ok := resource.(models.Pod)
	if !ok {
		return "Logs are only available for pods"
	}

	// Mock logs for now
	logs := fmt.Sprintf("Fetching logs for %s...\n\n", pod.Name)
	logs += "[2025-10-22 20:30:15] INFO: Container started\n"
	logs += "[2025-10-22 20:30:16] INFO: Initializing application\n"
	logs += "[2025-10-22 20:30:17] INFO: Loading configuration\n"
	logs += "[2025-10-22 20:30:18] INFO: Connecting to database\n"
	logs += "[2025-10-22 20:30:20] INFO: Server listening on port 8080\n"
	logs += "[2025-10-22 20:30:25] INFO: Health check passed\n"
	logs += "[2025-10-22 20:31:00] INFO: Processing request GET /api/health\n"
	logs += "[2025-10-22 20:31:30] INFO: Processing request GET /api/status\n"

	return logs
}

func (t *Tab) renderEventsView() string {
	resource := t.getSelectedResource()
	if resource == nil {
		return "Select a resource to view events"
	}

	// Color codes
	green := "\033[32m"
	yellow := "\033[33m"
	gray := "\033[38;5;245m"
	reset := "\033[0m"

	var events strings.Builder

	// Mock events - using ANSI codes directly to avoid lipgloss padding issues
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s2m ago  Successfully pulled image%s\n", green, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s5m ago  Created container%s\n", green, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s5m ago  Started container%s\n", green, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Warning %s%s1h ago  Back-off restarting failed container%s\n", yellow, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s2h ago  Scaled deployment to 2 replicas%s\n", green, reset, gray, reset))

	return events.String()
}

func (t *Tab) renderStatsView() string {
	statLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	statValueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	var stats strings.Builder

	switch t.currentResource {
	case ResourceEnvironments:
		total := len(t.environments)
		stats.WriteString(statLabelStyle.Render("Total Environments: ") + statValueStyle.Render(fmt.Sprintf("%d", total)) + "\n\n")

		totalPods := 0
		totalDeployments := 0
		for _, env := range t.environments {
			totalPods += len(env.Pods)
			totalDeployments += len(env.Deployments)
		}
		stats.WriteString(statLabelStyle.Render("Total Pods: ") + statValueStyle.Render(fmt.Sprintf("%d", totalPods)) + "\n")
		stats.WriteString(statLabelStyle.Render("Total Deployments: ") + statValueStyle.Render(fmt.Sprintf("%d", totalDeployments)) + "\n")

	case ResourcePods:
		var pods []models.Pod
		if t.selectedEnvironment != nil {
			pods = t.selectedEnvironment.Pods
		} else {
			pods = t.pods
		}

		total := len(pods)
		running := 0
		pending := 0
		failed := 0

		for _, pod := range pods {
			switch pod.Status {
			case "Running":
				running++
			case "Pending":
				pending++
			case "Failed":
				failed++
			}
		}

		stats.WriteString(statLabelStyle.Render("Total Pods: ") + statValueStyle.Render(fmt.Sprintf("%d", total)) + "\n\n")
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Render("● Running: ") + fmt.Sprintf("%d\n", running))
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Render("● Pending: ") + fmt.Sprintf("%d\n", pending))
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Render("● Failed: ") + fmt.Sprintf("%d\n", failed))

	case ResourceDeployments:
		var deployments []models.Deployment
		if t.selectedEnvironment != nil {
			deployments = t.selectedEnvironment.Deployments
		} else {
			deployments = t.deployments
		}

		total := len(deployments)
		totalReplicas := 0
		totalAvailable := 0

		for _, dep := range deployments {
			totalReplicas += dep.UpToDate
			totalAvailable += dep.Available
		}

		stats.WriteString(statLabelStyle.Render("Total Deployments: ") + statValueStyle.Render(fmt.Sprintf("%d", total)) + "\n\n")
		stats.WriteString(statLabelStyle.Render("Total Replicas: ") + statValueStyle.Render(fmt.Sprintf("%d", totalReplicas)) + "\n")
		stats.WriteString(statLabelStyle.Render("Available: ") + statValueStyle.Render(fmt.Sprintf("%d", totalAvailable)) + "\n")
	}

	return stats.String()
}
