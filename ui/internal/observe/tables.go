package observe

import (
	"fmt"
	"imperm-ui/pkg/models"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TableColumn defines a column in a generic table
type TableColumn struct {
	Header string
	Width  int
	Value  func(item interface{}) string
}

// renderGenericTable renders a table with the given columns and items
func renderGenericTable(
	items []interface{},
	columns []TableColumn,
	selectedIndex int,
	isLoading bool,
	emptyResourceName string,
	headerStyle, rowStyle, selectedStyle lipgloss.Style,
) string {
	var table strings.Builder

	// Render header
	var headerCols []string
	for _, col := range columns {
		headerCols = append(headerCols, headerStyle.Width(col.Width).Render(col.Header))
	}
	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCols...))
	table.WriteString("\n")

	// Render rows
	for i, item := range items {
		style := rowStyle
		if i == selectedIndex {
			style = selectedStyle
		}

		var rowCols []string
		for _, col := range columns {
			value := col.Value(item)
			rowCols = append(rowCols, style.Width(col.Width).Render(value))
		}

		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowCols...))
		table.WriteString("\n")
	}

	// Empty state
	if len(items) == 0 {
		if isLoading {
			table.WriteString(rowStyle.Render(fmt.Sprintf("Loading %s...", emptyResourceName)))
		} else {
			table.WriteString(rowStyle.Render(fmt.Sprintf("No %s found", emptyResourceName)))
		}
	}

	return table.String()
}

func (t *Tab) renderEnvironmentsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	// Convert environments to []interface{}
	items := make([]interface{}, len(t.environments))
	for i, env := range t.environments {
		items[i] = env
	}

	// Define columns
	columns := []TableColumn{
		{
			Header: "NAME",
			Width:  30,
			Value: func(item interface{}) string {
				env := item.(models.Environment)
				return truncate(env.Name, 28)
			},
		},
		{
			Header: "NAMESPACE",
			Width:  20,
			Value: func(item interface{}) string {
				env := item.(models.Environment)
				return truncate(env.Namespace, 18)
			},
		},
		{
			Header: "STATUS",
			Width:  15,
			Value: func(item interface{}) string {
				env := item.(models.Environment)
				return env.Status
			},
		},
		{
			Header: "AGE",
			Width:  15,
			Value: func(item interface{}) string {
				env := item.(models.Environment)
				return formatAge(env.Age)
			},
		},
		{
			Header: "PODS",
			Width:  10,
			Value: func(item interface{}) string {
				env := item.(models.Environment)
				return fmt.Sprintf("%d", len(env.Pods))
			},
		},
	}

	return renderGenericTable(items, columns, t.selectedIndex, t.isLoading, "environments", headerStyle, rowStyle, selectedStyle)
}

func (t *Tab) renderPodsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	// Determine which pods to display
	var pods []models.Pod
	if t.selectedEnvironment != nil {
		pods = t.selectedEnvironment.Pods
	} else {
		pods = t.pods
	}

	// Convert pods to []interface{}
	items := make([]interface{}, len(pods))
	for i, pod := range pods {
		items[i] = pod
	}

	// Define columns
	columns := []TableColumn{
		{Header: "NAME", Width: 30, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return truncate(pod.Name, 28)
		}},
		{Header: "NAMESPACE", Width: 15, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return truncate(pod.Namespace, 13)
		}},
		{Header: "READY", Width: 8, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return pod.Ready
		}},
		{Header: "STATUS", Width: 12, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return pod.Status
		}},
		{Header: "RESTARTS", Width: 10, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return fmt.Sprintf("%d", pod.Restarts)
		}},
		{Header: "CPU", Width: 10, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return pod.CPU
		}},
		{Header: "MEMORY", Width: 10, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return pod.Memory
		}},
		{Header: "AGE", Width: 10, Value: func(item interface{}) string {
			pod := item.(models.Pod)
			return formatAge(pod.Age)
		}},
	}

	return renderGenericTable(items, columns, t.selectedIndex, t.isLoading, "pods", headerStyle, rowStyle, selectedStyle)
}

func (t *Tab) renderDeploymentsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	// Determine which deployments to display
	var deployments []models.Deployment
	if t.selectedEnvironment != nil {
		deployments = t.selectedEnvironment.Deployments
	} else {
		deployments = t.deployments
	}

	// Convert deployments to []interface{}
	items := make([]interface{}, len(deployments))
	for i, deployment := range deployments {
		items[i] = deployment
	}

	// Define columns
	columns := []TableColumn{
		{Header: "NAME", Width: 35, Value: func(item interface{}) string {
			deployment := item.(models.Deployment)
			return truncate(deployment.Name, 33)
		}},
		{Header: "NAMESPACE", Width: 20, Value: func(item interface{}) string {
			deployment := item.(models.Deployment)
			return truncate(deployment.Namespace, 18)
		}},
		{Header: "READY", Width: 10, Value: func(item interface{}) string {
			deployment := item.(models.Deployment)
			return deployment.Ready
		}},
		{Header: "UP-TO-DATE", Width: 12, Value: func(item interface{}) string {
			deployment := item.(models.Deployment)
			return fmt.Sprintf("%d", deployment.UpToDate)
		}},
		{Header: "AVAILABLE", Width: 12, Value: func(item interface{}) string {
			deployment := item.(models.Deployment)
			return fmt.Sprintf("%d", deployment.Available)
		}},
		{Header: "AGE", Width: 15, Value: func(item interface{}) string {
			deployment := item.(models.Deployment)
			return formatAge(deployment.Age)
		}},
	}

	return renderGenericTable(items, columns, t.selectedIndex, t.isLoading, "deployments", headerStyle, rowStyle, selectedStyle)
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
