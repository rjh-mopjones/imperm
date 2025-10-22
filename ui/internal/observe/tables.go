package observe

import (
	"fmt"
	"imperm-ui/pkg/models"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (t *Tab) renderEnvironmentsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	var table strings.Builder

	// Header
	nameCol := headerStyle.Width(30).Render("NAME")
	namespaceCol := headerStyle.Width(20).Render("NAMESPACE")
	statusCol := headerStyle.Width(15).Render("STATUS")
	ageCol := headerStyle.Width(15).Render("AGE")
	podsCol := headerStyle.Width(10).Render("PODS")

	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, statusCol, ageCol, podsCol))
	table.WriteString("\n")

	// Rows
	for i, env := range t.environments {
		style := rowStyle
		if i == t.selectedIndex {
			style = selectedStyle
		}

		age := formatAge(env.Age)
		podCount := fmt.Sprintf("%d", len(env.Pods))

		nameCol := style.Width(30).Render(truncate(env.Name, 28))
		namespaceCol := style.Width(20).Render(truncate(env.Namespace, 18))
		statusCol := style.Width(15).Render(env.Status)
		ageCol := style.Width(15).Render(age)
		podsCol := style.Width(10).Render(podCount)

		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, statusCol, ageCol, podsCol))
		table.WriteString("\n")
	}

	if len(t.environments) == 0 {
		table.WriteString(rowStyle.Render("No environments found"))
	}

	return table.String()
}

func (t *Tab) renderPodsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	var table strings.Builder

	// Header
	nameCol := headerStyle.Width(30).Render("NAME")
	namespaceCol := headerStyle.Width(15).Render("NAMESPACE")
	readyCol := headerStyle.Width(8).Render("READY")
	statusCol := headerStyle.Width(12).Render("STATUS")
	restartsCol := headerStyle.Width(10).Render("RESTARTS")
	cpuCol := headerStyle.Width(10).Render("CPU")
	memoryCol := headerStyle.Width(10).Render("MEMORY")
	ageCol := headerStyle.Width(10).Render("AGE")

	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, statusCol, restartsCol, cpuCol, memoryCol, ageCol))
	table.WriteString("\n")

	// Determine which pods to display
	var pods []models.Pod
	if t.selectedEnvironment != nil {
		pods = t.selectedEnvironment.Pods
	} else {
		pods = t.pods
	}

	// Rows
	for i, pod := range pods {
		style := rowStyle
		if i == t.selectedIndex {
			style = selectedStyle
		}

		age := formatAge(pod.Age)

		nameCol := style.Width(30).Render(truncate(pod.Name, 28))
		namespaceCol := style.Width(15).Render(truncate(pod.Namespace, 13))
		readyCol := style.Width(8).Render(pod.Ready)
		statusCol := style.Width(12).Render(pod.Status)
		restartsCol := style.Width(10).Render(fmt.Sprintf("%d", pod.Restarts))
		cpuCol := style.Width(10).Render(pod.CPU)
		memoryCol := style.Width(10).Render(pod.Memory)
		ageCol := style.Width(10).Render(age)

		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, statusCol, restartsCol, cpuCol, memoryCol, ageCol))
		table.WriteString("\n")
	}

	if len(pods) == 0 {
		table.WriteString(rowStyle.Render("No pods found"))
	}

	return table.String()
}

func (t *Tab) renderDeploymentsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	var table strings.Builder

	// Header
	nameCol := headerStyle.Width(35).Render("NAME")
	namespaceCol := headerStyle.Width(20).Render("NAMESPACE")
	readyCol := headerStyle.Width(10).Render("READY")
	upToDateCol := headerStyle.Width(12).Render("UP-TO-DATE")
	availableCol := headerStyle.Width(12).Render("AVAILABLE")
	ageCol := headerStyle.Width(15).Render("AGE")

	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, upToDateCol, availableCol, ageCol))
	table.WriteString("\n")

	// Determine which deployments to display
	var deployments []models.Deployment
	if t.selectedEnvironment != nil {
		deployments = t.selectedEnvironment.Deployments
	} else {
		deployments = t.deployments
	}

	// Rows
	for i, deployment := range deployments {
		style := rowStyle
		if i == t.selectedIndex {
			style = selectedStyle
		}

		age := formatAge(deployment.Age)

		nameCol := style.Width(35).Render(truncate(deployment.Name, 33))
		namespaceCol := style.Width(20).Render(truncate(deployment.Namespace, 18))
		readyCol := style.Width(10).Render(deployment.Ready)
		upToDateCol := style.Width(12).Render(fmt.Sprintf("%d", deployment.UpToDate))
		availableCol := style.Width(12).Render(fmt.Sprintf("%d", deployment.Available))
		ageCol := style.Width(15).Render(age)

		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, upToDateCol, availableCol, ageCol))
		table.WriteString("\n")
	}

	if len(deployments) == 0 {
		table.WriteString(rowStyle.Render("No deployments found"))
	}

	return table.String()
}

func formatAge(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
