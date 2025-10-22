# Minikube Setup Guide

This guide will help you set up a local Kubernetes cluster using minikube to test the imperm middleware server.

## Prerequisites

- Docker Desktop (or another container runtime)
- Homebrew (for macOS)
- At least 2GB of free RAM

## Installation

### 1. Install minikube

```bash
# Install minikube via Homebrew
brew install minikube

# Verify installation
minikube version
```

### 2. Install kubectl (if not already installed)

```bash
# Install kubectl via Homebrew
brew install kubectl

# Verify installation
kubectl version --client
```

## Starting Your Local Cluster

### Start minikube

```bash
# Start minikube with default settings
minikube start

# Or with specific resources
minikube start --cpus=2 --memory=4096
```

This will:
- Download the Kubernetes cluster image (first time only)
- Create a virtual machine
- Configure kubectl to use the minikube cluster

### Verify Cluster is Running

```bash
# Check minikube status
minikube status

# Check kubectl can connect
kubectl get nodes

# You should see something like:
# NAME       STATUS   ROLES           AGE   VERSION
# minikube   Ready    control-plane   1m    v1.28.x
```

## Testing the Middleware Server

### 1. Build the Server

```bash
cd /Users/roryhedderman/GolandProjects/imperm/middleware
go build -o ../bin/imperm-server ./cmd
```

### 2. Run the Server (Non-Mock Mode)

```bash
# Run the server - it will connect to your minikube cluster
../bin/imperm-server

# You should see:
# Initializing Kubernetes client...
# Successfully connected to Kubernetes cluster
# Starting server on :8080
```

The server automatically uses your kubeconfig file (`~/.kube/config`) which minikube configures when you run `minikube start`.

### 3. Test Creating an Environment

```bash
# Create a simple environment (namespace only)
curl -X POST http://localhost:8080/api/environments/create \
  -H "Content-Type: application/json" \
  -d '{"name": "test-env", "with_options": false}'

# Create an environment with sample nginx deployment
curl -X POST http://localhost:8080/api/environments/create \
  -H "Content-Type: application/json" \
  -d '{"name": "demo-env", "with_options": true}'
```

### 4. Verify Environment Creation

```bash
# List all environments via API
curl http://localhost:8080/api/environments

# Or check directly with kubectl
kubectl get namespaces

# Check pods in the demo-env namespace
kubectl get pods -n demo-env
```

### 5. List Pods and Deployments

```bash
# List all pods
curl http://localhost:8080/api/pods

# List pods in specific namespace
curl http://localhost:8080/api/pods?namespace=demo-env

# List deployments in specific namespace
curl http://localhost:8080/api/deployments?namespace=demo-env
```

### 6. Destroy an Environment

```bash
# Delete the environment and all its resources
curl -X POST http://localhost:8080/api/environments/destroy \
  -H "Content-Type: application/json" \
  -d '{"name": "test-env"}'

# Verify deletion
kubectl get namespaces
```

## Useful minikube Commands

```bash
# Stop the cluster (preserves state)
minikube stop

# Start the stopped cluster
minikube start

# Delete the cluster entirely
minikube delete

# Open Kubernetes dashboard
minikube dashboard

# SSH into the minikube node
minikube ssh

# View cluster logs
minikube logs

# Check cluster IP
minikube ip

# List addons
minikube addons list

# Enable an addon (e.g., metrics-server for resource usage)
minikube addons enable metrics-server
```

## Troubleshooting

### "Failed to create Kubernetes client" Error

Check your kubeconfig:
```bash
kubectl config current-context
# Should show: minikube

kubectl config view
# Should show minikube cluster configuration
```

### Connection Timeout Errors

```bash
# Restart minikube
minikube stop
minikube start

# Or delete and recreate
minikube delete
minikube start
```

### Port Already in Use

If port 8080 is already taken, modify the port in `middleware/cmd/main.go` or use:
```bash
# Kill process using port 8080
lsof -ti:8080 | xargs kill -9
```

### Docker Driver Issues

```bash
# If you have issues with Docker Desktop, try the hyperkit driver
minikube start --driver=hyperkit

# Or virtualbox
brew install virtualbox
minikube start --driver=virtualbox
```

## Development Workflow

1. Start minikube once:
   ```bash
   minikube start
   ```

2. Run the middleware server:
   ```bash
   cd /Users/roryhedderman/GolandProjects/imperm
   ./bin/imperm-server
   ```

3. Run the UI client (in another terminal):
   ```bash
   cd /Users/roryhedderman/GolandProjects/imperm
   ./bin/imperm-ui --remote
   ```

4. When done for the day:
   ```bash
   minikube stop  # Preserves cluster state
   ```

5. Next session:
   ```bash
   minikube start  # Resumes from previous state
   ```

## Next Steps

- Enable metrics-server for real CPU/Memory stats: `minikube addons enable metrics-server`
- Explore the Kubernetes dashboard: `minikube dashboard`
- Try deploying custom applications to your environments
- Test the full UI workflow with real Kubernetes resources

## Resources

- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Kubernetes Concepts](https://kubernetes.io/docs/concepts/)
