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

	// Create MCP server with enhanced configuration
	s := server.NewMCPServer(
		"Go Development Tools",
		cfg.Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
		server.WithLogging(),
		server.WithHooks(hooks),
	)

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

// registerTools registers all available Go development tools with the MCP server.
func registerTools(s *server.MCPServer) { // Register go_build tool - simplified for v0.9.0
	buildTool := mcp.NewTool("go_build",
		mcp.WithDescription("Compile Go code into executable programs. Natural language aliases: compile, build, create executable. Examples: compile this Go code, build this program, create an executable from this code, make a binary from this Go file."),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code to compile")),
		mcp.WithString("outputPath", mcp.Description("Desired output location for the compiled executable")),
		mcp.WithString("buildTags", mcp.Description("Optional build tags for conditional compilation")),
		mcp.WithString("mainFile", mcp.Description("Main file name to compile"), mcp.DefaultString("main.go")),
	)
	s.AddTool(buildTool, tools.ExecuteGoBuildTool)
	// Register go_test tool
	testTool := mcp.NewTool("go_test",
		mcp.WithDescription("Run tests on Go code. Natural language aliases: test, unit test, run tests. Examples: test this Go code, run unit tests, check if tests pass, verify test coverage."),
		mcp.WithString("code", mcp.Description("Go source code")),
		mcp.WithString("testCode", mcp.Description("Go test code")),
		mcp.WithString("testPattern", mcp.Description("Test pattern to run specific tests")),
		mcp.WithBoolean("verbose", mcp.Description("Verbose output"), mcp.DefaultBool(false)),
		mcp.WithBoolean("coverage", mcp.Description("Enable coverage reporting"), mcp.DefaultBool(false)),
	)
	s.AddTool(testTool, tools.ExecuteGoTestTool)
	// Register go_run tool - replace WithArray with WithObject
	runTool := mcp.NewTool("go_run",
		mcp.WithDescription("Compile and run Go code. Natural language aliases: run, execute, start. Examples: run this Go code, execute this program, start this application with arguments."),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code")),
		// Replace WithArray with WithObject which should be more basic and available in v0.9.0
		mcp.WithObject("args", mcp.Description("Command-line arguments")),
		mcp.WithNumber("timeoutSecs", mcp.Description("Timeout in seconds"), mcp.DefaultNumber(30)),
	)
	s.AddTool(runTool, tools.ExecuteGoRunTool)
	// Register go_mod tool
	modTool := mcp.NewTool("go_mod",
		mcp.WithDescription("Manage Go module dependencies. Natural language aliases: dependencies, modules, manage dependencies. Examples: initialize a new module, update dependencies, add a new dependency, tidy up module dependencies."),
		mcp.WithString("command", mcp.Required(), mcp.Description("Subcommand (init, tidy, etc.)")),
		mcp.WithString("modulePath", mcp.Description("Project directory")),
		mcp.WithString("code", mcp.Description("Optional Go code")),
	)
	s.AddTool(modTool, tools.ExecuteGoModTool)
	// Register go_fmt tool
	fmtTool := mcp.NewTool("go_fmt",
		mcp.WithDescription("Format Go code. Natural language aliases: format, beautify, pretty-print. Examples: format this Go code, beautify this program, fix the formatting of this code, make this code look nice."),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code")),
	)
	s.AddTool(fmtTool, tools.ExecuteGoFmtTool)
	// Register go_analyze tool
	analyzeTool := mcp.NewTool("go_analyze",
		mcp.WithDescription("Analyze Go code for issues. Natural language aliases: lint, check, validate, inspect. Examples: analyze this Go code for issues, check for bugs, validate this function, inspect code quality."),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code")),
		mcp.WithBoolean("vet", mcp.Description("Run go vet"), mcp.DefaultBool(true)),
	)
	s.AddTool(analyzeTool, tools.ExecuteGoAnalyzeTool)

	log.Printf("Registered %d tools", 6)
}
