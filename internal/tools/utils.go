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