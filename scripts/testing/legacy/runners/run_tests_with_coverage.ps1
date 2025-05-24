# Previously located at: c:\Users\James\Documents\go-dev-mcp\scripts\testing\run_tests_with_coverage.ps1
#!/usr/bin/env pwsh
#
# Run tests with coverage reporting for Go Development MCP Server

param(
    [string]$OutputDir = "$PSScriptRoot\..\..\..\..\..\coverage",
    [switch]$ShowHtml = $false,
    [switch]$Short = $false,
    [string[]]$Packages = @("./internal/..."),
    [switch]$Verbose = $false,
    [int]$Timeout = 300
)

# Create output directory if it doesn't exist
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
    Write-Host "Created output directory: $OutputDir"
}

# Resolve to absolute path
$OutputDir = Resolve-Path $OutputDir

# Build command arguments
$testArgs = @("test")

if ($Short) {
    $testArgs += "-short"
}

if ($Verbose) {
    $testArgs += "-v"
}

$testArgs += "-coverprofile=$OutputDir\coverage.out"
$testArgs += "-covermode=atomic"
$testArgs += "-timeout=$($Timeout)s"
$testArgs += $Packages

# Run tests with coverage
Write-Host "Running tests with coverage..."
Write-Host "go $($testArgs -join ' ')"

$start = Get-Date
& go $testArgs
$testExitCode = $LASTEXITCODE
$duration = (Get-Date) - $start

Write-Host "Tests completed in $($duration.TotalSeconds) seconds with exit code: $testExitCode"

# Process coverage results if tests passed and coverage file exists
if (($testExitCode -eq 0) -and (Test-Path "$OutputDir\coverage.out")) {
    # Generate HTML coverage report
    Write-Host "Generating HTML coverage report..."
    & go tool cover -html="$OutputDir\coverage.out" -o="$OutputDir\coverage.html"
    
    # Generate function coverage report
    Write-Host "Generating function coverage report..."
    & go tool cover -func="$OutputDir\coverage.out" | Tee-Object -FilePath "$OutputDir\coverage_func.txt"
    
    # Extract total coverage
    $totalCoverage = Get-Content "$OutputDir\coverage_func.txt" | Select-String -Pattern "total:" | ForEach-Object { $_ -replace ".*total:.*?\s+(\d+\.\d+)%.*", '$1' }
    
    # Print summary
    Write-Host "`n-------------------------------------------------------"
    Write-Host "Coverage Summary"
    Write-Host "-------------------------------------------------------"
    Write-Host "Total Coverage: $totalCoverage%"
    Write-Host "Coverage reports saved to: $OutputDir"
    Write-Host " - HTML Report: $OutputDir\coverage.html"
    Write-Host " - Function Report: $OutputDir\coverage_func.txt"
    
    # Open HTML report if requested
    if ($ShowHtml) {
        Write-Host "Opening HTML coverage report..."
        Start-Process "$OutputDir\coverage.html"
    }
    
    # Create a badge-friendly JSON file
    $badgeData = @{
        schemaVersion = 1
        label = "coverage"
        message = "$totalCoverage%"
        color = "informational"
    }
    $badgeJson = $badgeData | ConvertTo-Json
    Set-Content -Path "$OutputDir\coverage-badge.json" -Value $badgeJson
    
    # Return success
    exit 0
} else {
    if ($testExitCode -ne 0) {
        Write-Host "Tests failed with exit code: $testExitCode" -ForegroundColor Red
    } else {
        Write-Host "Coverage file not found: $OutputDir\coverage.out" -ForegroundColor Red
    }
    
    # Return test exit code
    exit $testExitCode
}
