package observe

import (
	"fmt"
	"imperm-ui/pkg/models"
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

	_, ok := resource.(models.Pod)
	if !ok {
		return "Logs are only available for pods"
	}

	// Return the loaded logs
	if t.currentLogs == "" {
		return "Loading logs..."
	}

	return t.currentLogs
}

func (t *Tab) renderEventsView() string {
	resource := t.getSelectedResource()
	if resource == nil {
		return "Select a resource to view events"
	}

	if t.currentEvents == nil {
		return "Loading events..."
	}

	if len(t.currentEvents) == 0 {
		return "No events found"
	}

	// Color codes
	green := "\033[32m"
	yellow := "\033[33m"
	gray := "\033[38;5;245m"
	reset := "\033[0m"

	var events strings.Builder

	// Render actual events using ANSI codes directly to avoid lipgloss padding issues
	for _, event := range t.currentEvents {
		color := green
		if event.Type == "Warning" {
			color = yellow
		}

		// Calculate time ago
		timeAgo := formatAge(event.Timestamp)

		// Format: "● Type  TimeAgo  Message"
		events.WriteString(fmt.Sprintf("%s● %-8s%s%s%s  %s%s\n",
			color, event.Type, reset, gray, timeAgo, reset, event.Message))
	}

	return events.String()
}

func (t *Tab) renderStatsView() string {
	if t.currentStats == nil {
		return "Loading stats..."
	}

	statLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	statValueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	var stats strings.Builder

	switch t.currentResource {
	case ResourceEnvironments:
		stats.WriteString(statLabelStyle.Render("Total Environments: ") + statValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalEnvironments)) + "\n\n")
		stats.WriteString(statLabelStyle.Render("Total Pods: ") + statValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalPods)) + "\n")
		stats.WriteString(statLabelStyle.Render("Total Deployments: ") + statValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalDeployments)) + "\n")

	case ResourcePods:
		stats.WriteString(statLabelStyle.Render("Total Pods: ") + statValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalCount)) + "\n\n")
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Render("● Running: ") + fmt.Sprintf("%d\n", t.currentStats.RunningPods))
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Render("● Pending: ") + fmt.Sprintf("%d\n", t.currentStats.PendingPods))
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Render("● Failed: ") + fmt.Sprintf("%d\n", t.currentStats.FailedPods))

	case ResourceDeployments:
		stats.WriteString(statLabelStyle.Render("Total Deployments: ") + statValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalCount)) + "\n\n")
		stats.WriteString(statLabelStyle.Render("Total Replicas: ") + statValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalReplicas)) + "\n")
		stats.WriteString(statLabelStyle.Render("Available: ") + statValueStyle.Render(fmt.Sprintf("%d", t.currentStats.AvailableReplicas)) + "\n")
	}

	return stats.String()
}
