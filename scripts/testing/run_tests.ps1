#!/usr/bin/env pwsh
# This script uses Write-Host intentionally for colored console output in interactive usage
[Diagnostics.CodeAnalysis.SuppressMessageAttribute('PSAvoidUsingWriteHost', '')]
param()
<#
.SYNOPSIS
    Master test runner for Go Development MCP Server tests.

.DESCRIPTION
    This script can run all tests or specific categories of tests for the
    Go Development MCP Server. It provides a convenient interface for running
    tests from a single command and aggregates the results.

.PARAMETER TestType
    Type of tests to run. Valid values: all, basic, core, strategies

.PARAMETER VerboseOutput
    Show detailed test information.

.PARAMETER KeepTestDirs
    If specified, test directories will not be deleted after the test.

.PARAMETER ServerExecutable
    Path to the MCP server executable.

.EXAMPLE
    # Run all tests
    .\run_tests.ps1 -TestType all -VerboseOutput

.EXAMPLE
    # Run only the basic tests
    .\run_tests.ps1 -TestType basic
#>

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("all", "basic", "core", "strategies", "go")]
    [string]$TestType,

    [switch]$VerboseOutput,
    [switch]$KeepTestDirs,
    [string]$ServerExecutable = "..\..\build\server.exe",
    [switch]$UseGoTests,
    [switch]$WithCoverage,
    [switch]$WithRaceDetection
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Import utility functions
. "$PSScriptRoot\legacy\utils\test_utils.ps1"

# Track test results
$script:TestResults = @()
$ScriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path

# Check if we should use the new Go testing framework
$UseGoTests = $false  # Default to PowerShell tests for backward compatibility
if ($env:MCP_USE_GO_TESTS -eq "true" -or $PSBoundParameters.ContainsKey("UseGoTests")) {
    $UseGoTests = $true
}

function Invoke-TestScript {
    param(
        [string]$ScriptPath,
        [string]$Description,
        [hashtable]$Parameters
    )

    Write-Title "Running $Description"
    Write-Host "Script: $ScriptPath" -ForegroundColor Cyan

    $timer = Start-Timer
    try {
        # Reset last exit code before running test
        $global:LASTEXITCODE = $null

        # Run the script without capturing unused output
        & $ScriptPath @Parameters

        # Check result (null is considered success)
        $success = $LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq $null

        if ($success) {
            Write-Success "$Description test completed successfully"
        }
        else {
            Write-Failure "$Description test failed with exit code $LASTEXITCODE"
        }

        $script:TestResults += @{
            Name = $Description
            Success = $success
            ElapsedTime = $timer.Elapsed
        }
    }
    catch {
        Write-Failure "$Description test failed with error: $_"
        $script:TestResults += @{
            Name = $Description
            Success = $false
            ElapsedTime = $timer.Elapsed
            Error = $_
        }
    }
    Write-Host ""
}

function Invoke-GoTest {
    param(
        [string]$Description,
        [switch]$WithCoverage,
        [switch]$WithRaceDetection,
        [switch]$VerboseOutput
    )

    Write-Title "Running $Description"

    $timer = Start-Timer
    try {
        # Reset last exit code before running test
        $global:LASTEXITCODE = $null

        # Build arguments for the Go test runner
        $goTestParams = @{}
        if ($VerboseOutput) {
            $goTestParams.Add("Verbose", $true)
        }
        if ($WithCoverage) {
            $goTestParams.Add("Cover", $true)
        }
        if ($WithRaceDetection) {
            $goTestParams.Add("Race", $true)
        }
        if ($env:MCP_TEST_PARALLEL -ne "0") {
            $goTestParams.Add("Parallel", $true)
        }
          # Run the Go test runner
        & "$ScriptPath\legacy\runners\run_go_tests.ps1" @goTestParams

        # Check result
        $success = $LASTEXITCODE -eq 0

        if ($success) {
            Write-Success "$Description completed successfully"
        }
        else {
            Write-Failure "$Description failed with exit code $LASTEXITCODE"
        }

        $script:TestResults += @{
            Name = $Description
            Success = $success
            ElapsedTime = $timer.Elapsed
        }
    }
    catch {
        Write-Failure "$Description failed with error: $_"
        $script:TestResults += @{
            Name = $Description
            Success = $false
            ElapsedTime = $timer.Elapsed
            Error = $_
        }
    }

    Write-Host ""
}

# Common parameters for test scripts
$commonParams = @{}
if ($VerboseOutput) {
    $commonParams.Add("Verbose", $true)
}
if ($KeepTestDirs) {
    $commonParams.Add("KeepTestDirs", $true)
}
if ($ServerExecutable) {
    $commonParams.Add("ServerExecutable", $ServerExecutable)
}

# Run tests based on the test type
switch ($TestType) {
    "all" {
        # Run legacy basic tests
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\basic\simple_test.ps1" -Description "Legacy simple test" -Parameters $commonParams

        # Run legacy core tests
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\core\all_tools_test.ps1" -Description "Legacy all tools test" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\core\e2e_test.ps1" -Description "Legacy end-to-end test" -Parameters $commonParams

        # Run legacy direct strategy tests
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\direct\direct_strategy_test.ps1" -Description "Legacy direct strategy test" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\direct\direct_strat_verify.ps1" -Description "Legacy direct strategy verification" -Parameters $commonParams        # Run legacy hybrid strategy tests
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\hybrid\hybrid_strategy_test.ps1" -Description "Legacy hybrid strategy test" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\hybrid\hybrid_strat_verify.ps1" -Description "Legacy hybrid strategy verification" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\hybrid\hybrid_cli_test.ps1" -Description "Legacy hybrid CLI test" -Parameters $commonParams
          # Run Go tests if enabled
        if ($UseGoTests) {
            Invoke-GoTest -Description "Go Unit Tests" -WithCoverage:$WithCoverage -WithRaceDetection:$WithRaceDetection -VerboseOutput:$VerboseOutput
        }
    }
    "basic" {
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\basic\simple_test.ps1" -Description "Legacy simple test" -Parameters $commonParams
    }
    "core" {
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\core\all_tools_test.ps1" -Description "Legacy all tools test" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\core\e2e_test.ps1" -Description "Legacy end-to-end test" -Parameters $commonParams
    }
    "strategies" {
        # Run direct strategy tests
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\direct\direct_strategy_test.ps1" -Description "Legacy direct strategy test" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\direct\direct_strat_verify.ps1" -Description "Legacy direct strategy verification" -Parameters $commonParams        # Run hybrid strategy tests
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\hybrid\hybrid_strategy_test.ps1" -Description "Legacy hybrid strategy test" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\hybrid\hybrid_strat_verify.ps1" -Description "Legacy hybrid strategy verification" -Parameters $commonParams
        Invoke-TestScript -ScriptPath "$ScriptPath\legacy\strategies\hybrid\hybrid_cli_test.ps1" -Description "Legacy hybrid CLI test" -Parameters $commonParams
    }    "go" {
        # Run only the Go tests
        Invoke-GoTest -Description "Go Unit Tests" -WithCoverage:$WithCoverage -WithRaceDetection:$WithRaceDetection -VerboseOutput:$VerboseOutput
    }
}

# Print test summary
Write-Title "Test Summary"
$totalTests = $script:TestResults.Count
$passedTests = @($script:TestResults | Where-Object { $_.Success -eq $true }).Count
$failedTests = @($script:TestResults | Where-Object { $_.Success -eq $false }).Count

Write-Host "Test Results:" -ForegroundColor Cyan
foreach ($test in $script:TestResults) {
    $status = if ($test.Success) { "✅ PASS" } else { "❌ FAIL" }
    $time = $test.ElapsedTime.TotalSeconds.ToString("0.00")
    Write-Host "$status - $($test.Name) (${time}s)" -ForegroundColor $(if ($test.Success) { "Green" } else { "Red" })

    if (-not $test.Success -and $test.Error) {
        Write-Host "  Error: $($test.Error)" -ForegroundColor Red
    }
}

Write-Host "`nTotal Tests: $totalTests" -ForegroundColor Cyan
Write-Host "Passed: $passedTests" -ForegroundColor Green
Write-Host "Failed: $failedTests" -ForegroundColor Red

# Return exit code based on test success
if ($failedTests -gt 0) {
    exit 1
} else {
    exit 0
}
