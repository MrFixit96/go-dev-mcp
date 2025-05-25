# MCP Server Testing Framework

This package provides a standardized testing framework for the Go Development MCP Server. It includes test suite definitions, helpers, fixtures, and mocking utilities to facilitate writing comprehensive, parallel tests.

## Key Components

- **`suite.go`**: Base test suite implementation with common setup/teardown
- **`helpers.go`**: Helper functions for common test operations
- **`parallel.go`**: Utilities for controlling parallel test execution
- **`fixtures/`**: Test fixtures and data for consistent test environments
- **`mock/`**: Mock implementations for isolated testing

## Usage

### Basic Test Suite

```go
package yourpackage_test

import (
    "testing"

    "github.com/MrFixit96/go-dev-mcp/internal/testing"
    "github.com/stretchr/testify/suite"
)

// YourTestSuite defines a test suite
type YourTestSuite struct {
    testing.BaseSuite
    // Add suite-specific fields
}

// TestSomething is a test method
func (s *YourTestSuite) TestSomething() {
    // Your test code
    s.Equal("expected", "actual")
}

// TestYourSuite runs the test suite
func TestYourSuite(t *testing.T) {
    suite.Run(t, new(YourTestSuite))
}
```

### Parallel Testing

```go
func (s *YourTestSuite) TestParallel() {
    // Enable parallel execution
    testing.RunParallel(s.T())

    // Test code that can run in parallel
}
```

### Using Project Fixtures

```go
func (s *YourTestSuite) TestWithProject() {
    // Create a test project
    project := fixtures.SimpleProjectFixture(s.TempDir, "test-project")
    err := project.Setup()
    s.Require().NoError(err)
    defer project.Cleanup()

    // Use the project in your test
}
```

### Using Mocks

```go
func (s *YourTestSuite) TestWithMocks() {
    // Create a mock executor
    mockExecutor := &mock.ToolExecutor{}

    // Set up expectations
    mockExecutor.SetupSuccessResponse("go_run", "Hello, World!")

    // Use the mock in your test
}
```

## Best Practices

1. Always clean up resources in tests
2. Use table-driven tests for multiple test cases
3. Make tests independent of each other
4. Use descriptive test names
5. Test both success and error paths
6. Keep tests focused on one aspect
7. Use parallelism for independent tests

See the detailed documentation in `scripts/testing/TESTING.md` for more information.
