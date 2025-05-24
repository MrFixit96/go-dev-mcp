#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Wrapper for legacy strategy test runner.

.DESCRIPTION
    This script is a wrapper around the legacy strategy test runner script.
    It delegates all parameters to the legacy script for backward compatibility.

.PARAMETER TestType
    Type of tests to run. Valid values: direct, hybrid, both

.PARAMETER Verbose
    Show detailed test information.

.EXAMPLE
    # Run both types of strategy tests
    .\run_strategy_tests.ps1

.EXAMPLE
    # Run only direct strategy tests with verbose output
    .\run_strategy_tests.ps1 -TestType direct -Verbose
#>

param(
    [ValidateSet("direct", "hybrid", "both")]
    [string]$TestType = "both",
    [switch]$Verbose
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Output information about delegation to legacy script
Write-Host "Delegating to legacy strategy test runner..." -ForegroundColor Yellow
Write-Host "Path: $PSScriptRoot\legacy\runners\run_strategy_tests.ps1" -ForegroundColor Yellow

# Forward to legacy script
& "$PSScriptRoot\legacy\runners\run_strategy_tests.ps1" @PSBoundParameters
