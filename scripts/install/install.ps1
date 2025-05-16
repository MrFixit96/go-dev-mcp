# PowerShell installation script for Go Development MCP Server

param (
    [switch]$UserInstall = $false,
    [string]$InstallPath = ""
)

$ErrorActionPreference = "Stop"

# Determine installation path
if ($InstallPath -eq "") {
    if ($UserInstall) {
        $InstallPath = Join-Path $env:LOCALAPPDATA "Go-Dev-MCP"
    } else {
        $InstallPath = Join-Path $env:ProgramFiles "Go-Dev-MCP"
    }
}

# Create installation directory if it doesn't exist
if (-not (Test-Path $InstallPath)) {
    Write-Host "Creating installation directory: $InstallPath"
    New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
}

# Define binary path and config directory
$BinaryPath = Join-Path $InstallPath "go-dev-mcp.exe"
$ConfigDir = if ($UserInstall) {
    Join-Path $env:APPDATA "go-dev-mcp"
} else {
    Join-Path $env:ProgramData "go-dev-mcp"
}

# Create config directory if it doesn't exist
if (-not (Test-Path $ConfigDir)) {
    Write-Host "Creating configuration directory: $ConfigDir"
    New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
}

# Create default config.json if it doesn't exist
$ConfigPath = Join-Path $ConfigDir "config.json"
if (-not (Test-Path $ConfigPath)) {
    Write-Host "Creating default configuration file"
    $DefaultConfig = @{
        version = "1.0.0"
        logLevel = "info"
        sandboxType = "process"
        resourceLimits = @{
            cpuLimit = 2
            memoryLimit = 512
            timeoutSecs = 30
        }
    } | ConvertTo-Json -Depth 4
    Set-Content -Path $ConfigPath -Value $DefaultConfig -Encoding UTF8
}

# Copy binary to installation path
$SourceBinary = Join-Path $PSScriptRoot ".." "go-dev-mcp.exe"
if (Test-Path $SourceBinary) {
    Write-Host "Installing binary to: $BinaryPath"
    Copy-Item -Path $SourceBinary -Destination $BinaryPath -Force
} else {
    Write-Error "Binary not found at: $SourceBinary"
    exit 1
}

# Add to PATH if it's not already there
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", if ($UserInstall) { "User" } else { "Machine" })
if (-not $CurrentPath.Contains($InstallPath)) {
    Write-Host "Adding installation directory to PATH"
    [Environment]::SetEnvironmentVariable(
        "Path", 
        $CurrentPath + [IO.Path]::PathSeparator + $InstallPath, 
        if ($UserInstall) { "User" } else { "Machine" }
    )
}

# Create Claude Desktop integration file
$ClaudeConfigDir = Join-Path $env:LOCALAPPDATA "Anthropic" "Claude"
if (-not (Test-Path $ClaudeConfigDir)) {
    Write-Host "Creating Claude Desktop configuration directory"
    New-Item -ItemType Directory -Path $ClaudeConfigDir -Force | Out-Null
}

$ClaudeConfigPath = Join-Path $ClaudeConfigDir "claude_desktop_config.json"
if (Test-Path $ClaudeConfigPath) {
    Write-Host "Updating Claude Desktop configuration"
    $ClaudeConfig = Get-Content -Path $ClaudeConfigPath | ConvertFrom-Json -AsHashtable
    if (-not $ClaudeConfig.ContainsKey("mcpServers")) {
        $ClaudeConfig["mcpServers"] = @{}
    }
    $ClaudeConfig["mcpServers"]["go-dev"] = @{
        "command" = $BinaryPath
        "args" = @()
        "env" = @{}
        "disabled" = $false
        "autoApprove" = @()
    }
    $ClaudeConfig | ConvertTo-Json -Depth 10 | Set-Content -Path $ClaudeConfigPath -Encoding UTF8
} else {
    Write-Host "Creating new Claude Desktop configuration"
    $ClaudeConfig = @{
        "mcpServers" = @{
            "go-dev" = @{
                "command" = $BinaryPath
                "args" = @()
                "env" = @{}
                "disabled" = $false
                "autoApprove" = @()
            }
        }
    } | ConvertTo-Json -Depth 10
    Set-Content -Path $ClaudeConfigPath -Value $ClaudeConfig -Encoding UTF8
}

Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "Go Development MCP Server is now installed at: $BinaryPath"
Write-Host "Configuration file is located at: $ConfigPath"
Write-Host "The server is configured to work with Claude Desktop"