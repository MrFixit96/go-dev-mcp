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

	// Register tools directly with the MCP server
	registerTools(s)

	// Start the server with context
	log.Println("Server is ready to accept connections")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// registerTools directly registers all tools with the MCP server
func registerTools(s *server.MCPServer) { // Register go_build tool
	buildTool := mcp.NewTool("go_build",
		mcp.WithDescription("Compile Go code into executable programs."),
		mcp.WithString("code",
			mcp.Description("Complete Go source code to compile. Must include a main package and function.")),
		mcp.WithString("project_path",
			mcp.Description("Path to an existing Go project directory.")),
		mcp.WithString("workspace_path",
			mcp.Description("Path to a Go workspace directory (go.work file).")),
		mcp.WithString("module",
			mcp.Description("Specific module to build within a workspace.")),
		mcp.WithString("outputPath",
			mcp.Description("Path where the compiled executable should be saved.")),
		mcp.WithString("buildTags",
			mcp.Description("Build tags to use during compilation.")))

	s.AddTool(buildTool, tools.ExecuteGoBuildTool)
	// Register go_run tool
	runTool := mcp.NewTool("go_run",
		mcp.WithDescription("Run Go code directly."),
		mcp.WithString("code",
			mcp.Description("Complete Go source code to run. Must include a main package and function.")),
		mcp.WithString("project_path",
			mcp.Description("Path to an existing Go project directory.")),
		mcp.WithString("workspace_path",
			mcp.Description("Path to a Go workspace directory (go.work file).")),
		mcp.WithString("module",
			mcp.Description("Specific module to run within a workspace.")),
		mcp.WithNumber("timeoutSecs",
			mcp.Description("Maximum execution time in seconds before the program is terminated."),
			mcp.DefaultNumber(30)))

	s.AddTool(runTool, tools.ExecuteGoRunTool)
	// Register go_fmt tool
	fmtTool := mcp.NewTool("go_fmt",
		mcp.WithDescription("Format Go code according to standard Go formatting rules."),
		mcp.WithString("code",
			mcp.Description("Go source code to format.")),
		mcp.WithString("project_path",
			mcp.Description("Path to an existing Go project directory.")),
		mcp.WithString("workspace_path",
			mcp.Description("Path to a Go workspace directory (go.work file).")),
		mcp.WithString("module",
			mcp.Description("Specific module to format within a workspace.")))

	s.AddTool(fmtTool, tools.ExecuteGoFmtTool)
	// Register go_test tool
	testTool := mcp.NewTool("go_test",
		mcp.WithDescription("Run tests on Go code."),
		mcp.WithString("code",
			mcp.Description("Main Go source code to test.")),
		mcp.WithString("testCode",
			mcp.Description("Go test code that contains test functions.")),
		mcp.WithString("project_path",
			mcp.Description("Path to an existing Go project directory.")),
		mcp.WithString("workspace_path",
			mcp.Description("Path to a Go workspace directory (go.work file).")),
		mcp.WithString("module",
			mcp.Description("Specific module to test within a workspace.")),
		mcp.WithString("testPattern",
			mcp.Description("Pattern to filter which tests to run.")),
		mcp.WithBoolean("verbose",
			mcp.Description("Enable verbose output."),
			mcp.DefaultBool(false)),
		mcp.WithBoolean("coverage",
			mcp.Description("Enable coverage reporting."),
			mcp.DefaultBool(false)))

	s.AddTool(testTool, tools.ExecuteGoTestTool)
	// Register go_mod tool
	modTool := mcp.NewTool("go_mod",
		mcp.WithDescription("Manage Go module dependencies."),
		mcp.WithString("command",
			mcp.Description("Module command to execute (init, tidy, vendor, verify, why, graph, download)."),
			mcp.Required()),
		mcp.WithString("modulePath",
			mcp.Description("Module path for 'init' command.")),
		mcp.WithString("code",
			mcp.Description("Go source code for context.")),
		mcp.WithString("project_path",
			mcp.Description("Path to an existing Go project directory.")),
		mcp.WithString("workspace_path",
			mcp.Description("Path to a Go workspace directory (go.work file).")),
		mcp.WithString("module",
			mcp.Description("Specific module to manage within a workspace.")))

	s.AddTool(modTool, tools.ExecuteGoModTool)
	// Register go_analyze tool
	analyzeTool := mcp.NewTool("go_analyze",
		mcp.WithDescription("Analyze Go code for potential issues using go vet."),
		mcp.WithString("code",
			mcp.Description("Go source code to analyze.")),
		mcp.WithString("project_path",
			mcp.Description("Path to an existing Go project directory.")),
		mcp.WithString("workspace_path",
			mcp.Description("Path to a Go workspace directory (go.work file).")),
		mcp.WithString("module",
			mcp.Description("Specific module to analyze within a workspace.")),
		mcp.WithBoolean("vet",
			mcp.Description("Run go vet analysis."),
			mcp.DefaultBool(true)))

	s.AddTool(analyzeTool, tools.ExecuteGoAnalyzeTool)

	// Register go_workspace tool
	workspaceTool := mcp.NewTool("go_workspace",
		mcp.WithDescription("Manage Go workspaces for multi-module development."),
		mcp.WithString("command",
			mcp.Description("Workspace command to execute (init, use, sync, edit, vendor, info)."),
			mcp.Required()),
		mcp.WithString("workspace_path",
			mcp.Description("Path to the workspace directory where go.work file is or will be created.")),
		mcp.WithString("module_path",
			mcp.Description("Path to a module for 'use' command, or module name for 'edit' command.")),
		mcp.WithString("version",
			mcp.Description("Version constraint for 'edit' command (e.g., v1.2.3, latest).")),
		mcp.WithBoolean("recursive",
			mcp.Description("Search for modules recursively when using 'use' command."),
			mcp.DefaultBool(false)))

	s.AddTool(workspaceTool, tools.ExecuteGoWorkspaceTool)

	log.Printf("Registered comprehensive tools with MCP server")
}
