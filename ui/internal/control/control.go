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

type screenType int

const (
	screenMainActions screenType = iota
	screenOptionCategories
	screenOptionForm
)

type optionCategory struct {
	name   string
	fields []optionField
}

type optionField struct {
	name        string
	placeholder string
	value       string
}

type Tab struct {
	client               client.Client
	history              []models.EnvironmentHistory
	selectedAction       int
	actions              []string
	textInput            textinput.Model
	inputMode            bool
	createWithOpts       bool
	width                int
	height               int
	currentScreen        screenType
	selectedCategory     int
	optionCategories     []optionCategory
	currentCategoryIndex int
	selectedField        int
	fieldInputs          []textinput.Model
}

func NewTab(client client.Client) *Tab {
	ti := textinput.New()
	ti.Placeholder = "environment-name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	// Initialize option categories
	categories := []optionCategory{
		{
			name: "DeployOptions",
			fields: []optionField{
				{name: "Name", placeholder: "environment-name (leave empty for auto-generated)"},
				{name: "DeploymentType", placeholder: "e.g., rolling, blue-green"},
				{name: "Replicas", placeholder: "e.g., 3"},
			},
		},
		{
			name: "DockerOptions",
			fields: []optionField{
				{name: "Registry", placeholder: "e.g., docker.io"},
				{name: "Tag", placeholder: "e.g., latest"},
			},
		},
		{
			name: "IdaasOptions",
			fields: []optionField{
				{name: "Provider", placeholder: "e.g., okta, auth0"},
				{name: "ClientID", placeholder: "your-client-id"},
			},
		},
		{
			name: "KafkaOptions",
			fields: []optionField{
				{name: "Brokers", placeholder: "e.g., localhost:9092"},
				{name: "Topic", placeholder: "e.g., events"},
			},
		},
		{
			name: "ServiceOptions",
			fields: []optionField{
				{name: "Port", placeholder: "e.g., 8080"},
				{name: "Protocol", placeholder: "e.g., http, grpc"},
			},
		},
		{
			name: "SftpOptions",
			fields: []optionField{
				{name: "Host", placeholder: "e.g., sftp.example.com"},
				{name: "Port", placeholder: "e.g., 22"},
			},
		},
		{
			name: "BranchOptions",
			fields: []optionField{
				{name: "Branch", placeholder: "e.g., main, develop"},
				{name: "CommitSHA", placeholder: "e.g., abc123"},
			},
		},
	}

	return &Tab{
		client:           client,
		history:          []models.EnvironmentHistory{},
		selectedAction:   0,
		actions: []string{
			"Build Environment",
			"Build Environment with Options",
			"Destroy Environment",
		},
		textInput:        ti,
		inputMode:        false,
		currentScreen:    screenMainActions,
		selectedCategory: 0,
		optionCategories: categories,
		selectedField:    0,
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
				go t.client.CreateEnvironment(envName, false)
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
				t.currentScreen = screenOptionCategories
				t.selectedCategory = 0
			case 2: // Destroy Environment
				if len(t.history) > 0 {
					envName := t.history[len(t.history)-1].Name
					go t.client.DestroyEnvironment(envName)
					return t, t.loadHistory
				}
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
		go t.client.CreateEnvironment(envName, true)
		t.currentScreen = screenMainActions
		return t, t.loadHistory
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

func (t *Tab) initializeFieldInputs() {
	category := t.optionCategories[t.currentCategoryIndex]
	t.fieldInputs = make([]textinput.Model, len(category.fields))

	for i, field := range category.fields {
		ti := textinput.New()
		ti.Placeholder = field.placeholder
		ti.CharLimit = 256
		ti.Width = 50
		ti.SetValue(field.value)

		if i == 0 {
			ti.Focus()
		}

		t.fieldInputs[i] = ti
	}
}

func (t *Tab) saveFieldValues() {
	if t.currentCategoryIndex >= 0 && t.currentCategoryIndex < len(t.optionCategories) {
		for i, input := range t.fieldInputs {
			if i < len(t.optionCategories[t.currentCategoryIndex].fields) {
				t.optionCategories[t.currentCategoryIndex].fields[i].value = input.Value()
			}
		}
	}
}

func (t *Tab) getEnvironmentName() string {
	// Check if name is set in DeployOptions
	for _, category := range t.optionCategories {
		if category.name == "DeployOptions" {
			for _, field := range category.fields {
				if field.name == "Name" && field.value != "" {
					return field.value
				}
			}
		}
	}
	// Generate default name with timestamp
	return fmt.Sprintf("env-%d", len(t.history)+1)
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
		leftPanel.WriteString(titleStyle.Render("Environment Name:"))
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

	var content strings.Builder
	content.WriteString(titleStyle.Render(category.name))
	content.WriteString("\n\n")

	for i, field := range category.fields {
		// Put label and input on same line
		label := labelStyle.Render(field.name + ":")

		var inputView string
		if i < len(t.fieldInputs) {
			inputView = t.fieldInputs[i].View()
		}

		line := lipgloss.JoinHorizontal(lipgloss.Left, label, " ", inputView)
		content.WriteString(line)
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("[↑↓/Tab] Navigate  [Enter] Next Field  [Esc] Save & Back"))

	box := lipgloss.NewStyle().
		Width(t.width - 4).
		Height(t.height - 8).
		Padding(2).
		Render(content.String())

	return box
}

