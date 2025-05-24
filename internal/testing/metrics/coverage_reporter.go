// Package metrics provides coverage reporting utilities for MCP tests.
package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CoverageData represents code coverage data for a package.
type CoverageData struct {
	Package      string         `json:"package"`
	TotalLines   int            `json:"totalLines"`
	CoveredLines int            `json:"coveredLines"`
	Coverage     float64        `json:"coverage"`
	Timestamp    time.Time      `json:"timestamp"`
	Files        []FileCoverage `json:"files"`
}

// FileCoverage represents code coverage data for a specific file.
type FileCoverage struct {
	Filename     string  `json:"filename"`
	TotalLines   int     `json:"totalLines"`
	CoveredLines int     `json:"coveredLines"`
	Coverage     float64 `json:"coverage"`
}

// CoverageReport represents a comprehensive coverage report.
type CoverageReport struct {
	Timestamp       time.Time      `json:"timestamp"`
	TotalCoverage   float64        `json:"totalCoverage"`
	TotalLines      int            `json:"totalLines"`
	CoveredLines    int            `json:"coveredLines"`
	PackageCoverage []CoverageData `json:"packageCoverage"`
}

// GenerateCoverageReport generates a coverage report for the specified packages.
// If no packages are specified, all packages are included.
func GenerateCoverageReport(coverageFile string, packages ...string) (*CoverageReport, error) {
	// Ensure the coverage file exists
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("coverage file not found: %s", coverageFile)
	}

	// Parse the coverage file
	packageCoverage, err := parseCoverageFile(coverageFile)
	if err != nil {
		return nil, err
	}

	// Filter packages if specified
	if len(packages) > 0 {
		filteredCoverage := make([]CoverageData, 0)
		for _, pkg := range packageCoverage {
			for _, requestedPkg := range packages {
				if strings.Contains(pkg.Package, requestedPkg) {
					filteredCoverage = append(filteredCoverage, pkg)
					break
				}
			}
		}
		packageCoverage = filteredCoverage
	}

	// Calculate total coverage
	totalLines := 0
	coveredLines := 0
	for _, pkg := range packageCoverage {
		totalLines += pkg.TotalLines
		coveredLines += pkg.CoveredLines
	}

	totalCoverage := 0.0
	if totalLines > 0 {
		totalCoverage = float64(coveredLines) * 100.0 / float64(totalLines)
	}

	// Create the report
	report := &CoverageReport{
		Timestamp:       time.Now(),
		TotalCoverage:   totalCoverage,
		TotalLines:      totalLines,
		CoveredLines:    coveredLines,
		PackageCoverage: packageCoverage,
	}

	return report, nil
}

// parseCoverageFile parses the coverage file output by "go test -coverprofile".
func parseCoverageFile(coverageFile string) ([]CoverageData, error) {
	cmd := exec.Command("go", "tool", "cover", "-func", coverageFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run go tool cover: %w", err)
	}

	// Process output by package
	packageMap := make(map[string]*CoverageData)

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split the line into parts
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		// Extract package and file
		fileInfo := strings.Split(parts[0], ":")
		if len(fileInfo) < 2 {
			continue
		}

		fullPath := fileInfo[0]
		pkgPath := filepath.Dir(fullPath)
		fileName := filepath.Base(fullPath)

		// Extract coverage percentage
		if parts[len(parts)-1] == "(statements)" {
			// Total coverage line
			coverageStr := strings.TrimSuffix(parts[len(parts)-2], "%")
			coverage, err := strconv.ParseFloat(coverageStr, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse coverage percentage: %w", err)
			}

			// Create or update package data
			pkg, ok := packageMap[pkgPath]
			if !ok {
				pkg = &CoverageData{
					Package:   pkgPath,
					Files:     make([]FileCoverage, 0),
					Timestamp: time.Now(),
					Coverage:  coverage,
				}
				packageMap[pkgPath] = pkg
			} else {
				pkg.Coverage = coverage
			}
		} else {
			// Line for a specific function
			// Format: path/to/file.go:line.column	function	coverage%
			// We need to calculate file coverage separately
			// For now, we just count this as a covered line

			// Create or update package data
			pkg, ok := packageMap[pkgPath]
			if !ok {
				pkg = &CoverageData{
					Package:   pkgPath,
					Files:     make([]FileCoverage, 0),
					Timestamp: time.Now(),
				}
				packageMap[pkgPath] = pkg
			}

			// Find or create file coverage
			var fileCov *FileCoverage
			for i, fc := range pkg.Files {
				if fc.Filename == fileName {
					fileCov = &pkg.Files[i]
					break
				}
			}

			if fileCov == nil {
				pkg.Files = append(pkg.Files, FileCoverage{
					Filename:     fileName,
					TotalLines:   1,
					CoveredLines: 0, // We'll update this below
				})
				fileCov = &pkg.Files[len(pkg.Files)-1]
			} else {
				fileCov.TotalLines++
			}

			// Check if this line is covered
			if strings.HasPrefix(parts[len(parts)-1], "100.0%") {
				fileCov.CoveredLines++
				pkg.CoveredLines++
			}
			pkg.TotalLines++
		}
	}

	// Convert map to slice and calculate file coverage percentages
	result := make([]CoverageData, 0, len(packageMap))
	for _, pkg := range packageMap {
		// Calculate file coverage percentages
		for i := range pkg.Files {
			if pkg.Files[i].TotalLines > 0 {
				pkg.Files[i].Coverage = float64(pkg.Files[i].CoveredLines) * 100.0 / float64(pkg.Files[i].TotalLines)
			}
		}
		result = append(result, *pkg)
	}

	return result, nil
}

// GenerateHTMLCoverageReport generates an HTML coverage report.
func GenerateHTMLCoverageReport(coverageFile, outputFile string) error {
	cmd := exec.Command("go", "tool", "cover", "-html", coverageFile, "-o", outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate HTML coverage report: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// GenerateCoverageReportFiles generates and saves coverage report files.
func GenerateCoverageReportFiles(coverageFile, outputDir string, packages ...string) error {
	// Ensure the coverage file exists
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		return fmt.Errorf("coverage file not found: %s", coverageFile)
	}

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate the coverage report
	report, err := GenerateCoverageReport(coverageFile, packages...)
	if err != nil {
		return err
	}

	// Save the report as JSON
	jsonPath := filepath.Join(outputDir, "coverage_report.json")
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal coverage report to JSON: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	// Save the report as text
	textPath := filepath.Join(outputDir, "coverage_report.txt")
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Coverage Report - Generated at %s\n", report.Timestamp.Format(time.RFC3339)))
	sb.WriteString("=====================================\n\n")
	sb.WriteString(fmt.Sprintf("Total Coverage: %.2f%% (%d of %d lines covered)\n\n",
		report.TotalCoverage, report.CoveredLines, report.TotalLines))

	sb.WriteString("Package Coverage:\n")
	for _, pkg := range report.PackageCoverage {
		sb.WriteString(fmt.Sprintf("  %s: %.2f%% (%d of %d lines)\n",
			pkg.Package, pkg.Coverage, pkg.CoveredLines, pkg.TotalLines))

		sb.WriteString("  Files:\n")
		for _, file := range pkg.Files {
			sb.WriteString(fmt.Sprintf("    %s: %.2f%% (%d of %d lines)\n",
				file.Filename, file.Coverage, file.CoveredLines, file.TotalLines))
		}
		sb.WriteString("\n")
	}

	if err := os.WriteFile(textPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write text report: %w", err)
	}

	// Generate HTML coverage report
	htmlPath := filepath.Join(outputDir, "coverage_report.html")
	if err := GenerateHTMLCoverageReport(coverageFile, htmlPath); err != nil {
		return err
	}

	return nil
}
