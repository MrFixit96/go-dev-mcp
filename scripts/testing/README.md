# Go Development MCP Server Testing Framework

This directory contains our testing framework for the Go Development MCP Server. We use a **Go-based testing approach** as our primary testing methodology, which directly executes Go commands for more reliable and maintainable tests. The older PowerShell-based tests have been preserved in the `legacy` directory for reference purposes only but should not be used for new test development.

## Directory Structure

```
testing/
├── main.go            # Primary Go-based test runner for execution strategies
├── direct_runner.go   # Standalone direct execution strategy runner
├── hybrid_runner.go   # Standalone hybrid execution strategy runner
├── MIGRATION_STATUS.md # Migration progress tracking (currently 60% complete)
├── legacy/            # Legacy PowerShell test scripts (for reference only)
│   ├── basic/         # Legacy basic test scripts
│   ├── core/          # Legacy core functionality test scripts
│   ├── strategies/    # Legacy execution strategy test scripts
│   └── utils/         # Legacy shared utility functions
├── GETTING_STARTED.md # Quick reference guide
└── README.md          # This documentation file
```

## Go-based Testing Framework

We use a modern Go-based testing approach that offers several advantages:

### Test Runners

1. **Main Test Runner (`main.go`)**
   - Primary entry point for running all tests
   - Uses direct Go command execution without handlers
   - Simplified error handling and more predictable behavior
   - Runs with `go run main.go` or `go run main.go -type=direct|hybrid|both`

2. **Standalone Strategy Runners**
   - `direct_runner.go`: Tests direct code execution strategy
   - `hybrid_runner.go`: Tests hybrid execution with project path and code
   - Both use build tags for selective compilation

### Direct Execution Strategy

The Go-based tests execute Go commands directly using `os/exec` rather than going through handlers:

```go
cmd := exec.Command("go", "run", mainGoPath)
output, err := cmd.CombinedOutput()
```

This approach eliminates dependencies on external packages and context parameters, resulting in simpler and more maintainable code.

### Migration Status

Currently, approximately 60% of our tests have been migrated to the Go-based approach:

- ✅ Core tool functionality (run, build) - Complete
- ✅ Execution strategies - Complete
- ✅ Input handling - Complete
- 🔄 Integration between components - In Progress
- ❌ Mock server for E2E tests - Not Started
- ❌ API boundary testing - Not Started

For a detailed overview of the migration progress, see [MIGRATION_STATUS.md](MIGRATION_STATUS.md).

## Testing Approach

Our testing approach simulates real-world interactions with the server, focusing on:

1. **Testing all input modes**:
   - Code-only: When only inline code is provided
   - Project path-only: When only a directory path is provided
   - Hybrid: When both code and project path are provided

2. **Testing all server tools**:
   - `go_build`: Building Go code
   - `go_run`: Running Go code
   - `go_fmt`: Formatting Go code
   - `go_test`: Running tests for Go code
   - `go_mod`: Managing Go modules
   - `go_analyze`: Analyzing Go code for issues

3. **Verifying the Hybrid Strategy**:
   The hybrid execution strategy combines:
   - Project structure and dependencies from the `project_path`
   - Modified code from the `code` parameter

   This strategy is critical for providing context-aware code execution while allowing modifications.

## Running Go Tests

### Running with Go Commands

You can run the Go-based tests directly using the Go command-line tools:

```bash
# Run the main test runner with both direct and hybrid tests
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
go run main.go

# Run only direct execution tests
go run main.go -type=direct

# Run only hybrid execution tests
go run main.go -type=hybrid
```

### Using Standalone Test Runners

For specific tests with build tags:

```bash
# Run direct execution tests with the direct_runner
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
go run -tags=direct_test direct_runner.go

# Run hybrid execution tests with the hybrid_runner
go run -tags=hybrid_test hybrid_runner.go
```

### Using the PowerShell Script Runner for Go Tests

We also provide a PowerShell script to run the Go tests with additional options:

```powershell
# Run Go tests with verbose output
.\run_go_tests.ps1 -Verbose

# Run Go tests with race detection
.\run_go_tests.ps1 -Race

# Run Go tests with coverage
.\run_go_tests.ps1 -Cover
```

For more advanced testing with coverage reports:

```powershell
# Generate coverage report
.\run_tests_with_coverage.ps1
```

## Testing Approach

Our testing approach simulates real-world interactions with the server, focusing on:

1. **Testing all input modes**:
   - Code-only: When only inline code is provided
   - Project path-only: When only a directory path is provided
   - Hybrid: When both code and project path are provided

2. **Testing all server tools**:
   - `go_build`: Building Go code
   - `go_run`: Running Go code
   - `go_fmt`: Formatting Go code
   - `go_test`: Running tests for Go code
   - `go_mod`: Managing Go modules
   - `go_analyze`: Analyzing Go code for issues

3. **Verifying the Hybrid Strategy**:
   The hybrid execution strategy combines:
   - Project structure and dependencies from the `project_path`
   - Modified code from the `code` parameter
   
   This strategy is critical for providing context-aware code execution while allowing modifications.

## Legacy PowerShell Testing

> **IMPORTANT**: The following sections describe our legacy PowerShell testing approach. These scripts are kept for reference purposes only and should not be used for new test development. All new tests should use the Go-based approach described above.

### Legacy Master Test Runner

The legacy `run_tests.ps1` script provides a way to run legacy PowerShell tests by category:

```powershell
# Run all tests
.\legacy\run_tests.ps1 -TestType all

# Run only basic tests
.\legacy\run_tests.ps1 -TestType basic

# Run only core tests
.\legacy\run_tests.ps1 -TestType core

# Run only strategy tests
.\legacy\run_tests.ps1 -TestType strategies
```

Additional parameters:
- `-VerboseOutput`: Show detailed test information
- `-KeepTestDirs`: Keep temporary test directories for inspection
- `-ServerExecutable <path>`: Specify a custom server executable path

### Legacy Core Test Scripts

#### `legacy/core/all_tools_test.ps1`

A comprehensive legacy test suite that tests all six tools provided by the Go Development MCP Server with all applicable input modes.

**Usage**:

```powershell
.\legacy\core\all_tools_test.ps1 [-ServerExecutable <path>] [-KeepTestDirs] [-TestDir <path>] [-Verbose]
```

**Parameters**:

- `-ServerExecutable`: Path to the MCP server executable (default: "..\..\build\server.exe")
- `-KeepTestDirs`: If specified, test directories will not be deleted after testing
- `-TestDir`: Custom directory to use for test files
- `-Verbose`: Show detailed test information

#### `legacy/core/e2e_test.ps1`

An end-to-end legacy test script that verifies the server works correctly across multiple test cases.

**Usage**:

```powershell
.\legacy\core\e2e_test.ps1 [-ServerUrl <url>] [-TempDir <path>] [-KeepTempFiles] [-Verbose]
```

**Parameters**:

- `-ServerUrl`: URL of the running server (default: "http://localhost:8080")
- `-TempDir`: Custom directory to use for test files
- `-KeepTempFiles`: If specified, temporary files will not be deleted after the test
- `-Verbose`: Show detailed test information

### Legacy Strategy Test Scripts

#### `legacy/strategies/hybrid_strategy_test.ps1`

A detailed legacy test focused specifically on the hybrid execution strategy, verifying that modified code is correctly applied while maintaining project context.

**Usage**:

```powershell
.\legacy\strategies\hybrid_strategy_test.ps1 [-ServerExecutable <path>] [-TestDir <path>] [-KeepTestDirs] [-Verbose]
```

**Parameters**:

- `-ServerExecutable`: Path to the MCP server executable (default: "..\..\build\server.exe")
- `-TestDir`: Base directory for test files
- `-KeepTestDirs`: If specified, test directories won't be deleted after the test
- `-Verbose`: Show detailed step-by-step execution information

This script creates test projects with varying complexity, applies code modifications, and verifies the hybrid strategy correctly combines project context with modified code.

#### `legacy/strategies/hybrid_cli_test.ps1`

A CLI-focused legacy test for the hybrid execution strategy, calling the server executable directly.

#### `legacy/strategies/hybrid_strat_verify.ps1`

A simplified legacy test for verifying hybrid strategy functionality without requiring the server to be running.

### Legacy Basic Test Scripts

#### `legacy/basic/simple_test.ps1`

A minimal legacy test script that was formerly used as a starting point for quick tests or as a template for new test scripts.

### Legacy Utilities

#### `legacy/utils/test_utils.ps1`

Legacy shared utility functions for PowerShell testing scripts. This file contains functions for:

- Formatting and displaying test results
- Creating test projects with various configurations
- Running Go commands and capturing their output
- Validating test results with custom assertions

#### `legacy/utils/direct_test.go` and `legacy/utils/hybrid_test.go`

Legacy Go packages for testing the direct and hybrid execution strategies. These files have been updated to use direct Go command execution instead of the handler-based approach, preserving backward compatibility while eliminating dependencies on missing packages.

#### `legacy/utils/main.go`

A simple entry point for running legacy tests that use the direct and hybrid execution strategies without requiring the handlers package.

### Running Legacy PowerShell Tests

> **REMINDER**: Legacy tests are maintained for reference only and should not be used for new test development.

#### Legacy Quick Test

For a quick sanity check with legacy tests:

```powershell
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
.\legacy\basic\simple_test.ps1
```

#### Legacy Comprehensive Testing

To run comprehensive legacy tests with all tools and input modes:

```powershell
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
.\legacy\core\all_tools_test.ps1 -Verbose
```

#### Legacy Strategy-Specific Testing

To run legacy tests for the hybrid execution strategy specifically:

```powershell
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
.\legacy\strategies\hybrid_strategy_test.ps1 -Verbose
```

<!-- Section removed as it was duplicated -->

## Test Results

The Go and legacy tests output results with color-coded status indicators:

- ✅ PASS: Test completed successfully (green)
- ❌ FAIL: Test failed with details about the failure (red)
- ℹ️ INFO: Informational messages (white or cyan)

Each test provides timing information and a summary of passed and failed tests at the end.

## Adding New Tests

To add a new Go-based test:

1. Follow the patterns in `main.go`, `direct_runner.go`, or `hybrid_runner.go`
2. Use the direct execution approach with `os/exec` to run Go commands
3. Ensure your test covers specific scenarios or edge cases

> **Note**: All new tests should be implemented using Go, not PowerShell.

## Prerequisites

- Go 1.16 or newer (primary requirement)
- Go Development MCP Server (executable available in the build directory)
- PowerShell 5.1 or newer (only required for legacy tests)

## Troubleshooting

If Go-based tests fail:

1. Check that the server executable path is correct
2. Verify that Go is properly installed and in your PATH
3. Check for any required dependencies
4. Use the `-v` flag with `go run` for more verbose output
5. Inspect test output for specific error messages

For legacy PowerShell tests:

1. Use the `-VerboseOutput` flag for more detailed output
2. Use the `-KeepTestDirs` flag to preserve test directories for inspection

## Recent Improvements

### May 22, 2025 Updates

#### Documentation Modernization

- Restructured README to prioritize Go-based testing
- Moved PowerShell test documentation to legacy section
- Clarified that PowerShell tests are maintained for reference only
- Aligned documentation with current testing practices

#### Simplified Dependency Structure

- Removed dependency on handlers package and context parameters
- Moved all PowerShell tests and `test_utils.ps1` to the legacy directory
- Updated standalone test files to use direct Go command execution
- Eliminated all dependencies on old utility functions

#### Go-Native Testing Approach

- Replaced MCP handler-based execution with direct Go command execution
- Implemented simplified tool matching function directly in middleware
- Removed context usage from test runner for simplicity
- Built project with zero compilation errors

#### Improved Code Organization

- Maintained backward compatibility for legacy tests
- Ensured all test files follow the same pattern of direct command execution
- Simplified code structure by removing unnecessary abstractions
- Enhanced maintainability through reduced complexity
   
### Previous Enhancements

#### Improved Test Result Tracking

- Fixed test result counting in legacy scripts
- Added consistent exit code handling in all test scripts
- Proper scoping for test results collections

#### Enhanced Reliability

- Added server availability checking to legacy E2E tests
- Implemented file locking protection with retries
- Improved cleanup procedures with garbage collection

For more information on the Go Development MCP Server, refer to the main [README.md](../README.md).
