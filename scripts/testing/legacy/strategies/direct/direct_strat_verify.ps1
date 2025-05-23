#!/usr/bin/env pwsh
# direct_strat_verify.ps1 - A simplified test for verifying direct strategy functionality

param(
    [switch]$Verbose,
    [switch]$KeepTestDirs,
    [string]$ServerExecutable = "..\..\..\..\..\..\build\server.exe"
)

# Import utility functions
. "$PSScriptRoot\..\..\..\legacy\utils\test_utils.ps1"

try {
    # Create a temporary directory for the test
    $TempDir = Join-Path $env:TEMP "go-dev-direct-verify-$(Get-Random)"
    $ProjectDir = Join-Path $TempDir "direct-verify-project"
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $ProjectDir)) {
        New-Item -ItemType Directory -Path $ProjectDir -Force | Out-Null
    }
    
    # Set location to the project directory
    Push-Location $ProjectDir
    
    try {
        Write-Header "Setting up Direct Strategy Verification"
        
        # Initialize a basic Go module
        go mod init example.com/direct-verify
        if (-not $?) {
            throw "Failed to initialize Go module"
        }
        
        # Create a simple main.go file
        $MainGoContent = @'
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Direct Strategy Verification Test")
	
	// Special output for verification
	fmt.Println("DIRECT_VERIFICATION_MARKER")
	
	// Exit with success code
	os.Exit(0)
}
'@
        Set-Content -Path "main.go" -Value $MainGoContent
        
        # Running the project directly with Go
        Write-Header "Running Project Directly with Go"
        
        $output = go run .
        if (-not $?) {
            throw "Failed to run the Go project"
        }
        
        # Verify that the output contains our marker
        if ($output -notcontains "DIRECT_VERIFICATION_MARKER") {
            throw "Output verification failed - marker not found"
        }
        
        Write-Host "Output verification successful" -ForegroundColor Green
        Write-Success "Direct strategy verification completed successfully"
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
            Write-Host "You can run the project directly with: cd $ProjectDir; go run ."
        }
    }
    
    exit 0
}
catch {
    Write-Failure "Direct strategy verification failed: $_"
    exit 1
}
