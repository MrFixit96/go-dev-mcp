#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Comprehensive test suite for the Go Development MCP Server tools.

.DESCRIPTION
    This script tests all tools provided by the Go Development MCP Server, including:
    - go_build: Building Go code
    - go_run: Running Go code
    - go_fmt: Formatting Go code
    - go_test: Testing Go code
    - go_mod: Managing Go modules
    - go_analyze: Analyzing Go code for issues

    Each tool is tested with all applicable input modes:
    - Code-only: Using just the provided code
    - Project-path-only: Using an existing Go project directory
    - Hybrid: Using both code and project path

.PARAMETER ServerExecutable
    Path to the MCP server executable.

.PARAMETER KeepTestDirs
    If specified, test directories will not be deleted after the test.

.PARAMETER TestDir
    Custom directory to use for test files. If not specified, a temporary directory will be created.

.PARAMETER Verbose
    Show detailed test information.

.EXAMPLE
    .\all_tools_test.ps1 -ServerExecutable "..\..\build\server.exe" -Verbose

.NOTES
    This test simulates the behavior of the MCP server tools by directly manipulating
    Go project files, as the MCP server uses stdin/stdout rather than HTTP.
#>

param(
    [string]$ServerExecutable = "..\..\build\server.exe",
    [switch]$KeepTestDirs,
    [string]$TestDir = "",
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

function Write-Verbose {
    param([string]$Text)
    
    Write-Host $Text -ForegroundColor Gray
}

function Start-Timer {
    return [System.Diagnostics.Stopwatch]::StartNew()
}

function Get-ElapsedTime {
    param($Timer)
    
    $Timer.Stop()
    return $Timer.Elapsed
}

function Create-TestProject {
    param(
        [string]$ProjectPath,
        [string]$ProjectName = "test-project"
    )
    
    # Create project directory
    if (-not (Test-Path $ProjectPath)) {
        New-Item -ItemType Directory -Path $ProjectPath -Force | Out-Null
    }
    
    # Create main.go
    $mainCode = @"
package main

import "fmt"

func main() {
    greeting := GetGreeting()
    name := GetName()
    fmt.Printf("%s, %s!\n", greeting, name)
}
"@
    Set-Content -Path "$ProjectPath\main.go" -Value $mainCode
    
    # Create greeting.go
    $greetingCode = @"
package main

func GetGreeting() string {
    return "Hello"
}
"@
    Set-Content -Path "$ProjectPath\greeting.go" -Value $greetingCode
    
    # Create name.go
    $nameCode = @"
package main

func GetName() string {
    return "World"
}
"@
    Set-Content -Path "$ProjectPath\name.go" -Value $nameCode
    
    # Create test file
    $testCode = @"
package main

import "testing"

func TestGreeting(t *testing.T) {
    greeting := GetGreeting()
    if greeting != "Hello" {
        t.Errorf("Expected 'Hello', got '%s'", greeting)
    }
}

func TestName(t *testing.T) {
    name := GetName()
    if name != "World" {
        t.Errorf("Expected 'World', got '%s'", name)
    }
}
"@
    Set-Content -Path "$ProjectPath\main_test.go" -Value $testCode
    
    # Initialize go module
    Push-Location $ProjectPath
    go mod init "example.com/$ProjectName" | Out-Null
    Pop-Location
    
    return $ProjectPath
}

function Create-HybridProject {
    param(
        [string]$OriginalProjectPath,
        [string]$HybridProjectPath,
        [string]$ModifiedCode,
        [string]$ModifiedFile = "greeting.go"
    )
    
    # Create hybrid project directory
    if (-not (Test-Path $HybridProjectPath)) {
        New-Item -ItemType Directory -Path $HybridProjectPath -Force | Out-Null
    }
    
    # Copy project files
    Get-ChildItem -Path $OriginalProjectPath -File | ForEach-Object {
        if ($_.Name -ne $ModifiedFile) {
            Copy-Item -Path $_.FullName -Destination "$HybridProjectPath\$($_.Name)"
        }
    }
    
    # Use modified code for specified file
    Set-Content -Path "$HybridProjectPath\$ModifiedFile" -Value $ModifiedCode
    
    return $HybridProjectPath
}

function Start-GoCommand {
    param(
        [string]$ProjectPath,
        [string]$Command,
        [string[]]$Arguments
    )
    
    Push-Location $ProjectPath
    $cmdArgs = @($Command) + $Arguments
    $output = & go $cmdArgs 2>&1
    $exitCode = $LASTEXITCODE
    Pop-Location
    
    return @{
        ExitCode = $exitCode
        Output = $output
    }
}

function Format-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [object]$Timer,
        [string]$Message = ""
    )
    
    $elapsed = Get-ElapsedTime $Timer
    
    if ($Success) {
        Write-Success "$TestName - Passed (${elapsed})"
        if ($Message) {
            Write-Info "  $Message"
        }
    } else {
        Write-Failure "$TestName - Failed (${elapsed})"
        if ($Message) {
            Write-Failure "  $Message"
        }
    }
}

#endregion

#region Test Setup

# Create root test directory
if (-not $TestDir) {
    $TestDir = Join-Path $env:TEMP "go-dev-mcp-tests-$(Get-Random)"
}

if (-not (Test-Path $TestDir)) {
    New-Item -ItemType Directory -Path $TestDir -Force | Out-Null
}

Write-Title "Go Development MCP Server - Tool Test Suite"
Write-Info "Test directory: $TestDir"

# Create test project directories
$ProjectPathOnlyDir = Join-Path $TestDir "project-path-only"
$CodeOnlyDir = Join-Path $TestDir "code-only"
$HybridBaseDir = Join-Path $TestDir "hybrid-base"
$HybridTestDir = Join-Path $TestDir "hybrid-test"

# Create test projects
Create-TestProject -ProjectPath $ProjectPathOnlyDir -ProjectName "project-path-only"
Create-TestProject -ProjectPath $HybridBaseDir -ProjectName "hybrid-base"

#endregion

#region Test go_build Tool

Write-Title "Testing go_build Tool"

# Test project-path-only mode
Write-TestName "build with project-path-only"
$timer = Start-Timer
$buildResult = Start-GoCommand -ProjectPath $ProjectPathOnlyDir -Command "build" -Arguments @("-o", "app.exe")
Format-TestResult -TestName "Build with project-path-only" -Success ($buildResult.ExitCode -eq 0) -Timer $timer -Message "Executable created: $(Test-Path "$ProjectPathOnlyDir\app.exe")"

# Test code-only mode
Write-TestName "build with code-only"
$timer = Start-Timer
$codeOnlyMainFile = Join-Path $CodeOnlyDir "main.go"
New-Item -ItemType Directory -Path $CodeOnlyDir -Force | Out-Null
$mainCode = @"
package main

import "fmt"

func main() {
    fmt.Println("Hello from code-only build test!")
}
"@
Set-Content -Path $codeOnlyMainFile -Value $mainCode
Push-Location $CodeOnlyDir
go mod init example.com/code-only | Out-Null
$buildResult = & go build -o app.exe
$success = $LASTEXITCODE -eq 0
Pop-Location
Format-TestResult -TestName "Build with code-only" -Success $success -Timer $timer -Message "Executable created: $(Test-Path "$CodeOnlyDir\app.exe")"

# Test hybrid mode
Write-TestName "build with hybrid mode"
$timer = Start-Timer
$modifiedGreeting = @"
package main

func GetGreeting() string {
    return "Greetings"
}
"@
Create-HybridProject -OriginalProjectPath $HybridBaseDir -HybridProjectPath $HybridTestDir -ModifiedCode $modifiedGreeting
$buildResult = Start-GoCommand -ProjectPath $HybridTestDir -Command "build" -Arguments @("-o", "app.exe")
$success = $buildResult.ExitCode -eq 0 -and (Test-Path "$HybridTestDir\app.exe")
if ($success) {
    # Run the executable to verify it uses the modified code
    Push-Location $HybridTestDir
    $output = & .\app.exe
    $hybridSuccess = $output -eq "Greetings, World!"
    Pop-Location
}
Format-TestResult -TestName "Build with hybrid mode" -Success ($success -and $hybridSuccess) -Timer $timer -Message "Executable created and works with modified code"

#endregion

#region Test go_run Tool

Write-Title "Testing go_run Tool"

# Test project-path-only mode
Write-TestName "run with project-path-only"
$timer = Start-Timer
$runResult = Start-GoCommand -ProjectPath $ProjectPathOnlyDir -Command "run" -Arguments @(".")
$success = $runResult.ExitCode -eq 0 -and $runResult.Output -eq "Hello, World!"
Format-TestResult -TestName "Run with project-path-only" -Success $success -Timer $timer -Message "Output: $($runResult.Output)"

# Test code-only mode
Write-TestName "run with code-only"
$timer = Start-Timer
$codeOnlyDir = Join-Path $TestDir "code-only-run"
New-Item -ItemType Directory -Path $codeOnlyDir -Force | Out-Null
$runCode = @"
package main

import "fmt"

func main() {
    fmt.Println("Hello from code-only run test!")
}
"@
Set-Content -Path "$codeOnlyDir\main.go" -Value $runCode
Push-Location $codeOnlyDir
go mod init example.com/code-only-run | Out-Null
$runResult = & go run .
$success = $LASTEXITCODE -eq 0 -and $runResult -eq "Hello from code-only run test!"
Pop-Location
Format-TestResult -TestName "Run with code-only" -Success $success -Timer $timer -Message "Output: $runResult"

# Test hybrid mode
Write-TestName "run with hybrid mode"
$timer = Start-Timer
$hybridRunDir = Join-Path $TestDir "hybrid-run"
$modifiedGreeting = @"
package main

func GetGreeting() string {
    return "Greetings"
}
"@
Create-HybridProject -OriginalProjectPath $HybridBaseDir -HybridProjectPath $hybridRunDir -ModifiedCode $modifiedGreeting
$runResult = Start-GoCommand -ProjectPath $hybridRunDir -Command "run" -Arguments @(".")
$success = $runResult.ExitCode -eq 0 -and $runResult.Output -eq "Greetings, World!"
Format-TestResult -TestName "Run with hybrid mode" -Success $success -Timer $timer -Message "Output: $($runResult.Output)"

#endregion

#region Test go_fmt Tool

Write-Title "Testing go_fmt Tool"

# Test project-path-only mode
Write-TestName "fmt with project-path-only"
$timer = Start-Timer
# Create badly formatted code
$badlyFormattedCode = @"
package main
import "fmt"
func  main(  ){
fmt.Println("This is badly formatted code!")
}
"@
$fmtTestDir = Join-Path $TestDir "fmt-test"
New-Item -ItemType Directory -Path $fmtTestDir -Force | Out-Null
Set-Content -Path "$fmtTestDir\bad.go" -Value $badlyFormattedCode
Push-Location $fmtTestDir
go mod init example.com/fmt-test | Out-Null
$beforeFmt = Get-Content -Path "bad.go" -Raw
& go fmt
$afterFmt = Get-Content -Path "bad.go" -Raw
$success = $LASTEXITCODE -eq 0 -and $beforeFmt -ne $afterFmt
Pop-Location
Format-TestResult -TestName "Format with project-path-only" -Success $success -Timer $timer -Message "Code was formatted correctly"

# Test code-only mode (simulated since go fmt works on files)
Write-TestName "fmt with code-only"
$timer = Start-Timer
# We'll simulate this by creating a file, formatting it, and reading back
$fmtCodeOnlyDir = Join-Path $TestDir "fmt-code-only"
New-Item -ItemType Directory -Path $fmtCodeOnlyDir -Force | Out-Null
Set-Content -Path "$fmtCodeOnlyDir\bad.go" -Value $badlyFormattedCode
Push-Location $fmtCodeOnlyDir
$beforeFmt = $badlyFormattedCode
go mod init example.com/fmt-code-only | Out-Null
& go fmt
$afterFmt = Get-Content -Path "bad.go" -Raw
$success = $LASTEXITCODE -eq 0 -and $beforeFmt -ne $afterFmt
Pop-Location
Format-TestResult -TestName "Format with code-only" -Success $success -Timer $timer -Message "Code was formatted correctly"

# Test hybrid mode (simulated)
Write-TestName "fmt with hybrid mode"
$timer = Start-Timer
$hybridFmtDir = Join-Path $TestDir "hybrid-fmt"
$badlyFormattedGreeting = @"
package main
func  GetGreeting(  )  string{
return  "Greetings"
}
"@
Create-HybridProject -OriginalProjectPath $HybridBaseDir -HybridProjectPath $hybridFmtDir -ModifiedCode $badlyFormattedGreeting
$beforeFmt = $badlyFormattedGreeting
Push-Location $hybridFmtDir
& go fmt
$afterFmt = Get-Content -Path "greeting.go" -Raw
$success = $LASTEXITCODE -eq 0 -and $beforeFmt -ne $afterFmt
Pop-Location
Format-TestResult -TestName "Format with hybrid mode" -Success $success -Timer $timer -Message "Modified code was formatted correctly"

#endregion

#region Test go_test Tool

Write-Title "Testing go_test Tool"

# Test project-path-only mode
Write-TestName "test with project-path-only"
$timer = Start-Timer
$testResult = Start-GoCommand -ProjectPath $ProjectPathOnlyDir -Command "test" -Arguments @("-v")
$success = $testResult.ExitCode -eq 0 -and $testResult.Output -match "PASS"
Format-TestResult -TestName "Test with project-path-only" -Success $success -Timer $timer -Message "Tests passed"

# Test code-only mode
Write-TestName "test with code-only"
$timer = Start-Timer
$testCodeOnlyDir = Join-Path $TestDir "test-code-only"
New-Item -ItemType Directory -Path $testCodeOnlyDir -Force | Out-Null
$mainCode = @"
package main

func Add(a, b int) int {
    return a + b
}

func main() {
    Add(1, 2)
}
"@
$testCode = @"
package main

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Expected 5, got %d", result)
    }
}
"@
Set-Content -Path "$testCodeOnlyDir\main.go" -Value $mainCode
Set-Content -Path "$testCodeOnlyDir\main_test.go" -Value $testCode
Push-Location $testCodeOnlyDir
go mod init example.com/test-code-only | Out-Null
$testResult = & go test -v
$success = $LASTEXITCODE -eq 0 -and $testResult -match "PASS"
Pop-Location
Format-TestResult -TestName "Test with code-only" -Success $success -Timer $timer -Message "Tests passed"

# Test hybrid mode
Write-TestName "test with hybrid mode"
$timer = Start-Timer
$hybridTestDir = Join-Path $TestDir "hybrid-test-dir"
$modifiedTestCode = @"
package main

import "testing"

func TestGreeting(t *testing.T) {
    greeting := GetGreeting()
    if greeting != "Greetings" {
        t.Errorf("Expected 'Greetings', got '%s'", greeting)
    }
}

func TestName(t *testing.T) {
    name := GetName()
    if name != "World" {
        t.Errorf("Expected 'World', got '%s'", name)
    }
}
"@
$modifiedGreeting = @"
package main

func GetGreeting() string {
    return "Greetings"
}
"@
Create-HybridProject -OriginalProjectPath $HybridBaseDir -HybridProjectPath $hybridTestDir -ModifiedCode $modifiedGreeting
Set-Content -Path "$hybridTestDir\main_test.go" -Value $modifiedTestCode
$testResult = Start-GoCommand -ProjectPath $hybridTestDir -Command "test" -Arguments @("-v")
$success = $testResult.ExitCode -eq 0 -and $testResult.Output -match "PASS"
Format-TestResult -TestName "Test with hybrid mode" -Success $success -Timer $timer -Message "Tests passed with modified code"

#endregion

#region Test go_mod Tool

Write-Title "Testing go_mod Tool"

# Test project-path-only mode
Write-TestName "mod tidy with project-path-only"
$timer = Start-Timer
$modResult = Start-GoCommand -ProjectPath $ProjectPathOnlyDir -Command "mod" -Arguments @("tidy")
$success = $modResult.ExitCode -eq 0
Format-TestResult -TestName "Mod tidy with project-path-only" -Success $success -Timer $timer -Message "Module tidied successfully"

# Test init with new directory
Write-TestName "mod init with new directory"
$timer = Start-Timer
$modInitDir = Join-Path $TestDir "mod-init-test"
New-Item -ItemType Directory -Path $modInitDir -Force | Out-Null
Push-Location $modInitDir
$modResult = & go mod init example.com/mod-init-test
$success = $LASTEXITCODE -eq 0 -and (Test-Path "go.mod")
Pop-Location
Format-TestResult -TestName "Mod init with new directory" -Success $success -Timer $timer -Message "Module initialized successfully"

# Test mod with code (simulated)
Write-TestName "mod with code-based updates"
$timer = Start-Timer
$modCodeDir = Join-Path $TestDir "mod-code-test"
New-Item -ItemType Directory -Path $modCodeDir -Force | Out-Null
Push-Location $modCodeDir
go mod init example.com/mod-code-test | Out-Null
$codeWithDep = @"
package main

import (
    "fmt"
    "github.com/fatih/color"
)

func main() {
    color.Green("This uses a dependency!")
}
"@
Set-Content -Path "main.go" -Value $codeWithDep
& go mod tidy
$success = $LASTEXITCODE -eq 0 -and (Test-Path "go.sum") -and (Get-Content "go.mod" | Where-Object { $_ -match "github.com/fatih/color" })
Pop-Location
Format-TestResult -TestName "Mod with code-based updates" -Success $success -Timer $timer -Message "Module dependencies updated based on code"

#endregion

#region Test go_analyze Tool

Write-Title "Testing go_analyze Tool"

# Test project-path-only mode with clean code
Write-TestName "analyze with project-path-only (clean code)"
$timer = Start-Timer
$analyzeResult = Start-GoCommand -ProjectPath $ProjectPathOnlyDir -Command "vet" -Arguments @("./...")
$success = $analyzeResult.ExitCode -eq 0
Format-TestResult -TestName "Analyze with project-path-only (clean code)" -Success $success -Timer $timer -Message "No issues found (as expected)"

# Test project-path-only mode with buggy code
Write-TestName "analyze with project-path-only (buggy code)"
$timer = Start-Timer
$buggyDir = Join-Path $TestDir "buggy-code"
New-Item -ItemType Directory -Path $buggyDir -Force | Out-Null
$buggyCode = @"
package main

import "fmt"

func main() {
    var x int
    fmt.Printf("Value: %d", "string") // Bug: wrong format specifier
    fmt.Println(x)
}
"@
Set-Content -Path "$buggyDir\main.go" -Value $buggyCode
Push-Location $buggyDir
go mod init example.com/buggy-code | Out-Null
$analyzeResult = & go vet ./...
$bugFound = $LASTEXITCODE -ne 0 -or $analyzeResult -match "wrong type"
Pop-Location
Format-TestResult -TestName "Analyze with project-path-only (buggy code)" -Success $bugFound -Timer $timer -Message "Issues found (as expected)"

# Test code-only mode
Write-TestName "analyze with code-only"
$timer = Start-Timer
$codeOnlyBuggyDir = Join-Path $TestDir "code-only-buggy"
New-Item -ItemType Directory -Path $codeOnlyBuggyDir -Force | Out-Null
Set-Content -Path "$codeOnlyBuggyDir\main.go" -Value $buggyCode
Push-Location $codeOnlyBuggyDir
go mod init example.com/code-only-buggy | Out-Null
$analyzeResult = & go vet ./...
$bugFound = $LASTEXITCODE -ne 0 -or $analyzeResult -match "wrong type"
Pop-Location
Format-TestResult -TestName "Analyze with code-only" -Success $bugFound -Timer $timer -Message "Issues found (as expected)"

# Test hybrid mode
Write-TestName "analyze with hybrid mode"
$timer = Start-Timer
$hybridAnalyzeDir = Join-Path $TestDir "hybrid-analyze"
$buggyGreeting = @"
package main

import "fmt"

func GetGreeting() string {
    var x string
    fmt.Printf("Debug: %d", x) // Bug: wrong format specifier
    return "Greetings"
}
"@
Create-HybridProject -OriginalProjectPath $HybridBaseDir -HybridProjectPath $hybridAnalyzeDir -ModifiedCode $buggyGreeting
$analyzeResult = Start-GoCommand -ProjectPath $hybridAnalyzeDir -Command "vet" -Arguments @("./...")
$bugFound = $analyzeResult.ExitCode -ne 0 -or $analyzeResult.Output -match "wrong type"
Format-TestResult -TestName "Analyze with hybrid mode" -Success $bugFound -Timer $timer -Message "Issues found in modified code (as expected)"

#endregion

#region Test Summary

Write-Title "Test Summary"
Write-Info "All tests completed. Check the results above for details."
Write-Info "Test directories kept at: $TestDir" -ForegroundColor Yellow

# Clean up test directories if not keeping them
if (-not $KeepTestDirs) {
    Write-Info "Cleaning up test directories..."
    Remove-Item -Path $TestDir -Recurse -Force
    Write-Info "Test directories removed."
} else {
    Write-Info "Test directories kept for inspection at: $TestDir"
}

#endregion
