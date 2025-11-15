# Refactoring Example: Using prepareWorkspaceArgs Helper

This document shows how the new `prepareWorkspaceArgs()` helper function will be used to eliminate code duplication across tool files.

## Example: build.go Refactoring

### BEFORE (Current Implementation)
```go
func ExecuteGoBuildTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Extract parameters
	outputPath := mcp.ParseString(req, "outputPath", "")
	buildTags := mcp.ParseString(req, "buildTags", "")
	module := mcp.ParseString(req, "module", "")

	// Prepare build args
	args := []string{"build"}
	if buildTags != "" {
		args = append(args, "-tags", buildTags)
	}
	if outputPath != "" {
		args = append(args, "-o", outputPath)
	} else if input.Source == SourceCode {
		outputPath = "output"
		args = append(args, "-o", outputPath)
	}

	// Handle different source types - DUPLICATED CODE BLOCK
	switch input.Source {
	case SourceCode:
		args = append(args, input.MainFile)
	case SourceWorkspace:
		if module != "" {
			args = append(args, module)
		} else {
			args = append(args, "./...")
		}
	default:
		args = append(args, "./...")
	}

	// Execute using appropriate strategy
	strategy := GetExecutionStrategy(input, args...)
	result, err := strategy.Execute(ctx, input, args)
	// ... rest of function
}
```

### AFTER (Using Helper Function)
```go
func ExecuteGoBuildTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Resolve input
	input, err := ResolveInput(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Extract parameters
	outputPath := mcp.ParseString(req, "outputPath", "")
	buildTags := mcp.ParseString(req, "buildTags", "")
	module := mcp.ParseString(req, "module", "")

	// Prepare build args
	args := []string{"build"}
	if buildTags != "" {
		args = append(args, "-tags", buildTags)
	}
	if outputPath != "" {
		args = append(args, "-o", outputPath)
	} else if input.Source == SourceCode {
		outputPath = "output"
		args = append(args, "-o", outputPath)
	}

	// Handle different source types - USE HELPER FUNCTION
	args = prepareWorkspaceArgs(input, module, args)

	// Special handling for SourceCode (add main file)
	if input.Source == SourceCode {
		args = append(args, input.MainFile)
	}

	// Execute using appropriate strategy
	strategy := GetExecutionStrategy(input, args...)
	result, err := strategy.Execute(ctx, input, args)
	// ... rest of function
}
```

### Lines Removed
- Lines 38-54 (17 lines) replaced with 2 lines + 4-line special case
- Net reduction: ~11 lines per file
- **Total reduction across 6 files: ~66 lines of duplicated code**

---

## Example: test.go Refactoring

### BEFORE
```go
func ExecuteGoTestTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// ... parameter extraction ...

	// Prepare test args
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if coverage {
		args = append(args, "-cover")
	}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}

	// DUPLICATED CODE BLOCK
	switch input.Source {
	case SourceWorkspace:
		if module != "" {
			args = append(args, module)
		} else {
			args = append(args, "./...")
		}
	default:
		args = append(args, "./...")
	}

	// ... rest of function
}
```

### AFTER
```go
func ExecuteGoTestTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// ... parameter extraction ...

	// Prepare test args
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if coverage {
		args = append(args, "-cover")
	}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}

	// Use helper function to prepare workspace/module args
	args = prepareWorkspaceArgs(input, module, args)

	// ... rest of function
}
```

---

## Example: analyze.go Refactoring

### BEFORE
```go
func ExecuteGoAnalyzeTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// ... parameter extraction ...

	// Prepare vet args
	args := []string{"vet"}

	// DUPLICATED CODE BLOCK
	switch input.Source {
	case SourceWorkspace:
		if module != "" {
			args = append(args, module)
		} else {
			args = append(args, "./...")
		}
	case SourceCode:
		return executeCodeAnalysis(ctx, input.Code, runVet)
	default:
		args = append(args, "./...")
	}

	// ... rest of function
}
```

### AFTER
```go
func ExecuteGoAnalyzeTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// ... parameter extraction ...

	// Special handling for code analysis
	if input.Source == SourceCode {
		return executeCodeAnalysis(ctx, input.Code, runVet)
	}

	// Prepare vet args using helper
	args := []string{"vet"}
	args = prepareWorkspaceArgs(input, module, args)

	// ... rest of function
}
```

---

## Benefits of This Refactoring

1. **Code Reuse**: Eliminates ~66 lines of duplicated code across 6 files
2. **Maintainability**: Changes to workspace/module logic only need to be made in one place
3. **Testability**: The helper function has 100% test coverage with comprehensive edge cases
4. **Consistency**: Ensures all tools handle workspace/module arguments identically
5. **Readability**: Makes tool files cleaner and easier to understand
6. **Safety**: Defensive copying in helper prevents unintended side effects

## Files That Will Be Refactored (By Other Agents)

1. `/home/user/go-dev-mcp/internal/tools/build.go`
2. `/home/user/go-dev-mcp/internal/tools/analyze.go`
3. `/home/user/go-dev-mcp/internal/tools/fmt.go`
4. `/home/user/go-dev-mcp/internal/tools/test.go`
5. `/home/user/go-dev-mcp/internal/tools/run.go`
6. `/home/user/go-dev-mcp/internal/tools/mod.go` (minimal changes - already has different logic)

## Note on run.go

The `run.go` file has a slight variation where it uses `"."` instead of `"./..."` for workspace execution without a module. This will need special handling:

```go
// In run.go, you might need:
args = prepareWorkspaceArgs(input, module, args)

// Then override for the run command's specific behavior:
if input.Source == SourceWorkspace && module == "" {
	// run.go uses "." instead of "./..."
	args[len(args)-1] = "."
}
```

Or, consider if `run.go` should conform to the standard pattern and use `"./..."` like the other tools for consistency.
