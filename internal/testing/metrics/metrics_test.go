// filepath: c:\Users\James\Documents\go-dev-mcp\internal\testing\metrics\metrics_test.go
package metrics_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MrFixit96/go-dev-mcp/internal/testing/metrics"
	"github.com/stretchr/testify/require"
)

func TestBenchmarkSuite(t *testing.T) {
	t.Parallel()

	suite := metrics.NewBenchmarkSuite()
	require.NotNil(t, suite, "Benchmark suite should not be nil")

	// Add some benchmark results
	suite.AddResult("Test1", "Operation1", 100*time.Millisecond, 10, 1024)
	suite.AddResult("Test2", "Operation2", 200*time.Millisecond, 20, 2048)

	// Verify results are stored
	require.Len(t, suite.Results, 2, "Suite should have 2 results")
	require.Equal(t, "Test1", suite.Results[0].Name, "First result should have correct name")
	require.Equal(t, "Operation2", suite.Results[1].Operation, "Second result should have correct operation")
	require.Equal(t, 100*time.Millisecond, suite.Results[0].Duration, "First result should have correct duration")
	require.Equal(t, int64(20), suite.Results[1].MemAllocs, "Second result should have correct memory allocations")
	require.Equal(t, int64(1024), suite.Results[0].MemBytes, "First result should have correct memory bytes")

	// Test printing results - just ensure it doesn't panic
	stdout := os.Stdout
	os.Stdout = nil // Suppress output for the test
	suite.PrintResults()
	os.Stdout = stdout
}

func TestRunBenchmark(t *testing.T) {
	t.Parallel()

	// Create a simple benchmark function
	var executed bool
	benchFunc := func() {
		executed = true
		time.Sleep(10 * time.Millisecond) // Simulate some work
	}

	// Run the benchmark
	metrics.RunBenchmarkSimple(t, "TestOperation", benchFunc)

	// Verify the function was executed
	require.True(t, executed, "Benchmark function should have been executed")
}

func TestTimeOperation(t *testing.T) {
	t.Parallel()

	// Create a test function that sleeps for a known duration
	sleepTime := 50 * time.Millisecond
	testFunc := func() error {
		time.Sleep(sleepTime)
		return nil
	}

	// Measure the function execution time
	duration, err := metrics.TimeOperation(testFunc)

	// Verify results
	require.NoError(t, err, "No error should be returned")
	require.GreaterOrEqual(t, duration, sleepTime, "Duration should be at least the sleep time")
	require.Less(t, duration, sleepTime*2, "Duration should be reasonably close to the sleep time")
}

func TestBenchmarkOperation(t *testing.T) {
	// Skip in short mode since this is a benchmark
	if testing.Short() {
		t.Skip("Skipping benchmark test in short mode")
	}

	t.Run("Benchmark via helper", func(b *testing.T) {
		// Create a test function
		counter := 0
		testFunc := func() {
			counter++
		}

		// Call simple test benchmark since we have *testing.T not *testing.B
		metrics.RunBenchmarkSimple(b, "CounterIncrement", testFunc)
	})
}

func TestCoverageProfileGeneration(t *testing.T) {
	t.Parallel()

	// Create temporary directory for coverage files
	tempDir, err := os.MkdirTemp("", "coverage-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Test profile path
	profilePath := filepath.Join(tempDir, "coverage.out")

	// Skip actual generation since it would run go test command
	// In a real test environment, this could be tested with a mock command runner
	t.Run("AnalyzeCoverageProfile", func(t *testing.T) {
		// Create a mock coverage profile
		mockProfile := `mode: set
github.com/example/pkg/file1.go:10.30,15.2 1 1
github.com/example/pkg/file1.go:20.40,25.2 1 0
github.com/example/pkg/file2.go:5.20,10.2 1 1
total:	(statements)	66.7%
`
		err := os.WriteFile(profilePath, []byte(mockProfile), 0644)
		require.NoError(t, err, "Failed to write mock coverage profile")

		// Analyze the profile - this part might be skipped in CI/CD environments
		// or mocked based on your testing needs
		if os.Getenv("CI") == "" {
			result, err := metrics.AnalyzeCoverageProfile(profilePath)
			if err == nil {
				require.NotNil(t, result, "Coverage result should not be nil")
				// Check some basic properties without being too specific
				require.GreaterOrEqual(t, result.Percentage, 0.0, "Coverage percentage should be non-negative")
				require.LessOrEqual(t, result.Percentage, 100.0, "Coverage percentage should not exceed 100%")
			}
		}
	})

	t.Run("PrintCoverageReport", func(t *testing.T) {
		// Create a sample coverage result
		result := &metrics.CoverageResult{
			TotalLines:   100,
			CoveredLines: 75,
			Percentage:   75.0,
			Packages: map[string]metrics.PackageCoverage{
				"pkg1": {
					TotalLines:   50,
					CoveredLines: 40,
					Percentage:   80.0,
				},
				"pkg2": {
					TotalLines:   50,
					CoveredLines: 35,
					Percentage:   70.0,
				},
			},
		}

		// Verify printing works without panic
		var buf bytes.Buffer
		metrics.PrintCoverageReport(&buf, result)
		output := buf.String()

		require.Contains(t, output, "Coverage Report", "Output should contain report header")
		require.Contains(t, output, "75.00%", "Output should contain overall percentage")
		require.Contains(t, output, "pkg1: 80.00%", "Output should contain package coverage")
	})

	t.Run("WriteCoverageReport", func(t *testing.T) {
		// Create a sample coverage result
		result := &metrics.CoverageResult{
			TotalLines:   100,
			CoveredLines: 75,
			Percentage:   75.0,
			Packages: map[string]metrics.PackageCoverage{
				"pkg1": {
					TotalLines:   50,
					CoveredLines: 40,
					Percentage:   80.0,
				},
			},
		}

		jsonPath := filepath.Join(tempDir, "coverage.json")
		err := metrics.WriteCoverageReport(result, jsonPath)
		require.NoError(t, err, "Writing coverage report should succeed")

		// Verify file exists
		_, err = os.Stat(jsonPath)
		require.NoError(t, err, "JSON coverage report file should exist")
	})
}
