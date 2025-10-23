package terraform

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Variable represents a Terraform variable
type Variable struct {
	Name        string
	Description string
	Type        string
	Default     string
	Category    string // Extracted from description prefix
}

// OptionCategory represents a UI category for options
type OptionCategory struct {
	Name      string
	Variables []Variable
}

// Parser handles parsing Terraform files
type Parser struct {
	variables []Variable
}

// NewParser creates a new Terraform parser
func NewParser() *Parser {
	return &Parser{
		variables: []Variable{},
	}
}

// ParseFile parses a local Terraform file
func (p *Parser) ParseFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.parse(file)
}

// ParseURL fetches and parses a Terraform file from a URL
func (p *Parser) ParseURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch URL: status %d", resp.StatusCode)
	}

	return p.parse(resp.Body)
}

// parse parses Terraform variable blocks from a reader
func (p *Parser) parse(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	var currentVar *Variable
	inVariableBlock := false
	inDescriptionBlock := false
	braceDepth := 0
	descriptionLines := []string{}

	variableNameRegex := regexp.MustCompile(`variable\s+"([^"]+)"`)
	descriptionRegex := regexp.MustCompile(`description\s*=\s*"(.+)"`)
	typeRegex := regexp.MustCompile(`type\s*=\s*(.+)`)
	defaultRegex := regexp.MustCompile(`default\s*=\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "//") {
			continue
		}

		// Start of variable block
		if matches := variableNameRegex.FindStringSubmatch(line); matches != nil {
			if currentVar != nil {
				p.variables = append(p.variables, *currentVar)
			}
			currentVar = &Variable{
				Name: matches[1],
			}
			inVariableBlock = true
			braceDepth = 0
			descriptionLines = []string{}
			continue
		}

		if inVariableBlock {
			// Track brace depth
			braceDepth += strings.Count(line, "{")
			braceDepth -= strings.Count(line, "}")

			// Single line description
			if matches := descriptionRegex.FindStringSubmatch(trimmedLine); matches != nil {
				currentVar.Description = matches[1]
				currentVar.Category = extractCategory(matches[1])
			} else if strings.Contains(trimmedLine, "description") && strings.Contains(trimmedLine, "<<") {
				// Multi-line description (heredoc)
				inDescriptionBlock = true
				continue
			}

			// Collect multi-line description
			if inDescriptionBlock {
				if strings.Contains(trimmedLine, "EOF") || strings.Contains(trimmedLine, "EOT") {
					inDescriptionBlock = false
					currentVar.Description = strings.Join(descriptionLines, " ")
					currentVar.Category = extractCategory(currentVar.Description)
					descriptionLines = []string{}
				} else {
					descriptionLines = append(descriptionLines, trimmedLine)
				}
			}

			// Type
			if matches := typeRegex.FindStringSubmatch(trimmedLine); matches != nil {
				currentVar.Type = strings.TrimSpace(matches[1])
			}

			// Default
			if matches := defaultRegex.FindStringSubmatch(trimmedLine); matches != nil {
				currentVar.Default = strings.TrimSpace(matches[1])
			}

			// End of variable block
			if braceDepth == 0 && strings.Contains(line, "}") {
				if currentVar != nil {
					p.variables = append(p.variables, *currentVar)
					currentVar = nil
				}
				inVariableBlock = false
			}
		}
	}

	// Add last variable if exists
	if currentVar != nil {
		p.variables = append(p.variables, *currentVar)
	}

	return scanner.Err()
}

// extractCategory extracts the category from a description
// Format: "CategoryName - description text"
func extractCategory(description string) string {
	parts := strings.SplitN(description, " - ", 2)
	if len(parts) == 2 {
		category := strings.TrimSpace(parts[0])
		// Normalize category name
		category = strings.ReplaceAll(category, " ", "")
		return category
	}
	return "General"
}

// GetVariables returns all parsed variables
func (p *Parser) GetVariables() []Variable {
	return p.variables
}

// GetCategorizedOptions returns variables grouped by category
func (p *Parser) GetCategorizedOptions() []OptionCategory {
	categoryMap := make(map[string][]Variable)

	for _, v := range p.variables {
		category := v.Category
		if category == "" {
			category = "General"
		}
		categoryMap[category] = append(categoryMap[category], v)
	}

	// Convert map to slice
	var categories []OptionCategory
	for name, vars := range categoryMap {
		categories = append(categories, OptionCategory{
			Name:      name,
			Variables: vars,
		})
	}

	return categories
}

// GetVariablesByCategory returns variables for a specific category
func (p *Parser) GetVariablesByCategory(category string) []Variable {
	var result []Variable
	for _, v := range p.variables {
		if v.Category == category {
			result = append(result, v)
		}
	}
	return result
}
