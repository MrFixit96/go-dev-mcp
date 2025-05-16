package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoTestTool handles the go_test tool execution
func ExecuteGoTestTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	code := ""
	if c, ok := req.Params.Arguments["code"].(string); ok {
		code = c
	}

	testCode := ""
	if tc, ok := req.Params.Arguments["testCode"].(string); ok {
		testCode = tc
	}

	testPattern := ""
	if tp, ok := req.Params.Arguments["testPattern"].(string); ok {
		testPattern = tp
	}

	verbose := false
	if v, ok := req.Params.Arguments["verbose"].(bool); ok {
		verbose = v
	}

	coverage := false
	if c, ok := req.Params.Arguments["coverage"].(bool); ok {
		coverage = c
	}

	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-test-*")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create temp directory: %v", err)), nil
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple Go module
	modCmd := exec.Command("go", "mod", "init", "test")
	modCmd.Dir = tmpDir
	if output, err := modCmd.CombinedOutput(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize Go module: %v\n%s", err, output)), nil
	}

	// Write main code to file if provided
	if code != "" {
		sourceFile := filepath.Join(tmpDir, "main.go")
		if err := os.WriteFile(sourceFile, []byte(code), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to write source code: %v", err)), nil
		}
	}

	// Write test code to file
	testFile := filepath.Join(tmpDir, "main_test.go")
	if testCode == "" {
		// If no test code provided but main code exists, try to create a simple test
		if code != "" {
			testCode = fmt.Sprintf(`package main

import "testing"

func TestMain(t *testing.T) {
	t.Log("No specific tests provided. This is a placeholder test.")
	// Add your test assertions here
}`)
		} else {
			return mcp.NewToolResultError("No code or test code provided"), nil
		}
	}
	
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write test code: %v", err)), nil
	}

	// Prepare test command
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
	args = append(args, "./...")

	cmd := exec.Command("go", args...)
	cmd.Dir = tmpDir

	// Execute command
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// Parse test output to extract coverage information
	coverageInfo := ""
	if coverage && result.Successful {
		coverageInfo = extractCoverageInfo(result.Stdout)
	}

	var message string
	if result.Successful {
		message = "Tests passed"
	} else {
		message = "Tests failed"
	}

	responseContent := fmt.Sprintf(`{
		"success": %t,
		"message": "%s",
		"stdout": "%s",
		"stderr": "%s",
		"exitCode": %d,
		"duration": "%s",
		"coverage": "%s"
	}`, result.Successful, message, result.Stdout, result.Stderr, result.ExitCode, result.Duration.String(), coverageInfo)

	if result.Successful {
		return mcp.NewToolResultText(responseContent), nil
	} else {
		return mcp.NewToolResultError(responseContent), nil
	}
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