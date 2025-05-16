package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoBuildTool handles the go_build tool execution
func ExecuteGoBuildTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	code, ok := req.Params.Arguments["code"].(string)
	if !ok {
		return mcp.NewToolResultError("code must be a string"), nil
	}

	outputPath := ""
	if path, ok := req.Params.Arguments["outputPath"].(string); ok {
		outputPath = path
	}

	buildTags := ""
	if tags, ok := req.Params.Arguments["buildTags"].(string); ok {
		buildTags = tags
	}

	mainFile := "main.go"
	if file, ok := req.Params.Arguments["mainFile"].(string); ok && file != "" {
		mainFile = file
	}

	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-build-*")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create temp directory: %v", err)), nil
	}
	defer os.RemoveAll(tmpDir)

	// Write code to temporary file
	sourceFile := filepath.Join(tmpDir, mainFile)
	if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write source code: %v", err)), nil
	}

	// Prepare build command
	cmd := exec.Command("go", "build")
	if buildTags != "" {
		cmd.Args = append(cmd.Args, "-tags", buildTags)
	}
	if outputPath != "" {
		cmd.Args = append(cmd.Args, "-o", outputPath)
	} else {
		outputPath = filepath.Join(tmpDir, "output")
		cmd.Args = append(cmd.Args, "-o", outputPath)
	}
	cmd.Args = append(cmd.Args, sourceFile)
	cmd.Dir = tmpDir

	// Execute command
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	var message string
	if result.Successful {
		message = "Compilation successful"
	} else {
		message = "Compilation failed"
	}

	responseContent := fmt.Sprintf(`{
		"success": %t,
		"message": "%s",
		"stdout": "%s",
		"stderr": "%s",
		"exitCode": %d,
		"duration": "%s"
	}`, result.Successful, message, result.Stdout, result.Stderr, result.ExitCode, result.Duration.String())

	if result.Successful {
		return mcp.NewToolResultText(responseContent), nil
	} else {
		return mcp.NewToolResultError(responseContent), nil
	}
}