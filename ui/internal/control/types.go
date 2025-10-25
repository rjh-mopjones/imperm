package control

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"imperm-ui/pkg/client"
	"imperm-ui/pkg/models"
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
	currentOperation string
	operationLogs    []string
	operationStatus  string

	// Log panel focus and scrolling
	logPanelFocused bool
	logScrollOffset int

	// Status message
	statusMessage string
	statusTime    time.Time
	statusType    string // "success" or "error"
}

// Messages
type operationLogsMsg struct {
	logs    *models.OperationLogs
	envName string
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type tickMsg time.Time

type clearStatusMsg struct{}

type environmentCreatedMsg struct {
	envName string
	err     error
}
