// Package mcptesting provides a standardized testing framework for the Go Development MCP Server.
package mcptesting

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// BaseSuite is the foundation test suite for all MCP tests.
// It provides common functionality and test helpers.
type BaseSuite struct {
	suite.Suite
	TempDir string
}

// SetupSuite runs before all tests in the suite and sets up global resources.
func (s *BaseSuite) SetupSuite() {
	// Create a master temp directory for the entire suite
	tempDir, err := os.MkdirTemp("", "mcptest-*")
	if err != nil {
		s.FailNow("Failed to create temp directory for test suite: %v", err)
	}
	s.TempDir = tempDir
}

// TearDownSuite runs after all tests in the suite and cleans up global resources.
func (s *BaseSuite) TearDownSuite() {
	// Clean up the master temp directory if it exists
	if s.TempDir != "" {
		os.RemoveAll(s.TempDir)
	}
}

// NewTempDir creates a new temporary directory for a specific test.
func (s *BaseSuite) NewTempDir(testName string) string {
	testDir := filepath.Join(s.TempDir, testName)
	err := os.MkdirAll(testDir, 0755)
	s.Require().NoError(err, "Failed to create temp directory for test")
	return testDir
}

// RunSuite runs a test suite with the given testing context.
func RunSuite(t *testing.T, s suite.TestingSuite) {
	suite.Run(t, s)
}
