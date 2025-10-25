package observe

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (t *Tab) View() string {
	if t.width == 0 {
		return "Loading..."
	}

	// Show error if present
	if t.lastError != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(1, 2)
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'r' to retry", t.lastError))
	}

	// Styles
	cyanColor := lipgloss.Color("51")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cyanColor).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	// Title color changes based on focus
	titleColor := lipgloss.Color("245") // Grey when not focused
	if t.panelFocus == FocusTable {
		titleColor = cyanColor // Cyan when focused
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Padding(0, 0, 1, 0)

	tableHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("238")).
		Padding(0, 1)

	rowStyle := lipgloss.NewStyle().
		Padding(0, 1)

	// Cyan highlight for selected row
	selectedRowStyle := rowStyle.Copy().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("51")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	// Header bar with breadcrumbs
	var resourceName string
	var breadcrumb string
	switch t.currentResource {
	case ResourceEnvironments:
		resourceName = "Environments"
		breadcrumb = "Environments"
	case ResourcePods:
		resourceName = "Pods"
		if t.selectedEnvironment != nil {
			breadcrumb = fmt.Sprintf("Environments > %s > Pods", t.selectedEnvironment.Name)
		} else {
			breadcrumb = "Pods"
		}
	case ResourceDeployments:
		resourceName = "Deployments"
		if t.selectedEnvironment != nil {
			breadcrumb = fmt.Sprintf("Environments > %s > Deployments", t.selectedEnvironment.Name)
		} else {
			breadcrumb = "Deployments"
		}
	}

	autoRefreshIndicator := ""
	if t.autoRefresh {
		autoRefreshIndicator = " [AUTO]"
	}

	header := headerStyle.Render(fmt.Sprintf(" %s%s | Last update: %s ",
		breadcrumb,
		autoRefreshIndicator,
		t.lastUpdate.Format("15:04:05"),
	))

	// Build table
	var content strings.Builder
	content.WriteString(titleStyle.Render(resourceName))
	content.WriteString("\n")

	// Always reserve space for status message (so layout doesn't shift)
	if t.statusMessage != "" {
		var statusColor, bgColor string
		if t.statusType == "error" {
			statusColor = "196" // Red
			bgColor = "52"      // Dark red background
		} else {
			statusColor = "46"  // Green
			bgColor = "22"      // Dark green background
		}
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(statusColor)).
			Background(lipgloss.Color(bgColor)).
			Bold(true).
			Padding(0, 1).
			Margin(1, 0)
		content.WriteString(statusStyle.Render(t.statusMessage))
		content.WriteString("\n")
	} else {
		// Reserve space with just a newline (no visible bar)
		content.WriteString("\n\n\n")
	}

	switch t.currentResource {
	case ResourceEnvironments:
		content.WriteString(t.renderEnvironmentsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	case ResourcePods:
		content.WriteString(t.renderPodsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	case ResourceDeployments:
		content.WriteString(t.renderDeploymentsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	}

	// Right panel
	rightPanelHeight := t.height - 10
	rightPanelContent := t.renderRightPanel(rightPanelHeight)

	// Calculate widths for two-panel layout (50/50 split)
	tableWidth := t.width / 2         // 50% for table
	rightWidth := t.width - tableWidth // 50% for right panel

	// Reuse cyan color for highlights
	dimColor := lipgloss.Color("240")  // Dim gray

	// Style for focused/unfocused panels
	rightBorderColor := dimColor
	if t.panelFocus == FocusRightPanel {
		rightBorderColor = cyanColor
	}

	// Table panel (no border)
	tablePanel := lipgloss.NewStyle().
		Width(tableWidth - 2).
		Height(t.height - 8).
		Padding(1).
		Render(content.String())

	// Right panel with full box border
	rightPanel := lipgloss.NewStyle().
		Width(rightWidth - 4).
		Height(t.height - 10).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Render(rightPanelContent)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, rightPanel)

	// Help text
	var helpText string
	if t.panelFocus == FocusTable {
		helpText = "[→/l] Right Panel  [e/p/d] Views  [Enter] Drill-down  [↑↓/jk] Navigate  [x] Delete  [r] Refresh  [q] Quit"
	} else {
		helpText = "[←/h] Back  [→←/hl] Cycle Views  [↑↓/jk] Scroll  [1] Details  [2] Logs  [3] Events  [4] Stats  [q] Quit"
	}
	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		help,
	)
}
