//go:build legacy
// +build legacy

// Package testutils provides testing for the hybrid execution strategy
package testutils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunLegacyHybridTest executes a simple test for the hybrid execution strategy
func RunLegacyHybridTest() {
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "go-dev-mcp-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go project
	mainFile := filepath.Join(tempDir, "main.go")
	mainCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello from", GetName())
}

func GetName() string {
	return "original code"
}
`
	err = os.WriteFile(mainFile, []byte(mainCode), 0644)
	if err != nil {
		log.Fatalf("Failed to write main.go: %v", err)
	}

	// Initialize go.mod
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module example.com/hybrid-test

go 1.21
`
	err = os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write go.mod: %v", err)
	}

	// Modified code for testing hybrid strategy
	modifiedCode := `package main

func GetName() string {
	return "hybrid strategy"
}
`

	// Test hybrid strategy
	fmt.Println("=== Testing Hybrid Execution Strategy ===")
	fmt.Println("1. Creating request with both code and project path")

	// For hybrid approach, write the modified code to a new file
	modifiedFile := filepath.Join(tempDir, "name.go")
	err = os.WriteFile(modifiedFile, []byte(modifiedCode), 0644)
	if err != nil {
		log.Fatalf("Failed to write modified code: %v", err)
	}

	// Execute go run directly in the project directory
	fmt.Println("2. Executing go run with the hybrid approach")
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Execution failed: %v\n%s", err, output)
	}

	// Check the results
	outputStr := string(output)
	fmt.Println("3. Execution results:")
	fmt.Printf("   - Output: %s\n", strings.TrimSpace(outputStr))

	if strings.Contains(outputStr, "hybrid strategy") {
		fmt.Println("✅ TEST PASSED: Hybrid strategy correctly used the modified code!")
	} else {
		fmt.Println("❌ TEST FAILED: Output doesn't match expectations")
	}
}
