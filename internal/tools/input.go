package tools

import (
	"fmt"
	"log"
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
		// Validate and sanitize the path
		validatedPath, err := validatePath(workspacePath)
		if err != nil {
			return ctx, fmt.Errorf("invalid workspace path: %v", err)
		}
		ctx.WorkspacePath = validatedPath

		// Validate workspace path exists
		if _, err := os.Stat(validatedPath); os.IsNotExist(err) {
			return ctx, fmt.Errorf("workspace path does not exist: %s", validatedPath)
		}

		// Detect and validate workspace
		modules, err := detectWorkspaceModules(validatedPath)
		if err != nil {
			return ctx, fmt.Errorf("failed to detect workspace modules: %v", err)
		}
		ctx.WorkspaceModules = modules
		ctx.Source = SourceWorkspace
	}

	// Extract project_path if provided
	if path, ok := req.GetArguments()["project_path"].(string); ok && path != "" {
		// Validate and sanitize the path
		validatedPath, err := validatePath(path)
		if err != nil {
			return ctx, fmt.Errorf("invalid project path: %v", err)
		}
		ctx.ProjectPath = validatedPath

		// Validate path exists
		if _, err := os.Stat(validatedPath); os.IsNotExist(err) {
			return ctx, fmt.Errorf("project path does not exist: %s", validatedPath)
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

// detectWorkspaceModules detects and validates modules in a Go workspace.
// It searches for modules using two methods:
// 1. If a go.work file exists, it parses the file to extract module paths
// 2. If no go.work file exists, it walks the directory tree with a depth limit to find go.mod files
// The function returns relative paths for all discovered modules.
// Returns a slice of module paths (relative to workspace root) and any error encountered.
// OPTIMIZED: Uses filepath.WalkDir with depth limit instead of filepath.Walk
// This prevents scanning deep vendor/ or node_modules/ trees
func detectWorkspaceModules(workspacePath string) ([]string, error) {
	var modules []string

	// First, look for go.work file
	goWorkPath := filepath.Join(workspacePath, "go.work")
	if fileExists(goWorkPath) {
		// Parse go.work file to get module paths (no walk needed)
		workModules, err := ParseGoWorkFile(goWorkPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse go.work file: %v", err)
		}
		modules = append(modules, workModules...)
	} else {
		// Look for go.mod files in subdirectories with depth limit
		// WalkDir is more efficient than Walk (uses os.ReadDir internally)
		const maxDepth = 3 // Limit depth to avoid scanning deep trees like vendor/
		err := filepath.WalkDir(workspacePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				// Log the error but skip the problematic path
				log.Printf("Warning: skipping path during module detection: %v", err)
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Calculate current depth
			relPath, err := filepath.Rel(workspacePath, path)
			if err != nil {
				return nil
			}
			depth := 0
			if relPath != "." {
				depth = strings.Count(relPath, string(filepath.Separator)) + 1
			}

			// Skip directories that are too deep or commonly ignored
			if d.IsDir() {
				name := d.Name()
				// Skip hidden directories, vendor, node_modules, and other common non-Go directories
				if strings.HasPrefix(name, ".") ||
					name == "vendor" ||
					name == "node_modules" ||
					name == "testdata" {
					return filepath.SkipDir
				}
				// Skip if beyond max depth
				if depth >= maxDepth {
					return filepath.SkipDir
				}
			}

			if !d.IsDir() && d.Name() == "go.mod" {
				// Get relative path from workspace root
				relPath, err := filepath.Rel(workspacePath, filepath.Dir(path))
				if err != nil {
					// Log warning but continue processing other modules
					log.Printf("Warning: failed to get relative path for module at %s: %v", path, err)
					return nil
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

// ParseGoWorkFile parses a go.work file and returns the module paths.
// It handles both single-line use directives (use ./module) and multi-line use blocks:
//
//	use (
//	    ./module1
//	    ./module2
//	)
//
// The function ignores comments and empty lines, extracting only valid module paths.
// Returns a slice of module paths as specified in the go.work file.
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

// IsWorkspace checks if a given path contains a workspace (go.work file or multiple modules).
// It uses two criteria to determine if a path represents a Go workspace:
// 1. Presence of a go.work file in the directory
// 2. Multiple go.mod files in immediate subdirectories (indicating multi-module setup)
// Returns true if either condition is met, false otherwise.
// OPTIMIZED: Uses os.ReadDir for single-level traversal instead of filepath.Walk
// This changes complexity from O(all files in tree) to O(immediate subdirectories)
func IsWorkspace(path string) bool {
	// Check for go.work file first (cheap operation)
	goWorkPath := filepath.Join(path, "go.work")
	if fileExists(goWorkPath) {
		return true
	}

	// Only check immediate subdirectories, not entire tree
	// This is O(immediate subdirs) instead of O(all files in tree)
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Warning: error reading directory during workspace check: %v", err)
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
