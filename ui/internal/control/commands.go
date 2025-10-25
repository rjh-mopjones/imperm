package control

import (
	"imperm-ui/pkg/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// clearStatusAfterDelay clears the status message after 3 seconds
func (t *Tab) clearStatusAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// createEnvironment creates an environment asynchronously
func (t *Tab) createEnvironment(envName string, options *models.DeploymentOptions) tea.Cmd {
	return func() tea.Msg {
		// Start the async operation
		go func() {
			_ = t.client.CreateEnvironment(envName, options)
		}()
		// Return success immediately to show the message
		return environmentCreatedMsg{envName: envName, err: nil}
	}
}

// loadOperationLogs loads the operation logs for the current operation
func (t *Tab) loadOperationLogs() tea.Msg {
	if t.currentOperation == "" {
		return nil
	}

	logs, err := t.client.GetOperationLogs(t.currentOperation)
	if err != nil {
		return nil // Silently fail - logs might not be available yet
	}

	return operationLogsMsg{logs: logs, envName: t.currentOperation}
}

// tickCmd returns a command that ticks every 500ms for polling logs
func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
