package observe

import (
	tea "github.com/charmbracelet/bubbletea"
	"imperm-ui/internal/config"
	"imperm-ui/pkg/client"
	"imperm-ui/pkg/models"
)

func NewTab(client client.Client) *Tab {
	return &Tab{
		client:          client,
		currentResource: ResourceEnvironments,
		environments:    []models.Environment{},
		pods:            []models.Pod{},
		deployments:     []models.Deployment{},
		selectedIndex:   0,
		autoRefresh:     true,
		refreshInterval: config.ResourceRefreshInterval,
		isLoading:       true, // Start in loading state
		rightPanelView:  RightPanelLogs, // Default to Logs view for auto-updating
	}
}

func (t *Tab) Init() tea.Cmd {
	return tea.Batch(t.loadResources, t.tick())
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

// getCurrentPods returns the currently visible pods (either from selected environment or all pods)
func (t *Tab) getCurrentPods() []models.Pod {
	if t.selectedEnvironment != nil {
		return t.selectedEnvironment.Pods
	}
	return t.pods
}

// getCurrentDeployments returns the currently visible deployments (either from selected environment or all deployments)
func (t *Tab) getCurrentDeployments() []models.Deployment {
	if t.selectedEnvironment != nil {
		return t.selectedEnvironment.Deployments
	}
	return t.deployments
}

func (t *Tab) GetHeaderInfo() string {
	var resourceName string
	switch t.currentResource {
	case ResourceEnvironments:
		resourceName = "Environments"
	case ResourcePods:
		if t.selectedEnvironment != nil {
			resourceName = t.selectedEnvironment.Name + " > Pods"
		} else {
			resourceName = "Pods"
		}
	case ResourceDeployments:
		if t.selectedEnvironment != nil {
			resourceName = t.selectedEnvironment.Name + " > Deployments"
		} else {
			resourceName = "Deployments"
		}
	}

	autoRefreshIndicator := ""
	if t.autoRefresh {
		autoRefreshIndicator = " [AUTO]"
	}

	return resourceName + autoRefreshIndicator + " | Last update: " + t.lastUpdate.Format("15:04:05")
}
