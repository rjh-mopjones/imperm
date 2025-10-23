output "namespace_name" {
  description = "The name of the created namespace"
  value       = kubernetes_namespace.environment.metadata[0].name
}

output "namespace_id" {
  description = "The ID of the created namespace"
  value       = kubernetes_namespace.environment.id
}

output "deployment_created" {
  description = "Whether a sample deployment was created"
  value       = var.with_options
}

output "service_created" {
  description = "Whether a sample service was created"
  value       = var.with_options
}
