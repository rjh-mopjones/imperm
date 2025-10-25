package tfvars

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Variable represents a simple Terraform variable
type Variable struct {
	Name        string
	Description string
	Category    string
}

// ExtractFromFile reads a local .tf file and extracts variable names
func ExtractFromFile(filePath string) ([]Variable, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return extract(file)
}

// ExtractFromURL fetches a .tf file from a URL and extracts variable names
func ExtractFromURL(url string) ([]Variable, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL: status %d", resp.StatusCode)
	}

	return extract(resp.Body)
}

// extract pulls variable names and descriptions from a reader
func extract(reader io.Reader) ([]Variable, error) {
	scanner := bufio.NewScanner(reader)

	variableNameRegex := regexp.MustCompile(`variable\s+"([^"]+)"`)
	descriptionRegex := regexp.MustCompile(`description\s*=\s*"(.+)"`)

	var variables []Variable
	var currentVar *Variable

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "//") {
			continue
		}

		// Found a variable declaration
		if matches := variableNameRegex.FindStringSubmatch(line); matches != nil {
			if currentVar != nil {
				variables = append(variables, *currentVar)
			}
			currentVar = &Variable{
				Name: matches[1],
			}
			continue
		}

		// Found a description
		if currentVar != nil {
			if matches := descriptionRegex.FindStringSubmatch(trimmedLine); matches != nil {
				currentVar.Description = matches[1]
				currentVar.Category = extractCategory(matches[1])
			}
		}

		// End of variable block
		if currentVar != nil && strings.Contains(trimmedLine, "}") {
			variables = append(variables, *currentVar)
			currentVar = nil
		}
	}

	// Add last variable if exists
	if currentVar != nil {
		variables = append(variables, *currentVar)
	}

	return variables, scanner.Err()
}

// extractCategory extracts the category from a description
// Format: "CategoryName - description text"
func extractCategory(description string) string {
	parts := strings.SplitN(description, " - ", 2)
	if len(parts) == 2 {
		category := strings.TrimSpace(parts[0])
		category = strings.ReplaceAll(category, " ", "")
		return category
	}
	return "General"
}

// GroupByCategory groups variables by their category
func GroupByCategory(variables []Variable) map[string][]Variable {
	categoryMap := make(map[string][]Variable)

	for _, v := range variables {
		category := v.Category
		if category == "" {
			category = "General"
		}
		categoryMap[category] = append(categoryMap[category], v)
	}

	return categoryMap
}
