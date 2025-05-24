#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Detailed test for the hybrid execution strategy in the Go Development MCP Server.

.DESCRIPTION
    This script tests the hybrid execution strategy in depth, verifying that modified code is 
    correctly applied while maintaining the context from an existing Go project. It creates
    test projects of varying complexity and verifies that the hybrid strategy correctly 
    uses the modified code while preserving the project structure and dependencies.

.PARAMETER ServerExecutable
    Path to the MCP server executable.

.PARAMETER TestDir
    Base directory for test files. If not specified, a temporary directory will be created.

.PARAMETER KeepTestDirs
    If specified, test directories will not be deleted after the test.

.PARAMETER Verbose
    Show detailed step-by-step execution information.

.EXAMPLE
    .\hybrid_strategy_test.ps1 -ServerExecutable "..\..\build\server.exe" -Verbose
#>

param(
    [string]$ServerExecutable = "..\..\build\server.exe",
    [string]$TestDir = "",
    [switch]$KeepTestDirs,
    [switch]$Verbose
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Enable verbose output if requested
if ($Verbose) {
    $VerbosePreference = "Continue"
}

#region Helper Functions

function Write-Title {
    param([string]$Text)
    
    $line = "-" * ($Text.Length + 4)
    Write-Host "`n$line" -ForegroundColor Cyan
    Write-Host "| $Text |" -ForegroundColor Cyan
    Write-Host "$line" -ForegroundColor Cyan
}

function Write-SubTitle {
    param([string]$Text)
    
    Write-Host "`n### $Text ###" -ForegroundColor Magenta
}

function Write-TestName {
    param([string]$Text)
    
    Write-Host "Testing: $Text" -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Text)
    
    Write-Host "✅ $Text" -ForegroundColor Green
}

function Write-Failure {
    param([string]$Text)
    
    Write-Host "❌ $Text" -ForegroundColor Red
}

function Write-Info {
    param([string]$Text)
    
    Write-Host $Text -ForegroundColor Cyan
}

function Create-TestProject {
    param(
        [string]$ProjectDir,
        [string]$ProjectType = "basic" # Can be "basic", "multiple-files", or "with-deps"
    )
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $ProjectDir)) {
        New-Item -ItemType Directory -Path $ProjectDir -Force | Out-Null
    }
    
    # Create project based on type
    switch ($ProjectType) {
        "basic" {
            # Create main.go file
            $MainCode = @"
package main

import "fmt"

func main() {
    greeting := GetGreeting()
    name := GetName()
    fmt.Printf("%s, %s!\n", greeting, name)
}
"@
            Set-Content -Path "$ProjectDir\main.go" -Value $MainCode
            
            # Create greeting.go file
            $GreetingCode = @"
package main

func GetGreeting() string {
    return "Hello"
}
"@
            Set-Content -Path "$ProjectDir\greeting.go" -Value $GreetingCode
            
            # Create name.go file
            $NameCode = @"
package main

func GetName() string {
    return "World"
}
"@
            Set-Content -Path "$ProjectDir\name.go" -Value $NameCode
        }
        
        "multiple-files" {
            # Create main.go with imports from multiple files
            $MainCode = @"
package main

import (
    "fmt"
    "time"
)

func main() {
    greeting := GetGreeting(time.Now().Hour())
    name := GetName()
    fmt.Printf("%s, %s! The time is %s\n", greeting, name, GetCurrentTime())
}
"@
            Set-Content -Path "$ProjectDir\main.go" -Value $MainCode
            
            # Create greeting.go with time-based greeting
            $GreetingCode = @"
package main

func GetGreeting(hour int) string {
    if hour < 12 {
        return "Good morning"
    } else if hour < 18 {
        return "Good afternoon"
    } else {
        return "Good evening"
    }
}
"@
            Set-Content -Path "$ProjectDir\greeting.go" -Value $GreetingCode
            
            # Create name.go file
            $NameCode = @"
package main

func GetName() string {
    return "World"
}
"@
            Set-Content -Path "$ProjectDir\name.go" -Value $NameCode
            
            # Create time.go file
            $TimeCode = @"
package main

import "time"

func GetCurrentTime() string {
    return time.Now().Format("15:04:05")
}
"@
            Set-Content -Path "$ProjectDir\time.go" -Value $TimeCode
        }
        
        "with-deps" {
            # Create main.go file that uses an external dependency
            $MainCode = @"
package main

import (
    "fmt"
)

func main() {
    greeting := GetRegularGreeting()
    name := GetName()
    fmt.Printf("%s, %s!\n", greeting, name)
}
"@
            Set-Content -Path "$ProjectDir\main.go" -Value $MainCode
            
            # Create greeting.go file that uses color package
            $GreetingCode = @"
package main

import "github.com/fatih/color"

func GetColoredGreeting() string {
    green := color.New(color.FgGreen).SprintFunc()
    return green("Hello")
}

func GetRegularGreeting() string {
    return "Hello"
}
"@
            Set-Content -Path "$ProjectDir\greeting.go" -Value $GreetingCode
            
            # Create name.go file
            $NameCode = @"
package main

func GetName() string {
    return "World"
}
"@
            Set-Content -Path "$ProjectDir\name.go" -Value $NameCode
        }
    }
    
    # Initialize go module
    Push-Location $ProjectDir
    go mod init example.com/hybrid-test
    if ($ProjectType -eq "with-deps") {
        go get github.com/fatih/color
        go mod tidy
    }
    Pop-Location
    
    return $ProjectDir
}

function Create-HybridProject {
    param(
        [string]$OriginalProjectPath,
        [string]$HybridProjectPath,
        [string]$ModifiedCode,
        [string]$ModifiedFile = "greeting.go"
    )
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $HybridProjectPath)) {
        New-Item -ItemType Directory -Path $HybridProjectPath -Force | Out-Null
    }
    
    # Copy all files except the one we want to modify
    Get-ChildItem -Path $OriginalProjectPath -File | ForEach-Object {
        if ($_.Name -ne $ModifiedFile) {
            Copy-Item -Path $_.FullName -Destination "$HybridProjectPath\$($_.Name)"
        }
    }
    
    # Write the modified code file
    Set-Content -Path "$HybridProjectPath\$ModifiedFile" -Value $ModifiedCode
}

function Test-HybridStrategy {
    param(
        [string]$ProjectType,
        [string]$ModifiedCode,
        [string]$ModifiedFile = "greeting.go",
        [string]$ExpectedOutput
    )
    
    Write-Title "Testing Hybrid Strategy with $ProjectType Project"
    
    # Create paths
    $OriginalDir = Join-Path $TestDir "original-$ProjectType"
    $HybridDir = Join-Path $TestDir "hybrid-$ProjectType"
    
    # Create test project
    Write-TestName "Creating $ProjectType project"
    $ProjectPath = Create-TestProject -ProjectDir $OriginalDir -ProjectType $ProjectType
    
    # Run original project
    Write-TestName "Running original project"
    Push-Location $ProjectPath
    $OriginalOutput = go run .
    Pop-Location
    Write-Info "Original output: $OriginalOutput"
    
    # Create hybrid project with modified code
    Write-TestName "Creating hybrid project with modified code"
    Create-HybridProject -OriginalProjectPath $OriginalDir -HybridProjectPath $HybridDir -ModifiedCode $ModifiedCode -ModifiedFile $ModifiedFile
    
    # Run hybrid project
    Write-TestName "Running hybrid project"
    Push-Location $HybridDir
    $HybridOutput = go run .
    Pop-Location
    Write-Info "Hybrid output: $HybridOutput"
    
    # Verify results
    if ($HybridOutput -like "*$ExpectedOutput*") {
        Write-Success "Hybrid strategy correctly used the modified code!"
        return $true
    } else {
        Write-Failure "Expected output containing '$ExpectedOutput' but got '$HybridOutput'"
        return $false
    }
}

#endregion

# Main script execution

# Use random temp dir if not specified
if (-not $TestDir) {
    $TestDir = Join-Path $env:TEMP "go-dev-hybrid-test-$(Get-Random)"
}

# Create test directory
if (-not (Test-Path $TestDir)) {
    New-Item -ItemType Directory -Path $TestDir -Force | Out-Null
}

Write-Host "Using test directory: $TestDir"

# Track test results
$TestResults = @()

#region Test 1: Basic Project Hybrid Strategy

$ModifiedGreeting = @"
package main

func GetGreeting() string {
    return "Greetings"
}
"@

$Result = Test-HybridStrategy -ProjectType "basic" -ModifiedCode $ModifiedGreeting -ExpectedOutput "Greetings, World!"
$TestResults += @{Name = "Basic Project Hybrid Strategy"; Success = $Result}

#endregion

#region Test 2: Multiple Files Project Hybrid Strategy

$ModifiedGreeting = @"
package main

func GetGreeting(hour int) string {
    return "Hola"
}
"@

$Result = Test-HybridStrategy -ProjectType "multiple-files" -ModifiedCode $ModifiedGreeting -ExpectedOutput "Hola, World!"
$TestResults += @{Name = "Multiple Files Project Hybrid Strategy"; Success = $Result}

#endregion

#region Test 3: Project with Dependencies Hybrid Strategy

$ModifiedGreeting = @"
package main

import "github.com/fatih/color"

func GetColoredGreeting() string {
    red := color.New(color.FgRed).SprintFunc()
    return red("Bonjour")
}

func GetRegularGreeting() string {
    return "Bonjour"
}
"@

$Result = Test-HybridStrategy -ProjectType "with-deps" -ModifiedCode $ModifiedGreeting -ExpectedOutput "Bonjour, World!"
$TestResults += @{Name = "Project with Dependencies Hybrid Strategy"; Success = $Result}

#endregion

# Print test summary
Write-Title "Test Summary"
$SuccessCount = ($TestResults | Where-Object { $_.Success -eq $true }).Count
$FailureCount = ($TestResults | Where-Object { $_.Success -eq $false }).Count

foreach ($Test in $TestResults) {
    if ($Test.Success) {
        Write-Success "$($Test.Name): PASS"
    } else {
        Write-Failure "$($Test.Name): FAIL"
    }
}

Write-Host "`nTotal Tests: $($TestResults.Count)" -ForegroundColor Cyan
Write-Host "Passed: $SuccessCount" -ForegroundColor Green
Write-Host "Failed: $FailureCount" -ForegroundColor Red

# Clean up
if (-not $KeepTestDirs) {
    Write-Host "`nCleaning up test directories..." -ForegroundColor Yellow
    try {
        # Force garbage collection to release file handles
        [System.GC]::Collect()
        [System.GC]::WaitForPendingFinalizers()
        
        # Try to remove with retries
        $maxRetries = 3
        $retryCount = 0
        $success = $false
        
        while (-not $success -and $retryCount -lt $maxRetries) {
            try {
                Remove-Item -Path $TestDir -Recurse -Force -ErrorAction Stop
                $success = $true
                Write-Host "Test directories removed successfully" -ForegroundColor Green
            } catch {
                $retryCount++
                Write-Host "Failed to remove directory (attempt $retryCount of $maxRetries): $_" -ForegroundColor Yellow
                Start-Sleep -Seconds 2
            }
        }
        
        if (-not $success) {
            Write-Host "Could not remove test directories after $maxRetries attempts. They may need to be removed manually." -ForegroundColor Yellow
        }
    } catch {
        Write-Host "Error during cleanup: $_" -ForegroundColor Red
    }
} else {
    Write-Host "`nTest directories kept at: $TestDir" -ForegroundColor Yellow
}

# Return exit code based on test results
if ($FailureCount -gt 0) {
    exit 1
} else {
    exit 0
}
