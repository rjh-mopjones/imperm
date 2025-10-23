package terraform

// This file shows examples of different loader configurations
// Uncomment and modify the DefaultLoader function in loader.go to use these

/*
// Example 1: Load from GitHub
func DefaultLoader() (*Loader, error) {
	sources := []string{
		"https://raw.githubusercontent.com/your-org/imperm/main/terraform/modules/k8s-namespace/variables.tf",
	}

	config := LoaderConfig{
		Sources: sources,
	}

	loader := NewLoader(config)
	if err := loader.Load(); err != nil {
		return nil, err
	}

	return loader, nil
}
*/

/*
// Example 2: Load from multiple GitHub repos
func DefaultLoader() (*Loader, error) {
	sources := []string{
		"https://raw.githubusercontent.com/org/repo1/main/terraform/modules/base/variables.tf",
		"https://raw.githubusercontent.com/org/repo2/main/terraform/modules/extras/variables.tf",
	}

	config := LoaderConfig{
		Sources: sources,
	}

	loader := NewLoader(config)
	if err := loader.Load(); err != nil {
		return nil, err
	}

	return loader, nil
}
*/

/*
// Example 3: Fallback pattern - try GitHub first, fall back to local
func DefaultLoader() (*Loader, error) {
	sources := []string{
		"https://raw.githubusercontent.com/your-org/imperm/main/terraform/modules/k8s-namespace/variables.tf",
		"../terraform/modules/k8s-namespace", // Local fallback
	}

	config := LoaderConfig{
		Sources: sources,
	}

	loader := NewLoader(config)

	// Try each source until one succeeds
	var lastErr error
	for _, source := range sources {
		err := loader.loadSource(source)
		if err == nil {
			return loader, nil
		}
		lastErr = err
	}

	return nil, lastErr
}
*/

/*
// Example 4: Environment variable configuration
import "os"
import "strings"

func DefaultLoader() (*Loader, error) {
	// Get from environment or use default
	sourceEnv := os.Getenv("TERRAFORM_MODULE_SOURCES")

	var sources []string
	if sourceEnv != "" {
		sources = strings.Split(sourceEnv, ",")
	} else {
		sources = []string{
			"../terraform/modules/k8s-namespace",
		}
	}

	config := LoaderConfig{
		Sources: sources,
	}

	loader := NewLoader(config)
	if err := loader.Load(); err != nil {
		return nil, err
	}

	return loader, nil
}

// Usage:
// export TERRAFORM_MODULE_SOURCES="https://raw.githubusercontent.com/.../variables.tf,../terraform/modules/k8s-namespace"
// ./bin/imperm-ui
*/
