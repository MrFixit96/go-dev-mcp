# Run Tests Script for Go Development MCP Server
# This script runs the modern Go-based execution strategy tests

param(
    [ValidateSet("direct", "hybrid", "both")]
    [string]$TestType = "both",
    [switch]$Verbose
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Project directory (adjusted for new location in legacy/runners)
$ProjectDir = "C:\Users\James\Documents\go-dev-mcp"
$TestingDir = Join-Path $ProjectDir "scripts\testing"

# Output information
if ($Verbose) {
    Write-Host "Running Go-based execution strategy tests..." -ForegroundColor Cyan
    Write-Host "Project directory: $ProjectDir" -ForegroundColor Cyan
    Write-Host "Testing directory: $TestingDir" -ForegroundColor Cyan
    Write-Host "Test type: $TestType" -ForegroundColor Cyan
}

# Build the test runner if needed
if (-not (Test-Path "$ProjectDir\testing_runner.exe")) {
    Write-Host "Building test runner..." -ForegroundColor Yellow
    Push-Location $ProjectDir
    try {
        # Build main.go using an absolute path to avoid any confusion
        $mainGoPath = Join-Path $TestingDir "main.go"
        Write-Host "Building from: $mainGoPath" -ForegroundColor Yellow
        go build -o testing_runner.exe $mainGoPath
        if (-not $?) {
            Write-Error "Failed to build test runner"
            exit 1
        }
    }
    finally {
        Pop-Location
    }
}

# Run the tests
Push-Location $ProjectDir
try {
    Write-Host "Executing tests..." -ForegroundColor Green
    .\testing_runner.exe -type $TestType
}
finally {
    Pop-Location
}

Write-Host "Tests completed" -ForegroundColor Green
