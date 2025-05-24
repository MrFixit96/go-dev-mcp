// Package metrics provides report generation utilities for MCP tests.
package metrics

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Report represents a comprehensive test report.
type Report struct {
	Title       string            `json:"title"`
	GeneratedAt time.Time         `json:"generatedAt"`
	TestResults []TestResult      `json:"testResults"`
	Coverage    *CoverageReport   `json:"coverage,omitempty"`
	Benchmarks  []BenchmarkResult `json:"benchmarks,omitempty"`
	Summary     ReportSummary     `json:"summary"`
}

// ReportSummary contains overall statistics for the report.
type ReportSummary struct {
	TotalTests     int           `json:"totalTests"`
	PassedTests    int           `json:"passedTests"`
	FailedTests    int           `json:"failedTests"`
	SkippedTests   int           `json:"skippedTests"`
	PassPercentage float64       `json:"passPercentage"`
	TotalDuration  time.Duration `json:"totalDuration"`
	TotalCoverage  float64       `json:"totalCoverage,omitempty"`
}

// GenerateReport creates a comprehensive report from test results, coverage data, and benchmark results.
func GenerateReport(title string, testResults []TestResult, coverage *CoverageReport, benchmarks []BenchmarkResult) *Report {
	// Calculate summary statistics
	totalTests := len(testResults)
	passedTests := 0
	failedTests := 0
	skippedTests := 0
	var totalDuration time.Duration

	for _, result := range testResults {
		if result.Skipped {
			skippedTests++
		} else if result.Success {
			passedTests++
		} else {
			failedTests++
		}
		totalDuration += result.Duration
	}

	passPercentage := 0.0
	if totalTests-skippedTests > 0 {
		passPercentage = float64(passedTests) / float64(totalTests-skippedTests) * 100
	}

	totalCoverage := 0.0
	if coverage != nil {
		totalCoverage = coverage.TotalCoverage
	}

	summary := ReportSummary{
		TotalTests:     totalTests,
		PassedTests:    passedTests,
		FailedTests:    failedTests,
		SkippedTests:   skippedTests,
		PassPercentage: passPercentage,
		TotalDuration:  totalDuration,
		TotalCoverage:  totalCoverage,
	}

	return &Report{
		Title:       title,
		GeneratedAt: time.Now(),
		TestResults: testResults,
		Coverage:    coverage,
		Benchmarks:  benchmarks,
		Summary:     summary,
	}
}

// SaveAsJSON saves the report as a JSON file.
func (r *Report) SaveAsJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// SaveAsText saves the report as a plain text file.
func (r *Report) SaveAsText(path string) error {
	var b strings.Builder

	// Title and timestamp
	b.WriteString(fmt.Sprintf("%s\n", r.Title))
	b.WriteString(strings.Repeat("=", len(r.Title)))
	b.WriteString(fmt.Sprintf("\nGenerated at: %s\n\n", r.GeneratedAt.Format(time.RFC3339)))

	// Summary
	b.WriteString("Summary\n-------\n")
	b.WriteString(fmt.Sprintf("Total Tests: %d\n", r.Summary.TotalTests))
	b.WriteString(fmt.Sprintf("Passed: %d (%.1f%%)\n", r.Summary.PassedTests, r.Summary.PassPercentage))
	b.WriteString(fmt.Sprintf("Failed: %d\n", r.Summary.FailedTests))
	b.WriteString(fmt.Sprintf("Skipped: %d\n", r.Summary.SkippedTests))
	b.WriteString(fmt.Sprintf("Total Duration: %.2f seconds\n", r.Summary.TotalDuration.Seconds()))

	if r.Summary.TotalCoverage > 0 {
		b.WriteString(fmt.Sprintf("Total Coverage: %.2f%%\n", r.Summary.TotalCoverage))
	}
	b.WriteString("\n")

	// Test Results
	b.WriteString("Test Results\n------------\n")
	if len(r.TestResults) == 0 {
		b.WriteString("No test results available.\n\n")
	} else {
		// Group by package
		packageResults := make(map[string][]TestResult)
		for _, result := range r.TestResults {
			packageResults[result.Package] = append(packageResults[result.Package], result)
		}

		// Sort packages
		packages := make([]string, 0, len(packageResults))
		for pkg := range packageResults {
			packages = append(packages, pkg)
		}
		sort.Strings(packages)

		for _, pkg := range packages {
			b.WriteString(fmt.Sprintf("\nPackage: %s\n", pkg))

			results := packageResults[pkg]
			sort.Slice(results, func(i, j int) bool {
				return results[i].Name < results[j].Name
			})

			for _, result := range results {
				status := "PASS"
				if result.Skipped {
					status = "SKIP"
				} else if !result.Success {
					status = "FAIL"
				}

				b.WriteString(fmt.Sprintf("  [%s] %s (%.2fs)\n", status, result.Name, result.Duration.Seconds()))

				if !result.Success && result.ErrorMsg != "" {
					b.WriteString(fmt.Sprintf("    Error: %s\n", result.ErrorMsg))
				}

				if result.Skipped && result.SkipReason != "" {
					b.WriteString(fmt.Sprintf("    Reason: %s\n", result.SkipReason))
				}
			}
		}
	}
	b.WriteString("\n")

	// Coverage Report
	if r.Coverage != nil {
		b.WriteString("Coverage Report\n---------------\n")
		b.WriteString(fmt.Sprintf("Total Coverage: %.2f%% (%d of %d lines covered)\n\n",
			r.Coverage.TotalCoverage, r.Coverage.CoveredLines, r.Coverage.TotalLines))

		b.WriteString("Package Coverage:\n")
		for _, pkg := range r.Coverage.PackageCoverage {
			b.WriteString(fmt.Sprintf("  %s: %.2f%% (%d of %d lines)\n",
				pkg.Package, pkg.Coverage, pkg.CoveredLines, pkg.TotalLines))
		}
	} else {
		b.WriteString("Coverage Report\n---------------\n")
		b.WriteString("No coverage data available.\n")
	}
	b.WriteString("\n")

	// Benchmark Results
	if len(r.Benchmarks) > 0 {
		b.WriteString("Benchmark Results\n-----------------\n")
		for _, benchmark := range r.Benchmarks {
			b.WriteString(fmt.Sprintf("\nBenchmark: %s\n", benchmark.Name))
			b.WriteString(fmt.Sprintf("Duration: %.3f seconds\n", benchmark.Duration.Seconds()))

			for _, metric := range benchmark.Metrics {
				opsPerSec := float64(metric.Operations) / metric.Duration.Seconds()
				b.WriteString(fmt.Sprintf("  %s: %.2f ops/sec", metric.Name, opsPerSec))

				if metric.BytesPerOp > 0 {
					b.WriteString(fmt.Sprintf(", %d B/op", metric.BytesPerOp))
				}

				if metric.AllocsPerOp > 0 {
					b.WriteString(fmt.Sprintf(", %d allocs/op", metric.AllocsPerOp))
				}

				b.WriteString("\n")
			}
		}
	} else {
		b.WriteString("Benchmark Results\n-----------------\n")
		b.WriteString("No benchmark results available.\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

// SaveAsHTML saves the report as an HTML file.
func (r *Report) SaveAsHTML(path string) error {
	// Basic HTML template
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; line-height: 1.6; }
        h1, h2, h3 { color: #333; }
        .summary { background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .pass { color: green; }
        .fail { color: red; }
        .skip { color: orange; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20px; }
        th, td { text-align: left; padding: 12px; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .error-msg { color: red; font-family: monospace; white-space: pre-wrap; }
        .coverage-bar { height: 20px; background-color: #eee; border-radius: 3px; overflow: hidden; margin-bottom: 5px; }
        .coverage-fill { height: 100%; background-color: #4CAF50; }
        .benchmark { margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>Generated at: {{.GeneratedAt.Format "Jan 02, 2006 15:04:05 MST"}}</p>
    
    <div class="summary">
        <h2>Summary</h2>
        <p>Total Tests: {{.Summary.TotalTests}}</p>
        <p>Passed: <span class="pass">{{.Summary.PassedTests}} ({{printf "%.1f" .Summary.PassPercentage}}%)</span></p>
        <p>Failed: <span class="fail">{{.Summary.FailedTests}}</span></p>
        <p>Skipped: <span class="skip">{{.Summary.SkippedTests}}</span></p>
        <p>Total Duration: {{printf "%.2f" .Summary.TotalDuration.Seconds}} seconds</p>
        {{if gt .Summary.TotalCoverage 0.0}}
        <p>Total Coverage: {{printf "%.2f" .Summary.TotalCoverage}}%</p>
        <div class="coverage-bar">
            <div class="coverage-fill" style="width: {{printf "%.2f" .Summary.TotalCoverage}}%;"></div>
        </div>
        {{end}}
    </div>
    
    <h2>Test Results</h2>
    {{if eq (len .TestResults) 0}}
    <p>No test results available.</p>
    {{else}}
    <table>
        <tr>
            <th>Status</th>
            <th>Package</th>
            <th>Test</th>
            <th>Duration</th>
            <th>Details</th>
        </tr>
        {{range .TestResults}}
        <tr>
            <td>
            {{if .Skipped}}<span class="skip">SKIP</span>
            {{else if .Success}}<span class="pass">PASS</span>
            {{else}}<span class="fail">FAIL</span>
            {{end}}
            </td>
            <td>{{.Package}}</td>
            <td>{{.Name}}</td>
            <td>{{printf "%.2f" .Duration.Seconds}}s</td>
            <td>
                {{if and (not .Success) .ErrorMsg}}<div class="error-msg">{{.ErrorMsg}}</div>{{end}}
                {{if and .Skipped .SkipReason}}<div class="skip">{{.SkipReason}}</div>{{end}}
            </td>
        </tr>
        {{end}}
    </table>
    {{end}}
    
    {{if .Coverage}}
    <h2>Coverage Report</h2>
    <p>Total Coverage: {{printf "%.2f" .Coverage.TotalCoverage}}% ({{.Coverage.CoveredLines}} of {{.Coverage.TotalLines}} lines covered)</p>
    <table>
        <tr>
            <th>Package</th>
            <th>Coverage</th>
            <th>Lines Covered</th>
            <th>Total Lines</th>
        </tr>
        {{range .Coverage.PackageCoverage}}
        <tr>
            <td>{{.Package}}</td>
            <td>
                <div class="coverage-bar">
                    <div class="coverage-fill" style="width: {{printf "%.2f" .Coverage}}%;"></div>
                </div>
                {{printf "%.2f" .Coverage}}%
            </td>
            <td>{{.CoveredLines}}</td>
            <td>{{.TotalLines}}</td>
        </tr>
        {{end}}
    </table>
    {{else}}
    <h2>Coverage Report</h2>
    <p>No coverage data available.</p>
    {{end}}
    
    {{if gt (len .Benchmarks) 0}}
    <h2>Benchmark Results</h2>
    {{range .Benchmarks}}
    <div class="benchmark">
        <h3>{{.Name}}</h3>
        <p>Duration: {{printf "%.3f" .Duration.Seconds}} seconds</p>
        <table>
            <tr>
                <th>Metric</th>
                <th>Operations/sec</th>
                <th>Bytes/op</th>
                <th>Allocs/op</th>
            </tr>
            {{range .Metrics}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{printf "%.2f" (div .Operations .Duration.Seconds)}}</td>
                <td>{{if gt .BytesPerOp 0}}{{.BytesPerOp}}{{else}}-{{end}}</td>
                <td>{{if gt .AllocsPerOp 0}}{{.AllocsPerOp}}{{else}}-{{end}}</td>
            </tr>
            {{end}}
        </table>
    </div>
    {{end}}
    {{else}}
    <h2>Benchmark Results</h2>
    <p>No benchmark results available.</p>
    {{end}}
</body>
</html>`

	// Create custom template functions
	funcMap := template.FuncMap{
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}

	// Parse template with function map
	t := template.Must(template.New("report").Funcs(funcMap).Parse(tmpl))

	// Create file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer f.Close()

	// Execute template
	if err := t.Execute(f, r); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// GenerateReports creates and saves comprehensive test reports from results in the specified directory.
func GenerateReports(reportTitle string, testResultsDir, coverageFile, benchmarkResultsDir, outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load test results
	testResults := make([]TestResult, 0)
	if testResultsDir != "" {
		entries, err := os.ReadDir(testResultsDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
					path := filepath.Join(testResultsDir, entry.Name())
					data, err := os.ReadFile(path)
					if err != nil {
						continue
					}

					var result TestResult
					if err := json.Unmarshal(data, &result); err != nil {
						continue
					}

					testResults = append(testResults, result)
				}
			}
		}
	}

	// Load coverage report
	var coverage *CoverageReport
	if coverageFile != "" && fileExists(coverageFile) {
		report, err := GenerateCoverageReport(coverageFile)
		if err == nil {
			coverage = report
		}
	}

	// Load benchmark results
	benchmarks := make([]BenchmarkResult, 0)
	if benchmarkResultsDir != "" {
		entries, err := os.ReadDir(benchmarkResultsDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
					path := filepath.Join(benchmarkResultsDir, entry.Name())
					result, err := LoadBenchmarkResult(path)
					if err != nil {
						continue
					}
					benchmarks = append(benchmarks, *result)
				}
			}
		}
	}

	// Generate report
	report := GenerateReport(reportTitle, testResults, coverage, benchmarks)

	// Save reports in different formats
	timestamp := time.Now().Format("20060102-150405")
	baseName := fmt.Sprintf("test_report_%s", timestamp)

	jsonPath := filepath.Join(outputDir, baseName+".json")
	if err := report.SaveAsJSON(jsonPath); err != nil {
		return fmt.Errorf("failed to save JSON report: %w", err)
	}

	textPath := filepath.Join(outputDir, baseName+".txt")
	if err := report.SaveAsText(textPath); err != nil {
		return fmt.Errorf("failed to save text report: %w", err)
	}

	htmlPath := filepath.Join(outputDir, baseName+".html")
	if err := report.SaveAsHTML(htmlPath); err != nil {
		return fmt.Errorf("failed to save HTML report: %w", err)
	}

	return nil
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
