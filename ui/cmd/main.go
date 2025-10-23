package main

import (
	"flag"
	"fmt"
	"imperm-ui/pkg/client"
	"imperm-ui/internal"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	mockMode := flag.Bool("mock", false, "Run in mock mode (local client)")
	serverURL := flag.String("server", "http://localhost:8080", "Connect to Imperm server at URL")
	flag.Parse()

	var c client.Client

	if *mockMode {
		fmt.Println("Running in MOCK mode - using local mock client")
		c = client.NewMockClient()
	} else {
		fmt.Printf("Connecting to Imperm server at %s\n", *serverURL)
		c = client.NewHTTPClient(*serverURL)
	}

	// Create and run the Bubble Tea program
	model := ui.NewModel(c)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
