package terraform

import (
	"fmt"
	"path/filepath"
	"strings"
)

// LoaderConfig contains configuration for loading Terraform modules
type LoaderConfig struct {
	// Sources can be local file paths or URLs
	Sources []string
}

// Loader handles loading and parsing Terraform modules
type Loader struct {
	config LoaderConfig
	parser *Parser
}

// NewLoader creates a new Terraform loader
func NewLoader(config LoaderConfig) *Loader {
	return &Loader{
		config: config,
		parser: NewParser(),
	}
}

// Load parses all configured sources
func (l *Loader) Load() error {
	for _, source := range l.config.Sources {
		if err := l.loadSource(source); err != nil {
			return fmt.Errorf("failed to load source %s: %w", source, err)
		}
	}
	return nil
}

// loadSource determines if source is URL or file path and loads accordingly
func (l *Loader) loadSource(source string) error {
	if isURL(source) {
		return l.parser.ParseURL(source)
	}
	return l.parseLocalPath(source)
}

// parseLocalPath handles local file paths (can be a file or directory)
func (l *Loader) parseLocalPath(path string) error {
	// Check if it's a specific .tf file
	if strings.HasSuffix(path, ".tf") {
		return l.parser.ParseFile(path)
	}

	// Otherwise, assume it's a directory and look for variables.tf
	variablesPath := filepath.Join(path, "variables.tf")
	return l.parser.ParseFile(variablesPath)
}

// isURL checks if a source is a URL
func isURL(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// GetCategorizedOptions returns parsed options grouped by category
func (l *Loader) GetCategorizedOptions() []OptionCategory {
	return l.parser.GetCategorizedOptions()
}

// GetVariables returns all parsed variables
func (l *Loader) GetVariables() []Variable {
	return l.parser.GetVariables()
}

// DefaultLoader creates a loader with default configuration
// It looks for the k8s-namespace module in the standard location
// Can be configured via environment variable TERRAFORM_MODULE_SOURCE
func DefaultLoader() (*Loader, error) {
	// Try multiple possible paths depending on where the binary is run from
	sources := []string{
		"../terraform/modules/k8s-namespace",        // Running from ui/ directory
		"terraform/modules/k8s-namespace",           // Running from project root
		"./terraform/modules/k8s-namespace",         // Running from project root (explicit)
		"../../terraform/modules/k8s-namespace",     // Running from ui/cmd subdirectory
	}

	loader := NewLoader(LoaderConfig{Sources: sources})

	// Try to load from at least one source
	var lastErr error
	for _, source := range sources {
		err := loader.loadSource(source)
		if err == nil {
			// Successfully loaded from this source
			return loader, nil
		}
		lastErr = err
	}

	// All sources failed
	return nil, fmt.Errorf("failed to load from any source: %w", lastErr)
}

// LoaderFromURL creates a loader that fetches Terraform files from a GitHub raw URL
func LoaderFromURL(urls ...string) (*Loader, error) {
	config := LoaderConfig{
		Sources: urls,
	}

	loader := NewLoader(config)
	if err := loader.Load(); err != nil {
		return nil, err
	}

	return loader, nil
}
