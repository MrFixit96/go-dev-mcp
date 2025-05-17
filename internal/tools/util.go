package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// AddNLMetadata adds natural language metadata to help AI assistants interpret results.
// This allows the assistant to formulate more natural responses based on tool output.
func AddNLMetadata(response map[string]interface{}, toolName string) {
	// Only add this if it doesn't already exist
	if _, exists := response["nl_metadata"]; !exists {
		// Map of tool specific natural language templates
		templates := map[string]map[string]string{
			"go_build": {
				"success": "The code was successfully compiled",
				"error":   "There were errors compiling the code",
			},
			"go_test": {
				"success": "All tests passed successfully",
				"error":   "Some tests failed",
			},
			"go_run": {
				"success": "The program ran successfully",
				"error":   "The program execution failed",
			},
			"go_fmt": {
				"success": "The code was formatted according to Go standards",
				"error":   "There were issues formatting the code",
			},
			"go_mod": {
				"success": "The module operation was successful",
				"error":   "The module operation failed",
			},
			"go_analyze": {
				"success": "The code analysis completed without finding any issues",
				"error":   "The code analysis found potential issues",
			},
		}

		nlMetadata := map[string]string{}
		if template, ok := templates[toolName]; ok {
			if response["success"].(bool) {
				nlMetadata["result"] = template["success"]
			} else {
				nlMetadata["result"] = template["error"]
			}
		}

		response["nl_metadata"] = nlMetadata
	}
}

// FormatMCPResult converts a tool result to JSON and creates the appropriate MCP result object
func FormatMCPResult(result *ExecutionResult, data map[string]interface{}) (*mcp.CallToolResult, error) {
	// Add natural language metadata
	AddNLMetadata(data, "generic")

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	if data["success"].(bool) {
		return mcp.NewToolResultText(string(jsonBytes)), nil
	} else {
		return mcp.NewToolResultError(string(jsonBytes)), nil
	}
}

// extractFileNameFromPath extracts the filename from a path
func extractFileNameFromPath(path string) string {
	return filepath.Base(path)
}

// formatPathsForOutputs formats a slice of paths for output
func formatPathsForOutput(paths []string, basePath string) []string {
	formattedPaths := make([]string, len(paths))
	for i, path := range paths {
		if strings.HasPrefix(path, basePath) {
			// Make path relative to basePath
			relPath, err := filepath.Rel(basePath, path)
			if err == nil {
				formattedPaths[i] = relPath
				continue
			}
		}
		formattedPaths[i] = path
	}
	return formattedPaths
}
