package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoAnalyzeTool handles the go_analyze tool execution
func ExecuteGoAnalyzeTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	code, ok := req.Params.Arguments["code"].(string)
	if !ok {
		return mcp.NewToolResultError("code must be a string"), nil
	}

	runVet := true
	if vet, ok := req.Params.Arguments["vet"].(bool); ok {
		runVet = vet
	}

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

	responseContent := fmt.Sprintf(`{
		"success": %t,
		"message": "%s",
		"issues": %v,
		"vet": {
			"success": %t,
			"issues": %v
		}
	}`, success, message, issues, success, issues)

	if success {
		return mcp.NewToolResultText(responseContent), nil
	} else {
		// Even though there are issues, the analysis tool itself succeeded
		// So we return a success result with the analysis data
		return mcp.NewToolResultText(responseContent), nil
	}
}