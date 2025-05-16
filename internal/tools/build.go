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
	// Extract parameters using helper functions
	code := mcp.ParseString(req, "code", "")
	if code == "" {
		return mcp.NewToolResultError("No source code provided. Parameter 'code' is required"), nil
	}

	outputPath := mcp.ParseString(req, "outputPath", "")
	buildTags := mcp.ParseString(req, "buildTags", "")
	mainFile := mcp.ParseString(req, "mainFile", "main.go")

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

	// Format response with structured error handling
	if result.Successful {
		return formatBuildSuccess(result, outputPath), nil
	} else {
		return formatBuildError(result), nil
	}
}

// formatBuildSuccess creates a structured success response
func formatBuildSuccess(result *ExecutionResult, outputPath string) *mcp.CallToolResult {
	return mcp.NewToolResultText(fmt.Sprintf(`{
		"success": true,
		"message": "Compilation successful",
		"outputPath": "%s",
		"duration": "%s"
	}`, outputPath, result.Duration.String()))
}

// formatBuildError creates a structured error response
func formatBuildError(result *ExecutionResult) *mcp.CallToolResult {
	// Parse Go build errors for more context
	errorDetails := parseGoBuildErrors(result.Stderr)
	
	return mcp.NewToolResultError(fmt.Sprintf(`{
		"success": false,
		"message": "Compilation failed",
		"stderr": "%s",
		"exitCode": %d,
		"duration": "%s",
		"errorDetails": %s
	}`, result.Stderr, result.ExitCode, result.Duration.String(), errorDetails))
}

// parseGoBuildErrors extracts meaningful error information from Go build output
func parseGoBuildErrors(stderr string) string {
	// In a real implementation, this would parse the error output 
	// and structure it by file, line, error type, etc.
	// For now, we're just returning the raw stderr as JSON string
	return fmt.Sprintf(`"%s"`, stderr)
}