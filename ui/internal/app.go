package ui

import (
	"imperm-ui/internal/control"
	"imperm-ui/internal/observe"
	sharedui "imperm-ui/internal/ui"
	"imperm-ui/pkg/client"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tabType int

const (
	tabControl tabType = iota
	tabObserve
)

type Model struct {
	client      client.Client
	currentTab  tabType
	controlTab  *control.Tab
	observeTab  *observe.Tab
	width       int
	height      int
	initialized bool
}

func NewModel(client client.Client) *Model {
	return &Model{
		client:     client,
		currentTab: tabControl,
	}
}

func (m *Model) Init() tea.Cmd {
	// Initialize tabs lazily
	m.controlTab = control.NewTab(m.client)
	m.observeTab = observe.NewTab(m.client)

	var cmds []tea.Cmd
	cmds = append(cmds, m.controlTab.Init())
	cmds = append(cmds, m.observeTab.Init())

	m.initialized = true
	return tea.Batch(cmds...)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			// Switch tabs
			if m.currentTab == tabControl {
				m.currentTab = tabObserve
				// Auto-load environments when switching to observe tab
				return m, m.observeTab.Init()
			} else {
				m.currentTab = tabControl
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward window size to BOTH tabs
		_, cmd = m.controlTab.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = m.observeTab.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// Forward all messages to both tabs
	// Each tab will handle only the messages it cares about
	_, cmd = m.controlTab.Update(msg)
	cmds = append(cmds, cmd)
	_, cmd = m.observeTab.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.initialized {
		return "Initializing..."
	}

	// Tab bar
	tabBarStyle := lipgloss.NewStyle().
		Foreground(sharedui.ColorTextPale).
		Background(sharedui.ColorBorderDark).
		Padding(0, 2)

	activeTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(sharedui.ColorPrimary).
		Bold(true).
		Padding(0, 2)

	controlTabStyle := tabBarStyle
	observeTabStyle := tabBarStyle

	if m.currentTab == tabControl {
		controlTabStyle = activeTabStyle
	} else {
		observeTabStyle = activeTabStyle
	}

	tabs := lipgloss.JoinHorizontal(
		lipgloss.Top,
		controlTabStyle.Render(" Control "),
		observeTabStyle.Render(" Observe "),
	)

	// Add observe info when on observe tab
	var tabBarContent string
	if m.currentTab == tabObserve && m.observeTab != nil {
		infoStyle := lipgloss.NewStyle().
			Foreground(sharedui.ColorTextDimmer).
			Background(sharedui.ColorBorderDark).
			Padding(0, 2).
			MarginLeft(4)

		observeInfo := m.observeTab.GetHeaderInfo()
		tabBarContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			tabs,
			infoStyle.Render(observeInfo),
		)
	} else {
		tabBarContent = tabs
	}

	// Position tabs with left padding
	tabBar := lipgloss.NewStyle().
		PaddingLeft(4).
		Render(tabBarContent)

	// Content
	var content string
	switch m.currentTab {
	case tabControl:
		content = m.controlTab.View()
	case tabObserve:
		content = m.observeTab.View()
	}

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(sharedui.ColorTextDimmer).
		Padding(0, 1)

	footer := footerStyle.Render("[Tab] Switch  [q] Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
		footer,
	)
}
