package observe

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"imperm-ui/pkg/models"
)

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
			return logsLoadedMsg{logs: "", podName: ""}
		}

		pod, ok := resource.(models.Pod)
		if !ok {
			return logsLoadedMsg{logs: "Logs are only available for pods", podName: ""}
		}

		logs, err := t.client.GetPodLogs(pod.Namespace, pod.Name)
		if err != nil {
			return errMsg{err}
		}

		return logsLoadedMsg{logs: logs, podName: pod.Name}
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

func (t *Tab) deleteSelectedResource() tea.Cmd {
	return func() tea.Msg {
		// Perform deletion asynchronously
		go func() {
			switch t.currentResource {
			case ResourceEnvironments:
				if t.selectedIndex < len(t.environments) {
					env := t.environments[t.selectedIndex]
					_ = t.client.DestroyEnvironment(env.Name)
				}
			case ResourcePods:
				var pods []models.Pod
				if t.selectedEnvironment != nil {
					pods = t.selectedEnvironment.Pods
				} else {
					pods = t.pods
				}
				if t.selectedIndex < len(pods) {
					pod := pods[t.selectedIndex]
					_ = t.client.DeletePod(pod.Namespace, pod.Name)
				}
			case ResourceDeployments:
				var deployments []models.Deployment
				if t.selectedEnvironment != nil {
					deployments = t.selectedEnvironment.Deployments
				} else {
					deployments = t.deployments
				}
				if t.selectedIndex < len(deployments) {
					dep := deployments[t.selectedIndex]
					_ = t.client.DeleteDeployment(dep.Namespace, dep.Name)
				}
			}
		}()

		// Return message to trigger resource reload
		return resourceDeletedMsg{}
	}
}

func (t *Tab) clearStatusAfterDelay() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (t *Tab) tick() tea.Cmd {
	return tea.Tick(t.refreshInterval, func(tm time.Time) tea.Msg {
		return tickMsg(tm)
	})
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
