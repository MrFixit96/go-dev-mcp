package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoModTool handles the go_mod tool execution
func ExecuteGoModTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input using the standard pattern
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Extract parameters using new v0.29.0 API
	command, ok := req.GetArguments()["command"].(string)
	if !ok {
		return mcp.NewToolResultError("command must be a string"), nil
	}

	modulePath := ""
	if path, ok := req.GetArguments()["modulePath"].(string); ok {
		modulePath = path
	}

	module := mcp.ParseString(req, "module", "") // For workspace module selection

	// Validate command
	validCommands := map[string]bool{
		"init":     true,
		"tidy":     true,
		"vendor":   true,
		"verify":   true,
		"why":      true,
		"graph":    true,
		"download": true,
	}

	if !validCommands[command] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid command: %s. Supported commands: init, tidy, vendor, verify, why, graph, download", command)), nil
	}

	// Prepare command arguments
	args := []string{"mod", command}

	// For init command, if modulePath is provided, add it
	if command == "init" && modulePath != "" {
		args = append(args, modulePath)
	}

	// Handle workspace-specific module operations
	if input.Source == SourceWorkspace && module != "" {
		// For workspace operations with specific module, we need to change to that module directory
		// This will be handled by the execution strategy
	}
	// Execute using appropriate strategy
	strategy := GetExecutionStrategy(input, args...)
	result, err := strategy.Execute(ctx, input, args)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Execution error: %v", err)), nil
	}

	// For certain commands, try to read go.mod content from the working directory
	var goModContent string
	if command == "init" || command == "tidy" {
		var modFilePath string
		switch input.Source {
		case SourceWorkspace:
			if module != "" {
				modFilePath = filepath.Join(input.WorkspacePath, module, "go.mod")
			} else {
				modFilePath = filepath.Join(input.WorkspacePath, "go.mod")
			}
		case SourceProjectPath:
			modFilePath = filepath.Join(input.ProjectPath, "go.mod")
		}

		if modFilePath != "" {
			if content, err := os.ReadFile(modFilePath); err == nil {
				goModContent = string(content)
			}
		}
	}

	var message string
	if result.Successful {
		message = fmt.Sprintf("go mod %s succeeded", command)
	} else {
		message = fmt.Sprintf("go mod %s failed", command)
	}

	response := map[string]interface{}{
		"success":  result.Successful,
		"message":  message,
		"stdout":   result.Stdout,
		"stderr":   result.Stderr,
		"exitCode": result.ExitCode,
		"duration": result.Duration.String(),
		"source":   input.Source,
	}

	if goModContent != "" {
		response["goModContent"] = goModContent
	}

	if input.Source == SourceWorkspace {
		response["workspacePath"] = input.WorkspacePath
		response["workspaceModules"] = input.WorkspaceModules
		if module != "" {
			response["targetModule"] = module
		}
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_mod")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	if result.Successful {
		return mcp.NewToolResultText(string(jsonBytes)), nil
	} else {
		return mcp.NewToolResultError(string(jsonBytes)), nil
	}
}
