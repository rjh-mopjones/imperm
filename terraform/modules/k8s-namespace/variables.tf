variable "namespace_name" {
  description = "Name of the Kubernetes namespace to create"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", var.namespace_name))
    error_message = "Namespace name must be a valid DNS label (lowercase alphanumeric and hyphens only)."
  }
}

variable "constant_logger" {
  description = "Number of constant logger replicas (logs every 2s)"
  type        = number
  default     = 0
}

variable "fast_logger" {
  description = "Number of fast logger replicas (logs every 0.5s)"
  type        = number
  default     = 0
}

variable "error_logger" {
  description = "Number of error logger replicas (mixed INFO/ERROR logs)"
  type        = number
  default     = 0
}

variable "json_logger" {
  description = "Number of JSON logger replicas (JSON formatted logs)"
  type        = number
  default     = 0
}
