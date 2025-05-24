# MCP Server Testing Framework

This document describes the testing framework for the Go Development MCP Server, including the new Go-based testing approach and the transition from the previous PowerShell-based testing.

## Table of Contents

1. [Introduction](#introduction)
2. [Testing Framework Overview](#testing-framework-overview)
3. [Go Testing Framework](#go-testing-framework)
4. [Running Tests](#running-tests)
5. [Writing New Tests](#writing-new-tests)
6. [Test Organization](#test-organization)
7. [Parallel Testing](#parallel-testing)
8. [Best Practices](#best-practices)

## Introduction

The MCP Server testing framework has been redesigned to leverage Go's built-in testing capabilities, incorporate table-driven testing, and enable parallel test execution. This modernization aims to improve test reliability, speed, and maintainability.

## Testing Framework Overview

The testing framework consists of:

1. **Go Tests**: Modern Go tests using the `testing` package and the `testify` framework
2. **PowerShell Test Scripts**: Legacy tests written in PowerShell (being phased out)
3. **Test Runner**: A unified runner that supports both test types

## Go Testing Framework

The Go testing framework is built on standard Go testing patterns with additional structure:

- **Test Suites**: Using `testify/suite` for organized test grouping
- **Table-Driven Tests**: For comprehensive test cases
- **Parallel Execution**: Using `t.Parallel()` for concurrent testing
- **Fixtures**: Reusable test project templates and setups
- **Mocks**: For isolated testing of components

### Key Components

- **`internal/testing/suite.go`**: Base test suite implementation
- **`internal/testing/helpers.go`**: Common test helper functions
- **`internal/testing/fixtures/`**: Test fixtures and data
- **`internal/testing/mock/`**: Mock implementations for testing
- **`internal/testing/parallel.go`**: Parallel testing coordination

## Running Tests

### Using the Test Runner

The main test runner supports both PowerShell and Go tests:

```powershell
# Run all tests
.\scripts\testing\run_tests.ps1 -TestType all

# Run only Go tests
.\scripts\testing\run_tests.ps1 -TestType go -UseGoTests -WithCoverage -WithRaceDetection

# Run Go tests with coverage analysis
.\scripts\testing\run_tests.ps1 -TestType go -UseGoTests -WithCoverage
```

### Running Go Tests Directly

You can also run Go tests directly using standard Go tools:

```powershell
# Run all Go tests
go test ./internal/tools/...

# Run tests with verbose output
go test -v ./internal/tools/...

# Run tests with race detection
go test -race ./internal/tools/...

# Run a specific test
go test -v ./internal/tools/... -run TestRunToolSuite
```

### Environment Variables

- `MCP_USE_GO_TESTS`: Set to "true" to enable Go tests in the main runner
- `MCP_TEST_PARALLEL`: Set to control parallel test count (default: CPU cores - 1)

## Writing New Tests

### Test Suite Structure

New tests should follow this pattern:

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

// SetupSuite runs before all tests in the suite
func (s *YourTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	// Add your setup code
}

// TearDownSuite runs after all tests in the suite
func (s *YourTestSuite) TearDownSuite() {
	// Add your teardown code
	s.BaseSuite.TearDownSuite()
}

// TestYourFeature tests a specific feature
func (s *YourTestSuite) TestYourFeature() {
	// Enable parallel execution if appropriate
	testing.RunParallel(s.T())
	
	// Your test code
	s.Equal("expected", "actual")
}

// TestYourTestSuite runs the test suite
func TestYourTestSuite(t *testing.T) {
	suite.Run(t, new(YourTestSuite))
}
```

### Table-Driven Tests

For testing multiple similar cases:

```go
func (s *YourTestSuite) TestMultipleCases() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty input", "", ""},
		{"normal input", "hello", "HELLO"},
		{"special chars", "a!b@c#", "A!B@C#"},
	}
	
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := strings.ToUpper(tc.input)
			s.Equal(tc.expected, result)
		})
	}
}
```

### Using Test Fixtures

```go
// Create a test project
project := fixtures.SimpleProjectFixture(s.TempDir, "test-project")
err := project.Setup()
s.Require().NoError(err)
defer project.Cleanup()
```

## Test Organization

Tests are organized by functionality and test type:

- **Unit Tests**: Test individual functions or methods
- **Integration Tests**: Test interactions between components
- **End-to-End Tests**: Test complete workflows

Each test file should focus on a specific component or feature.

## Parallel Testing

To enable parallel testing:

```go
// At the beginning of each test method
testing.RunParallel(s.T())
```

Tests that use RunParallel:
1. Must be completely independent
2. Should not modify global state
3. Should use separate test directories

## Best Practices

1. **Use Table-Driven Tests**: For comprehensive testing of similar cases
2. **Write Isolated Tests**: Ensure tests don't depend on each other
3. **Clean Up Resources**: Always clean up temporary files and directories
4. **Use Assertions Properly**: Use `s.Assert()` for non-critical checks, `s.Require()` for critical checks
5. **Test Error Cases**: Always test error conditions, not just happy paths
6. **Keep Tests Focused**: Test one thing per test method
7. **Use Mocks When Appropriate**: Mock external dependencies for unit tests
8. **Include Edge Cases**: Test boundaries and special conditions

## Transitioning from PowerShell Tests

As we modernize our testing approach, we're gradually transitioning from PowerShell to Go tests. The `run_tests.ps1` script supports both for backward compatibility.

1. **Identify Tests to Migrate**: Look for PowerShell tests that would benefit from Go's testing capabilities
2. **Create Equivalent Go Tests**: Using the new framework
3. **Verify Both Pass**: Ensure both test versions pass before removing PowerShell tests
4. **Remove PowerShell Tests**: Once Go tests are stable and comprehensive
