//go:build direct_test
// +build direct_test

// Package main provides a direct execution strategy test entry point
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("=== Running Direct Execution Strategy Test ===")
	runDirectTest()
	fmt.Println("=== Direct Execution Strategy Test Completed ===")
}

// runDirectTest implements a test for the direct execution strategy
func runDirectTest() {
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

	// Execute go run directly
	cmd := exec.Command("go", "run", mainGoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run go command: %v\n%s", err, output)
	}

	// Parse and print the result
	fmt.Printf("Result: %s\n", output)
}
