# Imperm Architecture

## Project Structure

Imperm is split into two completely separate applications with clear boundaries:

```
imperm/
├── ui/                    # UI Application (Terminal Client)
│   ├── go.mod            # Separate Go module: imperm-ui
│   ├── cmd/              # UI entry point
│   ├── internal/         # UI-specific code
│   │   ├── app.go
│   │   ├── control/      # Control tab
│   │   └── observe/      # Observe tab
│   └── pkg/              # Shared packages (client, models)
│
├── middleware/            # Middleware Server
│   ├── go.mod            # Separate Go module: imperm-middleware
│   ├── cmd/              # Server entry point
│   ├── internal/         # Server-specific code
│   │   ├── api/          # HTTP API handlers
│   │   ├── k8s/          # Kubernetes integration (TODO)
│   │   └── store/        # State management (TODO)
│   └── pkg/              # Shared packages (client, models)
│
└── bin/                   # Compiled binaries
    ├── imperm-ui
    └── imperm-server
```

## Complete Separation

### UI Application (`ui/`)
- **Module**: `imperm-ui`
- **Purpose**: Terminal user interface
- **Dependencies**: Bubble Tea, Lipgloss
- **Can run**: Standalone (mock mode) OR connected to server

### Middleware Server (`middleware/`)
- **Module**: `imperm-middleware`
- **Purpose**: HTTP API server for Kubernetes management
- **Dependencies**: Standard library (+ K8s client when added)
- **Can run**: Mock mode OR connected to real Kubernetes

## Shared Code

Both applications share the same interface and models via `pkg/`:

### `pkg/client/`
- `Client` interface - Defines all operations
- `MockClient` - In-memory test implementation
- `HTTPClient` - HTTP API client implementation

### `pkg/models/`
- `Environment` - Kubernetes environment
- `Pod` - Kubernetes pod
- `Deployment` - Kubernetes deployment
- `EnvironmentHistory` - Historical data

## Building

```bash
make all      # Build both UI and server
make ui       # Build UI only
make server   # Server only
make clean    # Clean binaries
```

## Running

### Standalone UI (Mock Mode)
```bash
make run-ui
# OR: ./bin/imperm-ui --mock
```

### Client-Server Mode

**Terminal 1 - Start Server:**
```bash
make run-server
# OR: ./bin/imperm-server --mock --port 8080
```

**Terminal 2 - Start UI:**
```bash
make run-ui-remote
# OR: ./bin/imperm-ui --server http://localhost:8080
```

## Development Workflow

### Working on UI
```bash
cd ui
go run ./cmd --mock
```

### Working on Middleware/Server
```bash
cd middleware
go run ./cmd --mock --port 8080
```

### Adding Kubernetes Integration
1. Implement in `middleware/internal/k8s/`
2. Update `middleware/internal/api/handler.go` to use K8s client
3. No changes needed to UI!

## API Endpoints

When server is running, UI connects via HTTP:

- `GET /api/environments`
- `POST /api/environments/create`
- `POST /api/environments/destroy`
- `GET /api/environments/history`
- `GET /api/pods?namespace=X`
- `GET /api/deployments?namespace=X`
- `GET /health`

## Next Steps

1. **Kubernetes Integration**: Implement `middleware/internal/k8s/client.go`
2. **Authentication**: Add auth to server and update HTTP client
3. **Real-time Updates**: Add WebSocket support
4. **Persistence**: Implement `middleware/internal/store/`
