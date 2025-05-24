#!/usr/bin/env pwsh
# direct_strategy_test.ps1 - A test for verifying direct strategy functionality

param(
    [switch]$Verbose,
    [switch]$KeepTestDirs,
    [string]$ServerExecutable = "..\..\..\..\..\..\build\server.exe"
)

# Import utility functions
. "$PSScriptRoot\..\..\..\legacy\utils\test_utils.ps1"

try {
    # Create a temporary directory for the test
    $TempDir = Join-Path $env:TEMP "go-dev-direct-test-$(Get-Random)"
    $ProjectDir = Join-Path $TempDir "direct-test-project"
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $ProjectDir)) {
        New-Item -ItemType Directory -Path $ProjectDir -Force | Out-Null
    }
    
    # Set location to the project directory
    Push-Location $ProjectDir
    
    try {
        Write-Header "Setting up Direct Test Environment"
        
        # Initialize a basic Go module
        go mod init example.com/direct-test
        if (-not $?) {
            throw "Failed to initialize Go module"
        }
        
        # Create a simple main.go file
        $MainGoContent = @'
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello from Direct Strategy Test!")
	result := addNumbers(5, 7)
	fmt.Printf("5 + 7 = %d\n", result)
}

func addNumbers(a, b int) int {
	return a + b
}
'@
        Set-Content -Path "main.go" -Value $MainGoContent
        
        # Create a simple Go test file
        $TestGoContent = @'
package main

import "testing"

func TestAddNumbers(t *testing.T) {
	result := addNumbers(5, 7)
	expected := 12
	if result != expected {
		t.Errorf("Expected %d but got %d", expected, result)
	}
}
'@
        Set-Content -Path "main_test.go" -Value $TestGoContent
        
        # Run the project to verify functionality
        Write-Header "Running Project with Go"
        
        $output = go run .
        if (-not $?) {
            throw "Failed to run the Go project"
        }
        
        Write-Host $output -ForegroundColor Green
        
        # Run the test to verify testing functionality
        Write-Header "Running Tests with Go"
        
        $output = go test -v
        if (-not $?) {
            throw "Tests failed"
        }
        
        Write-Host $output -ForegroundColor Green
        
        Write-Success "Direct strategy test completed successfully"
    }
    finally {
        # Return to the original location
        Pop-Location
        
        # Clean up test directory unless asked to keep it
        if (-not $KeepTestDirs) {
            if (Test-Path $TempDir) {
                Remove-Item -Path $TempDir -Recurse -Force
                if ($Verbose) {
                    Write-Host "Removed temporary test directory: $TempDir" -ForegroundColor Gray
                }
            }
        }
        else {
            Write-Host "Test directory kept at: $TempDir" -ForegroundColor Yellow
        }
    }
    
    exit 0
}
catch {
    Write-Failure "Direct strategy test failed: $_"
    exit 1
}
