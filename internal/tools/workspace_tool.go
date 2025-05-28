package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteGoWorkspaceTool handles the go_workspace tool execution.
// It processes workspace management commands including init, use, sync, edit, vendor, and info.
// The function validates required parameters and dispatches to appropriate subcommand handlers.
// Returns a formatted tool result with command output or error information.
func ExecuteGoWorkspaceTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract command parameter
	command := mcp.ParseString(req, "command", "")
	if command == "" {
		return mcp.NewToolResultError("command parameter is required"), nil
	}

	// Extract workspace_path parameter
	workspacePath := mcp.ParseString(req, "workspace_path", "")
	if workspacePath == "" {
		return mcp.NewToolResultError("workspace_path parameter is required"), nil
	}

	// Execute the appropriate workspace command
	switch command {
	case "init":
		return executeWorkspaceInit(ctx, workspacePath, req)
	case "use":
		return executeWorkspaceUse(ctx, workspacePath, req)
	case "sync":
		return executeWorkspaceSync(ctx, workspacePath, req)
	case "edit":
		return executeWorkspaceEdit(ctx, workspacePath, req)
	case "vendor":
		return executeWorkspaceVendor(ctx, workspacePath, req)
	case "info":
		return executeWorkspaceInfo(ctx, workspacePath, req)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown workspace command: %s", command)), nil
	}
}

// executeWorkspaceInit initializes a new Go workspace.
// It creates the workspace directory if it doesn't exist and runs 'go work init'
// with any specified modules. The function supports both empty workspace creation
// and initialization with predefined modules.
// Returns a tool result with initialization status and workspace information.
func executeWorkspaceInit(ctx context.Context, workspacePath string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Ensure the workspace directory exists
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create workspace directory: %v", err)), nil
	}

	// Get modules to include in the workspace
	modules := []string{}
	if modulesArg, ok := req.GetArguments()["modules"]; ok {
		if modulesList, ok := modulesArg.([]interface{}); ok {
			for _, module := range modulesList {
				if moduleStr, ok := module.(string); ok {
					modules = append(modules, moduleStr)
				}
			}
		}
	}

	// Prepare go work init command
	args := []string{"work", "init"}
	args = append(args, modules...)

	// Execute command
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = workspacePath

	// Execute with timeout if set
	if deadline, ok := ctx.Deadline(); ok {
		execCtx, cancel := context.WithTimeout(ctx, time.Until(deadline))
		defer cancel()
		cmd = exec.CommandContext(execCtx, cmd.Path, cmd.Args[1:]...)
		cmd.Dir = workspacePath
	}
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Workspace init failed: %v", err)), nil
	}

	// Format response
	response := map[string]interface{}{
		"success":  result.Successful,
		"message":  "Workspace initialized successfully",
		"path":     workspacePath,
		"duration": result.Duration.String(),
		"stdout":   result.Stdout,
		"stderr":   result.Stderr,
		"modules":  modules,
	}

	// Add natural language metadata
	AddNLMetadata(response, "go_workspace")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// executeWorkspaceUse adds modules to an existing workspace.
// It validates that the workspace exists and executes 'go work use' to add specified modules.
// The function requires an existing go.work file and at least one module to add.
// Returns a tool result with the operation status and module information.
func executeWorkspaceUse(ctx context.Context, workspacePath string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if workspace exists
	goWorkPath := filepath.Join(workspacePath, "go.work")
	if !fileExists(goWorkPath) {
		return mcp.NewToolResultError("go.work file not found. Initialize the workspace first with 'init' command."), nil
	}

	// Get modules to add
	modules := []string{}
	if modulesArg, ok := req.GetArguments()["modules"]; ok {
		if modulesList, ok := modulesArg.([]interface{}); ok {
			for _, module := range modulesList {
				if moduleStr, ok := module.(string); ok {
					modules = append(modules, moduleStr)
				}
			}
		}
	}

	if len(modules) == 0 {
		return mcp.NewToolResultError("modules parameter is required for 'use' command"), nil
	}

	// Prepare go work use command
	args := []string{"work", "use"}
	args = append(args, modules...)

	// Execute command
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = workspacePath
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Workspace use failed: %v", err)), nil
	}

	response := map[string]interface{}{
		"success":  result.Successful,
		"message":  "Modules added to workspace successfully",
		"modules":  modules,
		"duration": result.Duration.String(),
		"stdout":   result.Stdout,
		"stderr":   result.Stderr,
	}

	AddNLMetadata(response, "go_workspace")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// executeWorkspaceSync synchronizes workspace dependencies.
// It runs 'go work sync' to ensure all workspace modules have consistent dependency versions.
// The function requires an existing workspace with a go.work file.
// Returns a tool result with synchronization status and output.
func executeWorkspaceSync(ctx context.Context, workspacePath string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if workspace exists
	goWorkPath := filepath.Join(workspacePath, "go.work")
	if !fileExists(goWorkPath) {
		return mcp.NewToolResultError("go.work file not found. Initialize the workspace first."), nil
	}

	// Execute go work sync
	cmd := exec.CommandContext(ctx, "go", "work", "sync")
	cmd.Dir = workspacePath
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Workspace sync failed: %v", err)), nil
	}

	response := map[string]interface{}{
		"success":  result.Successful,
		"message":  "Workspace synchronized successfully",
		"duration": result.Duration.String(),
		"stdout":   result.Stdout,
		"stderr":   result.Stderr,
	}

	AddNLMetadata(response, "go_workspace")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// executeWorkspaceEdit opens the go.work file for editing or modifies it programmatically.
// It runs 'go work edit -json' to retrieve the current workspace configuration in JSON format.
// This provides structured access to workspace settings and module configurations.
// Returns a tool result with the workspace configuration data.
func executeWorkspaceEdit(ctx context.Context, workspacePath string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if workspace exists
	goWorkPath := filepath.Join(workspacePath, "go.work")
	if !fileExists(goWorkPath) {
		return mcp.NewToolResultError("go.work file not found. Initialize the workspace first."), nil
	}

	// For now, just execute go work edit command
	cmd := exec.CommandContext(ctx, "go", "work", "edit", "-json")
	cmd.Dir = workspacePath
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Workspace edit failed: %v", err)), nil
	}

	response := map[string]interface{}{
		"success":       result.Successful,
		"message":       "Workspace configuration retrieved",
		"configuration": result.Stdout,
		"duration":      result.Duration.String(),
		"stderr":        result.Stderr,
	}

	AddNLMetadata(response, "go_workspace")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// executeWorkspaceVendor vendors all workspace dependencies.
// It runs 'go work vendor' to create a vendor directory containing all dependencies
// for all modules in the workspace. This enables offline builds and dependency isolation.
// Returns a tool result with vendoring status and operation details.
func executeWorkspaceVendor(ctx context.Context, workspacePath string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if workspace exists
	goWorkPath := filepath.Join(workspacePath, "go.work")
	if !fileExists(goWorkPath) {
		return mcp.NewToolResultError("go.work file not found. Initialize the workspace first."), nil
	}

	// Execute go work vendor
	cmd := exec.CommandContext(ctx, "go", "work", "vendor")
	cmd.Dir = workspacePath
	result, err := execute(cmd)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Workspace vendor failed: %v", err)), nil
	}

	response := map[string]interface{}{
		"success":  result.Successful,
		"message":  "Workspace dependencies vendored successfully",
		"duration": result.Duration.String(),
		"stdout":   result.Stdout,
		"stderr":   result.Stderr,
	}

	AddNLMetadata(response, "go_workspace")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// executeWorkspaceInfo provides information about the workspace.
// It collects and returns comprehensive workspace information including structure,
// module listings, and configuration details. The information is formatted as JSON
// for easy consumption and display.
// Returns a tool result with detailed workspace information.
func executeWorkspaceInfo(ctx context.Context, workspacePath string, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	strategy := &WorkspaceExecutionStrategy{}
	info, err := strategy.GetWorkspaceInfo(workspacePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get workspace info: %v", err)), nil
	}
	// Convert to JSON for better formatting
	infoJSON, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format workspace info: %v", err)), nil
	}

	response := map[string]interface{}{
		"success":       true,
		"message":       "Workspace information retrieved",
		"workspaceInfo": string(infoJSON),
		"info":          info,
	}

	AddNLMetadata(response, "go_workspace")

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
