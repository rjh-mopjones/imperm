package terraform

import (
	"strings"
	"testing"
)

func TestParseVariables(t *testing.T) {
	tfContent := `
variable "namespace_name" {
  description = "Deploy Options - Name of the Kubernetes namespace to create"
  type        = string
}

variable "constant_logger" {
  description = "Deploy Options - Number of constant logger replicas (logs every 2s)"
  type        = number
  default     = 0
}

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

variable "service_port" {
  description = "Service Options - Service port number"
  type        = number
  default     = 8080
}
`

	parser := NewParser()
	err := parser.parse(strings.NewReader(tfContent))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	vars := parser.GetVariables()
	if len(vars) != 5 {
		t.Errorf("Expected 5 variables, got %d", len(vars))
	}

	// Check categorization
	categories := parser.GetCategorizedOptions()
	if len(categories) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(categories))
	}

	// Verify specific variables
	found := false
	for _, v := range vars {
		if v.Name == "docker_registry" {
			found = true
			if v.Category != "DockerOptions" {
				t.Errorf("Expected category 'DockerOptions', got '%s'", v.Category)
			}
			if v.Default != `"docker.io"` {
				t.Errorf("Expected default 'docker.io', got '%s'", v.Default)
			}
		}
	}

	if !found {
		t.Error("docker_registry variable not found")
	}
}

func TestCategoryExtraction(t *testing.T) {
	tests := []struct {
		description string
		expected    string
	}{
		{"Deploy Options - Some description", "DeployOptions"},
		{"Docker Options - Container registry", "DockerOptions"},
		{"Service Options - Port number", "ServiceOptions"},
		{"No category here", "General"},
		{"", "General"},
	}

	for _, test := range tests {
		result := extractCategory(test.description)
		if result != test.expected {
			t.Errorf("For '%s', expected '%s', got '%s'", test.description, test.expected, result)
		}
	}
}
