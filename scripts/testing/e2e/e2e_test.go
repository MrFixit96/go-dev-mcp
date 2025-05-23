// Package e2e provides end-to-end tests for the MCP server.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MrFixit96/go-dev-mcp/internal/testing/mock"
)

// TestEnvironment represents the environment for E2E tests
type TestEnvironment struct {
	// ServerURL is the URL of the MCP server to test against
	ServerURL string
	// TempDir is the directory for temporary test files
	TempDir string
	// MockServer is the mock server for testing (nil when testing real server)
	MockServer *mock.MockServer
	// DeleteTempFiles indicates whether to clean up temporary files after tests
	DeleteTempFiles bool
}

// Setup initializes the test environment
func Setup(t *testing.T, useMockServer bool) *TestEnvironment {
	t.Helper()

	// Create a temp directory for test files
	tempDir, err := os.MkdirTemp("", "go-dev-mcp-test-")
	require.NoError(t, err, "Failed to create temp directory")

	// Determine server URL
	var serverURL string
	var mockServer *mock.MockServer

	if useMockServer {
		// Use mock server
		mockServer = mock.NewMockServer()
		serverURL = mockServer.URL()

		// Set up default handlers
		mockServer.AddToolHandler("go_fmt", mock.DefaultGoFmtHandler)
		mockServer.AddToolHandler("go_build", mock.DefaultGoBuildHandler)
		mockServer.AddToolHandler("go_run", mock.DefaultGoRunHandler)
		mockServer.AddToolHandler("go_test", mock.DefaultGoTestHandler)
		mockServer.AddToolHandler("go_mod", mock.DefaultGoModHandler)
	} else {
		// Use real server from environment variable or default
		serverURL = os.Getenv("MCP_SERVER_URL")
		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}

		// Check if server is available
		if !isServerAvailable(serverURL) {
			t.Skip("Server not available at ", serverURL)
		}
	}

	return &TestEnvironment{
		ServerURL:       serverURL,
		TempDir:         tempDir,
		MockServer:      mockServer,
		DeleteTempFiles: true,
	}
}

// Teardown cleans up the test environment
func (env *TestEnvironment) Teardown(t *testing.T) {
	t.Helper()

	// Close mock server if used
	if env.MockServer != nil {
		env.MockServer.Close()
	}

	// Remove temp directory if cleanup is enabled
	if env.DeleteTempFiles {
		err := os.RemoveAll(env.TempDir)
		assert.NoError(t, err, "Failed to remove temp directory")
	}
}

// CreateTestProject creates a simple Go project in the temp directory
func (env *TestEnvironment) CreateTestProject(t *testing.T) {
	t.Helper()

	// Create main.go
	helloWorldCode := `package main

import "fmt"

func main() {
	// A simple hello world program
	fmt.Println("Hello, World from Go Development MCP Server!")
}
`
	mainPath := filepath.Join(env.TempDir, "main.go")
	err := os.WriteFile(mainPath, []byte(helloWorldCode), 0644)
	require.NoError(t, err, "Failed to create main.go")

	// Create go.mod
	cmd := execCommand("go", "mod", "init", "example.com/hello")
	cmd.Dir = env.TempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to initialize Go module: %s", string(output))
}

// CallMCPTool calls an MCP tool and returns the result
func (env *TestEnvironment) CallMCPTool(toolName string, params map[string]interface{}) (*mcp.CallToolResult, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"params": map[string]interface{}{
			"name":  toolName,
			"input": params,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	resp, err := http.Post(
		fmt.Sprintf("%s/calltool", env.ServerURL),
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status code %d: %s", resp.StatusCode, string(respBody))
	}

	// Debug: Print raw response
	fmt.Printf("DEBUG: Raw server response: %s\n", string(respBody))

	// Parse response
	var result mcp.CallToolResult
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w\nRaw response: %s", err, string(respBody))
	}

	return &result, nil
}

// TestEndToEnd tests the MCP server end-to-end with all tools
func TestEndToEnd(t *testing.T) {
	// Run tests against both mock and real server if available
	runWithMock := true
	runWithReal := os.Getenv("RUN_WITH_REAL_SERVER") == "1"

	if runWithMock {
		t.Run("WithMockServer", func(t *testing.T) {
			runEndToEndTests(t, true)
		})
	}

	if runWithReal {
		t.Run("WithRealServer", func(t *testing.T) {
			runEndToEndTests(t, false)
		})
	}
}

func runEndToEndTests(t *testing.T, useMockServer bool) {
	env := Setup(t, useMockServer)
	defer env.Teardown(t)

	// Create test project
	env.CreateTestProject(t)

	// Test cases
	testCases := []struct {
		name       string
		tool       string
		params     map[string]interface{}
		validateFn func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "Format Go Code (Using Code Only)",
			tool: "go_fmt",
			params: map[string]interface{}{
				"code": `package main

import "fmt"

func main() {
	// A simple hello world program
	fmt.Println("Hello, World from Go Development MCP Server!")
}`,
			},
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				// Extract JSON from content
				require.NotEmpty(t, result.Content, "Result content is empty")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Content is not text content")

				// Parse JSON
				var response map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &response)
				require.NoError(t, err, "Failed to parse response JSON")

				// Validate response
				assert.True(t, response["success"].(bool), "Format operation should succeed")
				assert.NotEmpty(t, response["formattedCode"], "Formatted code should not be empty")
			},
		},
		{
			name: "Format Go Code (Using Project Path)",
			tool: "go_fmt",
			params: map[string]interface{}{
				"project_path": "${TempDir}",
			},
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				// Extract JSON from content
				require.NotEmpty(t, result.Content, "Result content is empty")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Content is not text content")

				// Parse JSON
				var response map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &response)
				require.NoError(t, err, "Failed to parse response JSON")

				// Validate response
				assert.True(t, response["success"].(bool), "Format operation should succeed")
				assert.NotEmpty(t, response["formattedCode"], "Formatted code should not be empty")
			},
		},
		{
			name: "Format Go Code (Hybrid Strategy)",
			tool: "go_fmt",
			params: map[string]interface{}{
				"code": `package main

import "fmt"

func main() {
	// A simple hello world program
	fmt.Println("Hello, World from Go Development MCP Server!")
}`,
				"project_path": "${TempDir}",
			},
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				// Extract JSON from content
				require.NotEmpty(t, result.Content, "Result content is empty")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Content is not text content")

				// Parse JSON
				var response map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &response)
				require.NoError(t, err, "Failed to parse response JSON")

				// Validate response
				assert.True(t, response["success"].(bool), "Format operation should succeed")

				// Check for metadata (strategy type)
				metadata, hasMetadata := response["metadata"].(map[string]interface{})
				if hasMetadata {
					strategy, hasStrategy := metadata["strategyType"].(string)
					if hasStrategy {
						assert.Equal(t, "hybrid", strategy, "Strategy type should be hybrid")
					}
				}
			},
		},
		{
			name: "Build Go Code",
			tool: "go_build",
			params: map[string]interface{}{
				"project_path": "${TempDir}",
			},
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				// Extract JSON from content
				require.NotEmpty(t, result.Content, "Result content is empty")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Content is not text content")

				// Parse JSON
				var response map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &response)
				require.NoError(t, err, "Failed to parse response JSON")

				// Validate response
				assert.True(t, response["success"].(bool), "Build operation should succeed")

				// Skip executable check for mock server
				if !useMockServer {
					outputPath, hasPath := response["outputPath"].(string)
					if hasPath {
						assert.FileExists(t, outputPath, "Build should create executable")
					}
				}
			},
		},
		{
			name: "Run Go Code",
			tool: "go_run",
			params: map[string]interface{}{
				"project_path": "${TempDir}",
			},
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				// Extract JSON from content
				require.NotEmpty(t, result.Content, "Result content is empty")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Content is not text content")

				// Parse JSON
				var response map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &response)
				require.NoError(t, err, "Failed to parse response JSON")

				// Validate response
				assert.True(t, response["success"].(bool), "Run operation should succeed")
				assert.Equal(t, float64(0), response["exitCode"], "Exit code should be 0")

				// Check output
				stdout, hasStdout := response["stdout"].(string)
				if hasStdout {
					assert.Contains(t, stdout, "Hello, World from Go Development MCP Server!", "Output should contain expected message")
				}
			},
		},
		{
			name: "Build Invalid Code (Error Handling)",
			tool: "go_build",
			params: map[string]interface{}{
				"code": `package main

func main() {
	fmt.Println(Hello World) // Syntax error - missing quotes
}`,
			},
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				// Extract JSON from content
				require.NotEmpty(t, result.Content, "Result content is empty")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Content is not text content")

				// Parse JSON
				var response map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &response)
				require.NoError(t, err, "Failed to parse response JSON")

				// Validate response
				assert.False(t, response["success"].(bool), "Build operation should fail")

				// Check for error message
				_, hasStderr := response["stderr"]
				assert.True(t, hasStderr, "Response should include stderr with error messages")
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Replace ${TempDir} placeholder with actual temp directory
			params := make(map[string]interface{})
			for k, v := range tc.params {
				if strVal, ok := v.(string); ok && strVal == "${TempDir}" {
					params[k] = env.TempDir
				} else {
					params[k] = v
				}
			}

			// Call the tool
			result, err := env.CallMCPTool(tc.tool, params)
			require.NoError(t, err, "Failed to call MCP tool")

			// Validate the result
			tc.validateFn(t, result)
		})
	}
}

// Helper functions

// isServerAvailable checks if the server is available
func isServerAvailable(url string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < 3; i++ {
		resp, err := client.Head(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(1 * time.Second)
	}

	return false
}

// execCommand is a wrapper for exec.Command that can be mocked in tests
var execCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
