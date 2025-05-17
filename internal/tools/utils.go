package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// ExecutionResult contains the result of a command execution
type ExecutionResult struct {
	Stdout     string
	Stderr     string
	ExitCode   int
	Duration   time.Duration
	Successful bool
	Command    string // Command that was executed
}

// NLMetadata represents natural language metadata for tools
type NLMetadata struct {
	Aliases  []string `json:"aliases"`
	Examples []string `json:"examples"`
}

// ToolNLMetadata maps tool names to their natural language metadata
var ToolNLMetadata = map[string]NLMetadata{
	"go_build": {
		Aliases: []string{
			"compile", "build", "create executable", "generate binary",
			"make binary", "construct program", "prepare executable",
			"generate program", "assemble code", "create program",
			"make executable", "generate exe", "turn into binary",
			"convert to executable", "transform into program",
		},
		Examples: []string{
			"compile this Go code",
			"build this program",
			"create an executable from this code",
			"make a binary from this Go file",
			"convert this source code into an executable",
			"generate a binary in the bin directory",
			"build this with the debug tag",
			"compile my Go application with race detection",
			"create a statically linked executable",
			"build with optimizations enabled",
		},
	},
	"go_test": {
		Aliases: []string{
			"test", "unit test", "run tests", "check tests", "verify tests",
			"execute tests", "validate tests", "run test suite", "run unit tests",
			"check test cases", "run test coverage", "test the code", "verify test cases",
			"test functions", "run test functions", "perform tests",
		},
		Examples: []string{
			"test this Go code",
			"run unit tests",
			"check if tests pass",
			"verify test coverage",
			"run tests with verbose output",
			"test this package and show coverage",
			"run only the TestLogin function",
			"test with race detection enabled",
			"run short tests only",
			"check if tests pass with verbose output",
			"I want to run the unit tests for this code",
		},
	},
	"go_run": {
		Aliases: []string{
			"run", "execute", "start", "launch", "run code", "execute code",
			"run program", "start program", "launch application", "start application",
			"execute program", "run application", "execute application", "invoke program",
			"run this", "execute this",
		},
		Examples: []string{
			"run this Go code",
			"execute this program",
			"start this application with arguments",
			"run this code with the input parameter set to 'test'",
			"execute the main package",
			"run this with environment variable DEBUG=true",
			"start the web server on port 8080",
			"run the application with configuration file config.json",
			"execute the script with verbose logging",
			"start this program and capture its output",
		},
	},
	"go_mod": {
		Aliases: []string{
			"dependencies", "modules", "manage dependencies", "dependency management",
			"update modules", "package dependencies", "go.mod", "module dependencies",
			"dependency tracking", "import management", "external packages", "third-party packages",
			"package management", "lib management", "library dependencies",
		},
		Examples: []string{
			"initialize a new module",
			"update dependencies",
			"add a new dependency",
			"tidy up module dependencies",
			"create a new go.mod file",
			"add github.com/gorilla/mux to the dependencies",
			"update all dependencies to their latest versions",
			"clean up unused dependencies with go mod tidy",
			"vendor all dependencies for this project",
			"initialize a module for a new project",
			"check for available updates to dependencies",
		},
	},
	"go_fmt": {
		Aliases: []string{
			"format", "beautify", "pretty-print", "indent", "format code", "beautify code",
			"pretty-print code", "clean up code", "standardize format", "fix formatting",
			"align code", "normalize code", "standardize indentation", "reorganize code",
			"improve readability", "gofmt",
		},
		Examples: []string{
			"format this Go code",
			"beautify this program",
			"fix the formatting of this code",
			"make this code look nice",
			"apply gofmt to this file",
			"standardize indentation in this code",
			"fix whitespace in this Go file",
			"reformat according to Go standards",
			"clean up this messy code",
			"apply standard Go formatting to this code",
			"pretty-print this function",
		},
	},
	"go_analyze": {
		Aliases: []string{
			"lint", "check", "validate", "inspect", "analyze", "examine",
			"vet", "go vet", "static analysis", "code check", "quality check",
			"verify code", "scrutinize", "review code", "find issues",
			"detect problems", "spot errors", "check for bugs",
		},
		Examples: []string{
			"analyze this Go code for issues",
			"check for bugs",
			"validate this function",
			"inspect code quality",
			"run go vet on this code",
			"find potential concurrency issues",
			"check for memory leaks in this code",
			"detect possible nil pointer dereferences",
			"analyze for inefficient patterns",
			"look for security vulnerabilities",
			"identify unused variables and imports",
			"check for proper error handling",
		},
	},
}

// execute runs a command and returns the execution result with better logging
func execute(cmd *exec.Cmd) (*ExecutionResult, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Capture the command for debugging
	cmdStr := cmd.String()
	log.Printf("Executing command: %s", cmdStr)

	// Execute with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use CommandContext instead of cmd.Run()
	execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	execCmd.Env = cmd.Env
	execCmd.Dir = cmd.Dir
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	start := time.Now()
	err := execCmd.Run()
	duration := time.Since(start)

	// Check if the context deadline exceeded
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Command timed out after %v: %s", duration, cmdStr)
		return &ExecutionResult{
			Stdout:     stdout.String(),
			Stderr:     "Command execution timed out",
			ExitCode:   -1,
			Duration:   duration,
			Successful: false,
			Command:    cmdStr,
		}, fmt.Errorf("command timed out after %v", duration)
	}

	result := &ExecutionResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Duration:   duration,
		Successful: err == nil,
		Command:    cmdStr,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}

	// Log execution results
	if result.Successful {
		log.Printf("Command succeeded in %v: %s", duration, cmdStr)
	} else {
		log.Printf("Command failed in %v with exit code %d: %s", duration, result.ExitCode, cmdStr)
	}

	return result, nil
}

// FormatCommandResult creates a standardized JSON response for tool executions
func FormatCommandResult(result *ExecutionResult, responseType string) *mcp.CallToolResult {
	response := map[string]interface{}{
		"success":   result.Successful,
		"exitCode":  result.ExitCode,
		"duration":  result.Duration.String(),
		"command":   result.Command,
		"type":      responseType,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if result.Stdout != "" {
		response["stdout"] = result.Stdout
	}

	if result.Stderr != "" {
		response["stderr"] = result.Stderr

		// Add structured error details if there's an error
		if !result.Successful {
			response["errorDetails"] = ParseGoErrors(result.Stderr)
		}
	}

	// Add natural language metadata based on the tool type
	AddNLMetadata(response, responseType)

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error marshaling response: %v", err))
	}

	if result.Successful {
		return mcp.NewToolResultText(string(jsonBytes))
	} else {
		return mcp.NewToolResultError(string(jsonBytes))
	}
}
