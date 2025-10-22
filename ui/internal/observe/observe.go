package observe

import (
	"imperm-ui/pkg/client"
	"imperm-ui/pkg/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type resourceType int

const (
	ResourceEnvironments resourceType = iota
	ResourcePods
	ResourceDeployments
)

type panelFocus int

const (
	FocusTable panelFocus = iota
	FocusRightPanel
)

type rightPanelView int

const (
	RightPanelDetails rightPanelView = iota
	RightPanelLogs
	RightPanelEvents
	RightPanelStats
)

type Tab struct {
	client           client.Client
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

	// Right panel data
	currentLogs   string
	currentEvents []models.Event
	currentStats  *models.ResourceStats
}

func NewTab(client client.Client) *Tab {
	return &Tab{
		client:          client,
		currentResource: ResourceEnvironments,
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

type logsLoadedMsg struct {
	logs string
}

type eventsLoadedMsg struct {
	events []models.Event
}

type statsLoadedMsg struct {
	stats *models.ResourceStats
}

func (t *Tab) loadResources() tea.Msg {
	envs, err := t.client.ListEnvironments()
	if err != nil {
		return errMsg{err}
	}

	pods, err := t.client.ListPods(t.filterNamespace)
	if err != nil {
		return errMsg{err}
	}

	deployments, err := t.client.ListDeployments(t.filterNamespace)
	if err != nil {
		return errMsg{err}
	}

	return resourcesLoadedMsg{
		environments: envs,
		pods:         pods,
		deployments:  deployments,
	}
}

func (t *Tab) loadLogs() tea.Cmd {
	return func() tea.Msg {
		resource := t.getSelectedResource()
		if resource == nil {
			return logsLoadedMsg{logs: ""}
		}

		pod, ok := resource.(models.Pod)
		if !ok {
			return logsLoadedMsg{logs: "Logs are only available for pods"}
		}

		logs, err := t.client.GetPodLogs(pod.Namespace, pod.Name)
		if err != nil {
			return errMsg{err}
		}

		return logsLoadedMsg{logs: logs}
	}
}

func (t *Tab) loadEvents() tea.Cmd {
	return func() tea.Msg {
		resource := t.getSelectedResource()
		if resource == nil {
			return eventsLoadedMsg{events: []models.Event{}}
		}

		var events []models.Event
		var err error

		switch r := resource.(type) {
		case models.Pod:
			events, err = t.client.GetPodEvents(r.Namespace, r.Name)
		case models.Deployment:
			events, err = t.client.GetDeploymentEvents(r.Namespace, r.Name)
		case models.Environment:
			// For environments, we could aggregate events from all pods/deployments
			// For now, return empty
			events = []models.Event{}
		}

		if err != nil {
			return errMsg{err}
		}

		return eventsLoadedMsg{events: events}
	}
}

func (t *Tab) loadStats() tea.Cmd {
	return func() tea.Msg {
		var resourceType string
		switch t.currentResource {
		case ResourceEnvironments:
			resourceType = "environments"
		case ResourcePods:
			resourceType = "pods"
		case ResourceDeployments:
			resourceType = "deployments"
		}

		stats, err := t.client.GetResourceStats(resourceType, t.filterNamespace)
		if err != nil {
			return errMsg{err}
		}

		return statsLoadedMsg{stats: stats}
	}
}

func (t *Tab) tick() tea.Cmd {
	return tea.Tick(t.refreshInterval, func(tm time.Time) tea.Msg {
		return tickMsg(tm)
	})
}

func (t *Tab) Init() tea.Cmd {
	return tea.Batch(t.loadResources, t.tick())
}

func (t *Tab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

	case tickMsg:
		if t.autoRefresh {
			return t, tea.Batch(t.loadResources, t.tick())
		}
		return t, t.tick()

	case resourcesLoadedMsg:
		t.environments = msg.environments
		t.pods = msg.pods
		t.deployments = msg.deployments
		t.lastUpdate = time.Now()

		// If we have a selected environment, update it with fresh data
		if t.selectedEnvironment != nil {
			for i := range t.environments {
				if t.environments[i].Name == t.selectedEnvironment.Name {
					t.selectedEnvironment = &t.environments[i]
					break
				}
			}
		}

		// Reset selection if out of bounds
		maxIndex := t.getMaxIndex()
		if t.selectedIndex >= maxIndex {
			t.selectedIndex = maxIndex - 1
			if t.selectedIndex < 0 {
				t.selectedIndex = 0
			}
		}

		// Load stats after resources are loaded
		return t, t.loadStats()

	case logsLoadedMsg:
		t.currentLogs = msg.logs

	case eventsLoadedMsg:
		t.currentEvents = msg.events

	case statsLoadedMsg:
		t.currentStats = msg.stats

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if t.panelFocus == FocusTable {
				// Already on left, do nothing
			} else {
				// Cycle left through right panel views
				if t.rightPanelView > 0 {
					t.rightPanelView--
					t.scrollOffset = 0 // Reset scroll when changing views
				} else {
					// Go back to table when pressing left on Details
					t.panelFocus = FocusTable
				}
			}
		case "right", "l":
			if t.panelFocus == FocusTable {
				// Move focus to right panel
				t.panelFocus = FocusRightPanel
				// Load data for the current view
				return t, t.loadDataForCurrentView()
			} else {
				// Cycle right through right panel views
				if t.rightPanelView < RightPanelStats {
					t.rightPanelView++
					t.scrollOffset = 0 // Reset scroll when changing views
					return t, t.loadDataForCurrentView()
				}
			}
		case "up", "k":
			if t.panelFocus == FocusTable {
				if t.selectedIndex > 0 {
					t.selectedIndex--
				}
			} else {
				// Scroll up in right panel
				if t.scrollOffset > 0 {
					t.scrollOffset--
				}
			}
		case "down", "j":
			if t.panelFocus == FocusTable {
				maxIndex := t.getMaxIndex()
				if t.selectedIndex < maxIndex-1 {
					t.selectedIndex++
				}
			} else {
				// Scroll down in right panel
				t.scrollOffset++
			}
		case "enter":
			// Drill down into environment (only when table focused)
			if t.panelFocus == FocusTable && t.currentResource == ResourceEnvironments && len(t.environments) > 0 {
				t.selectedEnvironment = &t.environments[t.selectedIndex]
				t.filterNamespace = t.selectedEnvironment.Namespace
				t.currentResource = ResourcePods
				t.selectedIndex = 0
				return t, t.loadResources
			}
		case "esc", "backspace":
			// Go back to all environments
			if t.selectedEnvironment != nil {
				t.selectedEnvironment = nil
				t.filterNamespace = ""
				t.currentResource = ResourceEnvironments
				t.selectedIndex = 0
				t.panelFocus = FocusTable
				return t, t.loadResources
			}
		case "e":
			// Switch to environments view
			t.currentResource = ResourceEnvironments
			t.selectedIndex = 0
			t.panelFocus = FocusTable
		case "p":
			// Switch to pods view
			t.currentResource = ResourcePods
			t.selectedIndex = 0
			t.panelFocus = FocusTable
		case "d":
			// Switch to deployments view
			t.currentResource = ResourceDeployments
			t.selectedIndex = 0
			t.panelFocus = FocusTable
		case "r":
			// Manual refresh
			return t, t.loadResources
		case "a":
			// Toggle auto-refresh
			t.autoRefresh = !t.autoRefresh
		case "1":
			// Quick switch to Details view (without changing focus)
			t.rightPanelView = RightPanelDetails
			t.scrollOffset = 0
			return t, t.loadDataForCurrentView()
		case "2":
			// Quick switch to Logs view (without changing focus)
			t.rightPanelView = RightPanelLogs
			t.scrollOffset = 0
			return t, t.loadDataForCurrentView()
		case "3":
			// Quick switch to Events view (without changing focus)
			t.rightPanelView = RightPanelEvents
			t.scrollOffset = 0
			return t, t.loadDataForCurrentView()
		case "4":
			// Quick switch to Stats view (without changing focus)
			t.rightPanelView = RightPanelStats
			t.scrollOffset = 0
			return t, t.loadDataForCurrentView()
		}
	}

	return t, nil
}

func (t *Tab) getMaxIndex() int {
	switch t.currentResource {
	case ResourceEnvironments:
		return len(t.environments)
	case ResourcePods:
		if t.selectedEnvironment != nil {
			return len(t.selectedEnvironment.Pods)
		}
		return len(t.pods)
	case ResourceDeployments:
		if t.selectedEnvironment != nil {
			return len(t.selectedEnvironment.Deployments)
		}
		return len(t.deployments)
	default:
		return 0
	}
}

func (t *Tab) getSelectedResource() interface{} {
	if t.getMaxIndex() == 0 || t.selectedIndex >= t.getMaxIndex() {
		return nil
	}

	switch t.currentResource {
	case ResourceEnvironments:
		return t.environments[t.selectedIndex]
	case ResourcePods:
		if t.selectedEnvironment != nil {
			return t.selectedEnvironment.Pods[t.selectedIndex]
		}
		return t.pods[t.selectedIndex]
	case ResourceDeployments:
		if t.selectedEnvironment != nil {
			return t.selectedEnvironment.Deployments[t.selectedIndex]
		}
		return t.deployments[t.selectedIndex]
	}
	return nil
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

func (t *Tab) loadDataForCurrentView() tea.Cmd {
	switch t.rightPanelView {
	case RightPanelLogs:
		return t.loadLogs()
	case RightPanelEvents:
		return t.loadEvents()
	case RightPanelStats:
		return t.loadStats()
	default:
		// Details view doesn't need to load data
		return nil
	}
}
