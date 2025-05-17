package tools

import (
	"context"
	"encoding/json"
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
	response := map[string]interface{}{
		"success":    true,
		"message":    "Compilation successful",
		"outputPath": outputPath,
		"duration":   result.Duration.String(),
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_build")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err))
	}

	return mcp.NewToolResultText(string(jsonBytes))
}

// formatBuildError creates a structured error response
func formatBuildError(result *ExecutionResult) *mcp.CallToolResult {
	// Parse Go build errors for more context
	errorDetails := parseGoBuildErrors(result.Stderr)

	response := map[string]interface{}{
		"success":      false,
		"message":      "Compilation failed",
		"stderr":       result.Stderr,
		"exitCode":     result.ExitCode,
		"duration":     result.Duration.String(),
		"errorDetails": errorDetails,
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_build")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err))
	}

	return mcp.NewToolResultError(string(jsonBytes))
}

// parseGoBuildErrors extracts meaningful error information from Go build output
func parseGoBuildErrors(stderr string) string {
	// In a real implementation, this would parse the error output
	// and structure it by file, line, error type, etc.
	// For now, we're just returning the raw stderr
	return stderr
}
