// filepath: c:\Users\James\Documents\go-dev-mcp\scripts\testing\legacy\utils\main.go
//go:build legacy
// +build legacy

// Package testutils provides a simple entry point for legacy tests
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	testType := flag.String("type", "both", "Test type to run: 'direct', 'hybrid', or 'both'")
	flag.Parse()
	switch *testType {
	case "direct":
		fmt.Println("=== Running Legacy Direct Execution Strategy Test ===")
		runLegacyDirectTest()
		fmt.Println("=== Legacy Direct Execution Strategy Test Completed ===")
	case "hybrid":
		fmt.Println("=== Running Legacy Hybrid Execution Strategy Test ===")
		runLegacyHybridTest()
		fmt.Println("=== Legacy Hybrid Execution Strategy Test Completed ===")
	case "both":
		fmt.Println("=== Running All Legacy Strategy Tests ===")

		fmt.Println("\n=== Running Legacy Direct Execution Strategy Test ===")
		runLegacyDirectTest()
		fmt.Println("=== Legacy Direct Execution Strategy Test Completed ===")
		fmt.Println("\n=== Running Legacy Hybrid Execution Strategy Test ===")
		runLegacyHybridTest()
		fmt.Println("=== Legacy Hybrid Execution Strategy Test Completed ===")

		fmt.Println("\n=== All Legacy Strategy Tests Completed ===")
	default:
		fmt.Printf("Unknown test type: %s. Use 'direct', 'hybrid', or 'both'\n", *testType)
	}
}

// runLegacyDirectTest runs a direct execution strategy test
func runLegacyDirectTest() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "legacy-direct-test-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go project
	mainGoContent := []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello from legacy direct execution strategy test!")
}
`)
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, mainGoContent, 0644); err != nil {
		fmt.Printf("Failed to write main.go: %v\n", err)
		os.Exit(1)
	}

	// Execute go run directly
	cmd := exec.Command("go", "run", mainGoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run go command: %v\n%s\n", err, output)
		os.Exit(1)
	}

	// Parse and print the result
	fmt.Printf("Result: %s\n", output)
}

// runLegacyHybridTest runs a hybrid execution strategy test
func runLegacyHybridTest() {
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "legacy-hybrid-test-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go project with a module
	mainGoContent := []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello from legacy hybrid execution strategy test!")
}
`)
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, mainGoContent, 0644); err != nil {
		fmt.Printf("Failed to write main.go: %v\n", err)
		os.Exit(1)
	}

	// Initialize a Go module
	goModContent := []byte(`module example.com/hybrid-test

go 1.21
`)
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), goModContent, 0644); err != nil {
		fmt.Printf("Failed to write go.mod: %v\n", err)
		os.Exit(1)
	}

	// Execute go run with the project
	cmd := exec.Command("go", "run", mainGoPath)
	cmd.Dir = tempDir // Set working directory to the temp directory
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to run go command: %v\n%s\n", err, output)
		os.Exit(1)
	}

	// Parse and print the result
	fmt.Printf("Result: %s\n", output)
}
