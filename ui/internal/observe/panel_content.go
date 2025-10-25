package observe

import (
	"fmt"
	"imperm-ui/internal/ui"
	"imperm-ui/pkg/models"
	"strings"
)

func (t *Tab) renderDetailsView() string {
	resource := t.getSelectedResource()
	if resource == nil {
		return "Select a resource to view details"
	}

	var details strings.Builder

	switch r := resource.(type) {
	case models.Environment:
		details.WriteString(ui.LabelStyle.Render("Name:") + " " + ui.ValueStyle.Render(r.Name) + "\n")
		details.WriteString(ui.LabelStyle.Render("Namespace:") + " " + ui.ValueStyle.Render(r.Namespace) + "\n")
		details.WriteString(ui.LabelStyle.Render("Status:") + " " + ui.ValueStyle.Render(r.Status) + "\n")
		details.WriteString(ui.LabelStyle.Render("Age:") + " " + ui.ValueStyle.Render(formatAge(r.Age)) + "\n")
		details.WriteString(ui.LabelStyle.Render("Pods:") + " " + ui.ValueStyle.Render(fmt.Sprintf("%d", len(r.Pods))) + "\n")
		details.WriteString(ui.LabelStyle.Render("Deployments:") + " " + ui.ValueStyle.Render(fmt.Sprintf("%d", len(r.Deployments))) + "\n")

	case models.Pod:
		details.WriteString(ui.LabelStyle.Render("Name:") + " " + ui.ValueStyle.Render(r.Name) + "\n")
		details.WriteString(ui.LabelStyle.Render("Namespace:") + " " + ui.ValueStyle.Render(r.Namespace) + "\n")
		details.WriteString(ui.LabelStyle.Render("Status:") + " " + ui.ValueStyle.Render(r.Status) + "\n")
		details.WriteString(ui.LabelStyle.Render("Ready:") + " " + ui.ValueStyle.Render(r.Ready) + "\n")
		details.WriteString(ui.LabelStyle.Render("Restarts:") + " " + ui.ValueStyle.Render(fmt.Sprintf("%d", r.Restarts)) + "\n")
		details.WriteString(ui.LabelStyle.Render("CPU:") + " " + ui.ValueStyle.Render(r.CPU) + "\n")
		details.WriteString(ui.LabelStyle.Render("Memory:") + " " + ui.ValueStyle.Render(r.Memory) + "\n")
		details.WriteString(ui.LabelStyle.Render("Age:") + " " + ui.ValueStyle.Render(formatAge(r.Age)) + "\n")

	case models.Deployment:
		details.WriteString(ui.LabelStyle.Render("Name:") + " " + ui.ValueStyle.Render(r.Name) + "\n")
		details.WriteString(ui.LabelStyle.Render("Namespace:") + " " + ui.ValueStyle.Render(r.Namespace) + "\n")
		details.WriteString(ui.LabelStyle.Render("Ready:") + " " + ui.ValueStyle.Render(r.Ready) + "\n")
		details.WriteString(ui.LabelStyle.Render("Up-to-Date:") + " " + ui.ValueStyle.Render(fmt.Sprintf("%d", r.UpToDate)) + "\n")
		details.WriteString(ui.LabelStyle.Render("Available:") + " " + ui.ValueStyle.Render(fmt.Sprintf("%d", r.Available)) + "\n")
		details.WriteString(ui.LabelStyle.Render("Age:") + " " + ui.ValueStyle.Render(formatAge(r.Age)) + "\n")
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

	var events strings.Builder

	// Render events using theme colors
	for _, event := range t.currentEvents {
		eventColor := ui.ColorSuccess
		if event.Type == "Warning" {
			eventColor = ui.ColorWarning
		}

		// Calculate time ago
		timeAgo := formatAge(event.Timestamp)

		// Format: "● Type  TimeAgo  Message"
		events.WriteString(fmt.Sprintf("\033[38;5;%sm● %-8s\033[0m\033[38;5;%sm%s\033[0m  %s\n",
			eventColor, event.Type, ui.ColorTextDim, timeAgo, event.Message))
	}

	return events.String()
}

func (t *Tab) renderStatsView() string {
	if t.currentStats == nil {
		return "Loading stats..."
	}

	var stats strings.Builder

	switch t.currentResource {
	case ResourceEnvironments:
		stats.WriteString(ui.StatLabelStyle.Render("Total Environments: ") + ui.StatValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalEnvironments)) + "\n\n")
		stats.WriteString(ui.StatLabelStyle.Render("Total Pods: ") + ui.StatValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalPods)) + "\n")
		stats.WriteString(ui.StatLabelStyle.Render("Total Deployments: ") + ui.StatValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalDeployments)) + "\n")

	case ResourcePods:
		stats.WriteString(ui.StatLabelStyle.Render("Total Pods: ") + ui.StatValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalCount)) + "\n\n")
		stats.WriteString(ui.SuccessStyle.Render("● Running: ") + fmt.Sprintf("%d\n", t.currentStats.RunningPods))
		stats.WriteString(ui.WarningStyle.Render("● Pending: ") + fmt.Sprintf("%d\n", t.currentStats.PendingPods))
		stats.WriteString(ui.ErrorStyle.Render("● Failed: ") + fmt.Sprintf("%d\n", t.currentStats.FailedPods))

	case ResourceDeployments:
		stats.WriteString(ui.StatLabelStyle.Render("Total Deployments: ") + ui.StatValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalCount)) + "\n\n")
		stats.WriteString(ui.StatLabelStyle.Render("Total Replicas: ") + ui.StatValueStyle.Render(fmt.Sprintf("%d", t.currentStats.TotalReplicas)) + "\n")
		stats.WriteString(ui.StatLabelStyle.Render("Available: ") + ui.StatValueStyle.Render(fmt.Sprintf("%d", t.currentStats.AvailableReplicas)) + "\n")
	}

	return stats.String()
}
