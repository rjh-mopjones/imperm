package control

import (
	"fmt"
	"imperm-ui/pkg/client"
	"imperm-ui/pkg/models"
	"imperm-ui/pkg/terraform"
	"strings"
	"time"

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

	// Operation logs
	currentOperation     string
	operationLogs        []string
	operationStatus      string
}

func NewTab(client client.Client) *Tab {
	ti := textinput.New()
	ti.Placeholder = "environment-name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	// Load option categories dynamically from Terraform modules
	categories := loadOptionsFromTerraform()

	return &Tab{
		client:           client,
		selectedAction:   0,
		actions: []string{
			"Build Environment",
			"Build Environment with Options",
		},
		textInput:        ti,
		inputMode:        false,
		currentScreen:    screenMainActions,
		selectedCategory: 0,
		optionCategories: categories,
		selectedField:    0,
	}
}

// loadOptionsFromTerraform loads option categories from Terraform modules
func loadOptionsFromTerraform() []optionCategory {
	// Try to load from default location (local Terraform module)
	loader, err := terraform.DefaultLoader()
	if err != nil {
		// If loading fails, fall back to hardcoded defaults
		// This ensures the UI always has options available
		// TODO: Add logging when available
		return getFallbackOptions()
	}

	tfCategories := loader.GetCategorizedOptions()
	if len(tfCategories) == 0 {
		// No categories found, use fallback
		return getFallbackOptions()
	}

	// Successfully loaded from Terraform!
	categories := make([]optionCategory, 0, len(tfCategories))

	for _, tfCat := range tfCategories {
		fields := make([]optionField, 0, len(tfCat.Variables))

		for _, tfVar := range tfCat.Variables {
			// Extract just the description part after " - "
			desc := tfVar.Description
			parts := strings.SplitN(desc, " - ", 2)
			if len(parts) == 2 {
				desc = parts[1]
			}

			// Create placeholder from description and default
			placeholder := desc
			if tfVar.Default != "" && tfVar.Default != "0" {
				placeholder = fmt.Sprintf("%s (default: %s)", desc, tfVar.Default)
			}

			// Convert snake_case to PascalCase for field name
			fieldName := toPascalCase(tfVar.Name)

			fields = append(fields, optionField{
				name:        fieldName,
				placeholder: placeholder,
				value:       "",
			})
		}

		categories = append(categories, optionCategory{
			name:   tfCat.Name,
			fields: fields,
		})
	}

	return categories
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// getFallbackOptions returns hardcoded options if Terraform loading fails
func getFallbackOptions() []optionCategory {
	return []optionCategory{
		{
			name: "DeployOptions",
			fields: []optionField{
				{name: "Name", placeholder: "environment-name (leave empty for auto-generated)"},
				{name: "Namespace", placeholder: "e.g., default, test-logging"},
				{name: "ConstantLogger", placeholder: "replicas (e.g., 3) - logs every 2s"},
				{name: "FastLogger", placeholder: "replicas (e.g., 2) - logs every 0.5s"},
				{name: "ErrorLogger", placeholder: "replicas (e.g., 1) - mixed INFO/ERROR logs"},
				{name: "JsonLogger", placeholder: "replicas (e.g., 2) - JSON formatted logs"},
			},
		},
		{
			name: "DockerOptions",
			fields: []optionField{
				{name: "DockerRegistry", placeholder: "Container registry URL (default: docker.io)"},
				{name: "DockerTag", placeholder: "Container image tag (default: latest)"},
				{name: "DockerPullPolicy", placeholder: "Image pull policy (default: IfNotPresent)"},
			},
		},
		{
			name: "ServiceOptions",
			fields: []optionField{
				{name: "ServicePort", placeholder: "Service port number (default: 8080)"},
				{name: "ServiceType", placeholder: "Kubernetes service type (default: ClusterIP)"},
			},
		},
	}
}

type operationLogsMsg struct {
	logs   *models.OperationLogs
	envName string
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type tickMsg time.Time

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

func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (t *Tab) Init() tea.Cmd {
	return tickCmd()
}

func (t *Tab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

	case operationLogsMsg:
		if msg.logs != nil && msg.envName == t.currentOperation {
			t.operationLogs = msg.logs.Logs
			t.operationStatus = msg.logs.Status

			// Clear current operation if completed or failed
			if msg.logs.Status == "completed" || msg.logs.Status == "failed" {
				go func() {
					time.Sleep(5 * time.Second)
					t.currentOperation = ""
					t.operationLogs = []string{}
					t.operationStatus = ""
				}()
			}
		}

	case tickMsg:
		// Poll for logs if we have a current operation
		if t.currentOperation != "" {
			return t, tea.Batch(tickCmd(), t.loadOperationLogs)
		}
		return t, tickCmd()

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
				// Create with nil options (no loggers)
				go t.client.CreateEnvironment(envName, nil)
				t.textInput.Reset()
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
		go t.client.CreateEnvironment(envName, options)
		t.currentScreen = screenMainActions
		return t, nil
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
	return fmt.Sprintf("env-%d", time.Now().Unix())
}

func (t *Tab) getDeploymentOptions(envName string) *models.DeploymentOptions {
	options := &models.DeploymentOptions{
		Name: envName,
	}

	// Parse DeployOptions category
	for _, category := range t.optionCategories {
		if category.name == "DeployOptions" {
			for _, field := range category.fields {
				if field.value == "" {
					continue
				}

				switch field.name {
				case "Namespace":
					options.Namespace = field.value
				case "ConstantLogger":
					// Parse replica count
					var replicas int
					fmt.Sscanf(field.value, "%d", &replicas)
					options.ConstantLogger = replicas
				case "FastLogger":
					var replicas int
					fmt.Sscanf(field.value, "%d", &replicas)
					options.FastLogger = replicas
				case "ErrorLogger":
					var replicas int
					fmt.Sscanf(field.value, "%d", &replicas)
					options.ErrorLogger = replicas
				case "JsonLogger":
					var replicas int
					fmt.Sscanf(field.value, "%d", &replicas)
					options.JsonLogger = replicas
				}
			}
		}
	}

	return options
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

	// Right panel - Operation Logs
	var rightPanel strings.Builder

	rightPanel.WriteString(titleStyle.Render("Terraform Logs"))
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

		// Show logs (last 20 lines to fit in panel)
		logStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Width(rightWidth - 6)

		startIdx := 0
		if len(t.operationLogs) > 20 {
			startIdx = len(t.operationLogs) - 20
		}

		for i := startIdx; i < len(t.operationLogs); i++ {
			rightPanel.WriteString(logStyle.Render(t.operationLogs[i]))
			rightPanel.WriteString("\n")
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

