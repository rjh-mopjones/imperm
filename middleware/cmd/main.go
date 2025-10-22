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
	flag.Parse()

	// Create API handler
	handler := api.NewHandler(*mockMode)

	// Setup routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Starting Imperm server on %s\n", addr)
	if *mockMode {
		fmt.Println("Running in MOCK mode - using simulated data")
	} else {
		fmt.Println("Running in PRODUCTION mode - connecting to Kubernetes")
	}

	log.Fatal(http.ListenAndServe(addr, mux))
}
