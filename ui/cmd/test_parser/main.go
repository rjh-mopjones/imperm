package main

import (
	"fmt"
	"imperm-ui/pkg/terraform"
)

func main() {
	fmt.Println("Loading Terraform modules...")

	// Load from default location
	loader, err := terraform.DefaultLoader()
	if err != nil {
		fmt.Printf("Error loading Terraform modules: %v\n", err)
		return
	}

	fmt.Println("\n=== All Variables ===")
	vars := loader.GetVariables()
	for _, v := range vars {
		fmt.Printf("\nVariable: %s\n", v.Name)
		fmt.Printf("  Category: %s\n", v.Category)
		fmt.Printf("  Description: %s\n", v.Description)
		fmt.Printf("  Type: %s\n", v.Type)
		if v.Default != "" {
			fmt.Printf("  Default: %s\n", v.Default)
		}
	}

	fmt.Println("\n\n=== Categorized Options ===")
	categories := loader.GetCategorizedOptions()
	for _, cat := range categories {
		fmt.Printf("\nCategory: %s (%d variables)\n", cat.Name, len(cat.Variables))
		for _, v := range cat.Variables {
			fmt.Printf("  - %s: %s\n", v.Name, v.Description)
		}
	}
}
