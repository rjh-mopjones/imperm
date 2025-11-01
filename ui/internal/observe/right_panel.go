package observe

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"imperm-ui/internal/ui"
)

func (t *Tab) renderRightPanel(height int) string {
	// Calculate actual content width accounting for border and padding
	// The right panel style uses: Width(layout.RightWidth - 4) with Padding(1) and RoundedBorder()
	// RoundedBorder adds 2 chars (left + right), Padding(1) adds 2 chars (left + right)
	layout := ui.CalculateSplitLayout(t.width, t.height)
	// Start with RightWidth - 4 (from style), then subtract border (2) and padding (2)
	panelWidth := layout.RightWidth - 8
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

	// Hard wrap the content to prevent lipgloss from truncating it
	// We need hard wrapping (character-based) not word wrapping for JSON
	// Cache wrapped content to avoid expensive re-wrapping
	var wrappedContent string
	if viewContent == t.cachedViewContent && panelWidth == t.cachedPanelWidth && t.rightPanelView == t.cachedPanelView {
		// Content and width haven't changed, use cached version
		wrappedContent = t.cachedWrappedContent
	} else {
		// Content or width changed, re-wrap and cache
		wrappedContent = hardWrap(viewContent, panelWidth)
		t.cachedWrappedContent = wrappedContent
		t.cachedPanelWidth = panelWidth
		t.cachedViewContent = viewContent
		t.cachedPanelView = t.rightPanelView
	}
	content.WriteString(wrappedContent)

	return content.String()
}

// hardWrap performs hard wrapping at character boundaries for content like JSON
// Optimized version with pre-allocation and efficient string building
func hardWrap(text string, width int) string {
	if width <= 0 || text == "" {
		return text
	}

	// Pre-calculate approximate result size to reduce allocations
	// Estimate: original length + (number of lines we'll add * newline size)
	estimatedWraps := len(text) / width
	estimatedSize := len(text) + estimatedWraps
	result := make([]byte, 0, estimatedSize)

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Process each line
		for len(line) > width {
			// Append width characters directly as bytes
			result = append(result, line[:width]...)
			result = append(result, '\n')
			line = line[width:]
		}
		// Append remaining part of line
		result = append(result, line...)

		// Add newline between original lines (but not after the last one)
		if i < len(lines)-1 {
			result = append(result, '\n')
		}
	}

	return string(result)
}
