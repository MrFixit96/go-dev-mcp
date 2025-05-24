#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Common utility functions for MCP server testing scripts.

.DESCRIPTION
    This file contains shared utility functions used across multiple testing scripts.
    It provides consistent formatting, project creation, and testing utilities.

.EXAMPLE
    # Include in your testing script
    . "$PSScriptRoot\..\utils\test_utils.ps1"
#>

#region Formatting Functions

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

function Write-DetailedInfo {
    param([string]$Text)
    
    if ($VerbosePreference -eq "Continue") {
        Write-Host $Text -ForegroundColor Gray
    }
}

function Start-Timer {
    return [System.Diagnostics.Stopwatch]::StartNew()
}

function Format-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [System.Diagnostics.Stopwatch]$Timer,
        [string]$Message = ""
    )
    
    $Timer.Stop()
    $elapsed = $Timer.Elapsed.TotalSeconds.ToString("0.000")
    
    if ($Success) {
        Write-Success "$TestName (${elapsed}s): $Message"
        return $true
    } else {
        Write-Failure "$TestName (${elapsed}s): $Message"
        return $false
    }
}

#endregion

#region Project Creation Functions

function Create-TestProject {
    param(
        [string]$ProjectPath,
        [string]$ProjectName = "test-project",
        [string]$ProjectType = "simple" # Can be "simple", "multiple-files", or "with-deps"
    )
    
    # Create directory if it doesn't exist
    if (-not (Test-Path $ProjectPath)) {
        New-Item -ItemType Directory -Path $ProjectPath -Force | Out-Null
    }
    
    # Create project based on type
    switch ($ProjectType) {
        "simple" {
            # Create main.go file
            $MainCode = @"
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
"@
            Set-Content -Path "$ProjectPath\main.go" -Value $MainCode
        }
        
        "multiple-files" {
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
            Set-Content -Path "$ProjectPath\main.go" -Value $MainCode
            
            # Create greeting.go file
            $GreetingCode = @"
package main

func GetGreeting() string {
    return "Hello"
}
"@
            Set-Content -Path "$ProjectPath\greeting.go" -Value $GreetingCode
            
            # Create name.go file
            $NameCode = @"
package main

func GetName() string {
    return "World"
}
"@
            Set-Content -Path "$ProjectPath\name.go" -Value $NameCode
        }
        
        "with-deps" {
            # Create main.go file that uses an external dependency
            $MainCode = @"
package main

import (
    "fmt"
    "github.com/fatih/color"
)

func main() {
    c := color.New(color.FgCyan)
    c.Println("Hello, World!")
}
"@
            Set-Content -Path "$ProjectPath\main.go" -Value $MainCode
        }
    }
    
    # Initialize go module
    Push-Location $ProjectPath
    go mod init example.com/$ProjectName
    if ($ProjectType -eq "with-deps") {
        go get github.com/fatih/color
        go mod tidy
    }
    Pop-Location
    
    return $ProjectPath
}

function Create-HybridProject {
    param(
        [string]$OriginalProjectPath,
        [string]$HybridProjectPath,
        [string]$ModifiedCode,
        [string]$ModifiedFile = "main.go"
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

#endregion

#region Go Command Functions

function Start-GoCommand {
    param(
        [string]$ProjectPath,
        [string]$Command,
        [string[]]$Arguments = @(),
        [int]$TimeoutSeconds = 10
    )
    
    Write-Verbose "Running 'go $Command $Arguments' in $ProjectPath"
    
    Push-Location $ProjectPath
    
    $startTime = Get-Date
    $process = Start-Process -FilePath "go" -ArgumentList (@($Command) + $Arguments) -NoNewWindow -Wait -PassThru
    $elapsed = (Get-Date) - $startTime
    
    Pop-Location
    
    $result = @{
        ExitCode = $process.ExitCode
        ElapsedMs = $elapsed.TotalMilliseconds
    }
    
    Write-Verbose "Command completed with exit code $($result.ExitCode) in $($result.ElapsedMs)ms"
    
    return $result
}

#endregion

#region Validation Functions

function Invoke-ServerTool {
    param(
        [string]$ServerExecutable,
        [string]$ToolName,
        [hashtable]$Params,
        [int]$TimeoutSeconds = 10
    )
    
    $requestObj = @{
        jsonrpc = "2.0"
        id = "test-$(Get-Random)"
        method = "calltool"
        params = @{
            name = $ToolName
            input = $Params
        }
    }
    
    $requestJson = $requestObj | ConvertTo-Json -Depth 10
    
    # Save request to temporary file
    $tempFile = [System.IO.Path]::GetTempFileName()
    Set-Content -Path $tempFile -Value $requestJson
    
    try {
        # Execute server with input from temporary file
        $output = Get-Content $tempFile | & $ServerExecutable
        
        # Parse response
        $response = $output | ConvertFrom-Json
        return $response
    }
    catch {
        Write-Failure "Failed to invoke server tool: $_"
        return $null
    }
    finally {
        # Clean up temporary file
        Remove-Item -Path $tempFile -Force
    }
}

#endregion
