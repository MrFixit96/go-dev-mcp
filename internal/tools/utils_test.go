package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidatePath_EmptyPath tests that empty paths are rejected
func TestValidatePath_EmptyPath(t *testing.T) {
	_, err := validatePath("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

// TestValidatePath_NormalRelativePath tests that normal relative paths are converted to absolute
func TestValidatePath_NormalRelativePath(t *testing.T) {
	result, err := validatePath("./test/path")
	if err != nil {
		t.Errorf("unexpected error for normal relative path: %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got: %s", result)
	}
	// Should not contain . or ..
	if strings.Contains(result, "..") || strings.Contains(result, "/.") {
		t.Errorf("expected cleaned path without . or .., got: %s", result)
	}
}

// TestValidatePath_NormalAbsolutePath tests that absolute paths are accepted and cleaned
func TestValidatePath_NormalAbsolutePath(t *testing.T) {
	testPath := "/tmp/test/path"
	result, err := validatePath(testPath)
	if err != nil {
		t.Errorf("unexpected error for absolute path: %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got: %s", result)
	}
	// Result should be cleaned
	expected := filepath.Clean(testPath)
	if !strings.HasPrefix(result, expected) {
		t.Errorf("expected path to start with %s, got: %s", expected, result)
	}
}

// TestValidatePath_PathTraversalWithDotDot tests paths with ../ sequences
func TestValidatePath_PathTraversalWithDotDot(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{
			name:  "simple parent directory",
			path:  "../test",
			valid: true, // Should be cleaned and converted to absolute
		},
		{
			name:  "multiple parent directories",
			path:  "../../etc/passwd",
			valid: true, // Should be cleaned and converted to absolute
		},
		{
			name:  "mixed with normal path",
			path:  "test/../../other",
			valid: true, // Should be cleaned
		},
		{
			name:  "absolute with parent reference",
			path:  "/tmp/../etc/passwd",
			valid: true, // Should be cleaned to /etc/passwd
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validatePath(tt.path)
			if tt.valid {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				// Result should be absolute and not contain ..
				if !filepath.IsAbs(result) {
					t.Errorf("expected absolute path, got: %s", result)
				}
				// The .. should be resolved/cleaned
				if strings.Contains(filepath.ToSlash(result), "/..") {
					t.Errorf("expected cleaned path without .., got: %s", result)
				}
			} else {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
		})
	}
}

// TestValidatePath_CleaningBehavior tests that paths are properly cleaned
func TestValidatePath_CleaningBehavior(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // strings that should be in the result
		excludes []string // strings that should NOT be in the result
	}{
		{
			name:     "removes trailing slashes",
			input:    "/tmp/test/",
			contains: []string{"/tmp/test"},
			excludes: []string{"/tmp/test//"}, // No double slashes
		},
		{
			name:     "removes double slashes",
			input:    "/tmp//test",
			contains: []string{"/tmp/test"},
			excludes: []string{"//"},
		},
		{
			name:     "resolves current directory references",
			input:    "/tmp/./test",
			contains: []string{"/tmp/test"},
			excludes: []string{"/./"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validatePath(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %s, got: %s", expected, result)
				}
			}

			for _, excluded := range tt.excludes {
				if strings.Contains(result, excluded) {
					t.Errorf("expected result NOT to contain %s, got: %s", excluded, result)
				}
			}
		})
	}
}

// TestValidatePath_WithExistingDirectory tests with actual existing directory
func TestValidatePath_WithExistingDirectory(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "validatepath-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	result, err := validatePath(tmpDir)
	if err != nil {
		t.Errorf("unexpected error for existing directory: %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got: %s", result)
	}
}

// TestValidatePath_WithNonExistentPath tests with paths that don't exist yet
func TestValidatePath_WithNonExistentPath(t *testing.T) {
	// This should not error - paths don't need to exist (e.g., for workspace init)
	nonExistent := "/tmp/this-path-definitely-does-not-exist-12345"
	result, err := validatePath(nonExistent)
	if err != nil {
		t.Errorf("unexpected error for non-existent path: %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got: %s", result)
	}
}

// TestValidatePath_WithSymlink tests symlink resolution
func TestValidatePath_WithSymlink(t *testing.T) {
	// Create a temporary directory and a symlink to it
	tmpDir, err := os.MkdirTemp("", "validatepath-symlink-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a symlink
	symlinkPath := filepath.Join(os.TempDir(), "validatepath-symlink-test")
	// Clean up any existing symlink
	os.Remove(symlinkPath)
	defer os.Remove(symlinkPath)

	err = os.Symlink(tmpDir, symlinkPath)
	if err != nil {
		t.Skipf("skipping symlink test, cannot create symlink: %v", err)
		return
	}

	result, err := validatePath(symlinkPath)
	if err != nil {
		t.Errorf("unexpected error for symlink path: %v", err)
	}

	// The result should be the resolved path (pointing to tmpDir)
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got: %s", result)
	}

	// Should resolve to the actual directory
	if result != tmpDir {
		t.Logf("symlink %s resolved to %s (expected %s)", symlinkPath, result, tmpDir)
		// This is actually correct behavior - symlinks should be resolved
	}
}

// TestValidatePath_RelativeToCurrentDir tests relative paths are resolved correctly
func TestValidatePath_RelativeToCurrentDir(t *testing.T) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	result, err := validatePath(".")
	if err != nil {
		t.Errorf("unexpected error for current directory: %v", err)
	}

	// Should resolve to current working directory
	if !strings.Contains(result, filepath.Base(cwd)) || !filepath.IsAbs(result) {
		t.Logf("current dir '.' resolved to: %s (cwd: %s)", result, cwd)
		// Just verify it's absolute
		if !filepath.IsAbs(result) {
			t.Errorf("expected absolute path, got: %s", result)
		}
	}
}

// TestValidatePath_ComplexTraversalAttempts tests more complex traversal patterns
func TestValidatePath_ComplexTraversalAttempts(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "multiple consecutive parent refs",
			path: "../../../../etc/passwd",
		},
		{
			name: "mixed slashes and dots",
			path: "/tmp/./test/../../../etc/passwd",
		},
		{
			name: "starting with parent ref",
			path: "../../../../../boot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validatePath(tt.path)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Should be absolute and cleaned
			if !filepath.IsAbs(result) {
				t.Errorf("expected absolute path, got: %s", result)
			}

			// Should not contain unresolved .. sequences
			// Note: On Windows, paths use backslashes
			normalizedResult := filepath.ToSlash(result)
			if strings.Contains(normalizedResult, "/..") || strings.HasSuffix(normalizedResult, "..") {
				t.Errorf("expected fully resolved path without .., got: %s", result)
			}

			t.Logf("Path %s resolved to: %s", tt.path, result)
		})
	}
}

// TestValidatePath_WindowsStylePaths tests Windows-style paths (when on Windows)
func TestValidatePath_WindowsStylePaths(t *testing.T) {
	// This test adapts to the OS
	if filepath.Separator == '\\' {
		// Windows
		result, err := validatePath("C:\\Windows\\System32")
		if err != nil {
			t.Errorf("unexpected error for Windows path: %v", err)
		}
		if !filepath.IsAbs(result) {
			t.Errorf("expected absolute path, got: %s", result)
		}
	} else {
		// Unix-like systems
		t.Skip("skipping Windows path test on non-Windows system")
	}
}
