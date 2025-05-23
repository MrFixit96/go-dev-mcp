// Package mcptesting provides testing helpers for MCP tests.
package mcptesting

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ProjectType defines the type of test project to create
type ProjectType string

const (
	// SimpleProject is a basic single-file Go project
	SimpleProject ProjectType = "simple"
	// MultiFileProject is a Go project with multiple source files
	MultiFileProject ProjectType = "multi-file"
	// WithDepsProject is a Go project with external dependencies
	WithDepsProject ProjectType = "with-deps"
)

// CreateProject creates a test Go project of the specified type in the given directory.
func CreateProject(projectPath, projectName string, projectType ProjectType) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return err
	}

	// Define project files based on type
	switch projectType {
	case SimpleProject:
		// Create main.go file
		mainCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainCode), 0644); err != nil {
			return err
		}

	case MultiFileProject:
		// Create main.go file
		mainCode := `package main

import "fmt"

func main() {
	greeting := GetGreeting()
	name := GetName()
	fmt.Printf("%s, %s!\n", greeting, name)
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainCode), 0644); err != nil {
			return err
		}

		// Create greeting.go file
		greetingCode := `package main

func GetGreeting() string {
	return "Hello"
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "greeting.go"), []byte(greetingCode), 0644); err != nil {
			return err
		}

		// Create name.go file
		nameCode := `package main

func GetName() string {
	return "World"
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "name.go"), []byte(nameCode), 0644); err != nil {
			return err
		}

	case WithDepsProject:
		// Create main.go file with external dependency
		mainCode := `package main

import (
	"fmt"
	"github.com/fatih/color"
)

func main() {
	c := color.New(color.FgCyan)
	c.Println("Hello, World!")
}
`
		if err := os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainCode), 0644); err != nil {
			return err
		}
	}

	// Initialize go module
	cmd := exec.Command("go", "mod", "init", "example.com/"+projectName)
	cmd.Dir = projectPath
	if err := cmd.Run(); err != nil {
		return err
	}

	// Install dependencies if needed
	if projectType == WithDepsProject {
		getDeps := exec.Command("go", "get", "github.com/fatih/color")
		getDeps.Dir = projectPath
		if err := getDeps.Run(); err != nil {
			return err
		}

		tidy := exec.Command("go", "mod", "tidy")
		tidy.Dir = projectPath
		if err := tidy.Run(); err != nil {
			return err
		}
	}

	return nil
}

// ExecResult represents the result of a command execution
type ExecResult struct {
	Success  bool
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// RunCommand executes a command with the given arguments and returns the result
func RunCommand(dir string, command string, args ...string) (*ExecResult, error) {
	start := time.Now()

	// Create command
	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	// Capture stdout and stderr
	stdoutBytes, err := cmd.Output()
	exitCode := 0
	stderr := ""

	// Handle error and extract stderr
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = string(exitErr.Stderr)
		} else {
			return nil, err
		}
	}

	duration := time.Since(start)
	success := exitCode == 0

	return &ExecResult{
		Success:  success,
		ExitCode: exitCode,
		Stdout:   string(stdoutBytes),
		Stderr:   stderr,
		Duration: duration,
	}, nil
}

// CreateHybridProject creates a modified version of an existing project
func CreateHybridProject(originalPath, hybridPath, modifiedCode, modifiedFile string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(hybridPath, 0755); err != nil {
		return err
	}

	// Walk through the original project and copy all files except the one to modify
	err := filepath.Walk(originalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(originalPath, path)
		if err != nil {
			return err
		}

		// Skip the file to be modified
		if relPath == modifiedFile {
			return nil
		}

		// Copy the file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(hybridPath, relPath)
		destDir := filepath.Dir(destPath)

		// Ensure destination directory exists
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		// Write the file
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Write the modified file
	return os.WriteFile(filepath.Join(hybridPath, modifiedFile), []byte(modifiedCode), 0644)
}
