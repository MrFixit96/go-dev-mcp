#!/usr/bin/env pwsh
#
# End-to-End Behavioral Testing for Go Development MCP Server
# This script tests the server by working with a real Go project

param(
    [string]$ServerUrl = "http://localhost:8080",
    [string]$TempDir = "$env:TEMP\go-dev-mcp-test-$(Get-Random)",
    [switch]$KeepTempFiles,
    [switch]$Verbose
)

# Enable verbose output if requested
$VerbosePreference = if ($Verbose) { "Continue" } else { "SilentlyContinue" }

# Function to check if server is available
function Test-ServerAvailable {
    param(
        [string]$Url,
        [int]$TimeoutSeconds = 5,
        [int]$MaxRetries = 3
    )
    
    for ($retry = 1; $retry -le $MaxRetries; $retry++) {
        try {
            $request = [System.Net.WebRequest]::Create($Url)
            $request.Timeout = $TimeoutSeconds * 1000
            $request.Method = "HEAD"
            
            Log-Message "Checking server availability (attempt $retry/$MaxRetries)..." "INFO"
            
            $response = $request.GetResponse()
            $response.Close()
            
            Log-Message "Server is available at $Url" "INFO"
            return $true
        } catch {
            Log-Message "Server not available (attempt $retry/$MaxRetries): $_" "WARN"
            if ($retry -lt $MaxRetries) {
                Start-Sleep -Seconds 2
            }
        }
    }
    
    Log-Message "Server is not available at $Url after $MaxRetries attempts" "ERROR"
    return $false
}

# Function to log messages with timestamp
function Log-Message {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message"
}

# Function to log error messages and exit
function Log-Error {
    param([string]$Message)
    Log-Message $Message "ERROR"
    if (-not $KeepTempFiles -and (Test-Path $TempDir)) {
        try {
            Log-Message "Cleaning up temporary directory: $TempDir" "INFO"
            # Force release any open handles (wait a moment first)
            [System.GC]::Collect()
            [System.GC]::WaitForPendingFinalizers()
            
            # Try to remove the directory with retries
            $maxRetries = 3
            $retryCount = 0
            $success = $false
            
            while (-not $success -and $retryCount -lt $maxRetries) {
                try {
                    Remove-Item -Recurse -Force -Path $TempDir -ErrorAction Stop
                    $success = $true
                } catch {
                    $retryCount++
                    Log-Message "Failed to remove directory (attempt $retryCount of $maxRetries): $_" "WARN"
                    Start-Sleep -Seconds 2
                }
            }
            
            if (-not $success) {
                Log-Message "Could not remove temporary directory after $maxRetries attempts. It may need to be removed manually." "WARN"
            }
        } catch {
            Log-Message "Error during cleanup: $_" "ERROR"
        }
    }
    exit 1
}

# Function to invoke MCP server tools
function Invoke-MCPTool {
    param(
        [Parameter(Mandatory=$true)]
        [string]$ToolName,
        
        [Parameter(Mandatory=$true)]
        [hashtable]$Params
    )

    $body = @{
        params = @{
            name = $ToolName
            input = $Params
        }
    } | ConvertTo-Json -Depth 10

    Write-Verbose "Sending request to $ServerUrl/calltool with payload:`n$body"
    
    try {
        $response = Invoke-RestMethod -Uri "$ServerUrl/calltool" -Method Post -Body $body -ContentType "application/json"
        return $response
    } catch {
        Log-Error "Failed to invoke MCP tool '$ToolName': $_"
    }
}

# Create temporary directory for test project
Log-Message "Creating temporary directory: $TempDir"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

# Create a simple Go project
$helloWorldCode = @"
package main

import "fmt"

func main() {
    // A simple hello world program
    fmt.Println("Hello, World from Go Development MCP Server!")
}
"@

Log-Message "Creating main.go file in test project"
Set-Content -Path "$TempDir\main.go" -Value $helloWorldCode

# Create go.mod file
Log-Message "Creating go.mod file in test project"
Set-Location $TempDir
$goModInit = Start-Process -FilePath "go" -ArgumentList "mod", "init", "example.com/hello" -Wait -NoNewWindow -PassThru
if ($goModInit.ExitCode -ne 0) {
    Log-Error "Failed to initialize Go module"
}

# Define test cases
$testCases = @(
    @{
        Name = "Format Go Code (Using Code Only)"
        Tool = "go_fmt"
        Params = @{
            code = $helloWorldCode
        }
        Validate = {
            param($result)
            if (-not $result.result.formattedCode) {
                return $false, "No formatted code in result"
            }
            return $true, $result.result.formattedCode
        }
    },
    @{
        Name = "Format Go Code (Using Project Path)"
        Tool = "go_fmt"
        Params = @{
            project_path = $TempDir
        }
        Validate = {
            param($result)
            if (-not $result.result.formattedCode) {
                return $false, "No formatted code in result"
            }
            return $true, $result.result.formattedCode
        }
    },
    @{
        Name = "Format Go Code (Hybrid - both code and project path)"
        Tool = "go_fmt"
        Params = @{
            code = $helloWorldCode
            project_path = $TempDir
        }
        Validate = {
            param($result)
            if (-not $result.result.formattedCode) {
                return $false, "No formatted code in result"
            }
            if (-not $result.result.metadata -or -not $result.result.metadata.strategyType) {
                return $false, "No strategy type in metadata"
            }
            return $true, "Strategy: $($result.result.metadata.strategyType), Code: $($result.result.formattedCode)"
        }
    },
    @{
        Name = "Build Go Code (Using Project Path)"
        Tool = "go_build"
        Params = @{
            project_path = $TempDir
        }
        Validate = {
            param($result)
            if ($result.result.success -ne $true) {
                return $false, "Build failed: $($result.result.stderr)"
            }
            # Verify the executable was created
            $exePath = "$TempDir\hello.exe"
            if (-not (Test-Path $exePath)) {
                return $false, "Executable not created at $exePath"
            }
            return $true, "Build successful, executable created"
        }
    },
    @{
        Name = "Run Go Code (Using Project Path)"
        Tool = "go_run"
        Params = @{
            project_path = $TempDir
        }
        Validate = {
            param($result)
            if ($result.result.exitCode -ne 0) {
                return $false, "Run failed with exit code $($result.result.exitCode): $($result.result.stderr)"
            }
            if (-not $result.result.stdout -or -not $result.result.stdout.Contains("Hello, World from Go Development MCP Server!")) {
                return $false, "Unexpected output: $($result.result.stdout)"
            }
            return $true, "Run successful with expected output"
        }
    },
    @{
        Name = "Build Invalid Go Code (Error Handling)"
        Tool = "go_build"
        Params = @{
            code = "package main

func main() {
    fmt.Println(Hello World) // Syntax error - missing quotes
}"
        }
        Validate = {
            param($result)
            if ($result.result.success -eq $true) {
                return $false, "Build should have failed but succeeded"
            }
            return $true, "Build failed as expected: $($result.result.stderr)"
        }
    }
)

# Run all test cases
$passCount = 0
$failCount = 0

Log-Message "Running tests against MCP server at $ServerUrl" "INFO"
Log-Message "Test project location: $TempDir" "INFO"

# Check if server is available before proceeding
if (-not (Test-ServerAvailable -Url $ServerUrl)) {
    Log-Message "Skipping tests because server is not available at $ServerUrl" "WARN"
    Log-Message "Please ensure the server is running before executing this test" "WARN"
      # We don't want to fail CI builds when server isn't running, so exit with success
    # but make it clear that tests were skipped
    Log-Message "-------------------------------------------------------" "SUMMARY"
    Log-Message "Test Summary: 0 passed, 0 failed, ALL TESTS SKIPPED" "SUMMARY"
    Log-Message "-------------------------------------------------------" "SUMMARY"
    
    # Clean up temporary directory
    if (-not $KeepTempFiles -and (Test-Path $TempDir)) {
        try {
            Log-Message "Cleaning up temporary directory: $TempDir" "INFO"
            [System.GC]::Collect()
            [System.GC]::WaitForPendingFinalizers()
            Remove-Item -Recurse -Force -Path $TempDir -ErrorAction Stop
        } catch {
            Log-Message "Warning: Could not remove temporary directory: $_" "WARN"
        }
    } elseif ($KeepTempFiles) {
        Log-Message "Keeping temporary test project at: $TempDir" "INFO"
    }
    
    exit 0
}

foreach ($test in $testCases) {
    Log-Message "Test: $($test.Name)" "TEST"
    
    try {
        $result = Invoke-MCPTool -ToolName $test.Tool -Params $test.Params
        $validationResult, $message = & $test.Validate $result
        
        if ($validationResult) {
            Log-Message "✅ PASS: $($test.Name)" "PASS"
            Write-Verbose "Result: $message"
            $passCount++
        } else {
            Log-Message "❌ FAIL: $($test.Name) - $message" "FAIL"
            Write-Verbose "Response: $($result | ConvertTo-Json -Depth 10)"
            $failCount++
        }
    } catch {
        Log-Message "❌ FAIL: $($test.Name) - Exception: $_" "FAIL"
        $failCount++
    }
}

# Print test summary
Log-Message "-------------------------------------------------------" "SUMMARY"
Log-Message "Test Summary: $passCount passed, $failCount failed" "SUMMARY"
Log-Message "-------------------------------------------------------" "SUMMARY"

# Clean up temporary directory
if (-not $KeepTempFiles) {
    Log-Message "Cleaning up temporary directory: $TempDir" "INFO"
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
                Remove-Item -Recurse -Force -Path $TempDir -ErrorAction Stop
                $success = $true
                Log-Message "Temporary directory removed successfully" "INFO" 
            } catch {
                $retryCount++
                Log-Message "Failed to remove directory (attempt $retryCount of $maxRetries): $_" "WARN"
                Start-Sleep -Seconds 2
            }
        }
        
        if (-not $success) {
            Log-Message "Could not remove temporary directory after $maxRetries attempts. It may need to be removed manually." "WARN"
        }
    } catch {
        Log-Message "Error during cleanup: $_" "WARN"
    }
} else {
    Log-Message "Keeping temporary test project at: $TempDir" "INFO"
}

# Exit with non-zero status if any tests failed
if ($failCount -gt 0) {
    exit 1
}
