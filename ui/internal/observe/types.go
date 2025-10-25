package observe

import (
	"time"

	"imperm-ui/pkg/client"
	"imperm-ui/pkg/models"
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
	currentLogs         string
	currentEvents       []models.Event
	currentStats        *models.ResourceStats
	lastPodName         string // Track last pod name for logs refresh
	lastDeploymentName  string // Track last deployment name for events refresh

	// Error tracking
	lastError error

	// Status message
	statusMessage string
	statusTime    time.Time
	statusType    string // "success" or "error"

	// Loading state
	isLoading bool
}

// Messages
type resourcesLoadedMsg struct {
	environments []models.Environment
	pods         []models.Pod
	deployments  []models.Deployment
}

type logsLoadedMsg struct {
	logs    string
	podName string
}

type eventsLoadedMsg struct {
	events []models.Event
}

type statsLoadedMsg struct {
	stats *models.ResourceStats
}

type resourceDeletedMsg struct{}
