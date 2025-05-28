# Unit Tests

This directory contains unit tests for the Go Development MCP Server components.

## Test Files

### `workspace_test.go`
- **Purpose**: Tests workspace detection and parsing functionality
- **Coverage**: 
  - `tools.IsWorkspace()` - Detects if a directory contains a `go.work` file
  - `tools.ParseGoWorkFile()` - Parses modules from `go.work` files
  - Workspace input resolution logic

### `workspace_tool_test.go`
- **Purpose**: Tests workspace tool functionality and command execution
- **Coverage**:
  - `tools.ExecuteGoWorkspaceTool()` - Executes workspace-related commands
  - Workspace command validation and parameter handling
  - Integration with Go workspace commands (`init`, `sync`, `use`, etc.)

## Running Tests

To run all unit tests in this directory:

```powershell
cd scripts/testing/unit
go test -v .
```

To run with coverage:

```powershell
go test -v -cover .
```

## Test Structure

These tests use:
- Go's standard `testing` package
- Table-driven test patterns for comprehensive coverage
- Temporary directories for isolated test environments
- Real Go workspace command execution for integration validation

The tests are designed to be fast, isolated, and provide comprehensive coverage of workspace functionality without requiring external dependencies.
