#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Wrapper for legacy Go test runner.

.DESCRIPTION
    This script is a wrapper around the legacy Go test runner script.
    It delegates all parameters to the legacy script for backward compatibility.

.PARAMETER Verbose
    Show detailed test information.

.PARAMETER Cover
    Run tests with coverage.

.PARAMETER Race
    Run tests with race detection.

.PARAMETER Parallel
    Run tests in parallel.

.PARAMETER Pattern
    Pattern to match test files.

.EXAMPLE
    # Run all Go tests
    .\run_go_tests.ps1

.EXAMPLE
    # Run Go tests with verbose output and coverage
    .\run_go_tests.ps1 -Verbose -Cover
#>

param(
    [switch]$Verbose,
    [switch]$Cover,
    [switch]$Race,
    [switch]$Parallel,
    [string]$Pattern
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Output information about delegation to legacy script
Write-Host "Delegating to legacy Go test runner..." -ForegroundColor Yellow
Write-Host "Path: $PSScriptRoot\legacy\runners\run_go_tests.ps1" -ForegroundColor Yellow

# Forward to legacy script
& "$PSScriptRoot\legacy\runners\run_go_tests.ps1" @PSBoundParameters
