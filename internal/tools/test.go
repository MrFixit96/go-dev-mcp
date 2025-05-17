package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoTestTool handles the go_test tool execution
func ExecuteGoTestTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Extract parameters
	testPattern := mcp.ParseString(req, "testPattern", "")
	verbose := mcp.ParseBoolean(req, "verbose", false)
	coverage := mcp.ParseBoolean(req, "coverage", false)

	// Prepare test args
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if coverage {
		args = append(args, "-cover")
	}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}

	// Always add ./... to run all tests in the directory
	args = append(args, "./...")
	// Execute using appropriate strategy
	strategy := GetExecutionStrategy(input, args...)
	result, err := strategy.Execute(ctx, input, args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// Create structured response with proper error handling
	if result.Successful {
		return formatTestSuccess(result, coverage), nil
	} else {
		return formatTestError(result), nil
	}
}

// generateSimpleTest creates a simple test case when none is provided
func generateSimpleTest() string {
	return `package main

import "testing"

func TestMain(t *testing.T) {
	t.Log("No specific tests provided. This is a placeholder test.")
	// Add your test assertions here
}`
}

// formatTestSuccess creates a structured success response for tests
func formatTestSuccess(result *ExecutionResult, withCoverage bool) *mcp.CallToolResult {
	// Parse test output to extract coverage and test statistics
	coverageInfo := ""
	if withCoverage && result.Successful {
		coverageInfo = extractCoverageInfo(result.Stdout)
	}

	testStats := parseTestStats(result.Stdout)

	response := map[string]interface{}{
		"success":   true,
		"message":   "Tests passed",
		"output":    result.Stdout,
		"duration":  result.Duration.String(),
		"coverage":  coverageInfo,
		"testStats": testStats,
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_test")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err))
	}

	return mcp.NewToolResultText(string(jsonBytes))
}

// formatTestError creates a structured error response for tests
func formatTestError(result *ExecutionResult) *mcp.CallToolResult {
	// Parse test errors for more context
	errorDetails := parseTestErrors(result.Stdout, result.Stderr)

	response := map[string]interface{}{
		"success":      false,
		"message":      "Tests failed",
		"output":       result.Stdout,
		"stderr":       result.Stderr,
		"exitCode":     result.ExitCode,
		"duration":     result.Duration.String(),
		"errorDetails": errorDetails,
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_test")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err))
	}

	return mcp.NewToolResultError(string(jsonBytes))
}

// extractCoverageInfo extracts coverage information from test output
func extractCoverageInfo(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "coverage:") {
			return strings.TrimSpace(line)
		}
	}
	return "Coverage information not available"
}

// parseTestStats extracts test statistics from output
func parseTestStats(output string) string {
	// In a real implementation, this would parse test output
	// for test count, run time, etc. into structured JSON
	return `{"count": "unknown", "passed": "unknown", "failed": "unknown"}`
}

// parseTestErrors extracts structured information from test failures
func parseTestErrors(stdout, stderr string) string {
	// In a real implementation, this would parse test failure output
	// into structured JSON with file, line, and error details
	return stderr
}
