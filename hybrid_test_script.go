package main

import (
"fmt"
"log"
"os"
"path/filepath"

"github.com/MrFixit96/go-dev-mcp/internal/tools"
)

func main() {
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
}`)

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
hybridResult, err := tools.ExecuteGoRunTool(tools.ToolRequest{
Input: map[string]interface{}{
"code":         modifiedCode,
"project_path": tempDir,
},
})
if err != nil {
log.Fatalf("Hybrid execution failed: %v", err)
}

fmt.Printf("Hybrid result: %s\n", hybridResult.Result.(tools.ExecutionResult).Stdout)

// Check if the hybrid strategy worked correctly
if hybridResult.Result.(tools.ExecutionResult).Stdout == "message from hybrid strategy\n" {
fmt.Println("\n✅ SUCCESS: Hybrid strategy correctly used the modified code!")
} else {
fmt.Println("\n❌ FAILURE: Hybrid strategy did not produce expected output")
}

// Check metadata for strategy information
if metadata, ok := hybridResult.Metadata.(map[string]interface{}); ok {
if strategyType, ok := metadata["strategyType"]; ok {
fmt.Printf("Strategy type used: %v\n", strategyType)
}
}
}
