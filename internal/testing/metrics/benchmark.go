package metrics

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/olekukonko/tablewriter"
)

// SimpleBenchmarkResult represents the result of a simple benchmark operation
type SimpleBenchmarkResult struct {
	Name      string
	Operation string
	Duration  time.Duration
	MemAllocs int64
	MemBytes  int64
}

// BenchmarkSuite manages a collection of simple benchmark results
type BenchmarkSuite struct {
	Results []SimpleBenchmarkResult
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		Results: make([]SimpleBenchmarkResult, 0),
	}
}

// AddResult adds a new benchmark result to the suite
func (s *BenchmarkSuite) AddResult(name, operation string, duration time.Duration, allocs, bytes int64) {
	s.Results = append(s.Results, SimpleBenchmarkResult{
		Name:      name,
		Operation: operation,
		Duration:  duration,
		MemAllocs: allocs,
		MemBytes:  bytes,
	})
}

// PrintResults prints the benchmark results in a formatted table
func (s *BenchmarkSuite) PrintResults() {
	if len(s.Results) == 0 {
		fmt.Println("No benchmark results to display")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	// Set headers
	headers := []string{"Test", "Operation", "Duration", "Allocs", "Bytes"}
	table.SetHeader(headers)

	// Configure table appearance
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("\t")
	table.SetRowSeparator("")

	for _, result := range s.Results {
		table.Append([]string{
			result.Name,
			result.Operation,
			result.Duration.String(),
			fmt.Sprintf("%d", result.MemAllocs),
			fmt.Sprintf("%.2f KB", float64(result.MemBytes)/1024),
		})
	}

	table.Render()
}

// BenchmarkOperation runs a function and records its performance
func BenchmarkOperation(b *testing.B, name string, fn func()) {
	b.Helper() // Mark this as a helper function to not affect line reporting

	// Run the function b.N times
	b.ResetTimer() // Reset timer to ignore setup time
	for i := 0; i < b.N; i++ {
		fn()
	}
}

// TimeOperation measures execution time of a function
func TimeOperation(fn func() error) (time.Duration, error) {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	return duration, err
}
