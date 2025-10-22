# Imperm

Kubernetes environment management with a beautiful terminal UI.

## Architecture

This is a **monorepo** containing two independent Go modules:

```
imperm/
├── ui/              # Terminal UI client (imperm-ui)
│   └── go.mod       # Separate Go module
│
└── middleware/      # HTTP API server (imperm-server)
    └── go.mod       # Separate Go module
```

### Why Monorepo?

- **Single checkout**: Clone once, get everything
- **Coordinated development**: Make changes across both in one PR
- **Shared code**: `pkg/` directory contains shared client and models
- **Independent modules**: Each has its own dependencies and versioning

## Quick Start

### Prerequisites

- Go 1.23+
- Make

### Build

```bash
# Build both applications
make all

# Or build individually
make ui
make server
```

Binaries will be in `bin/`:
- `bin/imperm-ui` - Terminal UI
- `bin/imperm-server` - HTTP API server

### Run

**Option 1: Standalone UI (Mock Mode)**

```bash
make run-ui
```

**Option 2: Client-Server Mode**

Terminal 1 - Start server:
```bash
make run-server
```

Terminal 2 - Start UI connected to server:
```bash
make run-ui-remote
```

## Project Structure

### UI (`ui/`)

Terminal user interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

```
ui/
├── cmd/           # Main entry point
├── internal/      # UI-specific code
│   ├── app.go
│   ├── control/   # Control tab (create/destroy environments)
│   └── observe/   # Observe tab (view pods, deployments)
└── pkg/           # Shared code (client, models)
```

**Features**:
- Two-panel layout with table and detail views
- Multiple right-panel views (Details, Logs, Events, Stats)
- Keyboard navigation with vim-style keybindings
- Real-time auto-refresh
- Can run standalone or connect to server

### Middleware (`middleware/`)

HTTP API server for Kubernetes management.

```
middleware/
├── cmd/           # Main entry point
├── internal/      # Server-specific code
│   ├── api/       # HTTP handlers
│   ├── k8s/       # Kubernetes client (TODO)
│   └── store/     # State management (TODO)
└── pkg/           # Shared code (client, models)
```

**API Endpoints**:
- `GET /api/environments`
- `POST /api/environments/create`
- `POST /api/environments/destroy`
- `GET /api/environments/history`
- `GET /api/pods?namespace=X`
- `GET /api/deployments?namespace=X`
- `GET /health`

## Development

### Directory Conventions

- **`cmd/`** - Application entry points (`package main`)
- **`internal/`** - Private code (Go prevents external imports)
- **`pkg/`** - Public libraries (can be imported by other projects)

### Working on UI

```bash
cd ui
go run ./cmd --mock
```

### Working on Server

```bash
cd middleware
go run ./cmd --mock --port 8080
```

### Running Tests

```bash
make test          # Run all tests
make test-ui       # UI tests only
make test-server   # Server tests only
```

### Managing Dependencies

```bash
make tidy          # Tidy both modules
```

Or manually:
```bash
cd ui && go mod tidy
cd middleware && go mod tidy
```

## Next Steps

1. **Kubernetes Integration**: Implement `middleware/internal/k8s/client.go`
2. **Real K8s Connection**: Update server to connect to actual Kubernetes clusters
3. **Authentication**: Add auth to server API
4. **Persistence**: Implement `middleware/internal/store/` for historical data
5. **Real-time Updates**: Add WebSocket support for live updates

## Commands Reference

```bash
make help          # Show all available commands
make all           # Build both applications
make clean         # Remove build artifacts
make run-ui        # Run UI in mock mode
make run-server    # Run server in mock mode
make run-ui-remote # Run UI connected to server
```

## License

[Your License Here]
