# Go Workspaces Implementation Prompt for Claude Sonnet 4

## Context and Background

You are tasked with implementing Go workspace support for the Go Development MCP Server. This is the next priority item on the roadmap. The system currently supports:

- Individual Go modules via project paths
- Code snippet execution via temporary directories  
- Hybrid execution combining both approaches
- Comprehensive toolchain integration (build, test, run, mod, format, analyze)

## Current Architecture Overview

The system uses a **Strategy Pattern** with three execution strategies:
- `CodeExecutionStrategy`: Executes provided code in temporary environments
- `ProjectExecutionStrategy`: Executes commands in existing Go project directories
- `HybridExecutionStrategy`: Combines both approaches

Input resolution is handled via `InputContext` with source types:
- `SourceCode`: Code provided directly
- `SourceProjectPath`: Existing Go project directory
- `SourceHybrid`: Both code and project path provided

## Go Workspaces Background

Go workspaces (introduced in Go 1.18) allow developers to work with multiple modules simultaneously. Key features:
- `go.work` file defines workspace root and module paths
- Enables local module development and testing
- Supports replace directives for local dependencies
- Allows seamless multi-module operations

## Requirements for Implementation

### Core Functionality Required

1. **Workspace Detection and Validation**
   - Detect `go.work` files in project hierarchies
   - Validate workspace structure and module references
   - Support both explicit workspace paths and auto-detection

2. **New Execution Strategy**
   - Implement `WorkspaceExecutionStrategy` following existing patterns
   - Handle commands that need workspace-wide context
   - Maintain compatibility with existing strategies

3. **Enhanced Input Resolution**
   - Add `SourceWorkspace` to `InputSource` enum
   - Extend `InputContext` with workspace-specific fields
   - Support workspace path parameter in tool calls

4. **Tool Integration** 
   - Extend all existing tools (build, test, run, mod, format, analyze) with workspace support
   - Add workspace-specific operations (go work init, go work use, go work sync)
   - Ensure proper working directory handling for multi-module operations

### Specific Implementation Tasks

#### 1. Core Workspace Types and Detection
```go
// Add to input.go
type InputSource int
const (
    // ... existing sources
    SourceWorkspace // New workspace source type
)

type InputContext struct {
    // ... existing fields
    WorkspacePath string   // Path to go.work file or workspace root
    WorkspaceModules []string // Discovered module paths within workspace
}
```

#### 2. Workspace Execution Strategy
Create `internal/tools/workspace.go` with:
- `WorkspaceExecutionStrategy` struct implementing `ExecutionStrategy` interface
- Workspace discovery and validation logic
- Multi-module command execution support

#### 3. Enhanced Tool Support
Extend each tool in `internal/tools/` to support:
- Workspace-aware operations
- Module selection within workspaces
- Proper working directory management

#### 4. New Workspace Management Tool
Implement `go_workspace` tool with subcommands:
- `init`: Initialize new workspace
- `use`: Add modules to existing workspace  
- `sync`: Synchronize workspace with module dependencies
- `edit`: Modify workspace configuration
- `vendor`: Vendor all workspace dependencies

#### 5. Testing Integration
Extend existing test framework with:
- Workspace-specific test fixtures
- Multi-module test scenarios
- Integration tests for workspace operations

## Technical Requirements

### Follow Existing Patterns
- Use the established strategy pattern architecture
- Maintain compatibility with existing `InputContext` resolution
- Follow the same error handling and response formatting patterns
- Implement comprehensive test coverage using the existing Go testing framework

### Code Quality Standards (Per Copilot Instructions)
- Apply parallel thinking approach with multiple solution paths
- Use "Skeleton of Thoughts" methodology for architecture design
- Implement minimum 100 iterations of refinement
- Follow idiomatic Go patterns and practices
- Ensure comprehensive error handling
- Include proper test coverage
- Validate against Go code smells and best practices

### Integration Points
- Extend `ResolveInput()` function in `input.go` to detect workspace sources
- Modify tool execution dispatching to use workspace strategy when appropriate
- Update configuration system to support workspace-specific settings
- Ensure proper cleanup of temporary workspace environments

## Expected Deliverables

1. **Core Implementation Files**
   - `internal/tools/workspace.go` - Workspace execution strategy and utilities
   - `internal/tools/workspace_tool.go` - Go workspace management tool
   - Updated `internal/tools/input.go` - Enhanced input resolution
   - Updated existing tool files with workspace support

2. **Testing Implementation**
   - Workspace-specific test fixtures in `internal/testing/fixtures/`
   - Integration tests in `internal/testing/`
   - End-to-end workspace tests following existing patterns

3. **Documentation Updates**
   - README.md workspace usage examples
   - Updated tool documentation with workspace parameters
   - Architecture documentation for workspace strategy

## Example Usage Scenarios

After implementation, users should be able to:

```go
// Initialize a new workspace
go_workspace(command: "init", workspace_path: "/path/to/workspace")

// Add modules to workspace
go_workspace(command: "use", workspace_path: "/path/to/workspace", modules: ["./module1", "./module2"])

// Build entire workspace
go_build(workspace_path: "/path/to/workspace")

// Test specific module in workspace
go_test(workspace_path: "/path/to/workspace", module: "./module1", verbose: true)

// Run workspace-wide operations
go_mod(command: "tidy", workspace_path: "/path/to/workspace")
```

## Implementation Approach

Use the parallel thinking methodology specified in the copilot instructions:

1. **Architecture Thread**: Design workspace strategy integration with existing patterns
2. **Implementation Thread**: Code the workspace detection, execution, and tool integration
3. **Error Handling Thread**: Implement comprehensive error handling for workspace operations
4. **Testing Thread**: Create comprehensive test coverage for all workspace scenarios

Process these threads simultaneously, iterating at least 100 times to refine the implementation until it meets all quality criteria and integrates seamlessly with the existing codebase.

## Success Criteria

- [ ] Workspace detection and validation working correctly
- [ ] All existing tools support workspace operations
- [ ] New `go_workspace` tool implemented with full subcommand support
- [ ] Comprehensive test coverage including edge cases
- [ ] Documentation updated with workspace examples
- [ ] Integration with existing strategy pattern maintained
- [ ] Error handling follows established patterns
- [ ] Code passes all quality checks and follows Go best practices

Implement this feature with the same level of sophistication and attention to detail as the existing codebase, ensuring it enhances the MCP server's capabilities while maintaining architectural consistency.
