package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MrFixit96/go-dev-mcp/internal/config"
	customServer "github.com/MrFixit96/go-dev-mcp/internal/server"
	"github.com/MrFixit96/go-dev-mcp/internal/tools"
)

// main is the entry point for the Go Development MCP Server.
// It initializes the server, registers tools, and handles graceful shutdown.
func main() {
	log.Println("Starting Go Development MCP Server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load configuration: %v. Using defaults.", err)
		cfg = config.DefaultConfig()
	}
	// Create hooks for enhanced server observability
	hooks := &server.Hooks{}
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Printf("Request received: %s, ID: %v", method, id)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		log.Printf("Request successful: %s, ID: %v", method, id)
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("Error processing request: %s, ID: %v, Error: %v", method, id, err)
	})

	// Get fuzzy matching middleware
	fuzzyMiddleware := customServer.FuzzyMatchMiddleware(cfg, nil)

	// Create MCP server with enhanced configuration
	s := server.NewMCPServer(
		"Go Development Tools",
		cfg.Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
		server.WithHooks(hooks),
		server.WithToolHandlerMiddleware(fuzzyMiddleware),
	)

	// Log that fuzzy matching middleware is applied
	log.Println("Fuzzy matching middleware applied")

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	// Register tools with proper mcp tooling
	registerTools(s)

	// Start the server with context
	log.Println("Server is ready to accept connections")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Project path parameter definition - reused across tools
func projectPathParam() mcp.PropertyOption {
	return mcp.Description(`Path to an existing Go project directory. When provided, the tool will operate on this directory instead of using the 'code' parameter.

- OS-specific path examples:
  - Windows: 'C:\Users\username\go\projects\myapp' or 'D:\projects\myapp'
  - macOS: '/Users/username/go/projects/myapp' or '~/go/projects/myapp'
  - Linux: '/home/username/go/projects/myapp' or '~/go/projects/myapp'

- Valid Go project characteristics:
  - Should contain a go.mod file (run 'go mod init' if not present)
  - Should have a proper Go module structure
  - For tools requiring a main package (build, run), ensure a main.go file exists

- Path guidelines:
  - Absolute paths are recommended and more reliable
  - Relative paths are resolved relative to the current working directory
  - Environment variables are not expanded (use full paths)
  - UNC paths on Windows are supported (\\\\server\\share\\path)

- Precedence behavior:
  - When both 'code' and 'project_path' are provided, 'project_path' takes precedence
  - The tool will operate on the project directory, ignoring the provided code

- Troubleshooting:
  - Ensure the path exists and is accessible
  - Verify you have read/write permissions to the directory
  - For Windows paths, use either forward slashes (/) or escaped backslashes (\\\\)
  - If using symlinks, ensure they are resolved correctly
  - Check that the directory contains valid Go code files`)
}

// Code parameter definition with modified requirement
func codeParam(required bool) mcp.PropertyOption {
	if required {
		// For required code, return the description
		// We'll add the Required() option separately in the WithString call
		return mcp.Description("Complete Go source code to compile. Must include a main package and function. Required when project_path is not provided.")
	}
	return mcp.Description("Complete Go source code to compile. Must include a main package and function. Not required when project_path is provided.")
}

// registerTools registers all available Go development tools with the MCP server.
func registerTools(s *server.MCPServer) {
	// Register go_build tool
	buildTool := mcp.NewTool("go_build",
		mcp.WithDescription("Compile Go code into executable programs. Takes source code or a project path and generates a binary executable that can be run on the target platform."),
		mcp.WithString("code", codeParam(false)),
		mcp.WithString("project_path", projectPathParam()),
		mcp.WithString("outputPath",
			mcp.Description("Path where the compiled executable should be saved. If not specified, a temporary location will be used. Example: './bin/myapp' or 'C:\\Users\\username\\app.exe'")),
		mcp.WithString("buildTags",
			mcp.Description("Optional build tags for conditional compilation. Multiple tags should be separated by spaces. Example: 'linux debug' or 'windows,gui'")),
		mcp.WithString("mainFile",
			mcp.Description("Name of the main file to create for the code. Default is 'main.go'. Only applies when using the 'code' parameter."),
			mcp.DefaultString("main.go")),
	)
	s.AddTool(buildTool, tools.ExecuteGoBuildTool)

	// Register go_test tool
	testTool := mcp.NewTool("go_test",
		mcp.WithDescription("Run tests on Go code. Executes test functions in code and reports results, with optional coverage information."),
		mcp.WithString("code",
			mcp.Description("Main Go source code to test. This is the code that the tests will be run against. Not required if testCode is provided or if project_path is used.")),
		mcp.WithString("project_path", projectPathParam()),
		mcp.WithString("testCode",
			mcp.Description("Go test code that contains test functions. Should include functions starting with 'Test' that take *testing.T as a parameter. Example: 'func TestAdd(t *testing.T) { ... }'")),
		mcp.WithString("testPattern",
			mcp.Description("Pattern to filter which tests to run. Use Go's testing pattern syntax, like 'TestAuth*' to run all tests starting with 'TestAuth'.")),
		mcp.WithBoolean("verbose",
			mcp.Description("Enable verbose output that shows each test as it runs with detailed information."),
			mcp.DefaultBool(false)),
		mcp.WithBoolean("coverage",
			mcp.Description("Generate code coverage statistics showing what percentage of code is tested."),
			mcp.DefaultBool(false)),
	)
	s.AddTool(testTool, tools.ExecuteGoTestTool)

	// Register go_run tool
	runTool := mcp.NewTool("go_run",
		mcp.WithDescription("Compile and run Go code in one step. The code is compiled in memory and executed immediately with any specified arguments."),
		mcp.WithString("code", codeParam(false)),
		mcp.WithString("project_path", projectPathParam()),
		mcp.WithObject("args",
			mcp.Description("Command-line arguments to pass to the program. Specify as an object where keys are argument positions and values are the arguments. Example: {\"0\": \"--verbose\", \"1\": \"filename.txt\"}")),
		mcp.WithNumber("timeoutSecs",
			mcp.Description("Maximum execution time in seconds before the program is terminated. Prevents infinite loops or long-running operations."),
			mcp.DefaultNumber(30)),
	)
	s.AddTool(runTool, tools.ExecuteGoRunTool)

	// Register go_mod tool
	modTool := mcp.NewTool("go_mod",
		mcp.WithDescription("Manage Go module dependencies. Handles operations related to go.mod files including initialization, adding dependencies, and tidying."),
		mcp.WithString("command", mcp.Required(),
			mcp.Description("The go mod subcommand to execute. Valid options include: 'init', 'tidy', 'download', 'vendor', 'verify', 'why', 'edit', 'graph'. Example: 'init' or 'tidy'")),
		mcp.WithString("project_path",
			mcp.Description("Existing project directory to perform module operations on. Required for operations on existing modules like 'tidy' or 'vendor'.")),
		mcp.WithString("modulePath",
			mcp.Description("Module path to use when initializing a new module. For 'init' command, this is the module name like 'github.com/username/project'.")),
		mcp.WithString("code",
			mcp.Description("Optional Go code to analyze for dependencies. When provided with 'tidy', it will ensure all imports in this code are properly reflected in go.mod.")),
	)
	s.AddTool(modTool, tools.ExecuteGoModTool)

	// Register go_fmt tool
	fmtTool := mcp.NewTool("go_fmt",
		mcp.WithDescription("Format Go code according to standard Go formatting rules. Makes code consistent with Go's style conventions."),
		mcp.WithString("code", codeParam(false)),
		mcp.WithString("project_path", projectPathParam()),
	)
	s.AddTool(fmtTool, tools.ExecuteGoFmtTool)

	// Register go_analyze tool
	analyzeTool := mcp.NewTool("go_analyze",
		mcp.WithDescription("Analyze Go code for potential issues, bugs, and style problems. Identifies common mistakes and optimization opportunities."),
		mcp.WithString("code", codeParam(false)),
		mcp.WithString("project_path", projectPathParam()),
		mcp.WithBoolean("vet",
			mcp.Description("Run go vet, which performs static analysis to find potential bugs, suspicious constructs, and other issues."),
			mcp.DefaultBool(true)),
	)
	s.AddTool(analyzeTool, tools.ExecuteGoAnalyzeTool)

	log.Printf("Registered %d tools", 6)
}
