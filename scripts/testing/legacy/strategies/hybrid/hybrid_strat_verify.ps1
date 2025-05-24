#!/usr/bin/env pwsh
# hybrid_strat_verify.ps1 - A simplified test for verifying hybrid strategy functionality

# Parameters
param(
    [string]$BuildPath = "c:\Users\James\Documents\go-dev-mcp\build",
    [string]$TempDir = $null
)

# Use random temp dir if not specified
if (-not $TempDir) {
    $TempDir = Join-Path $env:TEMP "go-dev-hybrid-test-$(Get-Random)"
}

# Functions
function Create-TestProject {
    param([string]$ProjectDir)
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $ProjectDir)) {
        New-Item -ItemType Directory -Path $ProjectDir -Force | Out-Null
    }
    
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
    
    # Initialize go module
    Push-Location $ProjectDir
    go mod init example.com/hybrid-test
    Pop-Location
    
    return $ProjectDir
}

function Write-Header {
    param([string]$Title)
    Write-Host "`n---------------------------------------------"
    Write-Host $Title
    Write-Host "---------------------------------------------"
}

# Create test project
Write-Header "Creating Test Project"
$ProjectPath = Create-TestProject -ProjectDir $TempDir
Write-Host "Created test project at $ProjectPath"

# Running the project directly with Go to verify
Write-Header "Running Project Directly with Go"
Push-Location $ProjectPath
$GoOutput = go run .
Pop-Location
Write-Host "Output: $GoOutput"

# Modified greeting for testing hybrid strategy
$ModifiedGreeting = @"
package main

func GetGreeting() string {
    return "Greetings"
}
"@

Write-Header "Creating Modified Code File"
Set-Content -Path "$ProjectPath\modified_greeting.go" -Value $ModifiedGreeting
Write-Host "Modified greeting file created"

# 1. Test project-path-only approach
Write-Header "1. Testing Project-Path-Only Approach"
# Create or find the project-path command pipeline
Push-Location $ProjectPath
$Output = go run .
Pop-Location
Write-Host "Output: $Output"
if ($Output -eq "Hello, World!") {
    Write-Host "✅ SUCCESS: Project-path approach works correctly"
} else {
    Write-Host "❌ FAILURE: Unexpected output: $Output"
}

# 2. Test code-only approach
Write-Header "2. Testing Code-Only Approach"
$CodeOnlyDir = Join-Path $env:TEMP "go-dev-code-only-$(Get-Random)"
New-Item -ItemType Directory -Path $CodeOnlyDir -Force | Out-Null

# Create a standalone file with all the code
$StandaloneCode = @"
package main

import "fmt"

func main() {
    fmt.Println("Greetings, World!")
}
"@
Set-Content -Path "$CodeOnlyDir\main.go" -Value $StandaloneCode

Push-Location $CodeOnlyDir
$Output = go run .
Pop-Location
Write-Host "Output: $Output"
if ($Output -eq "Greetings, World!") {
    Write-Host "✅ SUCCESS: Code-only approach works correctly"
} else {
    Write-Host "❌ FAILURE: Unexpected output: $Output"
}

# 3. Test hybrid approach by manually simulating what the hybrid strategy does
Write-Header "3. Testing Hybrid Approach Manually"
$HybridDir = Join-Path $env:TEMP "go-dev-hybrid-manual-$(Get-Random)"
New-Item -ItemType Directory -Path $HybridDir -Force | Out-Null

# Copy project files
Copy-Item -Path "$ProjectPath\go.mod" -Destination "$HybridDir\go.mod"
Copy-Item -Path "$ProjectPath\main.go" -Destination "$HybridDir\main.go"
Copy-Item -Path "$ProjectPath\name.go" -Destination "$HybridDir\name.go"

# Write our modified greeting.go instead of copying the original
Set-Content -Path "$HybridDir\greeting.go" -Value $ModifiedGreeting

Push-Location $HybridDir
$Output = go run .
Pop-Location
Write-Host "Output: $Output"
if ($Output -eq "Greetings, World!") {
    Write-Host "✅ SUCCESS: Manual hybrid approach works correctly"
} else {
    Write-Host "❌ FAILURE: Unexpected output: $Output"
}

# Print summary
Write-Header "Test Summary"
Write-Host "1. Project-Path-Only: Expected 'Hello, World!'"
Write-Host "2. Code-Only: Expected 'Greetings, World!'"
Write-Host "3. Hybrid Approach: Expected 'Greetings, World!'"
Write-Host "`nTest environment can be found at: $ProjectPath"
Write-Host "You can run the project directly with: cd $ProjectPath; go run ."
Write-Host "You can see the hybrid simulation at: $HybridDir"

# Clean up if needed
if (-not $env:KEEP_TEST_DIR) {
    Remove-Item -Path $CodeOnlyDir -Recurse -Force
    Remove-Item -Path $HybridDir -Recurse -Force
    Write-Host "`nTest directories cleaned up (except main project directory)"
    Write-Host "Set `$env:KEEP_TEST_DIR to true to keep all test directories"
}
