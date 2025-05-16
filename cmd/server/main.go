package main

import (
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
	
	// Create MCP server - use the simplest constructor for v0.19.0
	s := server.NewMCPServer(
		"Go Development Tools",
		cfg.Version,
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
func registerTools(s *server.MCPServer) {
	// Register go_build tool - simplified for v0.19.0
	buildTool := mcp.NewTool("go_build", 
		mcp.WithDescription("Compile Go code"),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code")),
		mcp.WithString("outputPath", mcp.Description("Desired output location")),
		mcp.WithString("buildTags", mcp.Description("Build tags")),
		mcp.WithString("mainFile", mcp.Description("Main file name"), mcp.DefaultString("main.go")),
	)
	s.AddTool(buildTool, tools.ExecuteGoBuildTool)
	
	// Register go_test tool
	testTool := mcp.NewTool("go_test",
		mcp.WithDescription("Run tests on Go code"),
		mcp.WithString("code", mcp.Description("Go source code")),
		mcp.WithString("testCode", mcp.Description("Go test code")),
		mcp.WithString("testPattern", mcp.Description("Test pattern to run specific tests")),
		mcp.WithBoolean("verbose", mcp.Description("Verbose output"), mcp.DefaultBool(false)),
		mcp.WithBoolean("coverage", mcp.Description("Enable coverage reporting"), mcp.DefaultBool(false)),
	)
	s.AddTool(testTool, tools.ExecuteGoTestTool)
	
	// Register go_run tool - replace WithArray with WithObject
	runTool := mcp.NewTool("go_run",
		mcp.WithDescription("Compile and run Go code"),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code")),
		// Replace WithArray with WithObject which should be more basic and available in v0.19.0
		mcp.WithObject("args", mcp.Description("Command-line arguments")),
		mcp.WithNumber("timeoutSecs", mcp.Description("Timeout in seconds"), mcp.DefaultNumber(30)),
	)
	s.AddTool(runTool, tools.ExecuteGoRunTool)
	
	// Register go_mod tool
	modTool := mcp.NewTool("go_mod",
		mcp.WithDescription("Manage Go module dependencies"),
		mcp.WithString("command", mcp.Required(), mcp.Description("Subcommand (init, tidy, etc.)")),
		mcp.WithString("modulePath", mcp.Description("Project directory")),
		mcp.WithString("code", mcp.Description("Optional Go code")),
	)
	s.AddTool(modTool, tools.ExecuteGoModTool)
	
	// Register go_fmt tool
	fmtTool := mcp.NewTool("go_fmt", 
		mcp.WithDescription("Format Go code"),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code")),
	)
	s.AddTool(fmtTool, tools.ExecuteGoFmtTool)
	
	// Register go_analyze tool
	analyzeTool := mcp.NewTool("go_analyze",
		mcp.WithDescription("Analyze Go code for issues"),
		mcp.WithString("code", mcp.Required(), mcp.Description("Go source code")),
		mcp.WithBoolean("vet", mcp.Description("Run go vet"), mcp.DefaultBool(true)),
	)
	s.AddTool(analyzeTool, tools.ExecuteGoAnalyzeTool)
	
	log.Printf("Registered %d tools", 6)
}