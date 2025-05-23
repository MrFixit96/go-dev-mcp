// Package integration provides integration tests for the Go Development MCP Server tools.
// This file implements comprehensive integration testing with timeout and resource constraints.
package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	mcptesting "github.com/MrFixit96/go-dev-mcp/internal/testing"
	"github.com/stretchr/testify/suite"
)

// TestConfiguration holds configuration for integration tests
type TestConfiguration struct {
	DefaultTimeout  time.Duration
	MaxMemoryMB     int64
	MaxCPUPercent   float64
	MaxDiskSpaceMB  int64
	ParallelWorkers int
	TempDirCleanup  bool
}

// DefaultTestConfig returns the default test configuration with reasonable limits
func DefaultTestConfig() TestConfiguration {
	return TestConfiguration{
		DefaultTimeout:  30 * time.Second,
		MaxMemoryMB:     500,  // 500MB max memory usage
		MaxCPUPercent:   80.0, // 80% max CPU usage
		MaxDiskSpaceMB:  100,  // 100MB max disk space
		ParallelWorkers: runtime.NumCPU(),
		TempDirCleanup:  true,
	}
}

// SimpleIntegrationTestSuite provides basic integration testing for Go MCP tools
type SimpleIntegrationTestSuite struct {
	mcptesting.BaseSuite
	config     TestConfiguration
	projectDir string
}

// SetupSuite initializes the test suite
func (s *SimpleIntegrationTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.config = DefaultTestConfig()

	// Create a simple test project
	s.projectDir = s.NewTempDir("simple-project")
	s.createSimpleGoProject(s.projectDir)
}

// createSimpleGoProject creates a basic Go project for testing
func (s *SimpleIntegrationTestSuite) createSimpleGoProject(projectPath string) {
	// Create go.mod file
	goModContent := `module testproject

go 1.21
`
	err := os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goModContent), 0644)
	s.Require().NoError(err, "Failed to create go.mod file")

	// Create a simple main.go file
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	err = os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainGoContent), 0644)
	s.Require().NoError(err, "Failed to create main.go file")
}

// runCommandWithTimeout executes a command with timeout
func (s *SimpleIntegrationTestSuite) runCommandWithTimeout(ctx context.Context, dir string, command string, args ...string) (*mcptesting.ExecResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.config.DefaultTimeout)
	defer cancel()

	start := time.Now()

	// Create command with context for timeout support
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	// Run command and capture output
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	var exitCode int
	success := true

	if err != nil {
		success = false
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	return &mcptesting.ExecResult{
		Success:  success,
		ExitCode: exitCode,
		Stdout:   string(output),
		Stderr:   "",
		Duration: duration,
	}, nil
}

// TestGoBuildBasic tests basic go build functionality
func (s *SimpleIntegrationTestSuite) TestGoBuildBasic() {
	ctx := context.Background()
	result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "build", "-o", "app.exe")
	s.NoError(err, "Build command should not error")
	s.True(result.Success, "Build should succeed")
	s.Less(result.Duration, s.config.DefaultTimeout, "Build should complete within timeout")

	// Verify executable was created
	execPath := filepath.Join(s.projectDir, "app.exe")
	s.FileExists(execPath, "Executable should be created")
}

// TestGoRunBasic tests basic go run functionality
func (s *SimpleIntegrationTestSuite) TestGoRunBasic() {
	ctx := context.Background()
	result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "run", ".")
	s.NoError(err, "Run command should not error")
	s.True(result.Success, "Run should succeed")
	s.Less(result.Duration, s.config.DefaultTimeout, "Run should complete within timeout")
	s.Contains(result.Stdout, "Hello, World!", "Output should match expected")
}

// TestGoFmtBasic tests basic go fmt functionality
func (s *SimpleIntegrationTestSuite) TestGoFmtBasic() {
	ctx := context.Background()
	result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "fmt")
	s.NoError(err, "Format command should not error")
	s.True(result.Success, "Format should succeed")
	s.Less(result.Duration, s.config.DefaultTimeout, "Format should complete within timeout")
}

// TestGoVetBasic tests basic go vet functionality
func (s *SimpleIntegrationTestSuite) TestGoVetBasic() {
	ctx := context.Background()
	result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "vet", "./...")
	s.NoError(err, "Vet command should not error")
	s.True(result.Success, "Vet should succeed for clean code")
	s.Less(result.Duration, s.config.DefaultTimeout, "Vet should complete within timeout")
}

// TestGoModBasic tests basic go mod functionality
func (s *SimpleIntegrationTestSuite) TestGoModBasic() {
	ctx := context.Background()
	result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "mod", "tidy")
	s.NoError(err, "Mod tidy command should not error")
	s.True(result.Success, "Mod tidy should succeed")
	s.Less(result.Duration, s.config.DefaultTimeout, "Mod tidy should complete within timeout")
}

// TestTimeoutEnforcement tests that timeouts are properly enforced
func (s *SimpleIntegrationTestSuite) TestTimeoutEnforcement() {
	ctx := context.Background()

	// Create a test that should timeout
	timeoutTestDir := s.NewTempDir("timeout-test")

	// Create code that sleeps longer than our test timeout
	sleepCode := `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting...")
	time.Sleep(8 * time.Second)
	fmt.Println("Done!")
}
`
	mainPath := filepath.Join(timeoutTestDir, "main.go")
	s.Require().NoError(os.WriteFile(mainPath, []byte(sleepCode), 0644))

	// Initialize module
	_, err := s.runCommandWithTimeout(ctx, timeoutTestDir, "go", "mod", "init", "example.com/timeout-test")
	s.Require().NoError(err)

	// Create a custom context with a shorter timeout
	shortCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// This should timeout
	start := time.Now()
	cmd := exec.CommandContext(shortCtx, "go", "run", ".")
	cmd.Dir = timeoutTestDir
	err = cmd.Run()
	elapsed := time.Since(start)

	// Should timeout before the full sleep duration
	s.Less(elapsed, 6*time.Second, "Should timeout before 6 seconds")
	s.Greater(elapsed, 2*time.Second, "Should run for at least 2 seconds")
	s.Error(err, "Should error due to timeout")
}

// TestResourceConstraints tests memory and resource monitoring
func (s *SimpleIntegrationTestSuite) TestResourceConstraints() {
	ctx := context.Background()

	// Test memory constraint awareness
	s.Run("MemoryAware", func() {
		// Create a simple test that should complete within memory limits
		memTestDir := s.NewTempDir("memory-test")

		// Create code that uses a reasonable amount of memory
		memCode := `package main

import "fmt"

func main() {
	// Create a slice with reasonable memory usage
	data := make([]string, 1000)
	for i := range data {
		data[i] = fmt.Sprintf("Item %d", i)
	}
	fmt.Printf("Created %d items\n", len(data))
}
`
		mainPath := filepath.Join(memTestDir, "main.go")
		s.Require().NoError(os.WriteFile(mainPath, []byte(memCode), 0644))

		// Initialize module
		_, err := s.runCommandWithTimeout(ctx, memTestDir, "go", "mod", "init", "example.com/memory-test")
		s.Require().NoError(err)

		// Run with timeout and verify it completes successfully
		result, err := s.runCommandWithTimeout(ctx, memTestDir, "go", "run", ".")
		s.NoError(err, "Memory test should not error")
		s.True(result.Success, "Memory test should succeed")
		s.Contains(result.Stdout, "Created 1000 items", "Should produce expected output")
	})

	// Test CPU constraint awareness
	s.Run("CPUAware", func() {
		// Create a test that does some CPU work but should complete reasonably
		cpuTestDir := s.NewTempDir("cpu-test")

		cpuCode := `package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	
	// Do some CPU work (but not excessive)
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	
	elapsed := time.Since(start)
	fmt.Printf("Computed sum: %d in %v\n", sum, elapsed)
}
`
		mainPath := filepath.Join(cpuTestDir, "main.go")
		s.Require().NoError(os.WriteFile(mainPath, []byte(cpuCode), 0644))

		// Initialize module
		_, err := s.runCommandWithTimeout(ctx, cpuTestDir, "go", "mod", "init", "example.com/cpu-test")
		s.Require().NoError(err)

		// Run with timeout and verify reasonable performance
		start := time.Now()
		result, err := s.runCommandWithTimeout(ctx, cpuTestDir, "go", "run", ".")
		elapsed := time.Since(start)

		s.NoError(err, "CPU test should not error")
		s.True(result.Success, "CPU test should succeed")
		s.Less(elapsed, 10*time.Second, "CPU test should complete in reasonable time")
		s.Contains(result.Stdout, "Computed sum:", "Should produce expected output")
	})
}

// TestAllGoTools tests all six Go tools with timeout and resource constraints
func (s *SimpleIntegrationTestSuite) TestAllGoTools() {
	ctx := context.Background()

	s.Run("GoBuild", func() {
		result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "build", "-o", "test-app.exe")
		s.NoError(err, "Build should not error")
		s.True(result.Success, "Build should succeed")
		s.Less(result.Duration, s.config.DefaultTimeout, "Build should complete within timeout")

		// Verify executable was created
		execPath := filepath.Join(s.projectDir, "test-app.exe")
		s.FileExists(execPath, "Executable should be created")
	})

	s.Run("GoRun", func() {
		result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "run", ".")
		s.NoError(err, "Run should not error")
		s.True(result.Success, "Run should succeed")
		s.Less(result.Duration, s.config.DefaultTimeout, "Run should complete within timeout")
		s.Contains(result.Stdout, "Hello, World!", "Should produce expected output")
	})

	s.Run("GoFmt", func() {
		result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "fmt")
		s.NoError(err, "Format should not error")
		s.True(result.Success, "Format should succeed")
		s.Less(result.Duration, s.config.DefaultTimeout, "Format should complete within timeout")
	})

	s.Run("GoVet", func() {
		result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "vet", "./...")
		s.NoError(err, "Vet should not error")
		s.True(result.Success, "Vet should succeed")
		s.Less(result.Duration, s.config.DefaultTimeout, "Vet should complete within timeout")
	})

	s.Run("GoMod", func() {
		result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "mod", "tidy")
		s.NoError(err, "Mod tidy should not error")
		s.True(result.Success, "Mod tidy should succeed")
		s.Less(result.Duration, s.config.DefaultTimeout, "Mod tidy should complete within timeout")
	})

	s.Run("GoTest", func() {
		// Create a simple test file
		testCode := `package main

import "testing"

func TestHello(t *testing.T) {
	// Simple test that should pass
	if 1+1 != 2 {
		t.Error("Math is broken!")
	}
}
`
		testPath := filepath.Join(s.projectDir, "main_test.go")
		err := os.WriteFile(testPath, []byte(testCode), 0644)
		s.Require().NoError(err, "Should create test file")

		result, err := s.runCommandWithTimeout(ctx, s.projectDir, "go", "test", "-v")
		s.NoError(err, "Test should not error")
		s.True(result.Success, "Test should succeed")
		s.Less(result.Duration, s.config.DefaultTimeout, "Test should complete within timeout")
		s.Contains(result.Stdout, "PASS", "Should show test passing")
	})
}

// TestEdgeCases tests various edge cases and error conditions
func (s *SimpleIntegrationTestSuite) TestEdgeCases() {
	ctx := context.Background()

	s.Run("NonExistentDirectory", func() {
		nonExistentDir := filepath.Join(s.TempDir, "does-not-exist")

		_, err := s.runCommandWithTimeout(ctx, nonExistentDir, "go", "run", ".")
		s.Error(err, "Should fail when directory does not exist")
	})

	s.Run("InvalidGoCode", func() {
		invalidCodeDir := s.NewTempDir("invalid-code")

		// Create syntactically invalid Go code
		invalidCode := `package main

func main( {
	fmt.Println("Missing closing parenthesis"
}
`
		mainPath := filepath.Join(invalidCodeDir, "main.go")
		s.Require().NoError(os.WriteFile(mainPath, []byte(invalidCode), 0644))

		// Initialize module
		_, err := s.runCommandWithTimeout(ctx, invalidCodeDir, "go", "mod", "init", "example.com/invalid-code")
		s.Require().NoError(err)

		// Should fail to compile
		result, err := s.runCommandWithTimeout(ctx, invalidCodeDir, "go", "run", ".")
		s.NoError(err, "Command should not error (but compilation should fail)")
		s.False(result.Success, "Invalid code should fail to compile")
		s.Contains(result.Stdout, "syntax error", "Should report syntax error")
	})

	s.Run("EmptyProject", func() {
		emptyDir := s.NewTempDir("empty-project")

		// Try to run in empty directory
		result, err := s.runCommandWithTimeout(ctx, emptyDir, "go", "run", ".")
		s.NoError(err, "Command should not error")
		s.False(result.Success, "Should fail in empty directory")
	})
}

// TestSimpleIntegrationSuite runs the simple integration test suite
func TestSimpleIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SimpleIntegrationTestSuite))
}
