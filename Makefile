.PHONY: all ui server clean run-ui run-server run-ui-remote help

# Default target
all: ui server

help:
	@echo "Imperm Monorepo - Available targets:"
	@echo ""
	@echo "Building:"
	@echo "  make all        - Build both UI and server"
	@echo "  make ui         - Build UI client only"
	@echo "  make server     - Build server only"
	@echo "  make clean      - Remove all build artifacts"
	@echo ""
	@echo "Running (Development):"
	@echo "  make run-ui            - Run UI in mock mode (standalone)"
	@echo "  make run-server        - Run server in mock mode"
	@echo "  make run-ui-remote     - Run UI connected to server"
	@echo "  make run-server-k8s    - Run server with real K8s (when implemented)"
	@echo ""
	@echo "Testing:"
	@echo "  make test       - Run tests for both modules"
	@echo "  make test-ui    - Run UI tests"
	@echo "  make test-server - Run server tests"

# Build targets
ui:
	@echo "Building UI client (imperm-ui)..."
	@cd ui && go build -o ../bin/imperm-ui ./cmd
	@echo "✓ Built: bin/imperm-ui"

server:
	@echo "Building server (imperm-server)..."
	@cd middleware && go build -o ../bin/imperm-server ./cmd
	@echo "✓ Built: bin/imperm-server"

# Clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "✓ Clean complete"

# Run targets
run-ui:
	@echo "Running UI in mock mode..."
	@cd ui && go run ./cmd --mock

run-ui-remote:
	@echo "Running UI connected to server at http://localhost:8080..."
	@cd ui && go run ./cmd --server http://localhost:8080

run-server:
	@echo "Running server in mock mode on port 8080..."
	@cd middleware && go run ./cmd --mock --port 8080

run-server-k8s:
	@echo "Running server with real Kubernetes..."
	@cd middleware && go run ./cmd --port 8080

# Test targets
test: test-ui test-server

test-ui:
	@echo "Running UI tests..."
	@cd ui && go test ./...

test-server:
	@echo "Running server tests..."
	@cd middleware && go test ./...

# Go module management
tidy:
	@echo "Tidying Go modules..."
	@cd ui && go mod tidy
	@cd middleware && go mod tidy
	@echo "✓ Modules tidied"
