// Package metrics provides utilities for collecting and reporting test metrics.
package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestResult represents the outcome of a test execution.
type TestResult struct {
	Name       string        `json:"name"`
	Package    string        `json:"package"`
	Success    bool          `json:"success"`
	Duration   time.Duration `json:"duration"`
	Output     string        `json:"output,omitempty"`
	ErrorMsg   string        `json:"errorMessage,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	Skipped    bool          `json:"skipped,omitempty"`
	SkipReason string        `json:"skipReason,omitempty"`
}

// Collector collects metrics about test executions.
type Collector struct {
	Results      []TestResult  `json:"results"`
	StartTime    time.Time     `json:"startTime"`
	EndTime      time.Time     `json:"endTime"`
	TotalTests   int           `json:"totalTests"`
	PassedTests  int           `json:"passedTests"`
	FailedTests  int           `json:"failedTests"`
	SkippedTests int           `json:"skippedTests"`
	TotalTime    time.Duration `json:"totalTime"`
	mu           sync.Mutex
}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{
		Results:   make([]TestResult, 0),
		StartTime: time.Now(),
	}
}

// RecordTestResult records the result of a test execution.
func (c *Collector) RecordTestResult(result TestResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Results = append(c.Results, result)
	c.TotalTests++

	if result.Skipped {
		c.SkippedTests++
	} else if result.Success {
		c.PassedTests++
	} else {
		c.FailedTests++
	}
}

// Start marks the start of the test execution.
func (c *Collector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.StartTime = time.Now()
}

// End marks the end of the test execution.
func (c *Collector) End() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.EndTime = time.Now()
	c.TotalTime = c.EndTime.Sub(c.StartTime)
}

// GenerateReport generates a report of the test metrics.
func (c *Collector) GenerateReport() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Test Execution Report\n"))
	builder.WriteString(fmt.Sprintf("====================\n\n"))
	builder.WriteString(fmt.Sprintf("Start Time: %s\n", c.StartTime.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("End Time: %s\n", c.EndTime.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("Duration: %.2f seconds\n\n", c.TotalTime.Seconds()))

	builder.WriteString(fmt.Sprintf("Summary:\n"))
	builder.WriteString(fmt.Sprintf("  Total Tests: %d\n", c.TotalTests))
	builder.WriteString(fmt.Sprintf("  Passed: %d (%.1f%%)\n", c.PassedTests, percentage(c.PassedTests, c.TotalTests)))
	builder.WriteString(fmt.Sprintf("  Failed: %d (%.1f%%)\n", c.FailedTests, percentage(c.FailedTests, c.TotalTests)))
	builder.WriteString(fmt.Sprintf("  Skipped: %d (%.1f%%)\n\n", c.SkippedTests, percentage(c.SkippedTests, c.TotalTests)))

	// Group results by package
	packages := make(map[string][]TestResult)
	for _, r := range c.Results {
		packages[r.Package] = append(packages[r.Package], r)
	}

	// Sort packages by name
	packageNames := make([]string, 0, len(packages))
	for pkg := range packages {
		packageNames = append(packageNames, pkg)
	}
	sort.Strings(packageNames)

	builder.WriteString(fmt.Sprintf("Results by Package:\n"))
	for _, pkg := range packageNames {
		results := packages[pkg]
		passed := 0
		failed := 0
		skipped := 0

		for _, r := range results {
			if r.Skipped {
				skipped++
			} else if r.Success {
				passed++
			} else {
				failed++
			}
		}

		builder.WriteString(fmt.Sprintf("\n  Package: %s\n", pkg))
		builder.WriteString(fmt.Sprintf("    Tests: %d (Passed: %d, Failed: %d, Skipped: %d)\n", len(results), passed, failed, skipped))

		// Sort test results by name
		sort.Slice(results, func(i, j int) bool {
			return results[i].Name < results[j].Name
		})

		for _, r := range results {
			status := "PASS"
			if r.Skipped {
				status = "SKIP"
			} else if !r.Success {
				status = "FAIL"
			}

			builder.WriteString(fmt.Sprintf("    - %s: %s (%.2fs)\n", status, r.Name, r.Duration.Seconds()))
			if !r.Success && r.ErrorMsg != "" {
				builder.WriteString(fmt.Sprintf("      Error: %s\n", r.ErrorMsg))
			}
			if r.Skipped && r.SkipReason != "" {
				builder.WriteString(fmt.Sprintf("      Reason: %s\n", r.SkipReason))
			}
		}
	}

	return builder.String()
}

// SaveReportToFile saves the test metrics report to a file.
func (c *Collector) SaveReportToFile(path string) error {
	report := c.GenerateReport()
	return os.WriteFile(path, []byte(report), 0644)
}

// SaveJSONToFile saves the test metrics as JSON to a file.
func (c *Collector) SaveJSONToFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// percentage calculates the percentage of part to total.
func percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) * 100 / float64(total)
}

// RunWithMetrics runs a test function and collects metrics about its execution.
func RunWithMetrics(t *testing.T, name string, testFunc func(t *testing.T)) {
	t.Helper()

	collector := NewCollector()
	collector.Start()
	// Create a wrapped testing.T that records the outcome
	wrapped := &recordingT{
		T:        t,
		name:     name,
		package_: getCurrentPackage(),
	}

	startTime := time.Now()

	// Run the test function
	func() {
		defer func() {
			if r := recover(); r != nil {
				wrapped.errorMsg = fmt.Sprintf("panic: %v", r)
				wrapped.success = false
			}
		}()

		testFunc(wrapped.T)
	}()

	wrapped.duration = time.Since(startTime)
	// Record the result
	collector.RecordTestResult(TestResult{
		Name:       wrapped.name,
		Package:    wrapped.package_,
		Success:    wrapped.success,
		Duration:   wrapped.duration,
		Output:     wrapped.output,
		ErrorMsg:   wrapped.errorMsg,
		Timestamp:  time.Now(),
		Skipped:    wrapped.skipped,
		SkipReason: wrapped.skipReason,
	})

	collector.End()

	// Save the report
	reportDir := os.Getenv("MCP_TEST_REPORT_DIR")
	if reportDir == "" {
		reportDir = "test_reports"
	}

	if err := os.MkdirAll(reportDir, 0755); err == nil {
		reportPath := filepath.Join(reportDir, strings.ReplaceAll(name, "/", "_")+".txt")
		jsonPath := filepath.Join(reportDir, strings.ReplaceAll(name, "/", "_")+".json")

		collector.SaveReportToFile(reportPath)
		collector.SaveJSONToFile(jsonPath)
	}
}

// getCurrentPackage returns the current package name.
func getCurrentPackage() string {
	pc, _, _, _ := runtime.Caller(2)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// recordingT is a wrapper around testing.T that records the test outcome.
type recordingT struct {
	*testing.T
	name       string
	package_   string
	success    bool
	skipped    bool
	skipReason string
	duration   time.Duration
	output     string
	errorMsg   string
}

func (r *recordingT) Errorf(format string, args ...interface{}) {
	r.success = false
	msg := fmt.Sprintf(format, args...)
	r.errorMsg = msg
	r.T.Errorf(format, args...)
}

func (r *recordingT) Fatalf(format string, args ...interface{}) {
	r.success = false
	msg := fmt.Sprintf(format, args...)
	r.errorMsg = msg
	r.T.Fatalf(format, args...)
}

func (r *recordingT) FailNow() {
	r.success = false
	r.T.FailNow()
}

func (r *recordingT) Skip(args ...interface{}) {
	r.skipped = true
	r.skipReason = fmt.Sprint(args...)
	r.T.Skip(args...)
}

func (r *recordingT) Skipf(format string, args ...interface{}) {
	r.skipped = true
	r.skipReason = fmt.Sprintf(format, args...)
	r.T.Skipf(format, args...)
}
