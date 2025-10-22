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
		case "up", "k":
			if o.selectedIndex > 0 {
				o.selectedIndex--
			}
		case "down", "j":
			maxIndex := o.getMaxIndex()
			if o.selectedIndex < maxIndex-1 {
				o.selectedIndex++
			}
		case "enter":
			// Drill down into environment
			if o.currentResource == resourceEnvironments && len(o.environments) > 0 {
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
				return o, o.loadResources
			}
		case "e":
			// Switch to environments view
			o.currentResource = resourceEnvironments
			o.selectedIndex = 0
		case "p":
			// Switch to pods view
			o.currentResource = resourcePods
			o.selectedIndex = 0
		case "d":
			// Switch to deployments view
			o.currentResource = resourceDeployments
			o.selectedIndex = 0
		case "r":
			// Manual refresh
			return o, o.loadResources
		case "a":
			// Toggle auto-refresh
			o.autoRefresh = !o.autoRefresh
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

func (o *observeTab) View() string {
	if o.width == 0 {
		return "Loading..."
	}

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	tableHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("238")).
		Padding(0, 1)

	rowStyle := lipgloss.NewStyle().
		Padding(0, 1)

	selectedRowStyle := rowStyle.Copy().
		Background(lipgloss.Color("237"))

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

	// Help text
	var helpText string
	if o.selectedEnvironment != nil {
		helpText = "[e] Environments  [p] Pods  [d] Deployments  [esc] Back  [↑↓/jk] Navigate  [r] Refresh  [a] Auto-refresh  [q] Quit"
	} else {
		helpText = "[e] Environments  [p] Pods  [d] Deployments  [Enter] Drill-down  [↑↓/jk] Navigate  [r] Refresh  [a] Auto-refresh  [q] Quit"
	}
	help := helpStyle.Render(helpText)

	// Combine all parts
	mainContent := lipgloss.NewStyle().
		Width(o.width - 4).
		Height(o.height - 8).
		Padding(1).
		Render(content.String())

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
