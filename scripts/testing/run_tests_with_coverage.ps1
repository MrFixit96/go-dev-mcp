#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Wrapper for legacy test runner with coverage.

.DESCRIPTION
    This script is a wrapper around the legacy test runner with coverage script.
    It delegates all parameters to the legacy script for backward compatibility.

.PARAMETER OutputDir
    Directory where coverage output files will be stored.

.PARAMETER Verbose
    Show detailed test information.

.EXAMPLE
    # Run tests with coverage
    .\run_tests_with_coverage.ps1

.EXAMPLE
    # Run tests with coverage and specify output directory
    .\run_tests_with_coverage.ps1 -OutputDir ".\my-coverage"
#>

param(
    [string]$OutputDir = ".\coverage",
    [switch]$Verbose
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Output information about delegation to legacy script
Write-Host "Delegating to legacy coverage test runner..." -ForegroundColor Yellow
Write-Host "Path: $PSScriptRoot\legacy\runners\run_tests_with_coverage.ps1" -ForegroundColor Yellow

# Forward to legacy script
& "$PSScriptRoot\legacy\runners\run_tests_with_coverage.ps1" @PSBoundParameters
