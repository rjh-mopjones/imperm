# Terraform Module Parser

This package provides functionality to dynamically load UI options from Terraform module variables.

## Overview

The UI options are automatically generated from Terraform variable definitions. Variables are categorized based on a prefix in their description field.

## How It Works

### 1. Categorization

Variables are grouped by category based on their description prefix:

```hcl
variable "docker_registry" {
  description = "Docker Options - Container registry URL"
  type        = string
  default     = "docker.io"
}
```

The format is: `"Category Name - description text"`

- `Docker Options` → Becomes `DockerOptions` category
- `Deploy Options` → Becomes `DeployOptions` category
- `Service Options` → Becomes `ServiceOptions` category

### 2. Loading Sources

The loader supports multiple sources:

#### Local Files
```go
loader, err := terraform.DefaultLoader()
```

This loads from the default location: `../terraform/modules/k8s-namespace/variables.tf`

#### GitHub URLs
```go
loader, err := terraform.LoaderFromURL(
    "https://raw.githubusercontent.com/your-org/repo/main/terraform/modules/k8s-namespace/variables.tf",
)
```

#### Multiple Sources
```go
config := terraform.LoaderConfig{
    Sources: []string{
        "../terraform/modules/k8s-namespace",
        "https://raw.githubusercontent.com/.../variables.tf",
    },
}
loader := terraform.NewLoader(config)
loader.Load()
```

### 3. Using in UI

The control UI automatically loads options from Terraform:

```go
categories := loadOptionsFromTerraform()
```

This creates dynamic UI categories based on the Terraform variables.

## Adding New Options

To add new UI options:

1. Add the variable to your Terraform module with a category prefix:

```hcl
variable "new_option" {
  description = "My Category - Description of the option"
  type        = string
  default     = "default-value"
}
```

2. The UI will automatically pick it up and create:
   - A category called "MyCategory" (if it doesn't exist)
   - A field with the description and default value

## Examples

See the test programs:
- `cmd/test_parser/main.go` - Demonstrates parsing local Terraform files
- `cmd/test_url_parser/main.go` - Demonstrates URL-based loading

Run them with:
```bash
go run cmd/test_parser/main.go
go run cmd/test_url_parser/main.go
```

## Future: GitHub Integration

When you host your Terraform modules on GitHub, simply update the loader to point to the raw GitHub URL:

```go
func DefaultLoader() (*Loader, error) {
    sources := []string{
        "https://raw.githubusercontent.com/your-org/imperm/main/terraform/modules/k8s-namespace/variables.tf",
    }
    // ...
}
```

The UI will automatically fetch and parse the variables from GitHub!
