//go:build hybrid_test
// +build hybrid_test

// filepath: c:\Users\James\Documents\go-dev-mcp\scripts\testing\hybrid_test.go
// Package main provides a hybrid execution strategy test entry point
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("=== Running Hybrid Execution Strategy Test ===")
	runHybridTest()
	fmt.Println("=== Hybrid Execution Strategy Test Completed ===")
}

// runHybridTest implements a test for the hybrid execution strategy
func runHybridTest() {
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

	// Execute go run directly in the project directory
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run go command: %v\n%s", err, output)
	}

	// Parse and print the result
	fmt.Printf("Result: %s\n", output)
}
