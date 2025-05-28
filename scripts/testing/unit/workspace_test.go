package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MrFixit96/go-dev-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestWorkspaceDetection(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test case 1: Not a workspace (no go.work file)
	isWS := tools.IsWorkspace(tempDir)
	if isWS {
		t.Error("Expected false for directory without go.work file")
	}

	// Test case 2: Create a go.work file
	workFile := filepath.Join(tempDir, "go.work")
	workContent := `go 1.21

use (
	./module1
	./module2
)
`
	err := os.WriteFile(workFile, []byte(workContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.work file: %v", err)
	}

	isWS = tools.IsWorkspace(tempDir)
	if !isWS {
		t.Error("Expected true for directory with go.work file")
	}

	// Test case 3: Parse workspace file
	modules, err := tools.ParseGoWorkFile(workFile)
	if err != nil {
		t.Fatalf("Failed to parse go.work file: %v", err)
	}

	expectedModules := []string{"./module1", "./module2"}
	if len(modules) != len(expectedModules) {
		t.Errorf("Expected %d modules, got %d", len(expectedModules), len(modules))
	}

	for i, module := range modules {
		if module != expectedModules[i] {
			t.Errorf("Expected module %s, got %s", expectedModules[i], module)
		}
	}
}

func TestInputResolution(t *testing.T) {
	// Create a temporary workspace
	tempDir := t.TempDir()

	// Create go.work file
	workFile := filepath.Join(tempDir, "go.work")
	workContent := `go 1.21

use ./testmodule
`
	err := os.WriteFile(workFile, []byte(workContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.work file: %v", err)
	}

	// Create module directory and go.mod
	moduleDir := filepath.Join(tempDir, "testmodule")
	err = os.MkdirAll(moduleDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}

	modFile := filepath.Join(moduleDir, "go.mod")
	modContent := `module example.com/testmodule

go 1.21
`
	err = os.WriteFile(modFile, []byte(modContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	// Create mock request with workspace_path
	req := mcp.CallToolRequest{
		Params: struct {
			Name      string      `json:"name"`
			Arguments interface{} `json:"arguments,omitempty"`
			Meta      *mcp.Meta   `json:"_meta,omitempty"`
		}{
			Name: "test_tool",
			Arguments: map[string]interface{}{
				"workspace_path": tempDir,
			},
		},
	}

	// Test workspace resolution
	input, err := tools.ResolveInput(req)
	if err != nil {
		t.Fatalf("Failed to resolve workspace input: %v", err)
	}

	if input.Source != tools.SourceWorkspace {
		t.Errorf("Expected source to be SourceWorkspace, got %v", input.Source)
	}

	if input.WorkspacePath != tempDir {
		t.Errorf("Expected workspace path %s, got %s", tempDir, input.WorkspacePath)
	}

	if len(input.WorkspaceModules) != 1 {
		t.Errorf("Expected 1 workspace module, got %d", len(input.WorkspaceModules))
	}

	if input.WorkspaceModules[0] != "./testmodule" {
		t.Errorf("Expected module ./testmodule, got %s", input.WorkspaceModules[0])
	}
}

func TestWorkspaceExecutionStrategy(t *testing.T) {
	// Create a temporary workspace
	tempDir := t.TempDir()

	// Create go.work file
	workFile := filepath.Join(tempDir, "go.work")
	workContent := `go 1.21

use ./testmodule
`
	err := os.WriteFile(workFile, []byte(workContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.work file: %v", err)
	}

	// Create module directory and files
	moduleDir := filepath.Join(tempDir, "testmodule")
	err = os.MkdirAll(moduleDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}

	modFile := filepath.Join(moduleDir, "go.mod")
	modContent := `module example.com/testmodule

go 1.21
`
	err = os.WriteFile(modFile, []byte(modContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	mainFile := filepath.Join(moduleDir, "main.go")
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from workspace module")
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.go file: %v", err)
	}

	// Test workspace strategy
	strategy := &tools.WorkspaceExecutionStrategy{}

	input := tools.InputContext{
		Source:           tools.SourceWorkspace,
		WorkspacePath:    tempDir,
		WorkspaceModules: []string{"./testmodule"},
	}

	// Test execution (go fmt command)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := strategy.Execute(ctx, input, []string{"fmt", "./..."})
	if err != nil {
		t.Errorf("Workspace execution failed: %v", err)
	}

	if !result.Successful {
		t.Errorf("Expected successful execution, got: %s", result.Stderr)
	}
}
