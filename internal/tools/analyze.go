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

// ExecuteGoAnalyzeTool handles the go_analyze tool execution
func ExecuteGoAnalyzeTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Use Parse method for optional boolean parameter
	runVet := mcp.ParseBoolean(req, "vet", true)
	module := mcp.ParseString(req, "module", "") // For workspace module selection

	// Prepare vet args
	args := []string{"vet"}

	// Handle different source types
	switch input.Source {
	case SourceWorkspace:
		// For workspace execution, handle module selection
		if module != "" {
			// Analyze specific module in workspace
			args = append(args, module)
		} else {
			// Analyze all modules in workspace
			args = append(args, "./...")
		}
	case SourceCode:
		// For code analysis, we need to use the existing temporary directory approach
		return executeCodeAnalysis(ctx, input.Code, runVet)
	default:
		// For project execution, analyze all packages
		args = append(args, "./...")
	}

	// Execute using appropriate strategy
	strategy := GetExecutionStrategy(input, args...)
	result, err := strategy.Execute(ctx, input, args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// Process results
	issues := []string{}
	message := "Analysis completed"
	success := result.Successful

	if result.Stdout != "" {
		issues = append(issues, result.Stdout)
	}
	if result.Stderr != "" {
		issues = append(issues, result.Stderr)
	}

	if !success {
		message = "Analysis found issues"
	}

	response := map[string]interface{}{
		"success":  success,
		"message":  message,
		"issues":   issues,
		"duration": result.Duration.String(),
		"source":   input.Source,
		"vet": map[string]interface{}{
			"success": success,
			"issues":  issues,
		},
	}

	if input.Source == SourceWorkspace {
		response["workspacePath"] = input.WorkspacePath
		response["workspaceModules"] = input.WorkspaceModules
		if module != "" {
			response["targetModule"] = module
		}
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_analyze")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	// Even though there might be issues, the analysis tool itself succeeded
	// So we always return a success result with the analysis data
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// executeCodeAnalysis handles analysis for code input using temporary directory
func executeCodeAnalysis(ctx context.Context, code string, runVet bool) (*mcp.CallToolResult, error) {
	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-analyze-*")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create temp directory: %v", err)), nil
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple Go module
	modCmd := exec.Command("go", "mod", "init", "analyze")
	modCmd.Dir = tmpDir
	if output, err := modCmd.CombinedOutput(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize Go module: %v\n%s", err, output)), nil
	}

	// Write code to temporary file
	sourceFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write source code: %v", err)), nil
	}

	// Run analysis
	issues := []string{}
	message := "Analysis completed"
	success := true

	if runVet {
		// Run go vet
		vetCmd := exec.Command("go", "vet", "./...")
		vetCmd.Dir = tmpDir
		vetResult, _ := execute(vetCmd)

		if vetResult.Stdout != "" || vetResult.Stderr != "" {
			if vetResult.Stdout != "" {
				issues = append(issues, vetResult.Stdout)
			}
			if vetResult.Stderr != "" {
				issues = append(issues, vetResult.Stderr)
			}

			if !vetResult.Successful {
				success = false
				message = "Analysis found issues"
			}
		}
	}

	response := map[string]interface{}{
		"success": success,
		"message": message,
		"issues":  issues,
		"source":  SourceCode,
		"vet": map[string]interface{}{
			"success": success,
			"issues":  issues,
		},
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_analyze")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
