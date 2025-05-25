# Getting Started with Go Development MCP Server Testing

This guide provides a quick overview of how to run tests for the Go Development MCP Server.

## Prerequisites

- PowerShell 5.1 or newer
- Go 1.16 or newer
- Go Development MCP Server (executable available in the build directory)

## Running Tests

### Option 1: Run Specific Test Categories

Using the master test runner script:

```powershell
# Run all tests
.\run_tests.ps1 -TestType all

# Run only basic tests
.\run_tests.ps1 -TestType basic

# Run only core tests
.\run_tests.ps1 -TestType core

# Run only strategy tests
.\run_tests.ps1 -TestType strategies
```

Additional parameters:

- `-VerboseOutput`: Show detailed test information
- `-KeepTestDirs`: Keep temporary test directories for inspection
- `-ServerExecutable <path>`: Specify a custom server executable path

### Option 2: Run Individual Tests

Run specific test scripts directly:

```powershell
# Basic test
.\basic\simple_test.ps1

# Core tests
.\core\all_tools_test.ps1
.\core\e2e_test.ps1

# Strategy tests
.\strategies\hybrid_strategy_test.ps1
.\strategies\hybrid_strat_verify.ps1
```

## Directory Structure

- `basic/`: Simple test cases for quick verification
- `core/`: Comprehensive test suites that test all tools and modes
- `strategies/`: Tests focused on specific execution strategies
- `utils/`: Shared utility functions and Go test packages

## Test Types

### Basic Tests

Simple tests to verify basic functionality. These are quick to run and useful for smoke testing.

### Core Tests

- **all_tools_test.ps1**: Tests all Go tools (build, run, format, test, mod, analyze) with all input modes
- **e2e_test.ps1**: End-to-end tests that make HTTP requests to a running server

### Strategy Tests

Tests that specifically focus on the hybrid execution strategy and other specialized testing needs.

## Troubleshooting

### Server Connection Issues

The e2e_test.ps1 script requires a running MCP server. If the server is not available, the test will
be skipped instead of failing. Make sure your server is running at http://localhost:8080 before running
the e2e tests.

### File Locking Issues

If you encounter "file in use" errors during cleanup, the testing framework will now attempt to:
1. Force garbage collection to release handles
2. Retry deletion multiple times
3. Continue execution even if deletion fails

If persistent file locking issues occur, restart your PowerShell session or use the `-KeepTestDirs`
parameter to skip cleanup.

### Basic Tests
Simple tests to verify basic functionality. These are quick to run and useful for smoke testing.

### Core Tests
- **all_tools_test.ps1**: Tests all Go tools (build, run, format, test, mod, analyze) with all input modes
- **e2e_test.ps1**: End-to-end tests that make HTTP requests to a running server

### Strategy Tests
Tests that specifically focus on the hybrid execution strategy and other specialized testing needs.

## Troubleshooting

### Server Connection Issues
The e2e_test.ps1 script requires a running MCP server. If the server is not available, the test will
be skipped instead of failing. Make sure your server is running at http://localhost:8080 before running
the e2e tests.

### File Locking Issues
If you encounter "file in use" errors during cleanup, the testing framework will now attempt to:
1. Force garbage collection to release handles
2. Retry deletion multiple times
3. Continue execution even if deletion fails

If persistent file locking issues occur, restart your PowerShell session or use the `-KeepTestDirs`
parameter to skip cleanup.

- **Basic tests**: Simple, quick-running tests for sanity checks
- **Core tests**: Comprehensive tests covering all tools and input modes
- **Strategy tests**: Tests focused on specific execution strategies like hybrid execution

## Troubleshooting

If tests fail:

1. Use the `-Verbose` flag for more detailed output
2. Use the `-KeepTestDirs` flag to preserve test directories for inspection
3. Check that the server executable path is correct
4. Verify that Go is properly installed and in your PATH

For more details, see the comprehensive [README.md](README.md).
