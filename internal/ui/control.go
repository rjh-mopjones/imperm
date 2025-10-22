package ui

import (
	"fmt"
	"imperm/internal/middleware"
	"imperm/internal/models"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type controlTab struct {
	client          middleware.Client
	history         []models.EnvironmentHistory
	selectedAction  int
	actions         []string
	textInput       textinput.Model
	inputMode       bool
	createWithOpts  bool
	width           int
	height          int
}

func newControlTab(client middleware.Client) *controlTab {
	ti := textinput.New()
	ti.Placeholder = "environment-name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	return &controlTab{
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

func (c *controlTab) loadHistory() tea.Msg {
	history, err := c.client.GetEnvironmentHistory()
	if err != nil {
		return errMsg{err}
	}
	return historyLoadedMsg{history}
}

func (c *controlTab) Init() tea.Cmd {
	return c.loadHistory
}

func (c *controlTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height

	case historyLoadedMsg:
		c.history = msg.history

	case tea.KeyMsg:
		if c.inputMode {
			switch msg.String() {
			case "enter":
				envName := c.textInput.Value()
				if envName != "" {
					withOpts := c.createWithOpts
					go c.client.CreateEnvironment(envName, withOpts)
					c.textInput.Reset()
				}
				c.inputMode = false
				return c, c.loadHistory
			case "esc":
				c.inputMode = false
				c.textInput.Reset()
			default:
				c.textInput, cmd = c.textInput.Update(msg)
				return c, cmd
			}
		} else {
			switch msg.String() {
			case "up", "k":
				if c.selectedAction > 0 {
					c.selectedAction--
				}
			case "down", "j":
				if c.selectedAction < len(c.actions)-1 {
					c.selectedAction++
				}
			case "enter":
				switch c.selectedAction {
				case 0: // Build Environment
					c.inputMode = true
					c.createWithOpts = false
					c.textInput.Focus()
				case 1: // Build Environment with Options
					c.inputMode = true
					c.createWithOpts = true
					c.textInput.Focus()
				case 2: // Destroy Environment
					if len(c.history) > 0 {
						// Destroy the most recent environment
						envName := c.history[len(c.history)-1].Name
						go c.client.DestroyEnvironment(envName)
						return c, c.loadHistory
					}
				}
			}
		}
	}

	return c, cmd
}

func (c *controlTab) View() string {
	if c.width == 0 {
		return "Loading..."
	}

	leftWidth := c.width / 2
	rightWidth := c.width - leftWidth

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

	for i, action := range c.actions {
		style := actionStyle
		if i == c.selectedAction {
			style = selectedActionStyle
		}
		leftPanel.WriteString(style.Render(action))
		leftPanel.WriteString("\n\n")
	}

	if c.inputMode {
		leftPanel.WriteString("\n")
		prompt := "Environment Name:"
		if c.createWithOpts {
			prompt = "Environment Name (with options):"
		}
		leftPanel.WriteString(titleStyle.Render(prompt))
		leftPanel.WriteString("\n")
		leftPanel.WriteString(c.textInput.View())
		leftPanel.WriteString("\n\n")
		leftPanel.WriteString(historyStyle.Render("Press Enter to confirm, Esc to cancel"))
	}

	// Right panel - History
	var rightPanel strings.Builder
	rightPanel.WriteString(titleStyle.Render("Environment History"))
	rightPanel.WriteString("\n\n")

	if len(c.history) == 0 {
		rightPanel.WriteString(historyStyle.Render("No environments launched yet"))
	} else {
		for i := len(c.history) - 1; i >= 0; i-- {
			entry := c.history[i]
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
		Height(c.height - 8).
		Padding(1).
		Render(leftPanel.String())

	rightBox := lipgloss.NewStyle().
		Width(rightWidth - 2).
		Height(c.height - 8).
		Padding(1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("240")).
		Render(rightPanel.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
}
