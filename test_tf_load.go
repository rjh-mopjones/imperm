package main

import (
	"fmt"
	"imperm-ui/pkg/terraform"
	"os"
)

func main() {
	pwd, _ := os.Getwd()
	fmt.Printf("Current directory: %s\n\n", pwd)

	fmt.Println("Attempting to load Terraform modules...")
	loader, err := terraform.DefaultLoader()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("\nFalling back to hardcoded options would be used in UI")
		return
	}

	fmt.Println("✓ Successfully loaded Terraform modules!\n")

	categories := loader.GetCategorizedOptions()
	fmt.Printf("Found %d categories:\n", len(categories))
	for _, cat := range categories {
		fmt.Printf("\n📁 %s (%d fields)\n", cat.Name, len(cat.Variables))
		for _, v := range cat.Variables {
			fmt.Printf("   - %s\n", v.Name)
		}
	}
}
