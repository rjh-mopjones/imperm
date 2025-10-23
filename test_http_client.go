package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Environment struct {
	Name        string
	Namespace   string
	Status      string
	Age         time.Time
	Pods        []Pod
	Deployments []Deployment
}

type Pod struct {
	Name      string
	Namespace string
	Status    string
	Ready     string
	Restarts  int
	Age       time.Time
	CPU       string
	Memory    string
}

type Deployment struct {
	Name      string
	Namespace string
	Ready     string
	UpToDate  int
	Available int
	Age       time.Time
}

func main() {
	fmt.Println("Testing HTTP client connection to server...")

	resp, err := http.Get("http://localhost:8080/api/environments")
	if err != nil {
		fmt.Printf("❌ Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✓ Response status: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		fmt.Printf("❌ Unexpected status code\n")
		return
	}

	var envs []Environment
	if err := json.NewDecoder(resp.Body).Decode(&envs); err != nil {
		fmt.Printf("❌ Error decoding JSON: %v\n", err)
		return
	}

	fmt.Printf("✓ Successfully decoded %d environments\n\n", len(envs))

	for _, env := range envs {
		fmt.Printf("Environment: %s\n", env.Name)
		fmt.Printf("  Namespace: %s\n", env.Namespace)
		fmt.Printf("  Status: %s\n", env.Status)
		fmt.Printf("  Age: %v\n", env.Age)
		fmt.Printf("  Pods: %d\n", len(env.Pods))
		fmt.Printf("  Deployments: %d\n", len(env.Deployments))
		fmt.Println()
	}
}
