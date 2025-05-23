// Package mcptesting provides parallel execution capabilities for MCP tests.
package mcptesting

import (
	"os"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

var (
	// globalMutex protects initialization of global resources
	globalMutex sync.Mutex

	// maxParallelTests controls the maximum number of parallel tests
	maxParallelTests = getMaxParallelTests()

	// isParallelEnabled indicates if parallel testing is enabled
	isParallelEnabled = true
)

// getMaxParallelTests determines the maximum number of parallel tests
// based on environment variables or system capabilities
func getMaxParallelTests() int {
	// Check if explicitly set in environment
	if envVal := os.Getenv("MCP_TEST_PARALLEL"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil && val > 0 {
			return val
		}
	}

	// Get number of CPUs and set a reasonable default
	numProcs := 4 // Default to 4 if we can't determine
	if n := runtime.NumCPU(); n > 0 {
		numProcs = n
	}

	// Use CPU count - 1 to avoid saturating the system
	// but ensure at least 1
	if numProcs > 1 {
		return numProcs - 1
	}
	return 1
}

// SetParallelEnabled sets whether parallel testing is enabled
func SetParallelEnabled(enabled bool) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	isParallelEnabled = enabled
}

// SetMaxParallelTests sets the maximum number of parallel tests
func SetMaxParallelTests(max int) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	if max > 0 {
		maxParallelTests = max
	}
}

// RunParallel marks a test to run in parallel if enabled
// and sets appropriate parallelism levels
func RunParallel(t *testing.T) {
	globalMutex.Lock()
	parallelEnabled := isParallelEnabled
	parallel := maxParallelTests
	globalMutex.Unlock()

	if parallelEnabled {
		t.Parallel()
		runtime.GOMAXPROCS(parallel)
	}
}
