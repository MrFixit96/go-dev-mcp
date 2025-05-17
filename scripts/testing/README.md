# End-to-End Behavioral Testing for Go Development MCP Server

This directory contains scripts for end-to-end behavioral testing of the Go Development MCP Server. These tests verify that the server works correctly by executing real-world scenarios with actual Go projects.

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

## Test Scripts

This directory contains the following test scripts:

### `all_tools_test.ps1`

A comprehensive test suite that tests all six tools provided by the Go Development MCP Server with all applicable input modes.

**Usage**:
```powershell
.\all_tools_test.ps1 [-ServerExecutable <path>] [-KeepTestDirs] [-TestDir <path>] [-Verbose]
```

**Parameters**:
- `-ServerExecutable`: Path to the MCP server executable (default: "..\..\build\server.exe")
- `-KeepTestDirs`: If specified, test directories will not be deleted after testing
- `-TestDir`: Custom directory to use for test files
- `-Verbose`: Show detailed test information

This script creates sample Go projects and tests each tool with various input combinations, verifying outputs match expectations.

### `hybrid_strategy_test.ps1`

A detailed test focused specifically on the hybrid execution strategy, verifying that modified code is correctly applied while maintaining project context.

**Usage**:
```powershell
.\hybrid_strategy_test.ps1 [-TestDir <path>] [-KeepTestDirs] [-Verbose]
```

**Parameters**:
- `-TestDir`: Base directory for test files
- `-KeepTestDirs`: If specified, test directories won't be deleted after the test
- `-Verbose`: Show detailed step-by-step execution information

This script creates test projects with varying complexity, applies code modifications, and verifies the hybrid strategy correctly combines project context with modified code.

### `simple_test.ps1`

A minimal test script that can be used as a starting point for quick tests or as a template for new test scripts.

## Test Coverage

The tests verify:

1. **Basic functionality**: Format, build, and run Go code
2. **Input sources**: Tests all input modes:
   - Code-only (inline code provided)
   - Project path-only (directory path provided)
   - Hybrid (both code and project path provided)
3. **Error handling**: Tests responses for invalid input
4. **Strategy selection**: Verifies appropriate execution strategies are selected

## Prerequisites

- PowerShell 5.1 or newer
- Go 1.16 or newer
- Go Development MCP Server (executable available in the build directory)

## Running the Tests

1. To test all tools with all input modes:
   ```powershell
   cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
   .\all_tools_test.ps1 -Verbose
   ```

2. To test the hybrid execution strategy specifically:
   ```powershell
   cd c:\Users\James\Documents\go-dev-mcp\scripts\testing
   .\hybrid_strategy_test.ps1 -Verbose
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

## Troubleshooting

If tests are failing:

1. Use the `-Verbose` flag to see detailed execution information
2. Use `-KeepTestDirs` to inspect the temporary project files
3. Check for Go environment issues (run `go env` to verify)
4. Verify the server executable is properly built
5. Isolate failures by running specific tests or sections
