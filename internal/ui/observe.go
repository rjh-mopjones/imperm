package ui

import (
	"fmt"
	"imperm/internal/middleware"
	"imperm/internal/models"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type resourceType int

const (
	resourceEnvironments resourceType = iota
	resourcePods
	resourceDeployments
)

type panelFocus int

const (
	focusTable panelFocus = iota
	focusRightPanel
)

type rightPanelView int

const (
	rightPanelDetails rightPanelView = iota
	rightPanelLogs
	rightPanelEvents
	rightPanelStats
)

type observeTab struct {
	client           middleware.Client
	currentResource  resourceType
	environments     []models.Environment
	pods             []models.Pod
	deployments      []models.Deployment
	selectedIndex    int
	width            int
	height           int
	lastUpdate       time.Time
	autoRefresh      bool
	refreshInterval  time.Duration

	// Drill-down state
	selectedEnvironment *models.Environment
	filterNamespace     string

	// Panel navigation
	panelFocus     panelFocus
	rightPanelView rightPanelView

	// Scrolling for right panel
	scrollOffset int
}

func newObserveTab(client middleware.Client) *observeTab {
	return &observeTab{
		client:          client,
		currentResource: resourceEnvironments,
		environments:    []models.Environment{},
		pods:            []models.Pod{},
		deployments:     []models.Deployment{},
		selectedIndex:   0,
		autoRefresh:     true,
		refreshInterval: 5 * time.Second,
	}
}

type tickMsg time.Time

type resourcesLoadedMsg struct {
	environments []models.Environment
	pods         []models.Pod
	deployments  []models.Deployment
}

func (o *observeTab) loadResources() tea.Msg {
	envs, err := o.client.ListEnvironments()
	if err != nil {
		return errMsg{err}
	}

	pods, err := o.client.ListPods(o.filterNamespace)
	if err != nil {
		return errMsg{err}
	}

	deployments, err := o.client.ListDeployments(o.filterNamespace)
	if err != nil {
		return errMsg{err}
	}

	return resourcesLoadedMsg{
		environments: envs,
		pods:         pods,
		deployments:  deployments,
	}
}

func (o *observeTab) tick() tea.Cmd {
	return tea.Tick(o.refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (o *observeTab) Init() tea.Cmd {
	return tea.Batch(o.loadResources, o.tick())
}

func (o *observeTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.width = msg.Width
		o.height = msg.Height

	case tickMsg:
		if o.autoRefresh {
			return o, tea.Batch(o.loadResources, o.tick())
		}
		return o, o.tick()

	case resourcesLoadedMsg:
		o.environments = msg.environments
		o.pods = msg.pods
		o.deployments = msg.deployments
		o.lastUpdate = time.Now()

		// If we have a selected environment, update it with fresh data
		if o.selectedEnvironment != nil {
			for i := range o.environments {
				if o.environments[i].Name == o.selectedEnvironment.Name {
					o.selectedEnvironment = &o.environments[i]
					break
				}
			}
		}

		// Reset selection if out of bounds
		maxIndex := o.getMaxIndex()
		if o.selectedIndex >= maxIndex {
			o.selectedIndex = maxIndex - 1
			if o.selectedIndex < 0 {
				o.selectedIndex = 0
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if o.panelFocus == focusTable {
				// Already on left, do nothing
			} else {
				// Cycle left through right panel views
				if o.rightPanelView > 0 {
					o.rightPanelView--
					o.scrollOffset = 0 // Reset scroll when changing views
				} else {
					// Go back to table when pressing left on Details
					o.panelFocus = focusTable
				}
			}
		case "right", "l":
			if o.panelFocus == focusTable {
				// Move focus to right panel
				o.panelFocus = focusRightPanel
			} else {
				// Cycle right through right panel views
				if o.rightPanelView < rightPanelStats {
					o.rightPanelView++
					o.scrollOffset = 0 // Reset scroll when changing views
				}
			}
		case "up", "k":
			if o.panelFocus == focusTable {
				if o.selectedIndex > 0 {
					o.selectedIndex--
				}
			} else {
				// Scroll up in right panel
				if o.scrollOffset > 0 {
					o.scrollOffset--
				}
			}
		case "down", "j":
			if o.panelFocus == focusTable {
				maxIndex := o.getMaxIndex()
				if o.selectedIndex < maxIndex-1 {
					o.selectedIndex++
				}
			} else {
				// Scroll down in right panel
				o.scrollOffset++
			}
		case "enter":
			// Drill down into environment (only when table focused)
			if o.panelFocus == focusTable && o.currentResource == resourceEnvironments && len(o.environments) > 0 {
				o.selectedEnvironment = &o.environments[o.selectedIndex]
				o.filterNamespace = o.selectedEnvironment.Namespace
				o.currentResource = resourcePods
				o.selectedIndex = 0
				return o, o.loadResources
			}
		case "esc", "backspace":
			// Go back to all environments
			if o.selectedEnvironment != nil {
				o.selectedEnvironment = nil
				o.filterNamespace = ""
				o.currentResource = resourceEnvironments
				o.selectedIndex = 0
				o.panelFocus = focusTable
				return o, o.loadResources
			}
		case "e":
			// Switch to environments view
			o.currentResource = resourceEnvironments
			o.selectedIndex = 0
			o.panelFocus = focusTable
		case "p":
			// Switch to pods view
			o.currentResource = resourcePods
			o.selectedIndex = 0
			o.panelFocus = focusTable
		case "d":
			// Switch to deployments view
			o.currentResource = resourceDeployments
			o.selectedIndex = 0
			o.panelFocus = focusTable
		case "r":
			// Manual refresh
			return o, o.loadResources
		case "a":
			// Toggle auto-refresh
			o.autoRefresh = !o.autoRefresh
		case "1":
			// Quick switch to Details view
			o.rightPanelView = rightPanelDetails
			o.panelFocus = focusRightPanel
		case "2":
			// Quick switch to Logs view
			o.rightPanelView = rightPanelLogs
			o.panelFocus = focusRightPanel
		case "3":
			// Quick switch to Events view
			o.rightPanelView = rightPanelEvents
			o.panelFocus = focusRightPanel
		case "4":
			// Quick switch to Stats view
			o.rightPanelView = rightPanelStats
			o.panelFocus = focusRightPanel
		}
	}

	return o, nil
}

func (o *observeTab) getMaxIndex() int {
	switch o.currentResource {
	case resourceEnvironments:
		return len(o.environments)
	case resourcePods:
		if o.selectedEnvironment != nil {
			return len(o.selectedEnvironment.Pods)
		}
		return len(o.pods)
	case resourceDeployments:
		if o.selectedEnvironment != nil {
			return len(o.selectedEnvironment.Deployments)
		}
		return len(o.deployments)
	default:
		return 0
	}
}

func (o *observeTab) getSelectedResource() interface{} {
	if o.getMaxIndex() == 0 || o.selectedIndex >= o.getMaxIndex() {
		return nil
	}

	switch o.currentResource {
	case resourceEnvironments:
		return o.environments[o.selectedIndex]
	case resourcePods:
		if o.selectedEnvironment != nil {
			return o.selectedEnvironment.Pods[o.selectedIndex]
		}
		return o.pods[o.selectedIndex]
	case resourceDeployments:
		if o.selectedEnvironment != nil {
			return o.selectedEnvironment.Deployments[o.selectedIndex]
		}
		return o.deployments[o.selectedIndex]
	}
	return nil
}

func (o *observeTab) View() string {
	if o.width == 0 {
		return "Loading..."
	}

	// Styles
	cyanColor := lipgloss.Color("51")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cyanColor).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	// Title color changes based on focus
	titleColor := lipgloss.Color("245") // Grey when not focused
	if o.panelFocus == focusTable {
		titleColor = cyanColor // Cyan when focused
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Padding(0, 0, 1, 0)

	tableHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("238")).
		Padding(0, 1)

	rowStyle := lipgloss.NewStyle().
		Padding(0, 1)

	// Cyan highlight for selected row
	selectedRowStyle := rowStyle.Copy().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("51")).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 0)

	// Header bar with breadcrumbs
	var resourceName string
	var breadcrumb string
	switch o.currentResource {
	case resourceEnvironments:
		resourceName = "Environments"
		breadcrumb = "Environments"
	case resourcePods:
		resourceName = "Pods"
		if o.selectedEnvironment != nil {
			breadcrumb = fmt.Sprintf("Environments > %s > Pods", o.selectedEnvironment.Name)
		} else {
			breadcrumb = "Pods"
		}
	case resourceDeployments:
		resourceName = "Deployments"
		if o.selectedEnvironment != nil {
			breadcrumb = fmt.Sprintf("Environments > %s > Deployments", o.selectedEnvironment.Name)
		} else {
			breadcrumb = "Deployments"
		}
	}

	autoRefreshIndicator := ""
	if o.autoRefresh {
		autoRefreshIndicator = " [AUTO]"
	}

	header := headerStyle.Render(fmt.Sprintf(" %s%s | Last update: %s ",
		breadcrumb,
		autoRefreshIndicator,
		o.lastUpdate.Format("15:04:05"),
	))

	// Build table
	var content strings.Builder
	content.WriteString(titleStyle.Render(resourceName))
	content.WriteString("\n\n")

	switch o.currentResource {
	case resourceEnvironments:
		content.WriteString(o.renderEnvironmentsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	case resourcePods:
		content.WriteString(o.renderPodsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	case resourceDeployments:
		content.WriteString(o.renderDeploymentsTable(tableHeaderStyle, rowStyle, selectedRowStyle))
	}

	// Right panel
	rightPanelHeight := o.height - 10
	rightPanelContent := o.renderRightPanel(rightPanelHeight)

	// Calculate widths for two-panel layout (50/50 split)
	tableWidth := o.width / 2         // 50% for table
	rightWidth := o.width - tableWidth // 50% for right panel

	// Reuse cyan color for highlights
	dimColor := lipgloss.Color("240")  // Dim gray

	// Style for focused/unfocused panels
	rightBorderColor := dimColor
	if o.panelFocus == focusRightPanel {
		rightBorderColor = cyanColor
	}

	// Table panel (no border)
	tablePanel := lipgloss.NewStyle().
		Width(tableWidth - 2).
		Height(o.height - 8).
		Padding(1).
		Render(content.String())

	// Right panel with full box border
	rightPanel := lipgloss.NewStyle().
		Width(rightWidth - 4).
		Height(o.height - 10).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Render(rightPanelContent)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, rightPanel)

	// Help text
	var helpText string
	if o.panelFocus == focusTable {
		helpText = "[→/l] Right Panel  [e/p/d] Views  [Enter] Drill-down  [↑↓/jk] Navigate  [1-4] Quick Switch  [r] Refresh  [q] Quit"
	} else {
		helpText = "[←/h] Back  [→←/hl] Cycle Views  [↑↓/jk] Scroll  [1] Details  [2] Logs  [3] Events  [4] Stats  [q] Quit"
	}
	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		mainContent,
		help,
	)
}

func (o *observeTab) renderEnvironmentsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	var table strings.Builder

	// Header
	nameCol := headerStyle.Width(30).Render("NAME")
	namespaceCol := headerStyle.Width(20).Render("NAMESPACE")
	statusCol := headerStyle.Width(15).Render("STATUS")
	ageCol := headerStyle.Width(15).Render("AGE")
	podsCol := headerStyle.Width(10).Render("PODS")

	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, statusCol, ageCol, podsCol))
	table.WriteString("\n")

	// Rows
	for i, env := range o.environments {
		style := rowStyle
		if i == o.selectedIndex {
			style = selectedStyle
		}

		age := formatAge(env.Age)
		podCount := fmt.Sprintf("%d", len(env.Pods))

		nameCol := style.Width(30).Render(truncate(env.Name, 28))
		namespaceCol := style.Width(20).Render(truncate(env.Namespace, 18))
		statusCol := style.Width(15).Render(env.Status)
		ageCol := style.Width(15).Render(age)
		podsCol := style.Width(10).Render(podCount)

		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, statusCol, ageCol, podsCol))
		table.WriteString("\n")
	}

	if len(o.environments) == 0 {
		table.WriteString(rowStyle.Render("No environments found"))
	}

	return table.String()
}

func (o *observeTab) renderPodsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	var table strings.Builder

	// Header
	nameCol := headerStyle.Width(30).Render("NAME")
	namespaceCol := headerStyle.Width(15).Render("NAMESPACE")
	readyCol := headerStyle.Width(8).Render("READY")
	statusCol := headerStyle.Width(12).Render("STATUS")
	restartsCol := headerStyle.Width(10).Render("RESTARTS")
	cpuCol := headerStyle.Width(10).Render("CPU")
	memoryCol := headerStyle.Width(10).Render("MEMORY")
	ageCol := headerStyle.Width(10).Render("AGE")

	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, statusCol, restartsCol, cpuCol, memoryCol, ageCol))
	table.WriteString("\n")

	// Determine which pods to display
	var pods []models.Pod
	if o.selectedEnvironment != nil {
		pods = o.selectedEnvironment.Pods
	} else {
		pods = o.pods
	}

	// Rows
	for i, pod := range pods {
		style := rowStyle
		if i == o.selectedIndex {
			style = selectedStyle
		}

		age := formatAge(pod.Age)

		nameCol := style.Width(30).Render(truncate(pod.Name, 28))
		namespaceCol := style.Width(15).Render(truncate(pod.Namespace, 13))
		readyCol := style.Width(8).Render(pod.Ready)
		statusCol := style.Width(12).Render(pod.Status)
		restartsCol := style.Width(10).Render(fmt.Sprintf("%d", pod.Restarts))
		cpuCol := style.Width(10).Render(pod.CPU)
		memoryCol := style.Width(10).Render(pod.Memory)
		ageCol := style.Width(10).Render(age)

		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, statusCol, restartsCol, cpuCol, memoryCol, ageCol))
		table.WriteString("\n")
	}

	if len(pods) == 0 {
		table.WriteString(rowStyle.Render("No pods found"))
	}

	return table.String()
}

func (o *observeTab) renderDeploymentsTable(headerStyle, rowStyle, selectedStyle lipgloss.Style) string {
	var table strings.Builder

	// Header
	nameCol := headerStyle.Width(35).Render("NAME")
	namespaceCol := headerStyle.Width(20).Render("NAMESPACE")
	readyCol := headerStyle.Width(10).Render("READY")
	upToDateCol := headerStyle.Width(12).Render("UP-TO-DATE")
	availableCol := headerStyle.Width(12).Render("AVAILABLE")
	ageCol := headerStyle.Width(15).Render("AGE")

	table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, upToDateCol, availableCol, ageCol))
	table.WriteString("\n")

	// Determine which deployments to display
	var deployments []models.Deployment
	if o.selectedEnvironment != nil {
		deployments = o.selectedEnvironment.Deployments
	} else {
		deployments = o.deployments
	}

	// Rows
	for i, deployment := range deployments {
		style := rowStyle
		if i == o.selectedIndex {
			style = selectedStyle
		}

		age := formatAge(deployment.Age)

		nameCol := style.Width(35).Render(truncate(deployment.Name, 33))
		namespaceCol := style.Width(20).Render(truncate(deployment.Namespace, 18))
		readyCol := style.Width(10).Render(deployment.Ready)
		upToDateCol := style.Width(12).Render(fmt.Sprintf("%d", deployment.UpToDate))
		availableCol := style.Width(12).Render(fmt.Sprintf("%d", deployment.Available))
		ageCol := style.Width(15).Render(age)

		table.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, nameCol, namespaceCol, readyCol, upToDateCol, availableCol, ageCol))
		table.WriteString("\n")
	}

	if len(deployments) == 0 {
		table.WriteString(rowStyle.Render("No deployments found"))
	}

	return table.String()
}

func formatAge(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (o *observeTab) renderRightPanel(height int) string {
	// View list items
	views := []struct {
		name string
		view rightPanelView
	}{
		{"Details", rightPanelDetails},
		{"Logs", rightPanelLogs},
		{"Events", rightPanelEvents},
		{"Stats", rightPanelStats},
	}

	// Styles for view list
	viewItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)

	selectedViewStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("86")).
		Bold(true).
		Padding(0, 1)

	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true).
		Padding(0, 1)

	// Build number row and view row separately
	var numberRow strings.Builder
	var viewRow strings.Builder

	for i, view := range views {
		// Add number
		number := fmt.Sprintf("[%d]", i+1)
		// Calculate width to match view name width
		viewNameWidth := lipgloss.Width(view.name) + 2 // +2 for padding
		numberRow.WriteString(numberStyle.Width(viewNameWidth).Align(lipgloss.Center).Render(number))
		if i < len(views)-1 {
			numberRow.WriteString(" ")
		}

		// Add view name
		style := viewItemStyle
		if view.view == o.rightPanelView {
			style = selectedViewStyle
		}
		viewRow.WriteString(style.Render(view.name))
		if i < len(views)-1 {
			viewRow.WriteString(" ")
		}
	}

	// Combine number row and view row
	var viewList strings.Builder
	viewList.WriteString(numberRow.String())
	viewList.WriteString("\n")
	viewList.WriteString(viewRow.String())

	// Use cyan for title when panel is focused
	titleColor := lipgloss.Color("245")
	if o.panelFocus == focusRightPanel {
		titleColor = lipgloss.Color("51") // Cyan
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor)

	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var content strings.Builder

	// Add view list
	content.WriteString(viewList.String())
	content.WriteString("\n")
	content.WriteString(separatorStyle.Render(strings.Repeat("─", 80)))
	content.WriteString("\n")

	// Add the current view name
	var panelName string
	switch o.rightPanelView {
	case rightPanelDetails:
		panelName = "Details"
	case rightPanelLogs:
		panelName = "Logs"
	case rightPanelEvents:
		panelName = "Events"
	case rightPanelStats:
		panelName = "Stats"
	}

	content.WriteString(titleStyle.Render(panelName))
	content.WriteString("\n")

	// Get the view content
	var viewContent string
	switch o.rightPanelView {
	case rightPanelDetails:
		viewContent = o.renderDetailsView()
	case rightPanelLogs:
		viewContent = o.renderLogsView()
	case rightPanelEvents:
		viewContent = o.renderEventsView()
	case rightPanelStats:
		viewContent = o.renderStatsView()
	}

	// Apply scrolling
	lines := strings.Split(viewContent, "\n")

	// Calculate available height for content (subtract numbers, view list, separator, title)
	availableHeight := height - 6

	// Ensure scroll offset is within bounds
	maxOffset := len(lines) - availableHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if o.scrollOffset > maxOffset {
		o.scrollOffset = maxOffset
	}

	// Get the visible lines
	endLine := o.scrollOffset + availableHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}

	visibleLines := lines[o.scrollOffset:endLine]
	content.WriteString(strings.Join(visibleLines, "\n"))

	return content.String()
}

func (o *observeTab) renderDetailsView() string {
	resource := o.getSelectedResource()
	if resource == nil {
		return "Select a resource to view details"
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Width(15)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255"))

	var details strings.Builder

	switch r := resource.(type) {
	case models.Environment:
		details.WriteString(labelStyle.Render("Name:") + " " + valueStyle.Render(r.Name) + "\n")
		details.WriteString(labelStyle.Render("Namespace:") + " " + valueStyle.Render(r.Namespace) + "\n")
		details.WriteString(labelStyle.Render("Status:") + " " + valueStyle.Render(r.Status) + "\n")
		details.WriteString(labelStyle.Render("Age:") + " " + valueStyle.Render(formatAge(r.Age)) + "\n")
		details.WriteString(labelStyle.Render("Pods:") + " " + valueStyle.Render(fmt.Sprintf("%d", len(r.Pods))) + "\n")
		details.WriteString(labelStyle.Render("Deployments:") + " " + valueStyle.Render(fmt.Sprintf("%d", len(r.Deployments))) + "\n")

	case models.Pod:
		details.WriteString(labelStyle.Render("Name:") + " " + valueStyle.Render(r.Name) + "\n")
		details.WriteString(labelStyle.Render("Namespace:") + " " + valueStyle.Render(r.Namespace) + "\n")
		details.WriteString(labelStyle.Render("Status:") + " " + valueStyle.Render(r.Status) + "\n")
		details.WriteString(labelStyle.Render("Ready:") + " " + valueStyle.Render(r.Ready) + "\n")
		details.WriteString(labelStyle.Render("Restarts:") + " " + valueStyle.Render(fmt.Sprintf("%d", r.Restarts)) + "\n")
		details.WriteString(labelStyle.Render("CPU:") + " " + valueStyle.Render(r.CPU) + "\n")
		details.WriteString(labelStyle.Render("Memory:") + " " + valueStyle.Render(r.Memory) + "\n")
		details.WriteString(labelStyle.Render("Age:") + " " + valueStyle.Render(formatAge(r.Age)) + "\n")

	case models.Deployment:
		details.WriteString(labelStyle.Render("Name:") + " " + valueStyle.Render(r.Name) + "\n")
		details.WriteString(labelStyle.Render("Namespace:") + " " + valueStyle.Render(r.Namespace) + "\n")
		details.WriteString(labelStyle.Render("Ready:") + " " + valueStyle.Render(r.Ready) + "\n")
		details.WriteString(labelStyle.Render("Up-to-Date:") + " " + valueStyle.Render(fmt.Sprintf("%d", r.UpToDate)) + "\n")
		details.WriteString(labelStyle.Render("Available:") + " " + valueStyle.Render(fmt.Sprintf("%d", r.Available)) + "\n")
		details.WriteString(labelStyle.Render("Age:") + " " + valueStyle.Render(formatAge(r.Age)) + "\n")
	}

	return details.String()
}

func (o *observeTab) renderLogsView() string {
	resource := o.getSelectedResource()
	if resource == nil {
		return "Select a pod to view logs"
	}

	pod, ok := resource.(models.Pod)
	if !ok {
		return "Logs are only available for pods"
	}

	// Mock logs for now
	logs := fmt.Sprintf("Fetching logs for %s...\n\n", pod.Name)
	logs += "[2025-10-22 20:30:15] INFO: Container started\n"
	logs += "[2025-10-22 20:30:16] INFO: Initializing application\n"
	logs += "[2025-10-22 20:30:17] INFO: Loading configuration\n"
	logs += "[2025-10-22 20:30:18] INFO: Connecting to database\n"
	logs += "[2025-10-22 20:30:20] INFO: Server listening on port 8080\n"
	logs += "[2025-10-22 20:30:25] INFO: Health check passed\n"
	logs += "[2025-10-22 20:31:00] INFO: Processing request GET /api/health\n"
	logs += "[2025-10-22 20:31:30] INFO: Processing request GET /api/status\n"

	return logs
}

func (o *observeTab) renderEventsView() string {
	resource := o.getSelectedResource()
	if resource == nil {
		return "Select a resource to view events"
	}

	// Color codes
	green := "\033[32m"
	yellow := "\033[33m"
	gray := "\033[38;5;245m"
	reset := "\033[0m"

	var events strings.Builder

	// Mock events - using ANSI codes directly to avoid lipgloss padding issues
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s2m ago  Successfully pulled image%s\n", green, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s5m ago  Created container%s\n", green, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s5m ago  Started container%s\n", green, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Warning %s%s1h ago  Back-off restarting failed container%s\n", yellow, reset, gray, reset))
	events.WriteString(fmt.Sprintf("%s● Normal  %s%s2h ago  Scaled deployment to 2 replicas%s\n", green, reset, gray, reset))

	return events.String()
}

func (o *observeTab) renderStatsView() string {
	statLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	statValueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)

	var stats strings.Builder

	switch o.currentResource {
	case resourceEnvironments:
		total := len(o.environments)
		stats.WriteString(statLabelStyle.Render("Total Environments: ") + statValueStyle.Render(fmt.Sprintf("%d", total)) + "\n\n")

		totalPods := 0
		totalDeployments := 0
		for _, env := range o.environments {
			totalPods += len(env.Pods)
			totalDeployments += len(env.Deployments)
		}
		stats.WriteString(statLabelStyle.Render("Total Pods: ") + statValueStyle.Render(fmt.Sprintf("%d", totalPods)) + "\n")
		stats.WriteString(statLabelStyle.Render("Total Deployments: ") + statValueStyle.Render(fmt.Sprintf("%d", totalDeployments)) + "\n")

	case resourcePods:
		var pods []models.Pod
		if o.selectedEnvironment != nil {
			pods = o.selectedEnvironment.Pods
		} else {
			pods = o.pods
		}

		total := len(pods)
		running := 0
		pending := 0
		failed := 0

		for _, pod := range pods {
			switch pod.Status {
			case "Running":
				running++
			case "Pending":
				pending++
			case "Failed":
				failed++
			}
		}

		stats.WriteString(statLabelStyle.Render("Total Pods: ") + statValueStyle.Render(fmt.Sprintf("%d", total)) + "\n\n")
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Render("● Running: ") + fmt.Sprintf("%d\n", running))
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Render("● Pending: ") + fmt.Sprintf("%d\n", pending))
		stats.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Render("● Failed: ") + fmt.Sprintf("%d\n", failed))

	case resourceDeployments:
		var deployments []models.Deployment
		if o.selectedEnvironment != nil {
			deployments = o.selectedEnvironment.Deployments
		} else {
			deployments = o.deployments
		}

		total := len(deployments)
		totalReplicas := 0
		totalAvailable := 0

		for _, dep := range deployments {
			totalReplicas += dep.UpToDate
			totalAvailable += dep.Available
		}

		stats.WriteString(statLabelStyle.Render("Total Deployments: ") + statValueStyle.Render(fmt.Sprintf("%d", total)) + "\n\n")
		stats.WriteString(statLabelStyle.Render("Total Replicas: ") + statValueStyle.Render(fmt.Sprintf("%d", totalReplicas)) + "\n")
		stats.WriteString(statLabelStyle.Render("Available: ") + statValueStyle.Render(fmt.Sprintf("%d", totalAvailable)) + "\n")
	}

	return stats.String()
}
