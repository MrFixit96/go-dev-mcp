package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecutionResult contains the result of a command execution
type ExecutionResult struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	Duration   time.Duration
	Successful bool
	Command    string
}

// NLMetadata represents natural language metadata for tools
type NLMetadata struct {
	Aliases  []string `json:"aliases"`
	Examples []string `json:"examples"`
}

// ToolNLMetadata maps tool names to their natural language metadata
var ToolNLMetadata = map[string]NLMetadata{
	"go_build": {
		Aliases: []string{"compile", "build", "create executable"},
		Examples: []string{
			"compile this Go code",
			"build this program",
			"create an executable from this code",
			"make a binary from this Go file",
		},
	},
	"go_test": {
		Aliases: []string{"test", "unit test", "run tests"},
		Examples: []string{
			"test this Go code",
			"run unit tests",
			"check if tests pass",
			"verify test coverage",
		},
	},
	"go_run": {
		Aliases: []string{"run", "execute", "start"},
		Examples: []string{
			"run this Go code",
			"execute this program",
			"start this application with arguments",
		},
	},
	"go_mod": {
		Aliases: []string{"dependencies", "modules", "manage dependencies"},
		Examples: []string{
			"initialize a new module",
			"update dependencies",
			"add a new dependency",
			"tidy up module dependencies",
		},
	},
	"go_fmt": {
		Aliases: []string{"format", "beautify", "pretty-print"},
		Examples: []string{
			"format this Go code",
			"beautify this program",
			"fix the formatting of this code",
			"make this code look nice",
		},
	},
	"go_analyze": {
		Aliases: []string{"lint", "check", "validate", "inspect"},
		Examples: []string{
			"analyze this Go code for issues",
			"check for bugs",
			"validate this function",
			"inspect code quality",
		},
	},
}

// execute runs a command and returns the execution result with better logging
func execute(cmd *exec.Cmd) (*ExecutionResult, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Capture the command for debugging
	cmdStr := cmd.String()
	log.Printf("Executing command: %s", cmdStr)

	// Execute with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use CommandContext instead of cmd.Run()
	execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	execCmd.Env = cmd.Env
	execCmd.Dir = cmd.Dir
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	start := time.Now()
	err := execCmd.Run()
	duration := time.Since(start)

	// Check if the context deadline exceeded
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Command timed out after %v: %s", duration, cmdStr)
		return &ExecutionResult{
			Stdout:     stdout.String(),
			Stderr:     "Command execution timed out",
			ExitCode:   -1,
			Duration:   duration,
			Successful: false,
			Command:    cmdStr,
		}, fmt.Errorf("command timed out after %v", duration)
	}

	result := &ExecutionResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Duration:   duration,
		Successful: err == nil,
		Command:    cmdStr,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}

	// Log execution results
	if result.Successful {
		log.Printf("Command succeeded in %v: %s", duration, cmdStr)
	} else {
		log.Printf("Command failed in %v with exit code %d: %s", duration, result.ExitCode, cmdStr)
	}

	return result, nil
}

// FormatCommandResult creates a standardized JSON response for tool executions
func FormatCommandResult(result *ExecutionResult, responseType string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"success":   result.Successful,
		"exitCode":  result.ExitCode,
		"duration":  result.Duration.String(),
		"command":   result.Command,
		"type":      responseType,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if result.Stdout != "" {
		response["stdout"] = result.Stdout
	}

	if result.Stderr != "" {
		response["stderr"] = result.Stderr

		// Add structured error details if there's an error
		if !result.Successful {
			response["errorDetails"] = ParseGoErrors(result.Stderr)
		}
	}

	// Add natural language metadata based on the tool type
	AddNLMetadata(response, responseType)

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err))
	}

	if result.Successful {
		return mcp.NewToolResultText(string(jsonBytes))
	} else {
		return mcp.NewToolResultError(string(jsonBytes))
	}
}

// AddNLMetadata adds natural language metadata to a response map
func AddNLMetadata(response map[string]interface{}, toolName string) map[string]interface{} {
	if metadata, exists := ToolNLMetadata[toolName]; exists {
		response["nlMetadata"] = metadata
	}
	return response
}
