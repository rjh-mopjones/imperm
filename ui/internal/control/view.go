package control

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (t *Tab) View() string {
	if t.width == 0 {
		return "Loading..."
	}

	switch t.currentScreen {
	case screenMainActions:
		return t.viewMainActions()
	case screenOptionCategories:
		return t.viewOptionCategories()
	case screenOptionForm:
		return t.viewOptionForm()
	}

	return "Unknown screen"
}

func (t *Tab) viewMainActions() string {
	leftWidth := t.width / 2
	rightWidth := t.width - leftWidth

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	actionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	selectedActionStyle := actionStyle.Copy().
		BorderForeground(lipgloss.Color("86")).
		Bold(true)

	historyStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color("245"))

	historyItemStyle := lipgloss.NewStyle().
		Padding(0, 1)

	// Left panel - Actions
	var leftPanel strings.Builder
	leftPanel.WriteString(titleStyle.Render("Actions"))
	leftPanel.WriteString("\n")

	// Always reserve space for status message (so layout doesn't shift)
	if t.statusMessage != "" {
		var statusColor string
		if t.statusType == "error" {
			// Check if it's a warning (unsupported operation)
			if strings.Contains(t.statusMessage, "Unsupported operation") {
				statusColor = "220" // Yellow/orange
			} else {
				statusColor = "196" // Red
			}
		} else {
			statusColor = "46" // Green
		}
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(statusColor)).
			Bold(true).
			Margin(1, 0)
		leftPanel.WriteString(statusStyle.Render(t.statusMessage))
		leftPanel.WriteString("\n")
	} else {
		// Reserve space with just a newline (no visible bar)
		leftPanel.WriteString("\n\n\n")
	}

	leftPanel.WriteString("\n")

	for i, action := range t.actions {
		style := actionStyle
		if i == t.selectedAction {
			style = selectedActionStyle
		}
		leftPanel.WriteString(style.Render(action))
		leftPanel.WriteString("\n\n")
	}

	if t.inputMode {
		leftPanel.WriteString("\n")
		leftPanel.WriteString(titleStyle.Render("Environment Name:"))
		leftPanel.WriteString("\n")
		leftPanel.WriteString(t.textInput.View())
		leftPanel.WriteString("\n\n")
		leftPanel.WriteString(historyStyle.Render("Press Enter to confirm, Esc to cancel"))
	} else if !t.logPanelFocused {
		// Show help text when not in input mode and logs not focused
		leftPanel.WriteString("\n")
		leftPanel.WriteString(historyStyle.Render("[↑↓/jk] Navigate  [→/l] View Logs  [Enter] Select"))
	} else {
		// Show help text when logs are focused
		leftPanel.WriteString("\n")
		leftPanel.WriteString(historyStyle.Render("[←/h/Esc] Back  [↑↓/jk] Scroll Logs"))
	}

	// Right panel - Operation Logs
	var rightPanel strings.Builder

	// Title color changes based on focus
	logTitleColor := lipgloss.Color("245") // Grey when not focused
	if t.logPanelFocused {
		logTitleColor = lipgloss.Color("51") // Cyan when focused
	}
	logTitleStyle := titleStyle.Copy().Foreground(logTitleColor)

	rightPanel.WriteString(logTitleStyle.Render("Terraform Logs"))
	rightPanel.WriteString("\n\n")

	if t.currentOperation != "" {
		statusColor := "86"
		if t.operationStatus == "failed" {
			statusColor = "196"
		} else if t.operationStatus == "completed" {
			statusColor = "46"
		}

		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(statusColor)).
			Bold(true)

		rightPanel.WriteString(historyItemStyle.Render(
			fmt.Sprintf("Environment: %s\nStatus: %s\n\n",
				t.currentOperation,
				statusStyle.Render(strings.ToUpper(t.operationStatus)),
			),
		))

		// Show logs (auto-scroll to bottom, showing most recent)
		logStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Width(rightWidth - 6)

		// Calculate available height for logs (subtract title, status, padding)
		availableLines := t.height - 12
		if availableLines < 5 {
			availableLines = 5
		}

		// If in scroll mode, use scroll offset; otherwise auto-follow (show last N lines)
		var startIdx, endIdx int
		if t.logPanelFocused {
			// Manual scroll mode - respect scroll offset
			maxOffset := len(t.operationLogs) - availableLines
			if maxOffset < 0 {
				maxOffset = 0
			}
			if t.logScrollOffset > maxOffset {
				t.logScrollOffset = maxOffset
			}
			startIdx = t.logScrollOffset
			endIdx = startIdx + availableLines
			if endIdx > len(t.operationLogs) {
				endIdx = len(t.operationLogs)
			}
		} else {
			// Auto-follow mode - show last N lines
			if len(t.operationLogs) > availableLines {
				startIdx = len(t.operationLogs) - availableLines
			} else {
				startIdx = 0
			}
			endIdx = len(t.operationLogs)
		}

		for i := startIdx; i < endIdx; i++ {
			rightPanel.WriteString(logStyle.Render(t.operationLogs[i]))
			rightPanel.WriteString("\n")
		}

		// Show scroll indicator only when in manual scroll mode
		if t.logPanelFocused && len(t.operationLogs) > availableLines {
			scrollInfo := fmt.Sprintf("[%d-%d/%d]", startIdx+1, endIdx, len(t.operationLogs))
			rightPanel.WriteString("\n")
			rightPanel.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(scrollInfo))
		}
	} else {
		// No operation running
		rightPanel.WriteString(historyStyle.Render("No operation running\n\nSelect an action to begin"))
	}

	// Combine panels
	leftBox := lipgloss.NewStyle().
		Width(leftWidth - 2).
		Height(t.height - 8).
		Padding(1).
		Render(leftPanel.String())

	rightBox := lipgloss.NewStyle().
		Width(rightWidth - 2).
		Height(t.height - 8).
		Padding(1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("240")).
		Render(rightPanel.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

func (t *Tab) viewOptionCategories() string {
	leftWidth := t.width / 2
	rightWidth := t.width - leftWidth

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	categoryStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(40)

	selectedCategoryStyle := categoryStyle.Copy().
		BorderForeground(lipgloss.Color("86")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	// Left panel - Categories
	var leftPanel strings.Builder
	leftPanel.WriteString(titleStyle.Render("Build Environment with Options"))
	leftPanel.WriteString("\n\n")
	leftPanel.WriteString(helpStyle.Render("Select an option category to configure:"))
	leftPanel.WriteString("\n\n")

	for i, category := range t.optionCategories {
		style := categoryStyle
		if i == t.selectedCategory {
			style = selectedCategoryStyle
		}

		// Show checkmark if category has values
		hasValues := false
		for _, field := range category.fields {
			if field.value != "" {
				hasValues = true
				break
			}
		}

		label := category.name
		if hasValues {
			label = "✓ " + label
		}

		leftPanel.WriteString(style.Render(label))
		leftPanel.WriteString("\n")
	}

	leftPanel.WriteString("\n")
	leftPanel.WriteString(helpStyle.Render("[↑↓/jk] Navigate  [Enter] Configure  [c] Create  [Esc] Back"))

	// Right panel - Configured options
	var rightPanel strings.Builder
	rightPanel.WriteString(titleStyle.Render("Configured Options"))
	rightPanel.WriteString("\n\n")

	categoryLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 2)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	hasAnyValues := false
	for _, category := range t.optionCategories {
		categoryHasValues := false
		var categoryContent strings.Builder

		for _, field := range category.fields {
			if field.value != "" {
				categoryHasValues = true
				hasAnyValues = true
				categoryContent.WriteString(fieldStyle.Render(fmt.Sprintf("%s: %s", field.name, valueStyle.Render(field.value))))
				categoryContent.WriteString("\n")
			}
		}

		if categoryHasValues {
			rightPanel.WriteString(categoryLabelStyle.Render(category.name))
			rightPanel.WriteString("\n")
			rightPanel.WriteString(categoryContent.String())
			rightPanel.WriteString("\n")
		}
	}

	if !hasAnyValues {
		rightPanel.WriteString(helpStyle.Render("No options configured yet"))
	}

	// Combine panels
	leftBox := lipgloss.NewStyle().
		Width(leftWidth - 2).
		Height(t.height - 8).
		Padding(1).
		Render(leftPanel.String())

	rightBox := lipgloss.NewStyle().
		Width(rightWidth - 2).
		Height(t.height - 8).
		Padding(1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("240")).
		Render(rightPanel.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}

func (t *Tab) viewOptionForm() string {
	if t.currentCategoryIndex < 0 || t.currentCategoryIndex >= len(t.optionCategories) {
		return "Invalid category"
	}

	leftWidth := t.width / 2
	rightWidth := t.width - leftWidth

	category := t.optionCategories[t.currentCategoryIndex]

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Bold(true).
		Width(25).
		Align(lipgloss.Right)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	// Left panel - Form
	var leftPanel strings.Builder
	leftPanel.WriteString(titleStyle.Render(category.name))
	leftPanel.WriteString("\n\n")

	for i, field := range category.fields {
		// Put label and input on same line
		label := labelStyle.Render(field.name + ":")

		var inputView string
		if i < len(t.fieldInputs) {
			inputView = t.fieldInputs[i].View()
		}

		line := lipgloss.JoinHorizontal(lipgloss.Left, label, " ", inputView)
		leftPanel.WriteString(line)
		leftPanel.WriteString("\n")
	}

	leftPanel.WriteString("\n")
	leftPanel.WriteString(helpStyle.Render("[↑↓/Tab] Navigate  [Enter] Next Field  [Esc] Save & Back"))

	// Right panel - Configured options (same as in viewOptionCategories)
	var rightPanel strings.Builder
	rightPanel.WriteString(titleStyle.Render("Configured Options"))
	rightPanel.WriteString("\n\n")

	categoryLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	fieldStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 2)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	hasAnyValues := false
	for catIdx, cat := range t.optionCategories {
		categoryHasValues := false
		var categoryContent strings.Builder

		for fieldIdx, field := range cat.fields {
			var displayValue string
			// If we're currently editing this category, show live input values
			if catIdx == t.currentCategoryIndex && fieldIdx < len(t.fieldInputs) {
				displayValue = t.fieldInputs[fieldIdx].Value()
			} else {
				displayValue = field.value
			}

			if displayValue != "" {
				categoryHasValues = true
				hasAnyValues = true
				categoryContent.WriteString(fieldStyle.Render(fmt.Sprintf("%s: %s", field.name, valueStyle.Render(displayValue))))
				categoryContent.WriteString("\n")
			}
		}

		if categoryHasValues {
			rightPanel.WriteString(categoryLabelStyle.Render(cat.name))
			rightPanel.WriteString("\n")
			rightPanel.WriteString(categoryContent.String())
			rightPanel.WriteString("\n")
		}
	}

	if !hasAnyValues {
		rightPanel.WriteString(helpStyle.Render("No options configured yet"))
	}

	// Combine panels
	leftBox := lipgloss.NewStyle().
		Width(leftWidth - 2).
		Height(t.height - 8).
		Padding(1).
		Render(leftPanel.String())

	rightBox := lipgloss.NewStyle().
		Width(rightWidth - 2).
		Height(t.height - 8).
		Padding(1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("240")).
		Render(rightPanel.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}
