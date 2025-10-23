variable "namespace_name" {
  description = "Name of the Kubernetes namespace to create"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", var.namespace_name))
    error_message = "Namespace name must be a valid DNS label (lowercase alphanumeric and hyphens only)."
  }
}

variable "with_options" {
  description = "Whether to create sample resources (deployment, service) in the namespace"
  type        = bool
  default     = false
}
