# Create test directory
$TestDir = "$env:TEMP\go-hybrid-test"
New-Item -ItemType Directory -Path $TestDir -Force

# Create greeting.go
$GreetingCode = @"
package main

func GetGreeting() string {
    return "Hello"
}
"@
Set-Content -Path "$TestDir\greeting.go" -Value $GreetingCode

# Create name.go
$NameCode = @"
package main

func GetName() string {
    return "World"
}
"@
Set-Content -Path "$TestDir\name.go" -Value $NameCode

# Create main.go
$MainCode = @"
package main

import "fmt"

func main() {
    greeting := GetGreeting()
    name := GetName()
    fmt.Printf("%s, %s!\n", greeting, name)
}
"@
Set-Content -Path "$TestDir\main.go" -Value $MainCode

# Initialize go module
Push-Location $TestDir
go mod init example.com/hybrid-test
$OriginalOutput = go run .
Pop-Location

Write-Host "Original output: $OriginalOutput"

# Modified greeting for hybrid test
$ModifiedGreeting = @"
package main

func GetGreeting() string {
    return "Greetings"
}
"@

# Create hybrid simulation
$HybridDir = "$env:TEMP\go-hybrid-sim"
New-Item -ItemType Directory -Path $HybridDir -Force

# Copy project structure
Copy-Item -Path "$TestDir\go.mod" -Destination "$HybridDir\go.mod"
Copy-Item -Path "$TestDir\main.go" -Destination "$HybridDir\main.go"
Copy-Item -Path "$TestDir\name.go" -Destination "$HybridDir\name.go"

# Use modified greeting
Set-Content -Path "$HybridDir\greeting.go" -Value $ModifiedGreeting

# Run hybrid simulation
Push-Location $HybridDir
$HybridOutput = go run .
Pop-Location

Write-Host "Hybrid output: $HybridOutput"

if ($HybridOutput -eq "Greetings, World!") {
    Write-Host "SUCCESS: Hybrid strategy verified!"
} else {
    Write-Host "FAILURE: Expected 'Greetings, World!' but got '$HybridOutput'"
}
