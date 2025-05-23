// package mock provides predefined response templates for MCP tool operations.
package mock

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ResponseTemplate defines a standard template for tool responses
type ResponseTemplate struct {
	// Success indicates whether the operation was successful
	Success bool
	// Message provides human-readable details about the operation
	Message string
	// AdditionalFields contains extra fields specific to each tool
	AdditionalFields map[string]interface{}
}

// ToolResponseTemplates maps tool names to their response templates
var ToolResponseTemplates = map[string]map[ResponseScenario]ResponseTemplate{
	"go_fmt": {
		ScenarioSuccess: {
			Success: true,
			Message: "Code formatted successfully",
			AdditionalFields: map[string]interface{}{
				"formattedCode": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World from Go Development MCP Server!")
}`,
				"codeChanged": true,
				"metadata": map[string]interface{}{
					"strategyType": "hybrid",
				},
			},
		},
		ScenarioFailure: {
			Success: false,
			Message: "Failed to format code",
			AdditionalFields: map[string]interface{}{
				"errors": []string{"syntax error: unexpected semicolon or newline"},
			},
		},
	},
	"go_build": {
		ScenarioSuccess: {
			Success: true,
			Message: "Build successful",
			AdditionalFields: map[string]interface{}{
				"outputPath": "/path/to/executable",
				"buildTime":  "1.25s",
			},
		},
		ScenarioFailure: {
			Success: false,
			Message: "Build failed",
			AdditionalFields: map[string]interface{}{
				"stderr": "main.go:5:2: undefined: fmt",
				"errors": []string{
					"main.go:5:2: undefined: fmt",
					"could not compile package",
				},
			},
		},
	},
	"go_run": {
		ScenarioSuccess: {
			Success: true,
			Message: "Execution successful",
			AdditionalFields: map[string]interface{}{
				"stdout":   "Hello, World from Go Development MCP Server!",
				"exitCode": 0,
				"runTime":  "0.032s",
			},
		},
		ScenarioFailure: {
			Success: false,
			Message: "Execution failed",
			AdditionalFields: map[string]interface{}{
				"stderr":   "panic: runtime error: index out of range",
				"exitCode": 1,
			},
		},
	},
	"go_test": {
		ScenarioSuccess: {
			Success: true,
			Message: "Tests passed successfully",
			AdditionalFields: map[string]interface{}{
				"output":    "ok  	package/path	0.015s",
				"coverage":  "coverage: 85.7% of statements",
				"testStats": map[string]int{"passed": 15, "failed": 0, "skipped": 2},
			},
		},
		ScenarioFailure: {
			Success: false,
			Message: "Tests failed",
			AdditionalFields: map[string]interface{}{
				"output":    "--- FAIL: TestSomething (0.00s)",
				"coverage":  "coverage: 43.2% of statements",
				"testStats": map[string]int{"passed": 10, "failed": 5, "skipped": 0},
				"failures":  []string{"TestSomething", "TestAnotherThing"},
			},
		},
	},
	"go_mod": {
		ScenarioSuccess: {
			Success: true,
			Message: "Module operation successful",
			AdditionalFields: map[string]interface{}{
				"output": "go: finding module for package github.com/example/package\ngo: downloading github.com/example/package v1.0.0",
			},
		},
		ScenarioFailure: {
			Success: false,
			Message: "Module operation failed",
			AdditionalFields: map[string]interface{}{
				"stderr": "go: module github.com/nonexistent/package: not found",
			},
		},
	},
}

// GetResponseForTool creates a CallToolResult for the given tool and scenario
func GetResponseForTool(toolName string, scenario ResponseScenario, customFields map[string]interface{}) (*mcp.CallToolResult, error) {
	// Get tool templates, fall back to generic templates if not found
	toolTemplates, ok := ToolResponseTemplates[toolName]
	if !ok {
		// Default generic responses
		switch scenario {
		case ScenarioSuccess:
			return createGenericResponse(true, fmt.Sprintf("%s completed successfully", toolName), nil)
		case ScenarioFailure:
			return createGenericResponse(false, fmt.Sprintf("%s operation failed", toolName), nil)
		default:
			return createGenericResponse(false, "Unknown tool or scenario", nil)
		}
	}

	// Get specific template for this scenario, fall back to success/failure if not found
	template, ok := toolTemplates[scenario]
	if !ok {
		// Fall back based on scenario type
		if scenario == ScenarioSuccess {
			template = toolTemplates[ScenarioSuccess]
		} else {
			// Default to failure template if available, otherwise generic failure
			if failTemplate, hasFailure := toolTemplates[ScenarioFailure]; hasFailure {
				template = failTemplate
			} else {
				return createGenericResponse(false, fmt.Sprintf("%s operation failed", toolName), nil)
			}
		}
	}

	// Merge template with custom fields
	responseMap := map[string]interface{}{
		"success": template.Success,
		"message": template.Message,
	}

	// Add all additional fields from template
	for k, v := range template.AdditionalFields {
		responseMap[k] = v
	}

	// Override with any custom fields
	for k, v := range customFields {
		responseMap[k] = v
	}

	// Convert to JSON
	jsonData, err := json.Marshal(responseMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response template: %w", err)
	}

	// Create MCP CallToolResult
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Text: string(jsonData)},
		},
	}

	return result, nil
}

// createGenericResponse creates a generic response when no template is found
func createGenericResponse(success bool, message string, customFields map[string]interface{}) (*mcp.CallToolResult, error) {
	responseMap := map[string]interface{}{
		"success": success,
		"message": message,
	}

	// Add custom fields if provided
	for k, v := range customFields {
		responseMap[k] = v
	}

	// Convert to JSON
	jsonData, err := json.Marshal(responseMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal generic response: %w", err)
	}

	// Create MCP CallToolResult
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Text: string(jsonData)},
		},
	}

	return result, nil
}

// CreateJSONRPCErrorResponse creates a JSON-RPC 2.0 compliant error response
func CreateJSONRPCErrorResponse(code int, message string, data interface{}) ([]byte, error) {
	response := struct {
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
		Error: struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		}{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: nil, // For notifications, ID is null
	}

	return json.Marshal(response)
}

// CreateJSONRPCSuccessResponse creates a JSON-RPC 2.0 compliant success response
func CreateJSONRPCSuccessResponse(id interface{}, result interface{}) ([]byte, error) {
	response := struct {
		JSONRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	return json.Marshal(response)
}
