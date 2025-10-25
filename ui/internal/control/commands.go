package control

import (
	"fmt"
	"time"

	"imperm-ui/internal/config"
	"imperm-ui/internal/messages"
	"imperm-ui/pkg/models"

	tea "github.com/charmbracelet/bubbletea"
)

// setStatus sets a status message and returns a command to clear it after a delay
func (t *Tab) setStatus(msgType, format string, args ...interface{}) tea.Cmd {
	t.statusMessage = fmt.Sprintf(format, args...)
	t.statusType = msgType
	t.statusTime = time.Now()
	return t.clearStatusAfterDelay()
}

// clearStatusAfterDelay clears the status message after 3 seconds
func (t *Tab) clearStatusAfterDelay() tea.Cmd {
	return tea.Tick(config.StatusMessageTimeout, func(t time.Time) tea.Msg {
		return messages.ClearStatusMsg{}
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

// tickCmd returns a command that ticks for polling logs
func tickCmd() tea.Cmd {
	return tea.Tick(config.LogPollingInterval, func(t time.Time) tea.Msg {
		return messages.TickMsg(t)
	})
}
