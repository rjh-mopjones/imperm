package control

import (
	"fmt"
	"imperm-ui/pkg/client"
	"imperm-ui/pkg/models"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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
			"Retain Environment",
			"Get Environment",
			"Delete Environment",
		},
		textInput:        ti,
		inputMode:        false,
		currentScreen:    screenMainActions,
		selectedCategory: 0,
		optionCategories: categories,
		selectedField:    0,
	}
}

func (t *Tab) Init() tea.Cmd {
	return tickCmd()
}

// Helper methods for form management

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
		Name:      envName,
		Variables: make(map[string]string),
	}

	// Collect all non-empty field values from all categories
	for _, category := range t.optionCategories {
		for _, field := range category.fields {
			if field.value != "" {
				// Use the original variable name as the key
				options.Variables[field.name] = field.value
			}
		}
	}

	return options
}
