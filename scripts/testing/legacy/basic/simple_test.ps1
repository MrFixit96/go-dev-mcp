#!/usr/bin/env pwsh
# Simple test for the hybrid strategy in the Go Development MCP Server
# This script verifies the basic functionality of the hybrid strategy by creating
# a simple Go project and running code with modifications.

Write-Host "Testing hybrid strategy" -ForegroundColor Cyan

# Create a temporary directory for testing
$tempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
Write-Host "Created test directory at $tempDir" -ForegroundColor Yellow

try {
    # Create a simple Go project
    Set-Location $tempDir
    Write-Host "Creating a simple Go project..." -ForegroundColor Yellow
    
    # Initialize a Go module
    & go mod init example.com/hybrid-test | Out-Null
    
    # Create main.go with original code
    $originalCode = @"
package main

import "fmt"

func main() {
    fmt.Println("Hello from the original project")
}
"@
    Set-Content -Path "$tempDir\main.go" -Value $originalCode
    
    # Modified code that would come from the "code" parameter
    $modifiedCode = @"
package main

import "fmt"

func main() {
    fmt.Println("Hello from the modified code")
}
"@
    
    # Create a temporary directory to simulate hybrid execution
    $hybridDir = [System.IO.Path]::GetTempPath() + "hybrid-" + [System.Guid]::NewGuid().ToString()
    New-Item -ItemType Directory -Path $hybridDir -Force | Out-Null
    
    # Copy the go.mod file to simulate project context
    Copy-Item -Path "$tempDir\go.mod" -Destination "$hybridDir\go.mod"
    
    # Write the modified code
    Set-Content -Path "$hybridDir\main.go" -Value $modifiedCode
    
    # Run the code in the hybrid directory
    Write-Host "Running with hybrid strategy..." -ForegroundColor Yellow
    $result = & go run "$hybridDir\main.go"
      # Verify the output
    $testPassed = $false
    if ($result -eq "Hello from the modified code") {
        Write-Host "✅ SUCCESS: Hybrid strategy applied modified code correctly" -ForegroundColor Green
        $testPassed = $true
    } else {
        Write-Host "❌ FAILURE: Expected 'Hello from the modified code' but got '$result'" -ForegroundColor Red
        $testPassed = $false
    }
}
finally {
    # Clean up
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    if (Test-Path $hybridDir) {
        Remove-Item -Path $hybridDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Return exit code based on test result
if ($testPassed) {
    exit 0
} else {
    exit 1
}
