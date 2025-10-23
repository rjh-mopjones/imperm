# Constant Logger - logs every 2 seconds
resource "kubernetes_deployment" "constant_logger" {
  count = var.constant_logger > 0 ? 1 : 0

  metadata {
    name      = "constant-logger"
    namespace = kubernetes_namespace.environment.metadata[0].name
    labels = {
      app        = "constant-logger"
      created-by = "imperm"
    }
  }

  spec {
    replicas = var.constant_logger

    selector {
      match_labels = {
        app = "constant-logger"
      }
    }

    template {
      metadata {
        labels = {
          app = "constant-logger"
        }
      }

      spec {
        container {
          name    = "logger"
          image   = "busybox:latest"
          command = ["/bin/sh"]
          args = [
            "-c",
            "while true; do echo \"[INFO] $(date '+%Y-%m-%d %H:%M:%S') - Constant logger message from pod $HOSTNAME\"; sleep 2; done"
          ]

          resources {
            limits = {
              cpu    = "100m"
              memory = "64Mi"
            }
            requests = {
              cpu    = "50m"
              memory = "32Mi"
            }
          }
        }
      }
    }
  }
}

# Fast Logger - logs every 0.5 seconds
resource "kubernetes_deployment" "fast_logger" {
  count = var.fast_logger > 0 ? 1 : 0

  metadata {
    name      = "fast-logger"
    namespace = kubernetes_namespace.environment.metadata[0].name
    labels = {
      app        = "fast-logger"
      created-by = "imperm"
    }
  }

  spec {
    replicas = var.fast_logger

    selector {
      match_labels = {
        app = "fast-logger"
      }
    }

    template {
      metadata {
        labels = {
          app = "fast-logger"
        }
      }

      spec {
        container {
          name    = "logger"
          image   = "busybox:latest"
          command = ["/bin/sh"]
          args = [
            "-c",
            "while true; do echo \"[INFO] $(date '+%Y-%m-%d %H:%M:%S.%N' | cut -b1-23) - Fast logger message from pod $HOSTNAME\"; sleep 0.5; done"
          ]

          resources {
            limits = {
              cpu    = "100m"
              memory = "64Mi"
            }
            requests = {
              cpu    = "50m"
              memory = "32Mi"
            }
          }
        }
      }
    }
  }
}

# Error Logger - mixed INFO and ERROR logs
resource "kubernetes_deployment" "error_logger" {
  count = var.error_logger > 0 ? 1 : 0

  metadata {
    name      = "error-logger"
    namespace = kubernetes_namespace.environment.metadata[0].name
    labels = {
      app        = "error-logger"
      created-by = "imperm"
    }
  }

  spec {
    replicas = var.error_logger

    selector {
      match_labels = {
        app = "error-logger"
      }
    }

    template {
      metadata {
        labels = {
          app = "error-logger"
        }
      }

      spec {
        container {
          name    = "logger"
          image   = "busybox:latest"
          command = ["/bin/sh"]
          args = [
            "-c",
            <<-EOT
            while true; do
              RAND=$((RANDOM % 3))
              if [ $RAND -eq 0 ]; then
                echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - Error occurred in pod $HOSTNAME: Something went wrong!"
              else
                echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') - Normal operation in pod $HOSTNAME"
              fi
              sleep 1
            done
            EOT
          ]

          resources {
            limits = {
              cpu    = "100m"
              memory = "64Mi"
            }
            requests = {
              cpu    = "50m"
              memory = "32Mi"
            }
          }
        }
      }
    }
  }
}

# JSON Logger - JSON formatted logs
resource "kubernetes_deployment" "json_logger" {
  count = var.json_logger > 0 ? 1 : 0

  metadata {
    name      = "json-logger"
    namespace = kubernetes_namespace.environment.metadata[0].name
    labels = {
      app        = "json-logger"
      created-by = "imperm"
    }
  }

  spec {
    replicas = var.json_logger

    selector {
      match_labels = {
        app = "json-logger"
      }
    }

    template {
      metadata {
        labels = {
          app = "json-logger"
        }
      }

      spec {
        container {
          name    = "logger"
          image   = "busybox:latest"
          command = ["/bin/sh"]
          args = [
            "-c",
            <<-EOT
            while true; do
              TIMESTAMP=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
              LEVEL="info"
              MESSAGE="JSON formatted log message"
              echo "{\"timestamp\":\"$TIMESTAMP\",\"level\":\"$LEVEL\",\"message\":\"$MESSAGE\",\"pod\":\"$HOSTNAME\",\"logger\":\"json-logger\"}"
              sleep 1
            done
            EOT
          ]

          resources {
            limits = {
              cpu    = "100m"
              memory = "64Mi"
            }
            requests = {
              cpu    = "50m"
              memory = "32Mi"
            }
          }
        }
      }
    }
  }
}
