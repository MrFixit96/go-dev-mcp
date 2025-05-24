// Package main provides the test runner for execution strategies
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Functions for direct execution strategy
func runDirectTest() {
	fmt.Println("=== Running Direct Execution Strategy Test ===")
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "direct-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go project
	mainGoContent := []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello from direct execution strategy test!")
}
`)
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, mainGoContent, 0644); err != nil {
		log.Fatalf("Failed to write main.go: %v", err)
	}
	// Run the test directly with the main.go file we created// Execute go run directly
	cmd := exec.Command("go", "run", mainGoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run go command: %v\n%s", err, output)
	}

	// Parse and print the result
	fmt.Printf("Result: %s\n", output)

	fmt.Println("=== Direct Execution Strategy Test Completed ===")
}

// Functions for hybrid execution strategy
func runHybridTest() {
	fmt.Println("=== Running Hybrid Execution Strategy Test ===")
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "hybrid-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go project
	mainGoContent := []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello from hybrid execution strategy test!")
}
`)
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, mainGoContent, 0644); err != nil {
		log.Fatalf("Failed to write main.go: %v", err)
	}

	// Initialize a Go module
	goModContent := []byte(`module example.com/hybrid-test

go 1.21
`)
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), goModContent, 0644); err != nil {
		log.Fatalf("Failed to write go.mod: %v", err)
	}
	// Run the test
	// Execute go run directly in the project directory
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run go command: %v\n%s", err, output)
	}

	// Parse and print the result
	fmt.Printf("Result: %s\n", output)

	fmt.Println("=== Hybrid Execution Strategy Test Completed ===")
}

func main() {
	testType := flag.String("type", "both", "Test type to run: 'direct', 'hybrid', or 'both'")
	flag.Parse()

	switch *testType {
	case "direct":
		runDirectTest()
	case "hybrid":
		runHybridTest()
	case "both":
		runDirectTest()
		fmt.Println()
		runHybridTest()
	default:
		fmt.Printf("Unknown test type: %s\n", *testType)
		fmt.Println("Available options: 'direct', 'hybrid', or 'both'")
		os.Exit(1)
	}
}
