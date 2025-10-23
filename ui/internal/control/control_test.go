package control

import (
	"testing"
)

func TestLoadOptionsFromTerraform(t *testing.T) {
	categories := loadOptionsFromTerraform()

	if len(categories) == 0 {
		t.Fatal("Expected at least one category, got zero")
	}

	// Should have at least DeployOptions, DockerOptions, ServiceOptions
	if len(categories) < 3 {
		t.Errorf("Expected at least 3 categories, got %d", len(categories))
	}

	// Check that categories have fields
	for _, cat := range categories {
		if len(cat.fields) == 0 {
			t.Errorf("Category %s has no fields", cat.name)
		}

		t.Logf("Category: %s (%d fields)", cat.name, len(cat.fields))
		for _, field := range cat.fields {
			t.Logf("  - %s: %s", field.name, field.placeholder)
		}
	}
}

func TestGetFallbackOptions(t *testing.T) {
	categories := getFallbackOptions()

	if len(categories) != 3 {
		t.Errorf("Expected 3 fallback categories, got %d", len(categories))
	}

	// Verify we have the expected categories
	expectedCategories := map[string]bool{
		"DeployOptions":  false,
		"DockerOptions":  false,
		"ServiceOptions": false,
	}

	for _, cat := range categories {
		if _, exists := expectedCategories[cat.name]; exists {
			expectedCategories[cat.name] = true
		} else {
			t.Errorf("Unexpected category: %s", cat.name)
		}

		if len(cat.fields) == 0 {
			t.Errorf("Category %s has no fields", cat.name)
		}
	}

	// Check all expected categories were found
	for name, found := range expectedCategories {
		if !found {
			t.Errorf("Expected category %s not found", name)
		}
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"docker_registry", "DockerRegistry"},
		{"namespace_name", "NamespaceName"},
		{"constant_logger", "ConstantLogger"},
		{"service_port", "ServicePort"},
		{"simple", "Simple"},
		{"", ""},
	}

	for _, test := range tests {
		result := toPascalCase(test.input)
		if result != test.expected {
			t.Errorf("toPascalCase(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}
