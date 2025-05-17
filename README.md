# Go Development MCP Server

The Go Development MCP Server is a comprehensive solution for integrating Go development workflows with AI assistants like Claude Desktop or other MCP-compatible tools. It enables AI assistants to compile, test, run, and analyze Go code directly through the Model Context Protocol (MCP).

## Features

- **Go Build**: Compile Go code and receive detailed feedback
- **Go Test**: Run tests on Go code with support for coverage analysis
- **Go Run**: Compile and execute Go programs with command-line arguments
- **Go Mod**: Manage Go module dependencies (init, tidy, download, etc.)
- **Go Format**: Format Go code according to standard conventions
- **Go Analyze**: Analyze Go code for issues using static analysis tools

### New in This Release

- **Project Path Support**: All tools now support working with existing Go project directories
- **Strategy Pattern**: Flexible execution strategies for code snippets vs. project directories
- **Enhanced Response Formatting**: Better structured responses with natural language metadata
- **Improved Error Handling**: More detailed and helpful error messages
- **End-to-End Testing**: Comprehensive behavioral testing scripts to verify functionality

## Testing

The server includes end-to-end behavioral testing capabilities to verify that it works correctly with real Go projects. The tests verify all input modes (code-only, project path, and hybrid) and ensure that the execution strategies work as expected.

### Running the Tests

```powershell
# Start the server in one terminal
cd go-dev-mcp
.\build\server.exe

# Run the tests in another terminal
cd go-dev-mcp\scripts\testing
.\e2e_test.ps1

# Test the hybrid execution strategy specifically
.\hybrid_strategy_test.ps1
```

See the [testing README](scripts/testing/README.md) for more details.

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Windows

#### Manual Installation

1. Clone the repository:
   ```
   git clone https://github.com/MrFixit96/go-dev-mcp.git
   cd go-dev-mcp
   ```

2. Build the executable:
   ```
   go build -o go-dev-mcp.exe ./cmd/server
   ```

3. Move the executable to a location in your PATH or reference it directly in your Claude Desktop configuration.

#### Using WinGet (Coming Soon)

```
winget install go-dev-mcp
```

### macOS

#### Using Homebrew (Coming Soon)

```
brew install go-dev-mcp
```

#### Manual Installation

1. Clone the repository:
   ```
   git clone https://github.com/MrFixit96/go-dev-mcp.git
   cd go-dev-mcp
   ```

2. Build the executable:
   ```
   go build -o go-dev-mcp ./cmd/server
   ```

3. Move the executable to a location in your PATH:
   ```
   sudo mv go-dev-mcp /usr/local/bin/
   ```

### Linux

1. Clone the repository:
   ```
   git clone https://github.com/MrFixit96/go-dev-mcp.git
   cd go-dev-mcp
   ```

2. Build the executable:
   ```
   go build -o go-dev-mcp ./cmd/server
   ```

3. Move the executable to a location in your PATH:
   ```
   sudo mv go-dev-mcp /usr/local/bin/
   ```

## Claude Desktop Integration

To integrate with Claude Desktop, update your `claude_desktop_config.json` file:

### Windows

```json
{
  "mcpServers": {
    "go-dev": {
      "command": "C:\\path\\to\\go-dev-mcp.exe",
      "args": [],
      "env": {},
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

### macOS and Linux

```json
{
  "mcpServers": {
    "go-dev": {
      "command": "/path/to/go-dev-mcp",
      "args": [],
      "env": {},
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

## Usage

### Working with Code Snippets

All tools accept Go code directly through the `code` parameter:

```
// Use go_build to compile code
go_build(code: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}")

// Run tests with go_test
go_test(code: "package main", testCode: "package main\n\nimport \"testing\"\n\nfunc TestHello(t *testing.T) {...}")
```

### Working with Project Directories

All tools now support working with existing Go project directories through the new `project_path` parameter:

```
// Compile a project
go_build(project_path: "/path/to/your/go/project")

// Run tests in a project
go_test(project_path: "/path/to/your/go/project", verbose: true, coverage: true)

// Format all files in a project
go_fmt(project_path: "/path/to/your/go/project")

// Analyze a project for issues
go_analyze(project_path: "/path/to/your/go/project", vet: true)
```

## Configuration

The server uses a configuration file located at:

- Windows: `%APPDATA%\go-dev-mcp\config.json`
- macOS: `~/Library/Application Support/go-dev-mcp/config.json`
- Linux: `~/.config/go-dev-mcp/config.json`

A default configuration file will be created on first run, which you can customize:

```json
{
  "version": "1.0.0",
  "logLevel": "info",
  "sandboxType": "process",
  "resourceLimits": {
    "cpuLimit": 2,
    "memoryLimit": 512,
    "timeoutSecs": 30
  }
}
```

## Security

The Go Development MCP Server runs commands in a sandboxed environment with:

- Process isolation
- Resource limits (CPU, memory, execution time)
- Temporary directory containment
- No network access by default

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Roadmap

- [ ] Add support for Go workspaces
- [ ] Implement Docker-based sandbox for stronger isolation
- [ ] Add debugging capabilities
- [ ] Support for Go race detector
- [ ] Improved error reporting with suggestions