package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoBuildTool handles the go_build tool execution
func ExecuteGoBuildTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	// Extract parameters
	outputPath := mcp.ParseString(req, "outputPath", "")
	buildTags := mcp.ParseString(req, "buildTags", "")
	
	// Prepare build args
	args := []string{"build"}
	if buildTags != "" {
		args = append(args, "-tags", buildTags)
	}
	if outputPath != "" {
		args = append(args, "-o", outputPath)
	} else if input.Source == SourceCode {
		// Only set default output for code input
		outputPath = "output"
		args = append(args, "-o", outputPath)
	}
	
	// For code execution, add the main file
	if input.Source == SourceCode {
		args = append(args, input.MainFile)
	} else {
		// For project execution, add ./... to build all packages
		args = append(args, "./...")
	}
	
	// Execute using appropriate strategy
	strategy := GetExecutionStrategy(input)
	result, err := strategy.Execute(ctx, input, args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}
	
	// Format response with structured error handling
	if result.Successful {
		return formatBuildSuccess(result, outputPath, input), nil
	} else {
		return formatBuildError(result), nil
	}
}

// formatBuildSuccess creates a structured success response
func formatBuildSuccess(result *ExecutionResult, outputPath string, input InputContext) *mcp.CallToolResult {
	// Determine the full output path
	var fullOutputPath string
	if filepath.IsAbs(outputPath) {
		fullOutputPath = outputPath
	} else if input.Source == SourceProjectPath {
		fullOutputPath = filepath.Join(input.ProjectPath, outputPath)
	} else {
		fullOutputPath = outputPath + " (in temporary directory)"
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    "Compilation successful",
		"outputPath": fullOutputPath,
		"duration":   result.Duration.String(),
		"source":     input.Source,
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
