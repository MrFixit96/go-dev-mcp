# Go Test Runner for MCP Server
# This script runs the Go tests using the new testing framework

param(
    [switch]$Verbose,
    [switch]$Race,
    [switch]$Cover,
    [switch]$Parallel,
    [int]$Timeout = 300,
    [string]$Pattern = ""
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Main project directory (adjusted for new location in legacy/runners)
$ProjectDir = "C:\Users\James\Documents\go-dev-mcp"
$TestingDir = Join-Path $ProjectDir "scripts\testing"

Write-Host "Running Go tests with modern testing framework..." -ForegroundColor Cyan
Write-Host "Project directory: $ProjectDir" -ForegroundColor Cyan
Write-Host "Testing directory: $TestingDir" -ForegroundColor Cyan

# Use go run to execute the main.go test runner directly (adjusting path for new location)
$mainGoPath = Join-Path $TestingDir "main.go"
$testArgs = @("run", $mainGoPath)

# Add type parameter if specified
if ($Pattern) {
    $testArgs += "-type=$Pattern"
}

# If verbose is enabled, we'll run with full output
if ($Verbose) {
    Write-Host "Running with verbose output..." -ForegroundColor Cyan
}

# Note: The following flags are not directly applicable when using go run,
# but we keep them for informational purposes
if ($Race) {
    Write-Host "Race detection enabled (informational only)" -ForegroundColor Yellow
}

if ($Cover) {
    Write-Host "Coverage tracking enabled (informational only)" -ForegroundColor Yellow
}

# Set parallelism environment variable if specified
if ($Parallel) {
    $env:MCP_TEST_PARALLEL = "4" # Use 4 parallel tests by default
} else {
    $env:MCP_TEST_PARALLEL = "1" # Run tests sequentially
}

# Display the test command
$testCommand = "go $($testArgs -join ' ')"
Write-Host "Running: $testCommand" -ForegroundColor Yellow

# Run the tests
& go $testArgs

# Save the exit code to return
$testExitCode = $LASTEXITCODE

# Display coverage report if requested
if ($Cover -and (Test-Path "coverage.out")) {
    Write-Host "`nCoverage Report:" -ForegroundColor Cyan
    & go tool cover -func=coverage.out
    
    # Generate HTML coverage report
    & go tool cover -html=coverage.out -o coverage.html
    Write-Host "HTML coverage report saved to coverage.html" -ForegroundColor Green
}

# Return the test exit code
exit $testExitCode
