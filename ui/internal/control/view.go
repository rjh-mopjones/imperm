package control

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"imperm-ui/internal/config"
	"imperm-ui/internal/ui"
)

// renderConfiguredOptionsPanel renders the "Configured Options" panel showing all configured values
func (t *Tab) renderConfiguredOptionsPanel(showLiveInputs bool) string {
	var panel strings.Builder
	panel.WriteString(ui.TitleStyle.Render("Configured Options"))
	panel.WriteString("\n\n")

	hasAnyValues := false
	for catIdx, cat := range t.optionCategories {
		categoryHasValues := false
		var categoryContent strings.Builder

		for fieldIdx, field := range cat.fields {
			var displayValue string
			// If showing live inputs and we're editing this category, show input values
			if showLiveInputs && catIdx == t.currentCategoryIndex && fieldIdx < len(t.fieldInputs) {
				displayValue = t.fieldInputs[fieldIdx].Value()
			} else {
				displayValue = field.value
			}

			if displayValue != "" {
				categoryHasValues = true
				hasAnyValues = true
				categoryContent.WriteString(ui.FieldStyle.Render(fmt.Sprintf("%s: %s", field.name, ui.ValueStyle.Render(displayValue))))
				categoryContent.WriteString("\n")
			}
		}

		if categoryHasValues {
			panel.WriteString(ui.CategoryLabelStyle.Render(cat.name))
			panel.WriteString("\n")
			panel.WriteString(categoryContent.String())
			panel.WriteString("\n")
		}
	}

	if !hasAnyValues {
		panel.WriteString(ui.HelpStyle.Render("No options configured yet"))
	}

	return panel.String()
}

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
	layout := ui.CalculateSplitLayout(t.width, t.height)

	selectedActionStyle := ui.BoxStyleUnselected.Copy().
		BorderForeground(ui.ColorPrimary).
		Bold(true)

	// Left panel - Actions
	var leftPanel strings.Builder
	leftPanel.WriteString(ui.TitleStyle.Render("Actions"))
	leftPanel.WriteString("\n")

	// Always reserve space for status message (so layout doesn't shift)
	leftPanel.WriteString(ui.RenderStatusMessage(t.statusMessage, t.statusType))
	leftPanel.WriteString("\n")

	leftPanel.WriteString("\n")

	for i, action := range t.actions {
		style := ui.BoxStyleUnselected
		if i == t.selectedAction {
			style = selectedActionStyle
		}
		leftPanel.WriteString(style.Render(action))
		leftPanel.WriteString("\n\n")
	}

	if t.inputMode {
		leftPanel.WriteString("\n")
		leftPanel.WriteString(ui.TitleStyle.Render("Environment Name:"))
		leftPanel.WriteString("\n")
		leftPanel.WriteString(t.textInput.View())
		leftPanel.WriteString("\n\n")
		leftPanel.WriteString(ui.InfoStyle.Render("Press Enter to confirm, Esc to cancel"))
	} else if !t.logPanelFocused {
		// Show help text when not in input mode and logs not focused
		leftPanel.WriteString("\n")
		leftPanel.WriteString(ui.InfoStyle.Render("[↑↓/jk] Navigate  [→/l] View Logs  [Enter] Select"))
	} else {
		// Show help text when logs are focused
		leftPanel.WriteString("\n")
		leftPanel.WriteString(ui.InfoStyle.Render("[←/h/Esc] Back  [↑↓/jk] Scroll Logs"))
	}

	// Right panel - Operation Logs
	var rightPanel strings.Builder

	// Title color changes based on focus
	logTitleColor := ui.ColorTextDim
	if t.logPanelFocused {
		logTitleColor = ui.ColorHighlight
	}
	logTitleStyle := ui.TitleStyle.Copy().Foreground(logTitleColor)

	rightPanel.WriteString(logTitleStyle.Render("Terraform Logs"))
	rightPanel.WriteString("\n\n")

	if t.currentOperation != "" {
		statusColor := ui.ColorRunning
		if t.operationStatus == "failed" {
			statusColor = ui.ColorError
		} else if t.operationStatus == "completed" {
			statusColor = ui.ColorSuccess
		}

		statusStyle := lipgloss.NewStyle().
			Foreground(statusColor).
			Bold(true)

		rightPanel.WriteString(ui.InfoItemStyle.Render(
			fmt.Sprintf("Environment: %s\nStatus: %s\n\n",
				t.currentOperation,
				statusStyle.Render(strings.ToUpper(t.operationStatus)),
			),
		))

		// Show logs (auto-scroll to bottom, showing most recent)
		logStyle := lipgloss.NewStyle().
			Foreground(ui.ColorText).
			Width(layout.RightWidth - config.LogWidthAdjustment)

		// Calculate available height for logs (subtract title, status, padding)
		availableLines := t.height - config.ContentHeightOffset
		if availableLines < config.MinLogLines {
			availableLines = config.MinLogLines
		}

		// Calculate scroll range
		startIdx, endIdx, adjustedOffset := ui.LogScrollRange(t.operationLogs, availableLines, t.logPanelFocused, t.logScrollOffset)
		t.logScrollOffset = adjustedOffset

		for i := startIdx; i < endIdx; i++ {
			rightPanel.WriteString(logStyle.Render(t.operationLogs[i]))
			rightPanel.WriteString("\n")
		}

		// Show scroll indicator only when in manual scroll mode
		if t.logPanelFocused && len(t.operationLogs) > availableLines {
			scrollInfo := fmt.Sprintf("[%d-%d/%d]", startIdx+1, endIdx, len(t.operationLogs))
			rightPanel.WriteString("\n")
			rightPanel.WriteString(lipgloss.NewStyle().
				Foreground(ui.ColorTextDimmer).
				Render(scrollInfo))
		}
	} else {
		// No operation running
		rightPanel.WriteString(ui.InfoStyle.Render("No operation running\n\nSelect an action to begin"))
	}

	// Combine panels
	return ui.RenderSplitPanels(layout, leftPanel.String(), rightPanel.String(), true)
}

func (t *Tab) viewOptionCategories() string {
	categoryStyle := ui.BoxStyleUnselected.Copy().
		Width(config.CategoryBoxWidth)

	selectedCategoryStyle := categoryStyle.Copy().
		BorderForeground(ui.ColorPrimary).
		Bold(true)

	// Left panel - Categories
	var leftPanel strings.Builder
	leftPanel.WriteString(ui.TitleStyle.Render("Build Environment with Options"))
	leftPanel.WriteString("\n\n")
	leftPanel.WriteString(ui.HelpStyle.Render("Select an option category to configure:"))
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
	leftPanel.WriteString(ui.HelpStyle.Render("[↑↓/jk] Navigate  [Enter] Configure  [c] Create  [Esc] Back"))

	// Right panel - Configured options
	rightPanelContent := t.renderConfiguredOptionsPanel(false)

	// Combine panels
	layout := ui.CalculateSplitLayout(t.width, t.height)
	return ui.RenderSplitPanels(layout, leftPanel.String(), rightPanelContent, true)
}

func (t *Tab) viewOptionForm() string {
	if t.currentCategoryIndex < 0 || t.currentCategoryIndex >= len(t.optionCategories) {
		return "Invalid category"
	}

	category := t.optionCategories[t.currentCategoryIndex]

	// Left panel - Form
	var leftPanel strings.Builder
	leftPanel.WriteString(ui.TitleStyle.Render(category.name))
	leftPanel.WriteString("\n\n")

	for i, field := range category.fields {
		// Put label and input on same line
		label := ui.FormLabelStyle.Render(field.name + ":")

		var inputView string
		if i < len(t.fieldInputs) {
			inputView = t.fieldInputs[i].View()
		}

		line := lipgloss.JoinHorizontal(lipgloss.Left, label, " ", inputView)
		leftPanel.WriteString(line)
		leftPanel.WriteString("\n")
	}

	leftPanel.WriteString("\n")
	leftPanel.WriteString(ui.HelpStyle.Render("[↑↓/Tab] Navigate  [Enter] Next Field  [Esc] Save & Back"))

	// Right panel - Configured options (with live input values)
	rightPanelContent := t.renderConfiguredOptionsPanel(true)

	// Combine panels
	layout := ui.CalculateSplitLayout(t.width, t.height)
	return ui.RenderSplitPanels(layout, leftPanel.String(), rightPanelContent, true)
}
