#!/bin/bash

# Script to deploy logging pods based on configured options
# This would be called by the server when creating an environment

NAMESPACE=${1:-"test-logging"}
CONSTANT_LOGGER_REPLICAS=${2:-0}
FAST_LOGGER_REPLICAS=${3:-0}
ERROR_LOGGER_REPLICAS=${4:-0}
JSON_LOGGER_REPLICAS=${5:-0}

echo "Deploying logging pods to namespace: $NAMESPACE"

# Create namespace if it doesn't exist
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Deploy constant-logger if replicas > 0
if [ "$CONSTANT_LOGGER_REPLICAS" -gt 0 ]; then
  echo "Deploying constant-logger with $CONSTANT_LOGGER_REPLICAS replicas..."
  cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: constant-logger
  namespace: $NAMESPACE
spec:
  replicas: $CONSTANT_LOGGER_REPLICAS
  selector:
    matchLabels:
      app: constant-logger
  template:
    metadata:
      labels:
        app: constant-logger
    spec:
      containers:
      - name: logger
        image: busybox
        command: ["/bin/sh"]
        args:
          - -c
          - |
            counter=0
            while true; do
              echo "[$(date '+%Y-%m-%d %H:%M:%S')] Log entry #\$counter - This is a constant log message from \$(hostname)"
              counter=\$((counter + 1))
              sleep 2
            done
EOF
fi

# Deploy fast-logger if replicas > 0
if [ "$FAST_LOGGER_REPLICAS" -gt 0 ]; then
  echo "Deploying fast-logger with $FAST_LOGGER_REPLICAS replicas..."
  cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fast-logger
  namespace: $NAMESPACE
spec:
  replicas: $FAST_LOGGER_REPLICAS
  selector:
    matchLabels:
      app: fast-logger
  template:
    metadata:
      labels:
        app: fast-logger
    spec:
      containers:
      - name: logger
        image: busybox
        command: ["/bin/sh"]
        args:
          - -c
          - |
            counter=0
            while true; do
              echo "[FAST] \$(date '+%H:%M:%S') - Event #\$counter | Status: OK | Pod: \$(hostname)"
              counter=\$((counter + 1))
              sleep 0.5
            done
EOF
fi

# Deploy error-logger if replicas > 0
if [ "$ERROR_LOGGER_REPLICAS" -gt 0 ]; then
  echo "Deploying error-logger with $ERROR_LOGGER_REPLICAS replicas..."
  cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: error-logger
  namespace: $NAMESPACE
spec:
  replicas: $ERROR_LOGGER_REPLICAS
  selector:
    matchLabels:
      app: error-logger
  template:
    metadata:
      labels:
        app: error-logger
    spec:
      containers:
      - name: logger
        image: busybox
        command: ["/bin/sh"]
        args:
          - -c
          - |
            counter=0
            while true; do
              if [ \$((counter % 5)) -eq 0 ]; then
                echo "[ERROR] \$(date) - Something went wrong! Error code: \$counter"
              else
                echo "[INFO] \$(date) - Normal operation, counter: \$counter"
              fi
              counter=\$((counter + 1))
              sleep 3
            done
EOF
fi

# Deploy json-logger if replicas > 0
if [ "$JSON_LOGGER_REPLICAS" -gt 0 ]; then
  echo "Deploying json-logger with $JSON_LOGGER_REPLICAS replicas..."
  cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: json-logger
  namespace: $NAMESPACE
spec:
  replicas: $JSON_LOGGER_REPLICAS
  selector:
    matchLabels:
      app: json-logger
  template:
    metadata:
      labels:
        app: json-logger
    spec:
      containers:
      - name: logger
        image: busybox
        command: ["/bin/sh"]
        args:
          - -c
          - |
            counter=0
            while true; do
              echo "{\"timestamp\":\"\$(date -Iseconds)\",\"level\":\"info\",\"message\":\"Processing request\",\"request_id\":\"\$counter\",\"pod\":\"\$(hostname)\",\"status\":\"success\"}"
              counter=\$((counter + 1))
              sleep 1
            done
EOF
fi

echo "Deployment complete!"
echo ""
echo "View pods with: kubectl get pods -n $NAMESPACE"
echo "View logs with: kubectl logs -l app=constant-logger -n $NAMESPACE --tail=20"
