// +build ignore

package main

import (
	"fmt"
	"imperm-ui/pkg/terraform"
	"os"
)

func main() {
	pwd, _ := os.Getwd()
	fmt.Printf("Working directory: %s\n\n", pwd)

	// Test direct path
	fmt.Println("Testing direct load...")
	loader := terraform.NewLoader(terraform.LoaderConfig{
		Sources: []string{"../terraform/modules/k8s-namespace"},
	})

	err := loader.Load()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else {
		cats := loader.GetCategorizedOptions()
		fmt.Printf("✓ Success! Found %d categories\n", len(cats))
		for _, cat := range cats {
			fmt.Printf("  - %s: %d variables\n", cat.Name, len(cat.Variables))
		}
	}
}
