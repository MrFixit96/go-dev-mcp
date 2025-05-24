// Package metrics provides benchmarking utilities for MCP tests.
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

// BenchmarkMetric represents a single benchmark measurement.
type BenchmarkMetric struct {
	Name        string                 `json:"name"`
	Operations  int                    `json:"operations"`
	Duration    time.Duration          `json:"duration"`
	BytesPerOp  int64                  `json:"bytesPerOp,omitempty"`
	AllocsPerOp int64                  `json:"allocsPerOp,omitempty"`
	MemoryUsage int64                  `json:"memoryUsage,omitempty"`
	CPUUsage    float64                `json:"cpuUsage,omitempty"`
	Parallelism int                    `json:"parallelism,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// BenchmarkResult represents a complete benchmark result with multiple metrics.
type BenchmarkResult struct {
	Name       string            `json:"name"`
	Metrics    []BenchmarkMetric `json:"metrics"`
	StartTime  time.Time         `json:"startTime"`
	EndTime    time.Time         `json:"endTime"`
	Duration   time.Duration     `json:"duration"`
	SystemInfo SystemInfo        `json:"systemInfo"`
}

// SystemInfo contains information about the system on which the benchmark was run.
type SystemInfo struct {
	GoVersion string `json:"goVersion"`
	GOOS      string `json:"goos"`
	GOARCH    string `json:"goarch"`
	NumCPU    int    `json:"numCPU"`
	Compiler  string `json:"compiler"`
	BuildTags string `json:"buildTags,omitempty"`
}

// BenchmarkCollector collects benchmark metrics.
type BenchmarkCollector struct {
	mu       sync.Mutex
	results  map[string]*BenchmarkResult
	current  *BenchmarkResult
	basePath string
}

// NewBenchmarkCollector creates a new benchmark collector.
func NewBenchmarkCollector(basePath string) *BenchmarkCollector {
	// If basePath is empty, use a default
	if basePath == "" {
		basePath = "benchmark_results"
	}

	return &BenchmarkCollector{
		results:  make(map[string]*BenchmarkResult),
		basePath: basePath,
	}
}

// StartBenchmark starts a new benchmark with the given name.
func (c *BenchmarkCollector) StartBenchmark(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create system info
	sysInfo := SystemInfo{
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		NumCPU:    runtime.NumCPU(),
		Compiler:  runtime.Compiler,
	}

	// Get build tags from environment
	if tags := os.Getenv("GOFLAGS"); tags != "" {
		sysInfo.BuildTags = tags
	}

	c.current = &BenchmarkResult{
		Name:       name,
		Metrics:    make([]BenchmarkMetric, 0),
		StartTime:  time.Now(),
		SystemInfo: sysInfo,
	}

	c.results[name] = c.current
}

// RecordMetric records a benchmark metric.
func (c *BenchmarkCollector) RecordMetric(metric BenchmarkMetric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		panic("No benchmark in progress")
	}

	// Set timestamp if not already set
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	c.current.Metrics = append(c.current.Metrics, metric)
}

// EndBenchmark ends the current benchmark.
func (c *BenchmarkCollector) EndBenchmark() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		panic("No benchmark in progress")
	}

	c.current.EndTime = time.Now()
	c.current.Duration = c.current.EndTime.Sub(c.current.StartTime)
}

// SaveResults saves all benchmark results to the base path.
func (c *BenchmarkCollector) SaveResults() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.MkdirAll(c.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create benchmark results directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")

	for name, result := range c.results {
		// Sanitize name for use in filename
		safeName := strings.ReplaceAll(name, " ", "_")
		safeName = strings.ReplaceAll(safeName, "/", "_")
		safeName = strings.ReplaceAll(safeName, "\\", "_")

		// Save result as JSON
		jsonPath := filepath.Join(c.basePath, fmt.Sprintf("%s_%s.json", safeName, timestamp))
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal benchmark result to JSON: %w", err)
		}

		if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write benchmark result: %w", err)
		}

		// Save result as text
		textPath := filepath.Join(c.basePath, fmt.Sprintf("%s_%s.txt", safeName, timestamp))
		textContent := c.formatResultAsText(result)

		if err := os.WriteFile(textPath, []byte(textContent), 0644); err != nil {
			return fmt.Errorf("failed to write benchmark result text: %w", err)
		}
	}

	return nil
}

// formatResultAsText formats a benchmark result as human-readable text.
func (c *BenchmarkCollector) formatResultAsText(result *BenchmarkResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Benchmark: %s\n", result.Name))
	b.WriteString(fmt.Sprintf("Start Time: %s\n", result.StartTime.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("End Time: %s\n", result.EndTime.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Duration: %.3f seconds\n\n", result.Duration.Seconds()))

	b.WriteString("System Information:\n")
	b.WriteString(fmt.Sprintf("  Go Version: %s\n", result.SystemInfo.GoVersion))
	b.WriteString(fmt.Sprintf("  OS/Arch: %s/%s\n", result.SystemInfo.GOOS, result.SystemInfo.GOARCH))
	b.WriteString(fmt.Sprintf("  CPUs: %d\n", result.SystemInfo.NumCPU))
	b.WriteString(fmt.Sprintf("  Compiler: %s\n", result.SystemInfo.Compiler))
	if result.SystemInfo.BuildTags != "" {
		b.WriteString(fmt.Sprintf("  Build Tags: %s\n", result.SystemInfo.BuildTags))
	}
	b.WriteString("\n")

	b.WriteString("Metrics:\n")

	// Sort metrics by name for consistent output
	metrics := make([]BenchmarkMetric, len(result.Metrics))
	copy(metrics, result.Metrics)
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	for _, metric := range metrics {
		b.WriteString(fmt.Sprintf("  %s:\n", metric.Name))
		b.WriteString(fmt.Sprintf("    Operations: %d\n", metric.Operations))
		b.WriteString(fmt.Sprintf("    Duration: %.3f seconds\n", metric.Duration.Seconds()))
		b.WriteString(fmt.Sprintf("    Ops/sec: %.2f\n", float64(metric.Operations)/metric.Duration.Seconds()))

		if metric.BytesPerOp > 0 {
			b.WriteString(fmt.Sprintf("    Bytes/op: %d\n", metric.BytesPerOp))
		}

		if metric.AllocsPerOp > 0 {
			b.WriteString(fmt.Sprintf("    Allocs/op: %d\n", metric.AllocsPerOp))
		}

		if metric.MemoryUsage > 0 {
			b.WriteString(fmt.Sprintf("    Memory Usage: %d bytes\n", metric.MemoryUsage))
		}

		if metric.CPUUsage > 0 {
			b.WriteString(fmt.Sprintf("    CPU Usage: %.2f%%\n", metric.CPUUsage))
		}

		if metric.Parallelism > 0 {
			b.WriteString(fmt.Sprintf("    Parallelism: %d\n", metric.Parallelism))
		}

		if len(metric.Extra) > 0 {
			b.WriteString("    Extra:\n")
			for k, v := range metric.Extra {
				b.WriteString(fmt.Sprintf("      %s: %v\n", k, v))
			}
		}
	}

	return b.String()
}

// RunBenchmark runs a benchmark function and records its results.
func RunBenchmark(b *testing.B, name string, fn func(b *testing.B)) {
	// Create collector with environment-based path
	basePath := os.Getenv("MCP_BENCHMARK_DIR")
	if basePath == "" {
		basePath = "benchmark_results"
	}

	collector := NewBenchmarkCollector(basePath)
	collector.StartBenchmark(name)

	// Run the benchmark
	b.ResetTimer()
	fn(b)
	// Record the results
	metric := BenchmarkMetric{
		Name:        name,
		Operations:  b.N,
		Duration:    b.Elapsed(),
		BytesPerOp:  0, // Will be filled by testing framework if available
		AllocsPerOp: 0, // Will be filled by testing framework if available
		Timestamp:   time.Now(),
	}

	collector.RecordMetric(metric)
	collector.EndBenchmark()

	// Save results
	if err := collector.SaveResults(); err != nil {
		b.Logf("Failed to save benchmark results: %v", err)
	}
}

// CompareBenchmarkResults compares two benchmark results and returns a report of the differences.
func CompareBenchmarkResults(baseline, current *BenchmarkResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Benchmark Comparison: %s\n", current.Name))
	b.WriteString("====================================\n\n")

	b.WriteString(fmt.Sprintf("Baseline: %s\n", baseline.StartTime.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Current:  %s\n\n", current.StartTime.Format(time.RFC3339)))

	// Map metrics by name for easy lookup
	baselineMetrics := make(map[string]BenchmarkMetric)
	for _, metric := range baseline.Metrics {
		baselineMetrics[metric.Name] = metric
	}

	// Compare metrics
	for _, currentMetric := range current.Metrics {
		b.WriteString(fmt.Sprintf("Metric: %s\n", currentMetric.Name))

		baselineMetric, found := baselineMetrics[currentMetric.Name]
		if !found {
			b.WriteString("  No baseline data available for comparison\n\n")
			continue
		}

		// Compare operations per second
		baselineOpsPerSec := float64(baselineMetric.Operations) / baselineMetric.Duration.Seconds()
		currentOpsPerSec := float64(currentMetric.Operations) / currentMetric.Duration.Seconds()
		opsPerSecDiff := (currentOpsPerSec - baselineOpsPerSec) / baselineOpsPerSec * 100

		b.WriteString(fmt.Sprintf("  Ops/sec: %.2f -> %.2f (%.2f%%)\n",
			baselineOpsPerSec, currentOpsPerSec, opsPerSecDiff))

		// Compare bytes per op if available
		if baselineMetric.BytesPerOp > 0 && currentMetric.BytesPerOp > 0 {
			bytesDiff := float64(currentMetric.BytesPerOp-baselineMetric.BytesPerOp) / float64(baselineMetric.BytesPerOp) * 100
			b.WriteString(fmt.Sprintf("  Bytes/op: %d -> %d (%.2f%%)\n",
				baselineMetric.BytesPerOp, currentMetric.BytesPerOp, bytesDiff))
		}

		// Compare allocs per op if available
		if baselineMetric.AllocsPerOp > 0 && currentMetric.AllocsPerOp > 0 {
			allocsDiff := float64(currentMetric.AllocsPerOp-baselineMetric.AllocsPerOp) / float64(baselineMetric.AllocsPerOp) * 100
			b.WriteString(fmt.Sprintf("  Allocs/op: %d -> %d (%.2f%%)\n",
				baselineMetric.AllocsPerOp, currentMetric.AllocsPerOp, allocsDiff))
		}

		b.WriteString("\n")
	}

	return b.String()
}

// LoadBenchmarkResult loads a benchmark result from a file.
func LoadBenchmarkResult(path string) (*BenchmarkResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result BenchmarkResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RunBenchmarkSimple runs a simple benchmark test with testing.T interface for compatibility
// This provides backward compatibility with the simpler test interface
func RunBenchmarkSimple(t *testing.T, name string, fn func()) {
	t.Helper() // Mark this as a helper function to not affect line reporting

	// Setup benchmark
	benchmarkName := fmt.Sprintf("Benchmark%s", name)

	t.Run(benchmarkName, func(t *testing.T) {
		// Skip in short mode
		if testing.Short() {
			t.Skip("Skipping benchmark in short mode")
		}

		start := time.Now()
		fn()
		duration := time.Since(start)

		t.Logf("%s took %s", benchmarkName, duration)
	})
}
