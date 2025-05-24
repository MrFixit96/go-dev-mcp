# Release Notes

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

# v0.1.0-alpha - Modernized MCP Implementation

This release modernizes the Go Development MCP Server with significant improvements to MCP library usage, natural language processing capabilities, and code quality.

## Added
- Enhanced natural language processing capabilities
- Fuzzy tool name matching based on string similarity
- Improved parameter documentation with detailed examples
- Context-aware matching system for better LLM interactions
- GitHub release workflow and dependabot configuration

## Changed
- Updated mark3labs/mcp-go dependency from v0.19.0 to v0.26.0
- Enhanced server configuration with modern options
- Improved parameter handling using helper functions
- Implemented proper JSON marshaling for tool responses
- Implemented comprehensive error handling

## Fixed
- Tool implementation compatibility with latest MCP API
- Configuration and initialization issues
- Error reporting with better context

## Usage
Download the appropriate binary for your platform and follow the installation instructions in the README.