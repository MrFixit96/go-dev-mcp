# Hybrid Strategy CLI Test
# This script tests the hybrid execution strategy by directly calling the server executable

# Create a temporary directory for test files
$TestDir = "$env:TEMP\go-dev-mcp-test-$(Get-Random)"
New-Item -ItemType Directory -Path $TestDir -Force

Write-Host "Creating test project in $TestDir"

# Create main.go file
$MainCode = @"
package main

import (
	"fmt"
)

func main() {
	greeting := GetGreeting()
	name := GetName()
	fmt.Printf("%s, %s!\n", greeting, name)
}
"@

# Create greeting.go file
$GreetingCode = @"
package main

func GetGreeting() string {
	return "Hello"
}
"@

# Create name.go file
$NameCode = @"
package main

func GetName() string {
	return "World"
}
"@

# Write files to disk
Set-Content -Path "$TestDir\main.go" -Value $MainCode
Set-Content -Path "$TestDir\greeting.go" -Value $GreetingCode
Set-Content -Path "$TestDir\name.go" -Value $NameCode

# Initialize go module
Push-Location $TestDir
go mod init example.com/hybrid-test
Pop-Location

# Create modified greeting file
$ModifiedGreeting = @"
package main

func GetGreeting() string {
	return "Greetings"
}
"@

# Create test input JSON for hybrid mode (JSON-RPC format)
$InputJson = @"
{
  "jsonrpc": "2.0",
  "id": "test-$(Get-Random)",
  "method": "calltool",
  "params": {
    "name": "go_run",
    "input": {
      "code": "$($ModifiedGreeting -replace "`n", "\\n" -replace '"', '\"')",
      "project_path": "$($TestDir -replace '\\', '\\\\')"
    }
  }
}
"@

# Save JSON to a file
$JsonFile = "$TestDir\input.json"
Set-Content -Path $JsonFile -Value $InputJson

Write-Host "Testing server with hybrid input mode..."
Write-Host "Running the project normally first..."

# Run the project directly first to verify baseline
Push-Location $TestDir
$BaselineOutput = go run .
Pop-Location
Write-Host "Baseline output: $BaselineOutput"

# Execute the server with our test input
$ServerPath = "c:\Users\James\Documents\go-dev-mcp\build\server.exe"
Write-Host "Sending JSON-RPC request to the server..."
$ServerOutput = Get-Content $JsonFile | & $ServerPath

Write-Host "Raw server output:"
$ServerOutput

# Parse the JSON response
try {
    $Response = $ServerOutput | ConvertFrom-Json
    Write-Host "Server response parsed"
    
    if ($Response.result -and $Response.result.result) {
        if ($Response.result.result.stdout) {
            Write-Host "Hybrid execution output: $($Response.result.result.stdout.Trim())"
            
            # Check if the output matches expectations
            if ($Response.result.result.stdout -match "Greetings, World") {
                Write-Host "✅ SUCCESS: Hybrid strategy worked correctly! Modified greeting was used."
            } else {
                Write-Host "❌ FAILURE: Output doesn't match expectations: $($Response.result.result.stdout)"
            }
            
            # Check for strategy metadata
            if ($Response.result.result.metadata -and $Response.result.result.metadata.strategyType) {
                Write-Host "Strategy used: $($Response.result.result.metadata.strategyType)"
            } else {
                Write-Host "⚠️ WARNING: No strategy metadata in response"
            }
        } else {
            Write-Host "❌ FAILURE: No stdout in result.result"
        }
    } else {
        Write-Host "❌ FAILURE: Unexpected response format"
    }
} catch {
    Write-Host "❌ FAILURE: Could not parse server response as JSON: $_"
}

# Clean up
if ($env:KEEP_TEST_DIR -ne "true") {
    Write-Host "Cleaning up test directory"
    Remove-Item -Recurse -Force $TestDir
} else {
    Write-Host "Test files kept at: $TestDir"
}
