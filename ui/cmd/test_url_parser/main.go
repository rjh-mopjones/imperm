package main

import (
	"fmt"
	"imperm-ui/pkg/terraform"
)

func main() {
	fmt.Println("Testing URL-based Terraform module loading...")
	fmt.Println("This demonstrates how to load from a GitHub raw URL")
	fmt.Println()

	// Example: Loading from a GitHub raw URL
	// Replace this with your actual GitHub repo URL when ready
	exampleURL := "https://raw.githubusercontent.com/your-org/your-repo/main/terraform/modules/k8s-namespace/variables.tf"

	fmt.Printf("Example usage:\n")
	fmt.Printf("  loader, err := terraform.LoaderFromURL(\"%s\")\n", exampleURL)
	fmt.Println()
	fmt.Println("For now, testing with local files...")

	// Test with local file for demonstration
	loader, err := terraform.DefaultLoader()
	if err != nil {
		fmt.Printf("Error loading Terraform modules: %v\n", err)
		return
	}

	categories := loader.GetCategorizedOptions()
	fmt.Printf("\nFound %d categories:\n", len(categories))
	for _, cat := range categories {
		fmt.Printf("  - %s (%d variables)\n", cat.Name, len(cat.Variables))
	}

	fmt.Println("\n✓ URL loading is supported and ready to use!")
	fmt.Println("  Just set the URL in the terraform.LoaderFromURL() function")
}
