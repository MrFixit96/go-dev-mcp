//go:build legacy
// +build legacy

// Package testutils provides direct testing for Go tools execution
package testutils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// RunLegacyDirectTest executes a test for direct execution strategy
func RunLegacyDirectTest() {
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

	// Create request for project path only
	req1 := mcp.CallToolRequest{
		Params: struct {
			Name      string      `json:"name"`
			Arguments interface{} `json:"arguments,omitempty"`
			Meta      *mcp.Meta   `json:"_meta,omitempty"`
		}{
			Name: "go_run",
			Arguments: map[string]interface{}{
				"project_path": tempDir,
			},
		},
	}

	// Execute go run directly
	cmd := exec.Command("go", "run", tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Project path execution failed: %v\n%s", err, output)
	}

	// Output the results
	fmt.Printf("Project path execution succeeded with output: %s\n",
		strings.TrimSpace(string(output)))

	// Now test with hybrid (both code and project path)
	fmt.Println("\n=== Testing with hybrid strategy (code + project path) ===")

	// For hybrid approach, write the modified code temporarily
	modifiedFile := filepath.Join(tempDir, "message.go")
	err = os.WriteFile(modifiedFile, []byte(modifiedCode), 0644)
	if err != nil {
		log.Fatalf("Failed to write modified code: %v", err)
	}

	// Execute go run directly with all files
	cmd2 := exec.Command("go", "run", ".")
	cmd2.Dir = tempDir
	hybridOutput, err := cmd2.CombinedOutput()
	if err != nil {
		log.Fatalf("Hybrid execution failed: %v\n%s", err, hybridOutput)
	}

	// Output the results
	fmt.Printf("Hybrid execution succeeded with output: %s\n",
		strings.TrimSpace(string(hybridOutput)))

	if strings.Contains(string(hybridOutput), "message from hybrid strategy") {
		fmt.Println("\n✅ SUCCESS: Hybrid strategy executed correctly!")
	} else {
		fmt.Println("\n❌ FAILURE: Hybrid strategy did not apply code changes")
	}
}
