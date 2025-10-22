# Imperm - Kubernetes Environment Manager

A k9s-inspired TUI (Terminal User Interface) for managing Kubernetes environments through a middleware layer. Built with [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Two-tab interface**: Control and Observe
- **Control Tab**: Build and destroy environments with a visual history
- **Observe Tab**: k9s-style view of environments, pods, and deployments with auto-refresh
- **Dual-Panel Layout**: Table on the left (60%), multi-view panel on the right (40%)
- **Right Panel Views**:
  - **Details**: Detailed information about selected resource
  - **Logs**: Real-time logs for pods
  - **Events**: Kubernetes events timeline
  - **Stats**: Summary statistics and health status
- **Resource Metrics**: View CPU and Memory usage for pods
- **Drill-down Navigation**: Press Enter on an environment to view its resources with breadcrumb navigation
- **Multiple Resource Views**: Toggle between Environments, Pods, and Deployments
- **Focus-based Navigation**: Navigate between panels with arrow keys, highlighted borders show focus
- **Mock Mode**: Test the UI without connecting to a real cluster
- **Middleware Architecture**: Designed to work with a middleware API layer instead of direct kubectl access

## Installation

```bash
go build -o imperm ./cmd/imperm
```

## Usage

### Mock Mode (Development)

```bash
./imperm --mock
```

### Navigation

#### Global
- `Tab` - Switch between Control and Observe tabs
- `q` or `Ctrl+C` - Quit

#### Control Tab
- `↑/↓` or `k/j` - Navigate actions
- `Enter` - Select action
- `Esc` - Cancel input mode

#### Observe Tab
- `←→` or `h/l` - Switch between table and right panel
- `↑/↓` or `k/j` - Navigate (table: rows, right panel: views)
- `Enter` - Drill down into selected environment
- `Esc` - Go back to all environments
- `e` - View Environments
- `p` - View Pods
- `d` - View Deployments
- `1-4` - Quick switch to right panel views (Details/Logs/Events/Stats)
- `r` - Manual refresh
- `a` - Toggle auto-refresh

#### Right Panel Views
- **Details** (`1`) - Show detailed information about selected resource
- **Logs** (`2`) - View logs for selected pod (pods only)
- **Events** (`3`) - Show Kubernetes events for selected resource
- **Stats** (`4`) - Display summary statistics for current view

## Project Structure

```
imperm/
├── cmd/
│   └── imperm/
│       └── main.go           # Application entry point
├── internal/
│   ├── middleware/
│   │   ├── client.go         # Middleware client interface
│   │   └── mock.go           # Mock implementation
│   ├── models/
│   │   └── environment.go    # Data models
│   └── ui/
│       ├── app.go            # Main application model
│       ├── control.go        # Control tab
│       └── observe.go        # Observe tab (k9s-style)
├── go.mod
├── go.sum
└── README.md
```

## Architecture

### Middleware Client Interface

The `middleware.Client` interface defines the contract for interacting with Kubernetes:

```go
type Client interface {
    ListEnvironments() ([]models.Environment, error)
    CreateEnvironment(name string, withOptions bool) error
    DestroyEnvironment(name string) error
    ListPods(namespace string) ([]models.Pod, error)
    GetPodLogs(namespace, podName string) (string, error)
    GetEnvironmentHistory() ([]models.EnvironmentHistory, error)
}
```

### Implementing a Real Middleware Client

To connect to your actual middleware API, create a new implementation of the `middleware.Client` interface:

```go
type HTTPClient struct {
    baseURL string
    // ... your HTTP client fields
}

func NewHTTPClient(baseURL string) *HTTPClient {
    return &HTTPClient{
        baseURL: baseURL,
    }
}

// Implement all Client interface methods
func (c *HTTPClient) ListEnvironments() ([]models.Environment, error) {
    // Make HTTP request to your middleware API
    // Parse response into []models.Environment
}

// ... implement other methods
```

Then update `cmd/imperm/main.go` to use your client:

```go
if *mockMode {
    client = middleware.NewMockClient()
} else {
    client = middleware.NewHTTPClient("https://your-middleware-api.com")
}
```

## Future Enhancements

- Add support for real middleware API endpoints
- Implement pod log viewing
- Add filtering and searching capabilities
- Support for additional Kubernetes resources (services, deployments, etc.)
- Configuration file support for middleware endpoint
- Export environment history to file
- Custom themes with Lipgloss

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for TUIs
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

MIT
