package ui

import (
	"imperm-ui/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// PanelLayout contains calculated dimensions for split panel layouts
type PanelLayout struct {
	LeftWidth   int
	RightWidth  int
	PanelHeight int
	TotalWidth  int
	TotalHeight int
}

// CalculateSplitLayout calculates dimensions for a two-panel split layout
func CalculateSplitLayout(totalWidth, totalHeight int) PanelLayout {
	leftWidth := totalWidth / config.SplitPanelRatio
	return PanelLayout{
		LeftWidth:   leftWidth,
		RightWidth:  totalWidth - leftWidth,
		PanelHeight: totalHeight - config.PanelHeightOffset,
		TotalWidth:  totalWidth,
		TotalHeight: totalHeight,
	}
}

// RenderSplitPanels renders left and right panels side by side
func RenderSplitPanels(layout PanelLayout, leftContent, rightContent string, rightBorder bool) string {
	leftBox := lipgloss.NewStyle().
		Width(layout.LeftWidth - config.PanelPadding).
		Height(layout.PanelHeight).
		Padding(1).
		Render(leftContent)

	rightStyle := lipgloss.NewStyle().
		Width(layout.RightWidth - config.PanelPadding).
		Height(layout.PanelHeight).
		Padding(1)

	if rightBorder {
		rightStyle = rightStyle.
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(ColorBorder)
	}

	rightBox := rightStyle.Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

// RenderStatusMessage renders a status message with appropriate styling and spacing
func RenderStatusMessage(message, messageType string) string {
	if message == "" {
		// Reserve space with just newlines (no visible bar)
		return "\n\n\n"
	}

	style := GetStatusStyle(messageType, message)
	return style.Render(message) + "\n"
}

// LogScrollRange calculates the start and end indices for displaying a scrollable log window
// When scrolling is false, shows the most recent lines (auto-follow mode)
// When scrolling is true, respects the scrollOffset for manual scrolling
func LogScrollRange(logs []string, availableLines int, scrolling bool, scrollOffset int) (startIdx, endIdx, adjustedOffset int) {
	if scrolling {
		// Manual scroll mode - respect scroll offset
		maxOffset := len(logs) - availableLines
		if maxOffset < 0 {
			maxOffset = 0
		}
		if scrollOffset > maxOffset {
			scrollOffset = maxOffset
		}
		startIdx = scrollOffset
		endIdx = startIdx + availableLines
		if endIdx > len(logs) {
			endIdx = len(logs)
		}
		return startIdx, endIdx, scrollOffset
	}

	// Auto-follow mode - show last N lines
	if len(logs) > availableLines {
		startIdx = len(logs) - availableLines
	} else {
		startIdx = 0
	}
	endIdx = len(logs)
	return startIdx, endIdx, scrollOffset
}
