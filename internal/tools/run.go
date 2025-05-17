package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoRunTool handles the go_run tool execution
func ExecuteGoRunTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	code, ok := req.Params.Arguments["code"].(string)
	if !ok {
		return mcp.NewToolResultError("code must be a string"), nil
	}

	// Get args if provided
	var cmdArgs []string
	if argsInterface, ok := req.Params.Arguments["args"].([]interface{}); ok {
		for _, arg := range argsInterface {
			if strArg, ok := arg.(string); ok {
				cmdArgs = append(cmdArgs, strArg)
			}
		}
	}

	// Get timeout if provided
	timeoutSecs := 30.0
	if timeout, ok := req.Params.Arguments["timeoutSecs"].(float64); ok && timeout > 0 {
		timeoutSecs = timeout
	}

	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-run-*")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create temp directory: %v", err)), nil
	}
	defer os.RemoveAll(tmpDir)

	// Write code to temporary file
	sourceFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write source code: %v", err)), nil
	}

	// Prepare run command
	runArgs := []string{"run", sourceFile}
	runArgs = append(runArgs, cmdArgs...)

	cmd := exec.Command("go", runArgs...)
	cmd.Dir = tmpDir

	// Create a context with timeout
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	// Set the command to use our context
	cmd = exec.CommandContext(runCtx, cmd.Path, cmd.Args[1:]...)
	cmd.Dir = tmpDir

	// Execute command
	result, err := execute(cmd)
	if runCtx.Err() == context.DeadlineExceeded {
		return mcp.NewToolResultError(fmt.Sprintf("Program execution timed out after %.0f seconds", timeoutSecs)), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	var message string
	if result.Successful {
		message = "Program executed successfully"
	} else {
		message = "Program execution failed"
	}

	response := map[string]interface{}{
		"success":  result.Successful,
		"message":  message,
		"stdout":   result.Stdout,
		"stderr":   result.Stderr,
		"exitCode": result.ExitCode,
		"duration": result.Duration.String(),
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_run")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	if result.Successful {
		return mcp.NewToolResultText(string(jsonBytes)), nil
	} else {
		return mcp.NewToolResultError(string(jsonBytes)), nil
	}
}
