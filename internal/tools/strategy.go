package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ExecutionResult stores the result of a command execution
type ExecutionResult struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	Duration   time.Duration
	Successful bool
}

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

// GetExecutionStrategy returns the appropriate strategy based on input context
func GetExecutionStrategy(input InputContext) ExecutionStrategy {
	switch input.Source {
	case SourceProjectPath:
		return &ProjectExecutionStrategy{}
	default:
		return &CodeExecutionStrategy{}
	}
}

// execute runs a command and captures its output
func execute(cmd *exec.Cmd) (*ExecutionResult, error) {
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)
	
	result := &ExecutionResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Duration:   duration,
		Successful: err == nil,
	}
	
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}
	
	return result, nil
}
