package config

import "time"

// Timing constants
const (
	// StatusMessageTimeout is how long status messages are displayed before auto-clearing
	StatusMessageTimeout = 3 * time.Second

	// LogPollingInterval is how often to poll for operation logs
	LogPollingInterval = 1 * time.Second // Reduced from 500ms to reduce CPU usage

	// ResourceRefreshInterval is how often to refresh resource lists in observe tab
	ResourceRefreshInterval = 10 * time.Second // Increased from 5s to reduce API calls and CPU usage
)

// Layout constants
const (
	// PanelHeightOffset is the vertical space reserved for headers/footers
	PanelHeightOffset = 8

	// PanelPadding is the standard padding for panels
	PanelPadding = 2

	// SplitPanelRatio divides the screen into left/right panels
	SplitPanelRatio = 2 // width / 2

	// ContentHeightOffset is the vertical space for calculating log content area
	ContentHeightOffset = 12

	// MinLogLines is the minimum number of log lines to display
	MinLogLines = 5

	// LogWidthAdjustment is the width reduction for log content
	LogWidthAdjustment = 6

	// CategoryBoxWidth is the standard width for category selection boxes
	CategoryBoxWidth = 40

	// FormLabelWidth is the standard width for form labels
	FormLabelWidth = 25

	// ScrollToBottom is a large value used to force scrolling to the bottom of content
	ScrollToBottom = 999999
)
