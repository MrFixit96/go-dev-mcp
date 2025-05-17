package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoFmtTool handles the go_fmt tool execution
func ExecuteGoFmtTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Prepare format args
	args := []string{"fmt"}

	// For project execution, add recursive flag
	if input.Source == SourceProjectPath {
		args = append(args, "./...")
	} else {
		// For code, we'll use "gofmt" directly as it's better for formatting individual files
		var formattedCode string

		// Create temporary directory for code-based formatting
		tmpDir, err := os.MkdirTemp("", "go-fmt-*")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create temp directory: %v", err)), nil
		}
		defer os.RemoveAll(tmpDir)

		// Write code to temporary file
		sourceFile := filepath.Join(tmpDir, "input.go")
		if err := os.WriteFile(sourceFile, []byte(input.Code), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write source code: %v", err)), nil
		}

		// Run gofmt
		cmd := exec.Command("gofmt", "-w", sourceFile)
		cmd.Dir = tmpDir

		result, err := execute(cmd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
		}

		// Read the formatted file
		formattedBytes, err := os.ReadFile(sourceFile)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to read formatted code: %v", err)), nil
		}
		formattedCode = string(formattedBytes)

		// Determine if code was changed
		codeChanged := formattedCode != input.Code

		response := map[string]interface{}{
			"success":     result.Successful,
			"message":     "Code formatted successfully",
			"code":        formattedCode,
			"stdout":      result.Stdout,
			"stderr":      result.Stderr,
			"codeChanged": codeChanged,
		}

		// Add natural language metadata
		AddNLMetadata(response, "go_fmt")

		jsonBytes, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
	// Execute using appropriate strategy for project path
	strategy := GetExecutionStrategy(input, args...)
	result, err := strategy.Execute(ctx, input, args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// For project formatting, the output is different
	if input.Source == SourceProjectPath {
		// Parse the formatted files from output
		filesFormatted := parseFormattedFiles(result.Stdout)

		response := map[string]interface{}{
			"success":        result.Successful,
			"message":        "Project formatting completed",
			"stdout":         result.Stdout,
			"stderr":         result.Stderr,
			"filesFormatted": filesFormatted,
		}

		// Add natural language metadata
		AddNLMetadata(response, "go_fmt")

		jsonBytes, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	return mcp.NewToolResultError("Invalid execution path"), nil
}

// parseFormattedFiles extracts information about formatted files from go fmt output
func parseFormattedFiles(output string) []string {
	// In a real implementation, this would parse go fmt output
	// to determine which files were formatted
	lines := strings.Split(output, "\n")
	var files []string

	for _, line := range lines {
		if strings.Contains(line, ".go") {
			files = append(files, strings.TrimSpace(line))
		}
	}

	return files
}
