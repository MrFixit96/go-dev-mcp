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

- **MCP v0.29.0 Compatibility**: Updated to use the latest Model Context Protocol v0.29.0
- **Project Path Support**: All tools now support working with existing Go project directories
- **Strategy Pattern**: Flexible execution strategies for code snippets vs. project directories
- **Enhanced Response Formatting**: Better structured responses with natural language metadata
- **Improved Error Handling**: More detailed and helpful error messages
- **End-to-End Testing**: Comprehensive behavioral testing scripts to verify functionality
- **Modern Testing Framework**: New Go-based testing framework with parallel test execution

## Testing

The server includes comprehensive testing capabilities to verify that it works correctly with real Go projects. Testing is provided through two frameworks:

1. **Go Testing Framework**: Modern, parallel test framework using Go's native testing facilities and testify
2. **PowerShell Testing**: Legacy end-to-end behavioral tests (for backward compatibility)

The tests verify all input modes (code-only, project path, and hybrid) and ensure that the execution strategies work as expected.

### Running the Tests

#### Go Tests (Recommended)

```powershell
# Run all Go tests
cd go-dev-mcp
.\scripts\testing\run_tests.ps1 -TestType go -UseGoTests -WithCoverage

# Run with race detection
.\scripts\testing\run_tests.ps1 -TestType go -UseGoTests -WithRaceDetection

# Run directly with Go
go test -v ./internal/tools/...
```

#### PowerShell Tests (Legacy)

```powershell
# Quick tests
cd go-dev-mcp\scripts\testing
.\basic\simple_test.ps1

# Comprehensive tests
cd go-dev-mcp\scripts\testing
.\core\all_tools_test.ps1 -Verbose

# Strategy-specific tests
cd go-dev-mcp\scripts\testing
.\strategies\hybrid_strategy_test.ps1 -Verbose
```

For detailed information about the testing framework, see the [Testing Documentation](scripts/testing/TESTING.md).

The testing scripts are organized into categories:

- **Basic tests**: Simple, quick-running tests for sanity checks
- **Core tests**: Comprehensive tests covering all tools and input modes
- **Strategy tests**: Tests focused on specific execution strategies like hybrid execution

See the [testing README](scripts/testing/README.md) for more details.

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Windows

#### Manual Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/MrFixit96/go-dev-mcp.git
   cd go-dev-mcp
   ```

2. Build the executable:

   ```bash
   go build -o go-dev-mcp.exe ./cmd/server
   ```

3. Move the executable to a location in your PATH or reference it directly in your Claude Desktop configuration.

#### Using WinGet (Coming Soon)

```powershell
winget install go-dev-mcp
```

### macOS

#### Using Homebrew (Coming Soon)

```bash
brew install go-dev-mcp
```

#### Manual Installation (macOS)

1. Clone the repository:

   ```bash
   git clone https://github.com/MrFixit96/go-dev-mcp.git
   cd go-dev-mcp
   ```

2. Build the executable:

   ```bash
   go build -o go-dev-mcp ./cmd/server
   ```

3. Move the executable to a location in your PATH:

   ```bash
   sudo mv go-dev-mcp /usr/local/bin/
   ```

### Linux

1. Clone the repository:

   ```bash
   git clone https://github.com/MrFixit96/go-dev-mcp.git
   cd go-dev-mcp
   ```

2. Build the executable:

   ```bash
   go build -o go-dev-mcp ./cmd/server
   ```

3. Move the executable to a location in your PATH:

   ```bash
   sudo mv go-dev-mcp /usr/local/bin/
   ```

## Claude Desktop Integration

To integrate with Claude Desktop, update your `claude_desktop_config.json` file:

### Windows Configuration

```json
{
  "mcpServers": {
    "go-dev": {
      "command": "C:\\path\\to\\go-dev-mcp.exe",
      "args": [],
      "env": {
        "GOCACHE": "%LOCALAPPDATA%\\go-build",
        "LOCALAPPDATA": "%LOCALAPPDATA%",
        "GOPATH": "%USERPROFILE%\\go",
        "GOROOT": "%GOROOT%",
        "PATH": "%PATH%",
        "DEBUG": "*"
      },
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

**Environment Variables Used:**

- `%LOCALAPPDATA%`: Resolves to `C:\Users\{username}\AppData\Local`
- `%USERPROFILE%`: Resolves to `C:\Users\{username}`
- `%GOROOT%`: Go installation directory (automatically set by Go installer)
- `%PATH%`: System PATH for Go binary access

**Alternative using Go Environment Variables:**

```json
{
  "mcpServers": {
    "go-dev": {
      "command": "C:\\path\\to\\go-dev-mcp.exe",
      "args": [],
      "env": {
        "DEBUG": "*"
      },
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

> **Note**: The alternative configuration relies on Go's default environment detection. Go automatically uses `%LOCALAPPDATA%\go-build` for GOCACHE and `%USERPROFILE%\go` for GOPATH when not explicitly set.

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

```go
// Use go_build to compile code
go_build(code: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}")

// Run tests with go_test
go_test(code: "package main", testCode: "package main\n\nimport \"testing\"\n\nfunc TestHello(t *testing.T) {...}")
```

### Working with Project Directories

All tools now support working with existing Go project directories through the new `project_path` parameter:

```go
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
