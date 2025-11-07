# Release Notes

## v0.3.0 - MCP v0.43.0 Upgrade

This release updates the Go Development MCP Server to be compatible with MCP v0.43.0, bringing new features and improvements from the upstream library.

### Added

- Support for custom HTTP headers in client requests (mcp-go v0.43.0)
- Session management with SessionWithResourceTemplates (mcp-go v0.43.0)
- HTTP and Stdio client Roots feature (mcp-go v0.43.0)
- Title field in Implementation struct per MCP spec (mcp-go v0.43.0)
- Enhanced JSON schema support via invopop/jsonschema integration

### Changed

- **Breaking Change**: `TextContent` struct now requires explicit `Type` field
  - Updated from: `mcp.TextContent{Text: "..."}`
  - Updated to: `mcp.TextContent{Type: mcp.ContentTypeText, Text: "..."}`
- Improved content type handling with proper type constants
- Updated all mock test handlers to use content type constants
- Enhanced dependency tree with better JSON schema support

### Fixed

- Custom header handling in client requests (mcp-go v0.43.0)
- Content type validation in tool results
- Test infrastructure compatibility with new mcp-go API

### Migration Guide

If you have custom code using `TextContent`, update it as follows:

```go
// Before (v0.29.0)
result := &mcp.CallToolResult{
    Content: []mcp.Content{
        mcp.TextContent{Text: "response"},
    },
}

// After (v0.43.0)
result := &mcp.CallToolResult{
    Content: []mcp.Content{
        mcp.TextContent{
            Type: mcp.ContentTypeText,
            Text: "response",
        },
    },
}
```

### Security

- All module checksums verified with `go mod verify`
- New dependencies audited:
  - `github.com/invopop/jsonschema@v0.13.0` - JSON Schema generation
  - `github.com/mailru/easyjson@v0.7.7` - Fast JSON serialization
  - `github.com/wk8/go-ordered-map/v2@v2.1.8` - Ordered map implementation
  - `github.com/buger/jsonparser@v1.1.1` - JSON parser
  - `github.com/bahlo/generic-list-go@v0.2.0` - Generic list utilities

---

## v0.2.0 - MCP v0.29.0 Migration

This release updates the Go Development MCP Server to be compatible with MCP v0.29.0, featuring improved tool result patterns and enhanced response handling.

### Added

- Compatibility with Model Context Protocol v0.29.0
- Updated tool result format using structured Content objects
- Enhanced error handling with proper MCP v0.29.0 patterns
- Improved build validation and compilation verification

### Changed

- Updated from `mcp.NewToolResponse(mcp.NewTextContent(...))` to `mcp.NewToolResultText(...)`
- Updated from `mcp.NewToolResponse(mcp.NewErrorContent(...))` to `mcp.NewToolResultError(...)`
- Refined error handling in fmt.go, run.go, and util.go
- Improved consistency across all tool implementations

### Fixed

- Tool result format compatibility with latest MCP API v0.29.0
- Build compilation issues with updated API patterns
- Response formatting consistency across all tools

---

## v0.1.0-alpha - Modernized MCP Implementation

This release modernizes the Go Development MCP Server with significant improvements to MCP library usage, natural language processing capabilities, and code quality.

### Features Added

- Enhanced natural language processing capabilities
- Fuzzy tool name matching based on string similarity
- Improved parameter documentation with detailed examples
- Context-aware matching system for better LLM interactions
- GitHub release workflow and dependabot configuration

### Changes Made

- Updated mark3labs/mcp-go dependency from v0.19.0 to v0.26.0
- Enhanced server configuration with modern options
- Improved parameter handling using helper functions
- Implemented proper JSON marshaling for tool responses
- Implemented comprehensive error handling

### Issues Fixed

- Tool implementation compatibility with latest MCP API
- Configuration and initialization issues
- Error reporting with better context

### Usage

Download the appropriate binary for your platform and follow the installation instructions in the README.
