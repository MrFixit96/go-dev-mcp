# Winget MCP Server Packaging Prompt Template

Use this template when requesting Claude to create Windows package deployment files for an MCP server. This enables distribution through the Winget package manager with support for both system-wide and user-level installations.

## Template

```
# Creating a Winget Package for {MCP-SERVER-NAME}

I need help creating Windows package deployment files for my {MCP-SERVER-NAME} MCP server, enabling distribution through the Winget package manager. The package should support both system-wide and user-level installations.

## Context
- The MCP server is named "{MCP-SERVER-NAME}" with version "{VERSION-NUMBER}"
- Source code is located at {SOURCE-REPOSITORY-URL}
- The server executable is a Go binary that needs to be available to Claude Desktop and other AI tools

## Requirements
1. Create a Winget manifest set according to current schema standards
2. Support both system-wide (Program Files) and per-user installation options
3. Include proper configuration file templates
4. Set up appropriate environment variables for Claude Desktop to detect the MCP server
5. Ensure the package is updatable through Winget's update mechanism

## Specific Questions
1. What's the best approach for a dual-installation mode package?
2. How should the manifest files be structured for maximum compatibility?
3. What placement strategy works best for MCP server executables to be discovered by Claude?
4. What PowerShell scripts might be needed for custom installation logic?
5. How can I automate manifest updates for new version releases?

Please provide step-by-step guidance on creating this package, including all necessary manifest files, installation scripts, and integration instructions. Include examples where applicable.
```

## Usage Instructions

1. Copy the template above and replace all placeholders in `{CURLY-BRACES}` with your specific information:
   - `{MCP-SERVER-NAME}`: The name of your MCP server (e.g., "Go-Dev-MCP")
   - `{VERSION-NUMBER}`: Current version of your MCP server (e.g., "1.0.0")
   - `{SOURCE-REPOSITORY-URL}`: GitHub or other repository URL

2. Submit the completed template to Claude 3.7 Sonnet or higher

3. Claude will respond with detailed guidance specific to your MCP server, including:
   - Winget manifest YAML files
   - PowerShell installation scripts
   - Configuration templates
   - Integration instructions for Claude Desktop

## Benefits of Using Winget

This approach leverages Winget (Windows Package Manager) to:
- Provide a modern installation experience
- Support silent installation for enterprise environments
- Allow for automatic updates
- Enable easy distribution through Microsoft's package ecosystem
- Support both per-user and system-wide installation modes
- Ensure proper integration with Claude Desktop and other AI tools

## Additional Resources

- [Windows Package Manager Documentation](https://learn.microsoft.com/en-us/windows/package-manager/)
- [Winget Manifest Creation Guide](https://learn.microsoft.com/en-us/windows/package-manager/package/manifest)
- [WingetCreate Tool](https://github.com/microsoft/winget-create)
