# End-to-End Behavioral Testing for Go Development MCP Server

This directory contains a comprehensive set of scripts for end-to-end behavioral testing of the Go Development MCP Server. These tests verify that the server works correctly by executing real-world scenarios with actual Go projects.

## Directory Structure

The testing scripts are organized into the following directories:

```
testing/
├── basic/             # Simple test cases for quick verification
├── core/              # Comprehensive test suites that test all tools and modes
├── strategies/        # Tests focused on specific execution strategies
├── utils/             # Shared utility functions and Go test packages
└── README.md          # This documentation file
```

## Test Approach

The Go Development MCP Server communicates via stdin/stdout rather than HTTP, allowing direct execution of Go tools with different input sources. Our testing approach simulates real-world interactions with the server, focusing on:

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

## Core Test Scripts

### `core/all_tools_test.ps1`

A comprehensive test suite that tests all six tools provided by the Go Development MCP Server with all applicable input modes.

**Usage**:

```powershell
.\core\all_tools_test.ps1 [-ServerExecutable <path>] [-KeepTestDirs] [-TestDir <path>] [-Verbose]
```

**Parameters**:

- `-ServerExecutable`: Path to the MCP server executable (default: "..\..\build\server.exe")
- `-KeepTestDirs`: If specified, test directories will not be deleted after testing
- `-TestDir`: Custom directory to use for test files
- `-Verbose`: Show detailed test information

### `core/e2e_test.ps1`

An end-to-end test script that verifies the server works correctly across multiple test cases.

**Usage**:

```powershell
.\core\e2e_test.ps1 [-ServerUrl <url>] [-TempDir <path>] [-KeepTempFiles] [-Verbose]
```

**Parameters**:

- `-ServerUrl`: URL of the running server (default: "http://localhost:8080")
- `-TempDir`: Custom directory to use for test files
- `-KeepTempFiles`: If specified, temporary files will not be deleted after the test
- `-Verbose`: Show detailed test information

## Strategy Test Scripts

### `strategies/hybrid_strategy_test.ps1`

A detailed test focused specifically on the hybrid execution strategy, verifying that modified code is correctly applied while maintaining project context.

**Usage**:

```powershell
.\strategies\hybrid_strategy_test.ps1 [-ServerExecutable <path>] [-TestDir <path>] [-KeepTestDirs] [-Verbose]
```

**Parameters**:

- `-ServerExecutable`: Path to the MCP server executable (default: "..\..\build\server.exe")
- `-TestDir`: Base directory for test files
- `-KeepTestDirs`: If specified, test directories won't be deleted after the test
- `-Verbose`: Show detailed step-by-step execution information

This script creates test projects with varying complexity, applies code modifications, and verifies the hybrid strategy correctly combines project context with modified code.

### `strategies/hybrid_cli_test.ps1`

A CLI-focused test for the hybrid execution strategy, calling the server executable directly.

### `strategies/hybrid_strat_verify.ps1`

A simplified test for verifying hybrid strategy functionality without requiring the server to be running.

## Basic Test Scripts

### `basic/simple_test.ps1`

A minimal test script that can be used as a starting point for quick tests or as a template for new test scripts.

## Utility Components

### `utils/test_utils.ps1`

Shared utility functions for testing scripts.

### `utils/direct_test.go` and `utils/hybrid_test.go`

Go packages for direct testing of the server's Go components.

## Running the Tests

### Quick Test

For a quick sanity check, run the simple test:

```powershell
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
.\basic\simple_test.ps1
```

### Comprehensive Testing

To test all tools with all input modes:

```powershell
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
.\core\all_tools_test.ps1 -Verbose
```

### Strategy-Specific Testing

To test the hybrid execution strategy specifically:

```powershell
cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
.\strategies\hybrid_strategy_test.ps1 -Verbose
```

## Understanding Test Results

The scripts output test results with color-coded status indicators:

- ✅ PASS: Test completed successfully (green)
- ❌ FAIL: Test failed with details about the failure (red)
- ℹ️ INFO: Informational messages (white or cyan)

Each test provides timing information and a summary of passed and failed tests at the end.

## Adding New Tests

To add a new test script:

1. Use `simple_test.ps1` as a starting point
2. Follow the patterns used in `all_tools_test.ps1` or `hybrid_strategy_test.ps1`
3. Ensure your test covers specific scenarios or edge cases

## Prerequisites

- PowerShell 5.1 or newer
- Go 1.16 or newer
- Go Development MCP Server (executable available in the build directory)

## Troubleshooting

If tests fail:

1. Check that the server executable path is correct
2. Verify that Go is properly installed and in your PATH
3. Check for any required dependencies
4. Use the `-Verbose` flag for more detailed output
5. Use the `-KeepTestDirs` flag to preserve test directories for inspection
