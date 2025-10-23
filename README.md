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

**Option 3: Terraform Mode (DEFAULT)**

Terraform mode is now the default! The server uses Terraform modules to provision Kubernetes namespaces and resources, providing infrastructure-as-code capabilities with state management.

```bash
# Run server (defaults to Terraform mode)
make run-server

# Or use other modes
make run-server-mock    # Mock mode
make run-server-k8s     # Direct K8s API mode
```

**Live Terraform Logs**: When creating or destroying environments from the UI, Terraform provisioning logs are streamed in real-time to the right panel of the control screen!

## Terraform Integration

### Overview

The middleware can use Terraform to provision Kubernetes resources instead of directly using the K8s API. This provides:
- **Infrastructure as Code**: All resources defined in Terraform configuration
- **State Management**: Terraform tracks resource state
- **Reproducibility**: Environments can be recreated from configuration
- **Version Control**: Terraform configs can be versioned

### Structure

```
terraform/
├── modules/           # Reusable Terraform modules
│   └── k8s-namespace/ # Module for creating K8s namespaces with resources
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── README.md
└── environments/      # Generated configs for each environment
    └── <env-name>/    # Created dynamically by middleware
```

### Prerequisites for Terraform Mode

1. Install Terraform:
```bash
brew install terraform
```

2. Ensure you have a valid kubeconfig:
```bash
export KUBECONFIG=~/.kube/config
# or set it in your environment
```

3. Verify Terraform installation:
```bash
terraform version
```

### How It Works

1. When you create an environment via the API, the middleware:
   - Generates a Terraform configuration in `terraform/environments/<env-name>/`
   - Runs `terraform init` to initialize the workspace
   - Runs `terraform apply` to create the resources

2. When you destroy an environment:
   - Runs `terraform destroy` to remove all resources
   - Cleans up the environment directory

### Running Modes

The middleware supports three modes:

| Mode | Flag | Description |
|------|------|-------------|
| Mock | `--mock` | Simulated data, no K8s connection |
| K8s | (default) | Direct Kubernetes API calls |
| Terraform | `--terraform` | Provision resources via Terraform |

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
│   ├── k8s/       # Kubernetes client
│   ├── terraform/ # Terraform client
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
