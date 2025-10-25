package control

import (
	"imperm-ui/pkg/tfvars"
	"strings"
)

// loadOptionsFromTerraform loads option categories from Terraform modules
func loadOptionsFromTerraform() []optionCategory {
	// Try multiple paths for the Terraform variables file
	paths := []string{
		"../terraform/modules/k8s-namespace/variables.tf",
		"terraform/modules/k8s-namespace/variables.tf",
		"./terraform/modules/k8s-namespace/variables.tf",
		"../../terraform/modules/k8s-namespace/variables.tf",
	}

	var variables []tfvars.Variable
	var err error

	for _, path := range paths {
		variables, err = tfvars.ExtractFromFile(path)
		if err == nil && len(variables) > 0 {
			break
		}
	}

	// If all paths fail, fall back to hardcoded options
	if err != nil || len(variables) == 0 {
		return getFallbackOptions()
	}

	// Group variables by category
	categoryMap := tfvars.GroupByCategory(variables)

	// Convert to optionCategory format
	categories := make([]optionCategory, 0, len(categoryMap))
	for catName, vars := range categoryMap {
		fields := make([]optionField, 0, len(vars))

		for _, v := range vars {
			// Extract just the description part after " - "
			desc := v.Description
			parts := strings.SplitN(desc, " - ", 2)
			if len(parts) == 2 {
				desc = parts[1]
			}

			fields = append(fields, optionField{
				name:        v.Name,
				placeholder: desc,
				value:       "",
			})
		}

		categories = append(categories, optionCategory{
			name:   catName,
			fields: fields,
		})
	}

	return categories
}

// getFallbackOptions returns hardcoded options if Terraform loading fails
func getFallbackOptions() []optionCategory {
	return []optionCategory{
		{
			name: "DeployOptions",
			fields: []optionField{
				{name: "name", placeholder: "environment-name (leave empty for auto-generated)"},
				{name: "namespace", placeholder: "e.g., default, test-logging"},
				{name: "constant_logger", placeholder: "replicas (e.g., 3) - logs every 2s"},
				{name: "fast_logger", placeholder: "replicas (e.g., 2) - logs every 0.5s"},
				{name: "error_logger", placeholder: "replicas (e.g., 1) - mixed INFO/ERROR logs"},
				{name: "json_logger", placeholder: "replicas (e.g., 2) - JSON formatted logs"},
			},
		},
		{
			name: "DockerOptions",
			fields: []optionField{
				{name: "docker_registry", placeholder: "Container registry URL (default: docker.io)"},
				{name: "docker_tag", placeholder: "Container image tag (default: latest)"},
				{name: "docker_pull_policy", placeholder: "Image pull policy (default: IfNotPresent)"},
			},
		},
		{
			name: "ServiceOptions",
			fields: []optionField{
				{name: "service_port", placeholder: "Service port number (default: 8080)"},
				{name: "service_type", placeholder: "Kubernetes service type (default: ClusterIP)"},
			},
		},
	}
}
