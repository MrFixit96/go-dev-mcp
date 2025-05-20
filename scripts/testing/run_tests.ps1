#!/usr/bin/env pwsh
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
    [ValidateSet("all", "basic", "core", "strategies")]
    [string]$TestType,
    
    [switch]$VerboseOutput,
    [switch]$KeepTestDirs,
    [string]$ServerExecutable = "..\..\build\server.exe"
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Import utility functions
. "$PSScriptRoot\utils\test_utils.ps1"

# Track test results
$script:TestResults = @()
$ScriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path

function Run-TestScript {
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
        
        # Run the script
        $output = & $ScriptPath @Parameters
        
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
        # Run basic tests
        Run-TestScript -ScriptPath "$ScriptPath\basic\simple_test.ps1" -Description "Simple test" -Parameters $commonParams
        
        # Run core tests
        Run-TestScript -ScriptPath "$ScriptPath\core\all_tools_test.ps1" -Description "All tools test" -Parameters $commonParams
        Run-TestScript -ScriptPath "$ScriptPath\core\e2e_test.ps1" -Description "End-to-end test" -Parameters $commonParams
        
        # Run strategy tests
        Run-TestScript -ScriptPath "$ScriptPath\strategies\hybrid_strategy_test.ps1" -Description "Hybrid strategy test" -Parameters $commonParams
        Run-TestScript -ScriptPath "$ScriptPath\strategies\hybrid_strat_verify.ps1" -Description "Hybrid strategy verification" -Parameters $commonParams
    }
    "basic" {
        Run-TestScript -ScriptPath "$ScriptPath\basic\simple_test.ps1" -Description "Simple test" -Parameters $commonParams
    }
    "core" {
        Run-TestScript -ScriptPath "$ScriptPath\core\all_tools_test.ps1" -Description "All tools test" -Parameters $commonParams
        Run-TestScript -ScriptPath "$ScriptPath\core\e2e_test.ps1" -Description "End-to-end test" -Parameters $commonParams
    }
    "strategies" {
        Run-TestScript -ScriptPath "$ScriptPath\strategies\hybrid_strategy_test.ps1" -Description "Hybrid strategy test" -Parameters $commonParams
        Run-TestScript -ScriptPath "$ScriptPath\strategies\hybrid_strat_verify.ps1" -Description "Hybrid strategy verification" -Parameters $commonParams
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
