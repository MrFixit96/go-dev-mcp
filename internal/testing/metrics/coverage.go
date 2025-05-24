// filepath: c:\Users\James\Documents\go-dev-mcp\internal\testing\metrics\coverage.go
package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CoverageResult represents a Go test coverage profile
type CoverageResult struct {
	TotalLines   int                        `json:"total_lines"`
	CoveredLines int                        `json:"covered_lines"`
	Percentage   float64                    `json:"percentage"`
	Packages     map[string]PackageCoverage `json:"packages"`
}

// PackageCoverage represents coverage for a single package
type PackageCoverage struct {
	TotalLines   int     `json:"total_lines"`
	CoveredLines int     `json:"covered_lines"`
	Percentage   float64 `json:"percentage"`
}

// GenerateCoverageProfile creates a coverage profile for the specified packages
func GenerateCoverageProfile(outputPath string, packages ...string) error {
	// Set default if no packages specified
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	// Prepare command
	args := []string{"test"}
	args = append(args, packages...)
	args = append(args, "-coverprofile="+outputPath)

	// Run the go test command with coverage
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate coverage profile: %w", err)
	}

	return nil
}

// AnalyzeCoverageProfile parses a coverage profile and returns statistics
func AnalyzeCoverageProfile(profilePath string) (*CoverageResult, error) {
	// Check if the profile exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("coverage profile not found: %s", profilePath)
	}

	// Run go tool cover to get coverage stats
	cmd := exec.Command("go", "tool", "cover", "-func="+profilePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze coverage profile: %w", err)
	}

	// Parse the coverage output
	result := &CoverageResult{
		Packages: make(map[string]PackageCoverage),
	}

	totalCovered, totalLines := 0, 0
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for the package-level metrics
		if strings.Contains(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				percentStr := strings.TrimSuffix(fields[len(fields)-1], "%")
				var percentage float64
				fmt.Sscanf(percentStr, "%f", &percentage)
				result.Percentage = percentage
			}
			continue
		}

		// Parse per-file coverage
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		filePath := strings.TrimSpace(parts[0])
		pkgName := extractPackageName(filePath)

		// Extract coverage percentage from each function
		funcParts := strings.Split(parts[1], "\t")
		if len(funcParts) < 2 {
			continue
		}

		coverageStr := strings.TrimSpace(funcParts[len(funcParts)-1])
		if coverageStr == "-" {
			continue // Skip functions with no statements
		}

		coverageStr = strings.TrimSuffix(coverageStr, "%")
		var coverage float64
		fmt.Sscanf(coverageStr, "%f", &coverage)

		// For simplicity, we're estimating that each function has on average 5 statements
		// A more precise approach would parse the actual coverage profile in detail
		// or use go tool cover -html to get exact statement counts
		estStatements := 5
		estCovered := int(float64(estStatements) * coverage / 100.0)

		// Update package coverage
		pkg, exists := result.Packages[pkgName]
		if !exists {
			pkg = PackageCoverage{}
		}
		pkg.TotalLines += estStatements
		pkg.CoveredLines += estCovered
		pkg.Percentage = float64(pkg.CoveredLines) / float64(pkg.TotalLines) * 100.0
		result.Packages[pkgName] = pkg

		// Update total coverage
		totalLines += estStatements
		totalCovered += estCovered
	}

	// Update total coverage
	result.TotalLines = totalLines
	result.CoveredLines = totalCovered
	if totalLines > 0 {
		result.Percentage = float64(totalCovered) / float64(totalLines) * 100.0
	}

	return result, nil
}

// GenerateCoverageHTML generates an HTML coverage report
func GenerateCoverageHTML(profilePath, htmlOutputPath string) error {
	cmd := exec.Command("go", "tool", "cover", "-html="+profilePath, "-o="+htmlOutputPath)
	return cmd.Run()
}

// extractPackageName extracts the package name from a file path
func extractPackageName(filePath string) string {
	// Simplistic package name extraction - in reality, you'd need to parse the file
	// or use go list to get the actual package name
	dir := filepath.Dir(filePath)
	return filepath.Base(dir)
}

// WriteCoverageReport outputs a coverage report in JSON format
func WriteCoverageReport(result *CoverageResult, outputPath string) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal coverage result: %w", err)
	}

	return os.WriteFile(outputPath, jsonData, 0644)
}

// PrintCoverageReport outputs a coverage report to the specified writer
func PrintCoverageReport(w io.Writer, result *CoverageResult) {
	fmt.Fprintf(w, "Coverage Report:\n")
	fmt.Fprintf(w, "----------------\n")
	fmt.Fprintf(w, "Total coverage: %.2f%% (%d/%d lines)\n\n",
		result.Percentage, result.CoveredLines, result.TotalLines)

	fmt.Fprintf(w, "Package Coverage:\n")
	for pkgName, pkg := range result.Packages {
		fmt.Fprintf(w, "  %s: %.2f%% (%d/%d lines)\n",
			pkgName, pkg.Percentage, pkg.CoveredLines, pkg.TotalLines)
	}
}

// RunCoverageReport generates and analyzes coverage for the given packages
func RunCoverageReport(outputDir string, packages ...string) (*CoverageResult, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate coverage profile
	profilePath := filepath.Join(outputDir, "coverage.out")
	if err := GenerateCoverageProfile(profilePath, packages...); err != nil {
		return nil, err
	}

	// Analyze coverage profile
	result, err := AnalyzeCoverageProfile(profilePath)
	if err != nil {
		return nil, err
	}

	// Generate HTML report
	htmlPath := filepath.Join(outputDir, "coverage.html")
	if err := GenerateCoverageHTML(profilePath, htmlPath); err != nil {
		return nil, fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// Write JSON report
	jsonPath := filepath.Join(outputDir, "coverage.json")
	if err := WriteCoverageReport(result, jsonPath); err != nil {
		return nil, fmt.Errorf("failed to write JSON report: %w", err)
	}

	fmt.Printf("Coverage reports generated in %s\n", outputDir)
	fmt.Printf("- HTML report: %s\n", htmlPath)
	fmt.Printf("- JSON report: %s\n", jsonPath)

	return result, nil
}
