# Kubernetes Namespace Module

This Terraform module creates a Kubernetes namespace with optional starter resources.

## Features

- Creates a Kubernetes namespace with Imperm labels and annotations
- Optionally creates sample deployment (nginx) and service
- Input validation for namespace naming conventions

## Usage

```hcl
module "environment" {
  source = "./modules/k8s-namespace"

  namespace_name = "dev-env-1"
  with_options   = true
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| namespace_name | Name of the Kubernetes namespace to create | `string` | n/a | yes |
| with_options | Whether to create sample resources | `bool` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| namespace_name | The name of the created namespace |
| namespace_id | The ID of the created namespace |
| deployment_created | Whether a sample deployment was created |
| service_created | Whether a sample service was created |
