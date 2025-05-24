package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoRunTool handles the go_run tool execution
func ExecuteGoRunTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	// Get args if provided
	var cmdArgs []string
	if argsObj, ok := req.GetArguments()["args"].(map[string]interface{}); ok {
		// Handle args as an object (per the mcp.WithObject parameter)
		for _, v := range argsObj {
			if strArg, ok := v.(string); ok {
				cmdArgs = append(cmdArgs, strArg)
			}
		}
	} else if argsArray, ok := req.GetArguments()["args"].([]interface{}); ok {
		// Handle args as an array (for backward compatibility)
		for _, arg := range argsArray {
			if strArg, ok := arg.(string); ok {
				cmdArgs = append(cmdArgs, strArg)
			}
		}
	}

	// Get timeout if provided using Parse method
	timeoutSecs := mcp.ParseFloat64(req, "timeoutSecs", 30.0)

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	// Prepare run args
	args := []string{"run"}

	if input.Source == SourceCode {
		args = append(args, input.MainFile)
	} else {
		args = append(args, "./...")
	}

	// Add command-line arguments
	args = append(args, cmdArgs...)
	// Execute using appropriate strategy
	strategy := GetExecutionStrategy(input, args...)
	result, err := strategy.Execute(execCtx, input, args)
	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		return mcp.NewToolResultError(fmt.Sprintf("Program execution timed out after %.0f seconds", timeoutSecs)), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// Format response
	var message string
	if result.Successful {
		message = "Program executed successfully"
	} else {
		message = "Program execution failed"
	}

	response := map[string]interface{}{
		"success":  result.Successful,
		"message":  message,
		"stdout":   result.Stdout,
		"stderr":   result.Stderr,
		"exitCode": result.ExitCode,
		"duration": result.Duration.String(),
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_run")
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
