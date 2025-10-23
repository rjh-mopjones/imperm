variable "namespace_name" {
  description = "Deploy Options - Name of the Kubernetes namespace to create"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", var.namespace_name))
    error_message = "Namespace name must be a valid DNS label (lowercase alphanumeric and hyphens only)."
  }
}

variable "constant_logger" {
  description = "Deploy Options - Number of constant logger replicas (logs every 2s)"
  type        = number
  default     = 0
}

variable "fast_logger" {
  description = "Deploy Options - Number of fast logger replicas (logs every 0.5s)"
  type        = number
  default     = 0
}

variable "error_logger" {
  description = "Deploy Options - Number of error logger replicas (mixed INFO/ERROR logs)"
  type        = number
  default     = 0
}

variable "json_logger" {
  description = "Deploy Options - Number of JSON logger replicas (JSON formatted logs)"
  type        = number
  default     = 0
}

# Docker Options
variable "docker_registry" {
  description = "Docker Options - Container registry URL"
  type        = string
  default     = "docker.io"
}

variable "docker_tag" {
  description = "Docker Options - Container image tag"
  type        = string
  default     = "latest"
}

variable "docker_pull_policy" {
  description = "Docker Options - Image pull policy (Always, IfNotPresent, Never)"
  type        = string
  default     = "IfNotPresent"
}

# Service Options
variable "service_port" {
  description = "Service Options - Service port number"
  type        = number
  default     = 8080
}

variable "service_type" {
  description = "Service Options - Kubernetes service type (ClusterIP, NodePort, LoadBalancer)"
  type        = string
  default     = "ClusterIP"
}
