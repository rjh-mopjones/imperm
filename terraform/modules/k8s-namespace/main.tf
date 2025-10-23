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

# Logger deployments are defined in loggers.tf
