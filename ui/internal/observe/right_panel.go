package observe

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (t *Tab) renderRightPanel(height int) string {
	// Calculate panel width (we'll use this for line truncation)
	panelWidth := (t.width / 2) - 8 // Half screen minus borders and padding
	if panelWidth < 40 {
		panelWidth = 40 // Minimum width to prevent issues
	}
	// View list items
	views := []struct {
		name string
		view rightPanelView
	}{
		{"Details", RightPanelDetails},
		{"Logs", RightPanelLogs},
		{"Events", RightPanelEvents},
		{"Stats", RightPanelStats},
	}

	// Styles for view list
	viewItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	selectedViewStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("86")).
		Bold(true).
		Padding(0, 1)

	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true).
		Padding(0, 1)

	// Build number row and view row separately
	var numberRow strings.Builder
	var viewRow strings.Builder

	for i, view := range views {
		// Add number
		number := fmt.Sprintf("[%d]", i+1)
		// Calculate width to match view name width
		viewNameWidth := lipgloss.Width(view.name) + 2 // +2 for padding
		numberRow.WriteString(numberStyle.Width(viewNameWidth).Align(lipgloss.Center).Render(number))
		if i < len(views)-1 {
			numberRow.WriteString(" ")
		}

		// Add view name
		style := viewItemStyle
		if view.view == t.rightPanelView {
			style = selectedViewStyle
		}
		viewRow.WriteString(style.Render(view.name))
		if i < len(views)-1 {
			viewRow.WriteString(" ")
		}
	}

	// Combine number row and view row
	var viewList strings.Builder
	viewList.WriteString(numberRow.String())
	viewList.WriteString("\n")
	viewList.WriteString(viewRow.String())

	// Use cyan for title when panel is focused
	titleColor := lipgloss.Color("245")
	if t.panelFocus == FocusRightPanel {
		titleColor = lipgloss.Color("51") // Cyan
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var content strings.Builder

	// Add view list
	content.WriteString(viewList.String())
	content.WriteString("\n")
	content.WriteString(separatorStyle.Render(strings.Repeat("─", 80)))
	content.WriteString("\n")

	// Add the current view name
	var panelName string
	switch t.rightPanelView {
	case RightPanelDetails:
		panelName = "Details"
	case RightPanelLogs:
		panelName = "Logs"
	case RightPanelEvents:
		panelName = "Events"
	case RightPanelStats:
		panelName = "Stats"
	}

	content.WriteString(titleStyle.Render(panelName))
	content.WriteString("\n")

	// Get the view content
	var viewContent string
	switch t.rightPanelView {
	case RightPanelDetails:
		viewContent = t.renderDetailsView()
	case RightPanelLogs:
		viewContent = t.renderLogsView()
	case RightPanelEvents:
		viewContent = t.renderEventsView()
	case RightPanelStats:
		viewContent = t.renderStatsView()
	}

	// Apply scrolling
	lines := strings.Split(viewContent, "\n")

	// Calculate available height for content (subtract numbers, view list, separator, title)
	availableHeight := height - 6
	if availableHeight < 5 {
		availableHeight = 5 // Minimum height to prevent issues
	}

	// Ensure scroll offset is within bounds
	maxOffset := len(lines) - availableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if t.scrollOffset > maxOffset {
		t.scrollOffset = maxOffset
	}

	// Get the visible lines
	endLine := t.scrollOffset + availableHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}

	visibleLines := lines[t.scrollOffset:endLine]

	// Ensure we don't exceed available height
	if len(visibleLines) > availableHeight {
		visibleLines = visibleLines[:availableHeight]
	}

	// Truncate each line to fit within panel width to prevent wrapping
	for i, line := range visibleLines {
		if len(line) > panelWidth {
			visibleLines[i] = line[:panelWidth-3] + "..."
		}
	}

	content.WriteString(strings.Join(visibleLines, "\n"))

	return content.String()
}
