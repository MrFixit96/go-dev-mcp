package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ExecutionStrategy defines the interface for different execution strategies
type ExecutionStrategy interface {
	Execute(ctx context.Context, input InputContext, args []string) (*ExecutionResult, error)
}

// CodeExecutionStrategy handles execution of provided code
type CodeExecutionStrategy struct{}

// Execute creates a temporary environment for code execution
func (s *CodeExecutionStrategy) Execute(ctx context.Context, input InputContext, args []string) (*ExecutionResult, error) {
	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-exec-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple Go module
	modCmd := exec.Command("go", "mod", "init", "temp")
	modCmd.Dir = tmpDir
	if output, err := modCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to initialize Go module: %v\n%s", err, output)
	}

	// Write main source file
	sourceFile := filepath.Join(tmpDir, input.MainFile)
	if err := os.WriteFile(sourceFile, []byte(input.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source code: %v", err)
	}

	// Write test file if test code is provided
	if input.TestCode != "" {
		testFileName := "main_test.go"
		if input.MainFile != "main.go" {
			// Generate test filename based on main file
			ext := filepath.Ext(input.MainFile)
			base := input.MainFile[:len(input.MainFile)-len(ext)]
			testFileName = base + "_test" + ext
		}

		testFile := filepath.Join(tmpDir, testFileName)
		if err := os.WriteFile(testFile, []byte(input.TestCode), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test code: %v", err)
		}
	}

	// Prepare command
	cmd := exec.Command("go", args...)
	cmd.Dir = tmpDir

	// Execute command with timeout if set
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		execCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd = exec.CommandContext(execCtx, cmd.Path, cmd.Args[1:]...)
		cmd.Dir = tmpDir
	}

	return execute(cmd)
}

// ProjectExecutionStrategy handles execution in an existing project directory
type ProjectExecutionStrategy struct{}

// Execute runs commands in the project directory
func (s *ProjectExecutionStrategy) Execute(ctx context.Context, input InputContext, args []string) (*ExecutionResult, error) {
	// Prepare command
	cmd := exec.Command("go", args...)
	cmd.Dir = input.ProjectPath

	// Execute command with timeout if set
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		execCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd = exec.CommandContext(execCtx, cmd.Path, cmd.Args[1:]...)
		cmd.Dir = input.ProjectPath
	}

	return execute(cmd)
}

// HybridExecutionStrategy handles hybrid scenarios where both code and project path are provided
type HybridExecutionStrategy struct{}

// Execute runs commands with knowledge of both code and project structure
func (s *HybridExecutionStrategy) Execute(ctx context.Context, input InputContext, args []string) (*ExecutionResult, error) {
	// Create temporary directory for the operation
	tmpDir, err := os.MkdirTemp("", "go-hybrid-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First, copy relevant files from the project to understand dependencies
	err = copyRelevantProjectFiles(input.ProjectPath, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to copy project files: %v", err)
	}

	// Write main source file with the provided code
	sourceFile := filepath.Join(tmpDir, input.MainFile)
	if err := os.WriteFile(sourceFile, []byte(input.Code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source code: %v", err)
	}

	// Write test file if test code is provided
	if input.TestCode != "" {
		testFileName := "main_test.go"
		if input.MainFile != "main.go" {
			// Generate test filename based on main file
			ext := filepath.Ext(input.MainFile)
			base := input.MainFile[:len(input.MainFile)-len(ext)]
			testFileName = base + "_test" + ext
		}

		testFile := filepath.Join(tmpDir, testFileName)
		if err := os.WriteFile(testFile, []byte(input.TestCode), 0644); err != nil {
			return nil, fmt.Errorf("failed to write test code: %v", err)
		}
	}

	// Update the module dependencies
	modTidyCmd := exec.Command("go", "mod", "tidy")
	modTidyCmd.Dir = tmpDir
	if output, err := modTidyCmd.CombinedOutput(); err != nil {
		// Non-fatal error, just log it
		fmt.Printf("Warning: failed to tidy module dependencies: %v\n%s", err, output)
	}

	// Prepare command
	cmd := exec.Command("go", args...)
	cmd.Dir = tmpDir

	// Execute command with timeout if set
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		execCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd = exec.CommandContext(execCtx, cmd.Path, cmd.Args[1:]...)
		cmd.Dir = tmpDir
	}

	return execute(cmd)
}

// copyRelevantProjectFiles copies essential files from a project to understand dependencies
func copyRelevantProjectFiles(srcDir, destDir string) error {
	// Copy go.mod and go.sum if they exist
	essentialFiles := []string{"go.mod", "go.sum"}
	for _, file := range essentialFiles {
		srcPath := filepath.Join(srcDir, file)
		if _, err := os.Stat(srcPath); err == nil {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %v", file, err)
			}

			destPath := filepath.Join(destDir, file)
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %v", file, err)
			}
		}
	}

	return nil
}

// CommandType represents different types of Go commands
type CommandType int

const (
	// CommandUnknown indicates an unknown command type
	CommandUnknown CommandType = iota
	// CommandBuild indicates a build operation
	CommandBuild
	// CommandTest indicates a test operation
	CommandTest
	// CommandRun indicates a run operation
	CommandRun
	// CommandMod indicates a module operation
	CommandMod
	// CommandFmt indicates a formatting operation
	CommandFmt
	// CommandVet indicates a static analysis operation
	CommandVet
)

// ProjectStructure represents different aspects of a Go project structure
type ProjectStructure struct {
	HasGoMod      bool // Whether the project has a go.mod file
	IsMultiModule bool // Whether the project has multiple modules
	HasMainPkg    bool // Whether the project has a main package
	IsTest        bool // Whether the project contains tests
}

// StrategySelector provides a sophisticated mechanism for selecting the appropriate execution strategy
type StrategySelector struct {
	Input            InputContext
	CommandType      CommandType
	ProjectStructure ProjectStructure
}

// NewStrategySelector creates a new strategy selector based on input context
func NewStrategySelector(input InputContext) *StrategySelector {
	selector := &StrategySelector{
		Input: input,
	}

	// Detect command type and project structure if project path is provided
	if input.Source == SourceProjectPath || input.Source == SourceHybrid {
		selector.detectProjectStructure()
	}

	return selector
}

// detectProjectStructure analyzes the project directory to determine its structure
func (s *StrategySelector) detectProjectStructure() {
	if s.Input.ProjectPath == "" {
		return
	}

	// Check for go.mod
	goModPath := filepath.Join(s.Input.ProjectPath, "go.mod")
	s.ProjectStructure.HasGoMod = fileExists(goModPath)

	// Check for main package
	mainGoPath := filepath.Join(s.Input.ProjectPath, "main.go")
	cmdDirPath := filepath.Join(s.Input.ProjectPath, "cmd")
	s.ProjectStructure.HasMainPkg = fileExists(mainGoPath) || dirExists(cmdDirPath)

	// Check for tests
	testFiles, _ := filepath.Glob(filepath.Join(s.Input.ProjectPath, "*_test.go"))
	s.ProjectStructure.IsTest = len(testFiles) > 0

	// Check for multi-module setup
	if s.ProjectStructure.HasGoMod {
		// Look for nested modules
		nestedModules, _ := filepath.Glob(filepath.Join(s.Input.ProjectPath, "**/go.mod"))
		s.ProjectStructure.IsMultiModule = len(nestedModules) > 1
	}
}

// SetCommandType sets the command type based on the command arguments
func (s *StrategySelector) SetCommandType(args []string) {
	if len(args) == 0 {
		s.CommandType = CommandUnknown
		return
	}

	switch args[0] {
	case "build":
		s.CommandType = CommandBuild
	case "test":
		s.CommandType = CommandTest
	case "run":
		s.CommandType = CommandRun
	case "mod":
		s.CommandType = CommandMod
	case "fmt":
		s.CommandType = CommandFmt
	case "vet":
		s.CommandType = CommandVet
	default:
		s.CommandType = CommandUnknown
	}
}

// SelectStrategy returns the most appropriate execution strategy based on the input context,
// command type, and project structure
func (s *StrategySelector) SelectStrategy() ExecutionStrategy {
	// If workspace path is provided, use workspace strategy
	if s.Input.Source == SourceWorkspace {
		return &WorkspaceExecutionStrategy{}
	}

	// If both code and project path are provided, use hybrid strategy
	if s.Input.Source == SourceHybrid {
		return &HybridExecutionStrategy{}
	}

	// If project path is provided, use project strategy
	if s.Input.Source == SourceProjectPath {
		return &ProjectExecutionStrategy{}
	}

	// Default to code execution strategy
	return &CodeExecutionStrategy{}
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists
func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetExecutionStrategy returns the appropriate strategy based on input context and command
// This is an enhanced version of the original function that supports more nuanced execution contexts
// beyond the current binary selection. It considers factors like project structure and command type.
func GetExecutionStrategy(input InputContext, args ...string) ExecutionStrategy {
	// Handle hybrid case (both code and project_path provided)
	if input.Source == SourceProjectPath && input.Code != "" {
		// Create a hybrid input source
		hybridInput := input
		hybridInput.Source = SourceHybrid

		// Create and configure a strategy selector
		selector := NewStrategySelector(hybridInput)
		if len(args) > 0 {
			selector.SetCommandType(args)
		}

		return selector.SelectStrategy()
	}

	// Create and configure a strategy selector for non-hybrid cases
	selector := NewStrategySelector(input)
	if len(args) > 0 {
		selector.SetCommandType(args)
	}
	return selector.SelectStrategy()
}
