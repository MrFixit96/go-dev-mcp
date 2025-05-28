package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// InputSource identifies the source type of input for Go tools
type InputSource int

const (
	// SourceUnknown indicates no recognizable input source
	SourceUnknown InputSource = iota
	// SourceCode indicates the input comes from provided code
	SourceCode
	// SourceProjectPath indicates the input is a Go project directory
	SourceProjectPath
	// SourceHybrid indicates both code and project path are provided
	SourceHybrid
	// SourceWorkspace indicates the input is a Go workspace directory
	SourceWorkspace
)

// InputContext holds information about the input to be processed
type InputContext struct {
	Source           InputSource
	Code             string
	ProjectPath      string
	WorkspacePath    string   // Path to go.work file or workspace root
	WorkspaceModules []string // Discovered module paths within workspace
	MainFile         string
	TestCode         string
}

// ResolveInput determines whether the request contains code or a project path
func ResolveInput(req mcp.CallToolRequest) (InputContext, error) {
	ctx := InputContext{Source: SourceUnknown}

	// Extract code if provided
	if code, ok := req.GetArguments()["code"].(string); ok && code != "" {
		ctx.Code = code
		ctx.Source = SourceCode
	}

	// Extract workspace_path if provided
	if workspacePath, ok := req.GetArguments()["workspace_path"].(string); ok && workspacePath != "" {
		ctx.WorkspacePath = workspacePath
		// Validate workspace path exists
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			return ctx, fmt.Errorf("workspace path does not exist: %s", workspacePath)
		}

		// Detect and validate workspace
		modules, err := detectWorkspaceModules(workspacePath)
		if err != nil {
			return ctx, fmt.Errorf("failed to detect workspace modules: %v", err)
		}
		ctx.WorkspaceModules = modules
		ctx.Source = SourceWorkspace
	}

	// Extract project_path if provided
	if path, ok := req.GetArguments()["project_path"].(string); ok && path != "" {
		ctx.ProjectPath = path
		// Validate path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return ctx, fmt.Errorf("project path does not exist: %s", path)
		}

		// If workspace_path is also provided, it takes precedence
		if ctx.Source != SourceWorkspace {
			// If both code and project_path are provided, use hybrid source
			if ctx.Code != "" {
				ctx.Source = SourceHybrid
			} else {
				ctx.Source = SourceProjectPath
			}
		}
	}

	// Extract test code if provided
	if testCode, ok := req.GetArguments()["testCode"].(string); ok && testCode != "" {
		ctx.TestCode = testCode
	}

	// Validate input
	if ctx.Source == SourceUnknown {
		return ctx, fmt.Errorf("at least one of 'code', 'project_path', or 'workspace_path' must be provided")
	}

	// Set default main file
	if mainFile, ok := req.GetArguments()["mainFile"].(string); ok && mainFile != "" {
		ctx.MainFile = mainFile
	} else {
		ctx.MainFile = "main.go"
	}

	return ctx, nil
}

// detectWorkspaceModules detects and validates modules in a Go workspace
func detectWorkspaceModules(workspacePath string) ([]string, error) {
	var modules []string

	// First, look for go.work file
	goWorkPath := filepath.Join(workspacePath, "go.work")
	if fileExists(goWorkPath) {
		// Parse go.work file to get module paths
		workModules, err := ParseGoWorkFile(goWorkPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse go.work file: %v", err)
		}
		modules = append(modules, workModules...)
	} else {
		// Look for go.mod files in subdirectories
		err := filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors, continue walking
			}

			if info.Name() == "go.mod" {
				// Get relative path from workspace root
				relPath, err := filepath.Rel(workspacePath, filepath.Dir(path))
				if err != nil {
					return nil // Skip this module
				}
				if relPath == "." {
					relPath = "./"
				} else {
					relPath = "./" + relPath
				}
				modules = append(modules, relPath)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk workspace directory: %v", err)
		}
	}

	// Return empty slice if no modules found - this is valid for an empty workspace
	return modules, nil
}

// ParseGoWorkFile parses a go.work file and returns the module paths
func ParseGoWorkFile(goWorkPath string) ([]string, error) {
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil, err
	}

	var modules []string
	lines := strings.Split(string(content), "\n")

	inUseBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		// Handle single-line use directive: "use ./module"
		if strings.HasPrefix(line, "use ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				modulePath := parts[1]
				modules = append(modules, modulePath)
			}
			continue
		}

		// Handle parenthesized use block: "use ("
		if strings.HasPrefix(line, "use (") || line == "use (" {
			inUseBlock = true
			continue
		}

		// Handle end of parenthesized block
		if inUseBlock && line == ")" {
			inUseBlock = false
			continue
		}

		// Handle module paths inside parenthesized block
		if inUseBlock {
			// Remove any trailing comments
			if idx := strings.Index(line, "//"); idx != -1 {
				line = strings.TrimSpace(line[:idx])
			}
			if line != "" {
				modules = append(modules, line)
			}
		}
	}

	return modules, nil
}

// IsWorkspace checks if a given path contains a workspace (go.work file or multiple modules)
func IsWorkspace(path string) bool {
	// Check for go.work file
	goWorkPath := filepath.Join(path, "go.work")
	if fileExists(goWorkPath) {
		return true
	}

	// Check for multiple go.mod files
	moduleCount := 0
	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "go.mod" {
			moduleCount++
			if moduleCount > 1 {
				return fmt.Errorf("found multiple modules") // Early termination
			}
		}
		return nil
	})

	return moduleCount > 1
}
