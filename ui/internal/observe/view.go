package observe

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"imperm-ui/internal/ui"
)

func (t *Tab) View() string {
	if t.width == 0 {
		return "Loading..."
	}

	// Show error if present
	if t.lastError != nil {
		return ui.ErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'r' to retry", t.lastError))
	}

	// Title color changes based on focus
	titleColor := ui.ColorTextDim
	if t.panelFocus == FocusTable {
		titleColor = ui.ColorHighlight
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Padding(0, 0, 1, 0)

	tableHeaderStyle := ui.TableHeaderStyle

	rowStyle := ui.TableRowStyle

	selectedRowStyle := ui.TableRowSelectedStyle

	// Get resource name for table title
	var resourceName string
	switch t.currentResource {
	case ResourceEnvironments:
		resourceName = "Environments"
	case ResourcePods:
		resourceName = "Pods"
	case ResourceDeployments:
		resourceName = "Deployments"
	}

	// Build table
	var content strings.Builder
	content.WriteString(titleStyle.Render(resourceName))
	content.WriteString("\n")

	// Always reserve space for status message (so layout doesn't shift)
	content.WriteString(ui.RenderStatusMessage(t.statusMessage, t.statusType))

	switch t.currentResource {
	case ResourceEnvironments:
		content.WriteString(t.renderEnvironmentsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	case ResourcePods:
		content.WriteString(t.renderPodsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	case ResourceDeployments:
		content.WriteString(t.renderDeploymentsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	}

	// Right panel
	layout := ui.CalculateSplitLayout(t.width, t.height)
	rightPanelHeight := t.height - 10
	rightPanelContent := t.renderRightPanel(rightPanelHeight)

	// Style for focused/unfocused panels - customize border color
	rightBorderColor := ui.ColorBorder
	if t.panelFocus == FocusRightPanel {
		rightBorderColor = ui.ColorHighlight
	}

	// Create right panel with custom border
	rightPanelStyled := lipgloss.NewStyle().
		Width(layout.RightWidth - 4).
		Height(t.height - 10).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Render(rightPanelContent)

	// Create table panel
	tablePanel := lipgloss.NewStyle().
		Width(layout.LeftWidth - 2).
		Height(layout.PanelHeight).
		Padding(1).
		Render(content.String())

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, rightPanelStyled)

	// Help text
	var helpText string
	if t.panelFocus == FocusTable {
		helpText = "[→/l] Right Panel  [e/p/d] Views  [Enter] Drill-down  [↑↓/jk] Navigate  [x] Delete  [r] Refresh  [q] Quit"
	} else {
		helpText = "[←/h] Back  [→←/hl] Cycle Views  [↑↓/jk] Scroll  [1] Details  [2] Logs  [3] Events  [4] Stats  [q] Quit"
	}
	help := ui.HelpStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainContent,
		help,
	)
}
