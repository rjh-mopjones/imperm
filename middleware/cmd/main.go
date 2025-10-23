package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"imperm-middleware/internal/api"
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	mockMode := flag.Bool("mock", false, "Run in mock mode (simulated K8s data)")
	k8sMode := flag.Bool("k8s", false, "Use direct Kubernetes API (instead of Terraform)")
	flag.Parse()

	// Determine mode - Terraform is now the default
	var mode api.HandlerMode
	if *mockMode {
		mode = api.ModeMock
	} else if *k8sMode {
		mode = api.ModeK8s
	} else {
		mode = api.ModeTerraform
	}

	// Create API handler
	handler := api.NewHandler(mode)

	// Setup routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting Imperm server on %s\n", addr)

	switch mode {
	case api.ModeMock:
		fmt.Println("Running in MOCK mode - using simulated data")
	case api.ModeTerraform:
		fmt.Println("Running in TERRAFORM mode - provisioning with Terraform")
	case api.ModeK8s:
		fmt.Println("Running in K8S mode - direct Kubernetes connection")
	}

	log.Fatal(http.ListenAndServe(addr, mux))
}
