# Go Development MCP Server Testing Framework Modernization Prompt

## Context

I'm working on modernizing the testing framework for a Go Development MCP Server that implements the Model Context Protocol. This server provides tools for Go language development that can be invoked by AI assistants like Claude. The testing framework is being migrated from PowerShell scripts to a modern Go-based approach with several improvements:

1. The package name has been changed from `testing` to `mcptesting` to avoid conflicts with the standard Go `testing` package
2. Utility functions for extracting arguments from `mcp.Params` have been created in `internal/tools/args.go`
3. Several test files have been updated to work with the map-based structure of `mcp.Params`

The codebase is in a transitional state with several files already updated but others still needing changes.

## Current Progress

- ✅ Changed package name from `testing` to `mcptesting` in:
  - `internal/testing/suite.go`
  - `internal/testing/helpers.go`
  - `internal/testing/parallel.go`
- ✅ Created utility functions for extracting arguments in `internal/tools/args.go`
- ✅ Updated some tests to use the new argument extraction utilities
- ✅ Implemented mock objects in `internal/testing/mock/tools.go`
- ✅ Created test fixture helpers in `internal/testing/fixtures/projects.go`
- ✅ Successfully updated the following test files:
  - `internal/tools/input_test.go`
  - `internal/tools/integration_test.go`
  - `internal/tools/strategy_test.go`
  - `internal/tools/run_test.go`

## Required Tasks

Please complete the modernization of the testing framework by:

1. Updating any remaining tool implementations to consistently use the new argument extraction utilities
   - Identify tools still using the old approach (directly accessing `map[string]interface{}`)
   - Refactor them to use the `ExtractStringArg`, `ExtractBoolArg`, etc. functions

2. Creating proper Go tests for PowerShell tests that still need migration:
   - Implement end-to-end tests with a mock server (per `e2e_test.ps1`)
   - Create strategy verification tests (per `hybrid_strat_verify.ps1`)

3. Adding test metrics collection:
   - Implement utilities for measuring test coverage
   - Add performance benchmarking for critical operations

4. Ensuring complete edge case coverage:
   - Tests for error conditions and recovery
   - Tests for resource limits and timeouts
   - Tests for concurrent execution

5. Implementing testing patterns following best practices from modern Go testing frameworks:
   - Table-driven tests for variations of the same functionality
   - Test fixtures for complex setup requirements
   - Subtests for better organization and parallel execution
   - Using testify's suite and assertions consistently

## Argument Extraction Utilities

The following utilities in `internal/tools/args.go` should be used consistently across all tool implementations:

```go
// ExtractArguments extracts the arguments map from mcp.Params
func ExtractArguments(params mcp.Params) (map[string]interface{}, error)

// ExtractStringArg extracts a string argument from arguments map
func ExtractStringArg(args map[string]interface{}, key string) (string, bool)

// ExtractBoolArg extracts a boolean argument from arguments map
func ExtractBoolArg(args map[string]interface{}, key string) (bool, bool)

// ExtractIntArg extracts an integer argument from arguments map
func ExtractIntArg(args map[string]interface{}, key string) (int, bool)

// ExtractStringMapArg extracts a map[string]interface{} argument from arguments map
func ExtractStringMapArg(args map[string]interface{}, key string) (map[string]interface{}, bool)

// ExtractStringListArg extracts a []string argument from arguments map
func ExtractStringListArg(args map[string]interface{}, key string) ([]string, bool)
```

## Testing Framework Structure

The test suite uses a hierarchical structure:

1. `mcptesting.BaseSuite` - The foundation test suite in `internal/testing/suite.go`
2. Tool-specific test suites that extend BaseSuite
3. Test fixtures in `internal/testing/fixtures` for reusable test scenarios
4. Mock objects in `internal/testing/mock` for isolation testing

## Mock Server Requirements

The mock server implementation should:

1. Simulate the MCP protocol API endpoints
2. Allow injection of predetermined responses for testing
3. Record requests for later verification
4. Support both successful responses and error conditions
5. Include mock implementations for all six tools:
   - go_build
   - go_run
   - go_test
   - go_mod
   - go_fmt
   - go_analyze

## Best Practices to Implement

Please implement these modern Go testing best practices:

1. **Table-Driven Tests**: Use `[]struct` to define test cases with inputs and expected outputs
2. **Parallel Test Execution**: Use `t.Parallel()` for independent tests to improve performance
3. **HTTP Test Server**: Use `httptest.NewServer` for API endpoint testing
4. **Assertions**: Use testify's `assert` and `require` packages for clearer test logic
5. **Mock Objects**: Use the mock package for isolation testing
6. **Test Fixtures**: Use fixtures for reusable test data and setup
7. **Subtests**: Use `t.Run()` for organizing related test cases
8. **Test Suites**: Use testify's suite for sharing setup/teardown code
9. **Context**: Use context for timeout management and cancelation
10. **Resource Cleanup**: Ensure proper cleanup of all resources in defer blocks

## Implementation Strategy

1. First analyze any remaining tool implementations that don't use the arg extraction utilities
2. Refactor each tool implementation to use the appropriate extraction functions
3. Implement missing Go tests based on existing PowerShell tests
4. Add metrics collection to the test framework
5. Create a comprehensive set of edge case tests
6. Verify that all tests pass and the code coverage meets the target (90%)

## Deliverables

Please provide:

1. Updated implementation code for tools to use the arg extraction utilities
2. New Go test files for mock server and end-to-end testing
3. Test metrics collection utilities
4. Strategy verification tests
5. Expanded edge case test coverage

Thank you for helping complete this testing framework modernization effort.
