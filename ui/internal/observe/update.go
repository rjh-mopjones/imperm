package observe

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"imperm-ui/pkg/models"
)

func (t *Tab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height

	case tickMsg:
		if t.autoRefresh {
			// Reload resources and also reload current view data (logs, events, etc.)
			return t, tea.Batch(t.loadResources, t.loadDataForCurrentView(), t.tick())
		}
		return t, t.tick()

	case resourcesLoadedMsg:
		t.environments = msg.environments
		t.pods = msg.pods
		t.deployments = msg.deployments
		t.lastUpdate = time.Now()
		t.lastError = nil // Clear any previous errors
		t.isLoading = false // Data loaded successfully

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

		// Load stats and current view data after resources are loaded
		return t, tea.Batch(t.loadStats(), t.loadDataForCurrentView())

	case logsLoadedMsg:
		// If the pod name changed, reset scroll to bottom for new pod
		if t.lastPodName != "" && t.lastPodName != msg.podName {
			t.scrollOffset = 999999 // Will be clamped to max in render
		} else if t.lastPodName == msg.podName && t.currentLogs != "" {
			// Same pod, check if we were at/near the bottom
			oldLines := len(strings.Split(t.currentLogs, "\n"))
			// If we were within 5 lines of the bottom, auto-scroll to new bottom
			availableHeight := t.height - 16 // Approximate content height
			maxOldOffset := oldLines - availableHeight
			if maxOldOffset < 0 {
				maxOldOffset = 0
			}
			if t.scrollOffset >= maxOldOffset-5 {
				// User was at/near bottom, scroll to new bottom
				t.scrollOffset = 999999 // Will be clamped to max in render
			}
		} else {
			// First load, scroll to bottom
			t.scrollOffset = 999999 // Will be clamped to max in render
		}
		t.currentLogs = msg.logs
		t.lastPodName = msg.podName

	case eventsLoadedMsg:
		t.currentEvents = msg.events

	case statsLoadedMsg:
		t.currentStats = msg.stats

	case errMsg:
		// Display error as a status message
		t.statusMessage = fmt.Sprintf("❌ Error: %v", msg.err)
		t.statusType = "error"
		t.statusTime = time.Now()
		t.lastError = nil // Don't show full-screen error
		t.isLoading = false // Stop loading on error
		return t, t.clearStatusAfterDelay()

	case clearStatusMsg:
		// Clear the status message
		t.statusMessage = ""
		return t, nil

	case resourceDeletedMsg:
		// Success message already shown immediately, just reload resources
		return t, t.loadResources

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
					// Reload data for the newly selected resource
					return t, t.loadDataForCurrentView()
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
					// Reload data for the newly selected resource
					return t, t.loadDataForCurrentView()
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
			return t, t.loadDataForCurrentView()
		case "p":
			// Switch to pods view
			t.currentResource = ResourcePods
			t.selectedIndex = 0
			t.panelFocus = FocusTable
			return t, t.loadDataForCurrentView()
		case "d":
			// Switch to deployments view
			t.currentResource = ResourceDeployments
			t.selectedIndex = 0
			t.panelFocus = FocusTable
			return t, t.loadDataForCurrentView()
		case "r":
			// Manual refresh - also clears errors
			t.lastError = nil
			t.isLoading = true
			return t, t.loadResources
		case "a":
			// Toggle auto-refresh
			t.autoRefresh = !t.autoRefresh
		case "x":
			// Delete selected resource
			if t.panelFocus == FocusTable {
				// Get the resource info to show message immediately
				var resourceName, resourceType string
				switch t.currentResource {
				case ResourceEnvironments:
					if t.selectedIndex < len(t.environments) {
						resourceName = t.environments[t.selectedIndex].Name
						resourceType = "environment"
					}
				case ResourcePods:
					var pods []models.Pod
					if t.selectedEnvironment != nil {
						pods = t.selectedEnvironment.Pods
					} else {
						pods = t.pods
					}
					if t.selectedIndex < len(pods) {
						resourceName = pods[t.selectedIndex].Name
						resourceType = "pod"
					}
				case ResourceDeployments:
					var deployments []models.Deployment
					if t.selectedEnvironment != nil {
						deployments = t.selectedEnvironment.Deployments
					} else {
						deployments = t.deployments
					}
					if t.selectedIndex < len(deployments) {
						resourceName = deployments[t.selectedIndex].Name
						resourceType = "deployment"
					}
				}

				// Show success message immediately
				if resourceName != "" {
					t.statusMessage = fmt.Sprintf("✓ Deleted %s: %s", resourceType, resourceName)
					t.statusType = "success"
					t.statusTime = time.Now()
				}

				return t, tea.Batch(t.deleteSelectedResource(), t.clearStatusAfterDelay())
			}
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
