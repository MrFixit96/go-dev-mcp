package tools

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WorkspaceExecutionStrategy handles execution of commands in Go workspaces
type WorkspaceExecutionStrategy struct{}

// Execute runs commands in the workspace context
func (s *WorkspaceExecutionStrategy) Execute(ctx context.Context, input InputContext, args []string) (*ExecutionResult, error) {
	// Validate workspace path
	if input.WorkspacePath == "" {
		return nil, fmt.Errorf("workspace path is required for workspace execution")
	}

	// Ensure the workspace directory exists and is valid
	if _, err := os.Stat(input.WorkspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace path does not exist: %s", input.WorkspacePath)
	}

	// Check if this is a valid workspace
	if !s.isValidWorkspace(input.WorkspacePath) {
		return nil, fmt.Errorf("path is not a valid Go workspace: %s", input.WorkspacePath)
	}

	// For workspace operations, we need to adapt the execution context
	workingDir, modifiedArgs, err := s.adaptWorkspaceExecution(args, input.WorkspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to adapt workspace execution: %v", err)
	}

	// Prepare command
	cmd := exec.CommandContext(ctx, "go", modifiedArgs...)
	cmd.Dir = workingDir

	return execute(cmd)
}

// adaptWorkspaceExecution determines the proper working directory and command args for workspace operations.
// It analyzes the command type and workspace structure to determine the optimal execution context.
// For code-related commands (fmt, vet, test, build), it finds an appropriate module directory.
// For workspace commands (work), it ensures execution from the workspace root.
// Returns the working directory, modified arguments, and any error encountered.
func (s *WorkspaceExecutionStrategy) adaptWorkspaceExecution(args []string, workspacePath string) (string, []string, error) {
	if len(args) == 0 {
		return workspacePath, args, nil
	}

	// For commands that work on code (fmt, vet, test, build), we need to run them from within a module
	switch args[0] {
	case "fmt", "vet", "test", "build":
		// Check if the command targets "./..." which means "all packages"
		if len(args) > 1 && args[1] == "./..." {
			// Get the first available module to run the command in
			modules, err := s.GetWorkspaceModules(workspacePath)
			if err != nil {
				return workspacePath, args, fmt.Errorf("failed to get workspace modules: %v", err)
			}

			if len(modules) > 0 {
				// Find the first module that exists
				for _, module := range modules {
					// Determine the module path
					var modulePath string
					if module == "./" {
						modulePath = workspacePath
					} else {
						// Remove the "./" prefix for path construction
						cleanModule := strings.TrimPrefix(module, "./")
						modulePath = filepath.Join(workspacePath, cleanModule)
					}

					// Check if the module directory exists and has a go.mod file
					if stat, err := os.Stat(modulePath); err == nil && stat.IsDir() {
						if goModPath := filepath.Join(modulePath, "go.mod"); fileExists(goModPath) {
							// Run the command from within this module directory
							return modulePath, args, nil
						}
					}
				}
			}
		}
	case "work":
		// Workspace commands should run from the workspace root
		return workspacePath, args, nil
	}

	// Default: run from workspace root
	return workspacePath, args, nil
}

// isValidWorkspace checks if the given path is a valid Go workspace.
// It validates workspace structure by looking for either:
// 1. A go.work file in the directory, or
// 2. Multiple go.mod files in immediate subdirectories indicating a multi-module setup
// Returns true if the path represents a valid workspace, false otherwise.
// OPTIMIZED: Uses os.ReadDir for single-level traversal instead of filepath.Walk
// This changes complexity from O(all files in tree) to O(immediate subdirectories)
func (s *WorkspaceExecutionStrategy) isValidWorkspace(path string) bool {
	// Check for go.work file first (cheap operation)
	goWorkPath := filepath.Join(path, "go.work")
	if fileExists(goWorkPath) {
		return true
	}

	// Only check immediate subdirectories, not entire tree
	// This is O(immediate subdirs) instead of O(all files in tree)
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Warning: error reading directory during workspace validation: %v", err)
		return false
	}

	moduleCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Check if this subdirectory has a go.mod
		goModPath := filepath.Join(path, entry.Name(), "go.mod")
		if fileExists(goModPath) {
			moduleCount++
			if moduleCount > 1 {
				return true // Early exit - found multiple modules
			}
		}
	}
	return false
}

// GetWorkspaceModules returns the list of modules in the workspace.
// It delegates to detectWorkspaceModules to discover all modules within the workspace.
// Returns a slice of module paths and any error encountered during detection.
func (s *WorkspaceExecutionStrategy) GetWorkspaceModules(workspacePath string) ([]string, error) {
	return detectWorkspaceModules(workspacePath)
}

// GetWorkspaceInfo provides detailed information about the workspace.
// It validates the workspace structure and collects comprehensive information including:
// - Workspace path
// - Whether a go.work file exists
// - List of all discovered modules
// Returns a WorkspaceInfo struct with the collected data or an error if validation fails.
func (s *WorkspaceExecutionStrategy) GetWorkspaceInfo(workspacePath string) (*WorkspaceInfo, error) {
	if !s.isValidWorkspace(workspacePath) {
		return nil, fmt.Errorf("not a valid workspace: %s", workspacePath)
	}

	info := &WorkspaceInfo{
		Path:      workspacePath,
		HasGoWork: fileExists(filepath.Join(workspacePath, "go.work")),
	}

	modules, err := detectWorkspaceModules(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect modules: %v", err)
	}
	info.Modules = modules

	return info, nil
}

// WorkspaceInfo contains information about a Go workspace
type WorkspaceInfo struct {
	Path      string   `json:"path"`
	HasGoWork bool     `json:"hasGoWork"`
	Modules   []string `json:"modules"`
}
