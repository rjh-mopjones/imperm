package control

import (
	"imperm-ui/internal/messages"

	tea "github.com/charmbracelet/bubbletea"
)

func (t *Tab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

	case operationLogsMsg:
		if msg.logs != nil && msg.envName == t.currentOperation {
			// Limit log buffer to prevent memory bloat from long-running operations
			// Keep only the most recent 1000 log lines
			const maxLogLines = 1000
			logs := msg.logs.Logs
			if len(logs) > maxLogLines {
				// Keep the most recent logs
				logs = logs[len(logs)-maxLogLines:]
			}
			t.operationLogs = logs
			t.operationStatus = msg.logs.Status
			// Logs are now persistent and won't be cleared automatically
		}

	case messages.TickMsg:
		// Poll for logs if we have a current operation
		if t.currentOperation != "" {
			return t, tea.Batch(tickCmd(), t.loadOperationLogs)
		}
		return t, tickCmd()

	case messages.ClearStatusMsg:
		// Clear the status message
		t.statusMessage = ""
		return t, nil

	case environmentCreatedMsg:
		// Only handle errors here since success is shown immediately
		if msg.err != nil {
			return t, t.setStatus("error", "❌ Failed to create environment '%s': %v", msg.envName, msg.err)
		}

	case tea.KeyMsg:
		switch t.currentScreen {
		case screenMainActions:
			return t.updateMainActions(msg)
		case screenOptionCategories:
			return t.updateOptionCategories(msg)
		case screenOptionForm:
			return t.updateOptionForm(msg)
		}
	}

	return t, cmd
}

func (t *Tab) updateMainActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if t.inputMode {
		switch msg.String() {
		case "enter":
			envName := t.textInput.Value()
			if envName != "" {
				t.currentOperation = envName
				t.operationLogs = []string{}
				t.operationStatus = "running"
				t.textInput.Reset()
				t.inputMode = false
				// Create with nil options (no loggers)
				return t, tea.Batch(t.createEnvironment(envName, nil), t.setStatus("success", "✓ Started creating environment '%s'", envName))
			}
			t.inputMode = false
			return t, nil
		case "esc":
			t.inputMode = false
			t.textInput.Reset()
		default:
			t.textInput, cmd = t.textInput.Update(msg)
			return t, cmd
		}
	} else if t.logPanelFocused {
		// Handle log panel navigation
		switch msg.String() {
		case "up", "k":
			if t.logScrollOffset > 0 {
				t.logScrollOffset--
			}
		case "down", "j":
			t.logScrollOffset++
		case "left", "h", "esc":
			// Exit log panel focus
			t.logPanelFocused = false
			t.logScrollOffset = 0
		}
	} else {
		switch msg.String() {
		case "up", "k":
			if t.selectedAction > 0 {
				t.selectedAction--
			}
		case "down", "j":
			if t.selectedAction < len(t.actions)-1 {
				t.selectedAction++
			}
		case "right", "l":
			// Focus log panel if there are logs
			if t.currentOperation != "" && len(t.operationLogs) > 0 {
				t.logPanelFocused = true
				t.logScrollOffset = 0
			}
		case "enter":
			switch t.selectedAction {
			case 0: // Build Environment
				t.inputMode = true
				t.createWithOpts = false
				t.textInput.Focus()
			case 1: // Build Environment with Options
				t.currentScreen = screenOptionCategories
				t.selectedCategory = 0
			case 2: // Retain Environment
				return t, t.setStatus("error", "⚠️  Unsupported operation: Retain Environment")
			case 3: // Get Environment
				return t, t.setStatus("error", "⚠️  Unsupported operation: Get Environment")
			case 4: // Delete Environment
				return t, t.setStatus("error", "⚠️  Unsupported operation: Delete Environment")
			}
		}
	}

	return t, cmd
}

func (t *Tab) updateOptionCategories(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if t.selectedCategory > 0 {
			t.selectedCategory--
		}
	case "down", "j":
		if t.selectedCategory < len(t.optionCategories)-1 {
			t.selectedCategory++
		}
	case "enter":
		// Enter the selected category form
		t.currentCategoryIndex = t.selectedCategory
		t.currentScreen = screenOptionForm
		t.selectedField = 0
		t.initializeFieldInputs()
	case "esc":
		// Go back to main actions
		t.currentScreen = screenMainActions
	case "c":
		// Create environment with configured options
		envName := t.getEnvironmentName()
		options := t.getDeploymentOptions(envName)
		t.currentOperation = envName
		t.operationLogs = []string{}
		t.operationStatus = "running"
		t.currentScreen = screenMainActions
		return t, tea.Batch(t.createEnvironment(envName, options), t.setStatus("success", "✓ Started creating environment '%s'", envName))
	}

	return t, nil
}

func (t *Tab) updateOptionForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "up":
		if t.selectedField > 0 {
			t.fieldInputs[t.selectedField].Blur()
			t.selectedField--
			t.fieldInputs[t.selectedField].Focus()
		}
	case "down", "tab":
		if t.selectedField < len(t.fieldInputs)-1 {
			t.fieldInputs[t.selectedField].Blur()
			t.selectedField++
			t.fieldInputs[t.selectedField].Focus()
		}
	case "enter":
		// Save current field value and move to next
		t.saveFieldValues()
		if t.selectedField < len(t.fieldInputs)-1 {
			t.fieldInputs[t.selectedField].Blur()
			t.selectedField++
			t.fieldInputs[t.selectedField].Focus()
		}
	case "esc":
		// Save and go back to category selection
		t.saveFieldValues()
		t.currentScreen = screenOptionCategories
	default:
		// Update the focused input (allows typing hjkl and other characters)
		if t.selectedField < len(t.fieldInputs) {
			t.fieldInputs[t.selectedField], cmd = t.fieldInputs[t.selectedField].Update(msg)
		}
		return t, cmd
	}

	return t, nil
}
