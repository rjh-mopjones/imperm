package ui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorPrimary    = lipgloss.Color("86")  // Cyan/teal (main accent)
	ColorHighlight  = lipgloss.Color("51")  // Bright cyan (focus/active)
	ColorText       = lipgloss.Color("255") // White (primary text)
	ColorTextDim    = lipgloss.Color("245") // Gray (secondary/help text)
	ColorTextDimmer = lipgloss.Color("241") // Darker gray
	ColorTextPale   = lipgloss.Color("230") // Pale white
	ColorBorder     = lipgloss.Color("240") // Border gray
	ColorBorderDark = lipgloss.Color("235") // Dark border
	ColorBorderDim  = lipgloss.Color("236") // Dimmer border
	ColorBackground = lipgloss.Color("237") // Background gray
	ColorBackgroundAlt = lipgloss.Color("238") // Alt background
	ColorError      = lipgloss.Color("196") // Red (errors)
	ColorSuccess    = lipgloss.Color("46")  // Green (success)
	ColorWarning    = lipgloss.Color("220") // Yellow/orange (warnings)
	ColorRunning    = lipgloss.Color("86")  // Orange/yellow (running status)
)

// Common reusable styles
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Padding(1, 0)

	TitleStyleFocused = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Padding(0, 0, 1, 0)

	// Help/hint text
	HelpStyle = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Padding(1, 0)

	HintStyle = lipgloss.NewStyle().
		Foreground(ColorTextDimmer).
		Padding(0, 1)

	// Borders
	BorderStyle = lipgloss.NewStyle().
		BorderForeground(ColorBorder)

	BorderStyleFocused = lipgloss.NewStyle().
		BorderForeground(ColorHighlight)

	// Boxes/panels
	BoxStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	BoxStyleSelected = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Bold(true)

	// Status messages
	SuccessStyle = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Bold(true).
		Margin(1, 0)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true).
		Margin(1, 0)

	WarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarning).
		Bold(true).
		Margin(1, 0)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorTextPale).
		Background(ColorBackgroundAlt).
		Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
		Padding(0, 1)

	TableRowSelectedStyle = lipgloss.NewStyle().
		Background(ColorBackground).
		Foreground(ColorHighlight).
		Bold(true).
		Padding(0, 1)

	// Label/Value pairs for details views
	LabelStyle = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Width(15)

	ValueStyle = lipgloss.NewStyle().
		Foreground(ColorText)

	// Category and field styles for forms
	CategoryLabelStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	FieldStyle = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Padding(0, 2)

	FormLabelStyle = lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Bold(true).
		Width(25).
		Align(lipgloss.Right)

	// Action/Category box styles
	BoxStyleUnselected = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	// Info text styles
	InfoStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(ColorTextDim)

	InfoItemStyle = lipgloss.NewStyle().
		Padding(0, 1)

	// Stat display styles
	StatLabelStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	StatValueStyle = lipgloss.NewStyle().
		Foreground(ColorText).
		Bold(true)
)

// GetStatusColor returns the appropriate color for a status type
func GetStatusColor(statusType string) lipgloss.Color {
	switch statusType {
	case "error":
		return ColorError
	case "success":
		return ColorSuccess
	case "warning":
		return ColorWarning
	case "running":
		return ColorRunning
	default:
		return ColorTextDim
	}
}

// GetStatusStyle returns the appropriate style for a status message
func GetStatusStyle(statusType string, message string) lipgloss.Style {
	// Check for unsupported operation warnings
	if statusType == "error" && len(message) > 0 {
		// Simple check for warning-like messages (check for ⚠ emoji)
		if len(message) >= 3 && message[0:3] == "⚠" {
			return WarningStyle
		}
	}

	switch statusType {
	case "error":
		return ErrorStyle
	case "success":
		return SuccessStyle
	case "warning":
		return WarningStyle
	default:
		return HelpStyle
	}
}
