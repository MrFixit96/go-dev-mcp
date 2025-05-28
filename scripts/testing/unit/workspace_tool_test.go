package unit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/MrFixit96/go-dev-mcp/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestExecuteGoWorkspaceTool(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		workspacePath string
		expectedError bool
		setupFunc     func(string) error
		cleanupFunc   func(string) error
	}{
		{
			name:          "missing command parameter",
			command:       "",
			workspacePath: "/tmp/test-workspace",
			expectedError: true,
		},
		{
			name:          "missing workspace_path parameter",
			command:       "init",
			workspacePath: "",
			expectedError: true,
		},
		{
			name:          "unknown command",
			command:       "unknown",
			workspacePath: "/tmp/test-workspace",
			expectedError: true,
		},
		{
			name:          "workspace init success",
			command:       "init",
			workspacePath: "/tmp/test-workspace-init",
			expectedError: false,
			cleanupFunc: func(path string) error {
				return os.RemoveAll(path)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tt.workspacePath); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Cleanup
			defer func() {
				if tt.cleanupFunc != nil {
					tt.cleanupFunc(tt.workspacePath)
				}
			}()

			// Create mock request
			req := createMockRequest(map[string]interface{}{
				"command":        tt.command,
				"workspace_path": tt.workspacePath,
			})

			// Execute
			result, err := tools.ExecuteGoWorkspaceTool(context.Background(), req)
			// Verify
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectedError {
				if !result.IsError {
					t.Errorf("Expected error but got success")
				}
			} else {
				if result.IsError {
					t.Errorf("Expected success but got error: %v", result.Content)
				}
			}
		})
	}
}

func TestWorkspaceCommands(t *testing.T) {
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "test-workspace")

	t.Run("init command", func(t *testing.T) {
		req := createMockRequest(map[string]interface{}{
			"command":        "init",
			"workspace_path": workspacePath,
		})
		result, err := tools.ExecuteGoWorkspaceTool(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			textContent, ok := result.Content[0].(mcp.TextContent)
			if ok {
				t.Errorf("Init command failed: %v", textContent.Text)
			} else {
				t.Errorf("Init command failed: %v", result.Content)
			}
		}

		// Check if go.work file was created
		goWorkPath := filepath.Join(workspacePath, "go.work")
		if !fileExists(goWorkPath) {
			t.Errorf("go.work file was not created")
		}
	})

	t.Run("info command", func(t *testing.T) {
		req := createMockRequest(map[string]interface{}{
			"command":        "info",
			"workspace_path": workspacePath,
		})
		result, err := tools.ExecuteGoWorkspaceTool(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			textContent, ok := result.Content[0].(mcp.TextContent)
			if ok {
				t.Errorf("Info command failed: %v", textContent.Text)
			} else {
				t.Errorf("Info command failed: %v", result.Content)
			}
		}

		// Verify response structure
		var response map[string]interface{}
		textContent, ok := result.Content[0].(mcp.TextContent)
		if !ok {
			t.Errorf("Expected TextContent, got %T", result.Content[0])
			return
		}
		if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
			t.Errorf("Failed to parse response JSON: %v", err)
		}

		if success, ok := response["success"].(bool); !ok || !success {
			t.Errorf("Expected success=true in response")
		}
	})

	t.Run("sync command", func(t *testing.T) {
		req := createMockRequest(map[string]interface{}{
			"command":        "sync",
			"workspace_path": workspacePath,
		})

		result, err := tools.ExecuteGoWorkspaceTool(context.Background(), req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		// Sync might fail if no modules are present, but should not error
		if result.IsError {
			// This is expected if workspace has no modules
			textContent, ok := result.Content[0].(mcp.TextContent)
			if ok {
				t.Logf("Sync failed as expected (no modules): %v", textContent.Text)
			} else {
				t.Logf("Sync failed as expected (no modules): %v", result.Content)
			}
		}
	})
}

func TestWorkspaceUseCommand(t *testing.T) {
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "test-workspace")
	modulePath := filepath.Join(tempDir, "test-module")

	// Create a simple module
	if err := os.MkdirAll(modulePath, 0755); err != nil {
		t.Fatalf("Failed to create module directory: %v", err)
	}

	// Create go.mod file
	goModContent := "module example.com/test-module\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(modulePath, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Initialize workspace first
	initReq := createMockRequest(map[string]interface{}{
		"command":        "init",
		"workspace_path": workspacePath,
	})

	_, err := tools.ExecuteGoWorkspaceTool(context.Background(), initReq)
	if err != nil {
		t.Fatalf("Failed to initialize workspace: %v", err)
	}

	// Test use command
	useReq := createMockRequest(map[string]interface{}{
		"command":        "use",
		"workspace_path": workspacePath,
		"modules":        []interface{}{modulePath},
	})
	result, err := tools.ExecuteGoWorkspaceTool(context.Background(), useReq)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.IsError {
		textContent, ok := result.Content[0].(mcp.TextContent)
		if ok {
			t.Errorf("Use command failed: %v", textContent.Text)
		} else {
			t.Errorf("Use command failed: %v", result.Content)
		}
	}
}

func createMockRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: struct {
			Name      string      `json:"name"`
			Arguments interface{} `json:"arguments,omitempty"`
			Meta      *mcp.Meta   `json:"_meta,omitempty"`
		}{
			Name:      "go_workspace",
			Arguments: args,
		},
	}
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Benchmark tests for workspace operations
func BenchmarkWorkspaceInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempDir := b.TempDir()
		workspacePath := filepath.Join(tempDir, "bench-workspace")

		req := createMockRequest(map[string]interface{}{
			"command":        "init",
			"workspace_path": workspacePath,
		})

		_, err := tools.ExecuteGoWorkspaceTool(context.Background(), req)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkWorkspaceInfo(b *testing.B) {
	tempDir := b.TempDir()
	workspacePath := filepath.Join(tempDir, "bench-workspace")

	// Initialize workspace once
	initReq := createMockRequest(map[string]interface{}{
		"command":        "init",
		"workspace_path": workspacePath,
	})
	_, err := tools.ExecuteGoWorkspaceTool(context.Background(), initReq)
	if err != nil {
		b.Fatalf("Failed to initialize workspace: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := createMockRequest(map[string]interface{}{
			"command":        "info",
			"workspace_path": workspacePath,
		})

		_, err := tools.ExecuteGoWorkspaceTool(context.Background(), req)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
