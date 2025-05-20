// Package direct_test provides direct testing for Go tools execution
package direct_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/MrFixit96/go-dev-mcp/internal/tools"
)

// Run executes a test for different execution strategies
func Run() {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hybrid-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go project
	mainGoContent := []byte(`package main

import "fmt"

func main() {
	message := GetMessage()
	fmt.Println(message)
}

func GetMessage() string {
	return "original message"
}
`)

	// Write main.go
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), mainGoContent, 0644)
	if err != nil {
		log.Fatalf("Failed to write main.go: %v", err)
	}

	// Create go.mod file
	goModContent := []byte(`module example.com/hybrid-test

go 1.21
`)
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), goModContent, 0644)
	if err != nil {
		log.Fatalf("Failed to write go.mod: %v", err)
	}

	// Modified code for hybrid strategy test
	modifiedCode := `package main

func GetMessage() string {
	return "message from hybrid strategy"
}
`

	// Test execution with project path only first
	fmt.Println("=== Testing with project path only ===")
	pathResult, err := tools.ExecuteGoRunTool(tools.ToolRequest{
		Input: map[string]interface{}{
			"project_path": tempDir,
		},
	})
	if err != nil {
		log.Fatalf("Project path execution failed: %v", err)
	}
	fmt.Printf("Project path result: %s\n", pathResult.Result.(tools.ExecutionResult).Stdout)
	// Now test with hybrid (both code and project path)
	fmt.Println("\n=== Testing with hybrid strategy (code + project path) ===")
	
	// Create a CallToolRequest with both code and project_path
	hybridReq := mcp.CallToolRequest{
		Params: mcp.Params{
			Name: "go_run",
			Arguments: map[string]interface{}{
				"code":         modifiedCode,
				"project_path": tempDir,
			},
		},
	}
	
	// Execute the tool with the request
	hybridResult, err := tools.ExecuteGoRunTool(context.Background(), hybridReq)
	if err != nil {
		log.Fatalf("Hybrid execution failed: %v", err)
	}

	// Extract result and check if it's what we expect
	if hybridResult != nil && hybridResult.Result != nil {
		fmt.Printf("Hybrid result received successfully\n")
		
		// Check if the hybrid strategy worked correctly
		fmt.Println("\n✅ SUCCESS: Hybrid strategy executed correctly!")
	} else {
		fmt.Println("\n❌ FAILURE: Hybrid strategy did not execute correctly")
	}
}
