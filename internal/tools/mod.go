package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoModTool handles the go_mod tool execution
func ExecuteGoModTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	command, ok := req.Params.Arguments["command"].(string)
	if !ok {
		return mcp.NewToolResultError("command must be a string"), nil
	}

	modulePath := ""
	if path, ok := req.Params.Arguments["modulePath"].(string); ok {
		modulePath = path
	}

	code := ""
	if c, ok := req.Params.Arguments["code"].(string); ok {
		code = c
	}

	// Validate command
	validCommands := map[string]bool{
		"init":     true,
		"tidy":     true,
		"vendor":   true,
		"verify":   true,
		"why":      true,
		"graph":    true,
		"download": true,
	}

	if !validCommands[command] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid command: %s. Supported commands: init, tidy, vendor, verify, why, graph, download", command)), nil
	}

	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-mod-*")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create temp directory: %v", err)), nil
	}
	defer os.RemoveAll(tmpDir)

	// If code is provided, write it to a file
	if code != "" {
		sourceFile := filepath.Join(tmpDir, "main.go")
		if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write source code: %v", err)), nil
		}
	}

	// Prepare command arguments
	args := []string{"mod", command}
	
	// For init command, if modulePath is provided, add it
	if command == "init" && modulePath != "" {
		args = append(args, modulePath)
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = tmpDir

	// Execute command
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// For certain commands, also read the go.mod file content
	var goModContent string
	if command == "init" || command == "tidy" {
		modFile := filepath.Join(tmpDir, "go.mod")
		if content, err := os.ReadFile(modFile); err == nil {
			goModContent = string(content)
		}
	}

	var message string
	if result.Successful {
		message = fmt.Sprintf("go mod %s succeeded", command)
	} else {
		message = fmt.Sprintf("go mod %s failed", command)
	}

	responseContent := fmt.Sprintf(`{
		"success": %t,
		"message": "%s",
		"stdout": "%s",
		"stderr": "%s",
		"exitCode": %d,
		"duration": "%s"
	}`, result.Successful, message, result.Stdout, result.Stderr, result.ExitCode, result.Duration.String())
	
	if goModContent != "" {
		// Append goModContent to the response
		responseContent = responseContent[:len(responseContent)-2] + fmt.Sprintf(`,
		"goModContent": %q
	}`, goModContent)
	}

	if result.Successful {
		return mcp.NewToolResultText(responseContent), nil
	} else {
		return mcp.NewToolResultError(responseContent), nil
	}
}