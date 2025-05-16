package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoFmtTool handles the go_fmt tool execution
func ExecuteGoFmtTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	code, ok := req.Params.Arguments["code"].(string)
	if !ok {
		return mcp.NewToolResultError("code must be a string"), nil
	}

	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-fmt-*")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create temp directory: %v", err)), nil
	}
	defer os.RemoveAll(tmpDir)

	// Write code to temporary file
	sourceFile := filepath.Join(tmpDir, "input.go")
	if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write source code: %v", err)), nil
	}

	// Run gofmt
	cmd := exec.Command("gofmt", "-w", sourceFile)
	cmd.Dir = tmpDir

	// Execute command
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// Read the formatted file
	formattedCode, err := os.ReadFile(sourceFile)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read formatted code: %v", err)), nil
	}

	var message string
	if result.Successful {
		message = "Code formatted successfully"
	} else {
		message = "Formatting failed"
	}
	
	// Determine if code was changed
	codeChanged := string(formattedCode) != code
	
	responseContent := fmt.Sprintf(`{
		"success": %t,
		"message": "%s",
		"code": %q,
		"stdout": "%s",
		"stderr": "%s",
		"codeChanged": %t
	}`, result.Successful, message, string(formattedCode), result.Stdout, result.Stderr, codeChanged)
	
	return mcp.NewToolResultText(responseContent), nil
}