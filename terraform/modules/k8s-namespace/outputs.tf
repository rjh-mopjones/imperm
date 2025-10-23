output "namespace_name" {
  description = "The name of the created namespace"
  value       = kubernetes_namespace.environment.metadata[0].name
}

output "namespace_id" {
  description = "The ID of the created namespace"
  value       = kubernetes_namespace.environment.id
}

output "constant_logger_created" {
  description = "Whether constant logger was created"
  value       = var.constant_logger > 0
}

output "fast_logger_created" {
  description = "Whether fast logger was created"
  value       = var.fast_logger > 0
}

output "error_logger_created" {
  description = "Whether error logger was created"
  value       = var.error_logger > 0
}

output "json_logger_created" {
  description = "Whether JSON logger was created"
  value       = var.json_logger > 0
}
