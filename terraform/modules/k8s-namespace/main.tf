terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
  }
}

resource "kubernetes_namespace" "environment" {
  metadata {
    name = var.namespace_name

    labels = {
      managed-by  = "imperm"
      environment = var.namespace_name
    }

    annotations = {
      created-at = timestamp()
      created-by = "imperm-terraform"
    }
  }
}

# Optional: Create a sample deployment if with_options is true
resource "kubernetes_deployment" "sample_app" {
  count = var.with_options ? 1 : 0

  metadata {
    name      = "sample-app"
    namespace = kubernetes_namespace.environment.metadata[0].name

    labels = {
      app        = "sample"
      created-by = "imperm"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "sample"
      }
    }

    template {
      metadata {
        labels = {
          app = "sample"
        }
      }

      spec {
        container {
          name  = "nginx"
          image = "nginx:latest"

          port {
            container_port = 80
          }

          resources {
            limits = {
              cpu    = "500m"
              memory = "512Mi"
            }
            requests = {
              cpu    = "250m"
              memory = "256Mi"
            }
          }
        }
      }
    }
  }
}

# Optional: Create a service for the sample app
resource "kubernetes_service" "sample_service" {
  count = var.with_options ? 1 : 0

  metadata {
    name      = "sample-service"
    namespace = kubernetes_namespace.environment.metadata[0].name
  }

  spec {
    selector = {
      app = "sample"
    }

    port {
      port        = 80
      target_port = 80
    }

    type = "ClusterIP"
  }
}
