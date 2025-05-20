// Package hybrid_test provides testing for the hybrid execution strategy
package hybrid_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/MrFixit96/go-dev-mcp/internal/tools"
)

// Run executes a simple test for the hybrid execution strategy
func Run() {
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
	fmt.Println("1. Creating input context with both code and project path")

	// Create input context with both code and project path (hybrid mode)
	inputCtx := tools.InputContext{
		Code:        modifiedCode,
		ProjectPath: tempDir,
		Source:      tools.SourceHybrid, // Using hybrid source type
	}

	// Get the appropriate strategy
	fmt.Println("2. Determining execution strategy")
	strategy := tools.GetExecutionStrategy(inputCtx, "run")

	fmt.Printf("3. Selected strategy type: %T\n", strategy)
	// Execute the strategy
	fmt.Println("4. Executing the hybrid strategy")
	result, err := strategy.Execute(context.Background(), inputCtx, []string{"run"})
	if err != nil {
		log.Fatalf("Execution failed: %v", err)
	}

	// Check the results
	fmt.Println("5. Execution results:")
	fmt.Printf("   - Exit code: %d\n", result.ExitCode)
	fmt.Printf("   - Stdout: %s\n", result.Stdout)

	if result.ExitCode == 0 && result.Stdout == "Hello from hybrid strategy\n" {
		fmt.Println("✅ TEST PASSED: Hybrid strategy correctly used the modified code!")
	} else {
		fmt.Println("❌ TEST FAILED: Output doesn't match expectations")
	}
}
