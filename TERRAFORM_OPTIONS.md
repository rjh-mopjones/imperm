# Dynamic Terraform Options

The UI options are now dynamically loaded from Terraform module variables!

## How It Works

### 1. Primary Mode: Load from Terraform
When you run `./bin/imperm-ui` from the project root, it automatically loads options from:
```
terraform/modules/k8s-namespace/variables.tf
```

### 2. Fallback Mode: Hardcoded Options
If Terraform files can't be found (e.g., running from different directory), the UI falls back to hardcoded options that match the Terraform definitions.

## Current Options

The UI now displays these option categories (from Terraform):

### DeployOptions
- **NamespaceName**: Name of the Kubernetes namespace
- **ConstantLogger**: Number of constant logger replicas (logs every 2s)
- **FastLogger**: Number of fast logger replicas (logs every 0.5s)
- **ErrorLogger**: Number of error logger replicas (mixed INFO/ERROR)
- **JsonLogger**: Number of JSON logger replicas (JSON formatted)

### DockerOptions (NEW!)
- **DockerRegistry**: Container registry URL (default: docker.io)
- **DockerTag**: Container image tag (default: latest)
- **DockerPullPolicy**: Image pull policy (default: IfNotPresent)

### ServiceOptions (NEW!)
- **ServicePort**: Service port number (default: 8080)
- **ServiceType**: Kubernetes service type (default: ClusterIP)

## Adding New Options

To add new UI options:

1. Edit `terraform/modules/k8s-namespace/variables.tf`
2. Add a variable with category prefix in description:

```hcl
variable "my_new_option" {
  description = "Category Name - Description of your option"
  type        = string
  default     = "default-value"
}
```

3. Rebuild the UI: `make`
4. The option automatically appears in the UI!

## Category Format

Use this format in variable descriptions:
```
"Category Name - description text"
```

Examples:
- `"Docker Options - Container registry"` → DockerOptions category
- `"Deploy Options - Namespace name"` → DeployOptions category
- `"Custom Options - My setting"` → CustomOptions category

## Future: GitHub URL Loading

When you push your Terraform modules to GitHub, update `ui/pkg/terraform/loader.go`:

```go
sources := []string{
    "https://raw.githubusercontent.com/your-org/imperm/main/terraform/modules/k8s-namespace/variables.tf",
    "terraform/modules/k8s-namespace", // Local fallback
}
```

The UI will fetch and parse variables directly from GitHub!

## Testing

Test the parser:
```bash
cd ui
go run cmd/test_parser/main.go
```

You should see all categories and variables parsed from the Terraform file.

## Technical Details

- **Parser**: `ui/pkg/terraform/parser.go` - Parses .tf files
- **Loader**: `ui/pkg/terraform/loader.go` - Handles file/URL loading
- **UI Integration**: `ui/internal/control/control.go` - Converts to UI options
- **Fallback**: Hardcoded options ensure UI always works

The system is resilient - it always shows options, whether loaded from Terraform or using fallback values.
