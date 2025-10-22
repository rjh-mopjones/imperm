package ui

import (
	"imperm-ui/pkg/client"
	"imperm-ui/internal/control"
	"imperm-ui/internal/observe"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tabType int

const (
	tabControl tabType = iota
	tabObserve
)

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

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
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("235")).
		Padding(0, 2)

	activeTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("86")).
		Bold(true).
		Padding(0, 2)

	controlTabStyle := tabBarStyle
	observeTabStyle := tabBarStyle

	if m.currentTab == tabControl {
		controlTabStyle = activeTabStyle
	} else {
		observeTabStyle = activeTabStyle
	}

	tabBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		controlTabStyle.Render(" Control "),
		observeTabStyle.Render(" Observe "),
	)

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
		Foreground(lipgloss.Color("241")).
		Padding(0, 1)

	footer := footerStyle.Render("[Tab] Switch  [q] Quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
		footer,
	)
}
