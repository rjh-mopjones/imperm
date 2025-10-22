package control

import (
	"fmt"
	"imperm-ui/pkg/client"
	"imperm-ui/pkg/models"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Tab struct {
	client          client.Client
	history         []models.EnvironmentHistory
	selectedAction  int
	actions         []string
	textInput       textinput.Model
	inputMode       bool
	createWithOpts  bool
	width           int
	height          int
}

func NewTab(client client.Client) *Tab {
	ti := textinput.New()
	ti.Placeholder = "environment-name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	return &Tab{
		client:         client,
		history:        []models.EnvironmentHistory{},
		selectedAction: 0,
		actions: []string{
			"Build Environment",
			"Build Environment with Options",
			"Destroy Environment",
		},
		textInput: ti,
		inputMode: false,
	}
}

type historyLoadedMsg struct {
	history []models.EnvironmentHistory
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

func (t *Tab) loadHistory() tea.Msg {
	history, err := t.client.GetEnvironmentHistory()
	if err != nil {
		return errMsg{err}
	}
	return historyLoadedMsg{history}
}

func (t *Tab) Init() tea.Cmd {
	return t.loadHistory
}

func (t *Tab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

	case historyLoadedMsg:
		t.history = msg.history

	case tea.KeyMsg:
		if t.inputMode {
			switch msg.String() {
			case "enter":
				envName := t.textInput.Value()
				if envName != "" {
					withOpts := t.createWithOpts
					go t.client.CreateEnvironment(envName, withOpts)
					t.textInput.Reset()
				}
				t.inputMode = false
				return t, t.loadHistory
			case "esc":
				t.inputMode = false
				t.textInput.Reset()
			default:
				t.textInput, cmd = t.textInput.Update(msg)
				return t, cmd
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
			case "enter":
				switch t.selectedAction {
				case 0: // Build Environment
					t.inputMode = true
					t.createWithOpts = false
					t.textInput.Focus()
				case 1: // Build Environment with Options
					t.inputMode = true
					t.createWithOpts = true
					t.textInput.Focus()
				case 2: // Destroy Environment
					if len(t.history) > 0 {
						// Destroy the most recent environment
						envName := t.history[len(t.history)-1].Name
						go t.client.DestroyEnvironment(envName)
						return t, t.loadHistory
					}
				}
			}
		}
	}

	return t, cmd
}

func (t *Tab) View() string {
	if t.width == 0 {
		return "Loading..."
	}

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
	leftPanel.WriteString("\n\n")

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
		prompt := "Environment Name:"
		if t.createWithOpts {
			prompt = "Environment Name (with options):"
		}
		leftPanel.WriteString(titleStyle.Render(prompt))
		leftPanel.WriteString("\n")
		leftPanel.WriteString(t.textInput.View())
		leftPanel.WriteString("\n\n")
		leftPanel.WriteString(historyStyle.Render("Press Enter to confirm, Esc to cancel"))
	}

	// Right panel - History
	var rightPanel strings.Builder
	rightPanel.WriteString(titleStyle.Render("Environment History"))
	rightPanel.WriteString("\n\n")

	if len(t.history) == 0 {
		rightPanel.WriteString(historyStyle.Render("No environments launched yet"))
	} else {
		for i := len(t.history) - 1; i >= 0; i-- {
			entry := t.history[i]
			opts := ""
			if entry.WithOptions {
				opts = " [with options]"
			}
			text := fmt.Sprintf("• %s%s\n  %s - %s",
				entry.Name,
				opts,
				entry.LaunchedAt.Format("15:04:05"),
				entry.Status,
			)
			rightPanel.WriteString(historyItemStyle.Render(text))
			rightPanel.WriteString("\n\n")
		}
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
