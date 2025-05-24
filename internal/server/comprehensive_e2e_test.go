// filepath: c:\Users\James\Documents\go-dev-mcp\internal\server\comprehensive_e2e_test.go
package server_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MrFixit96/go-dev-mcp/internal/testing/mock"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestE2EComprehensive runs comprehensive end-to-end tests for all tools
// using table-driven tests and covering various scenarios including edge cases.
func (s *E2ETestSuite) TestE2EComprehensive() {
	// Define test cases as a table
	testCases := []struct {
		name           string
		tool           string
		params         map[string]interface{}
		setupMock      func() // Optional function to set up custom mock handlers
		resetMock      bool   // Whether to reset to default handler after test
		validateResult func(*mcp.CallToolResult) bool
		expectedError  bool
	}{
		{
			name: "GoFmt with Valid Code",
			tool: "go_fmt",
			params: map[string]interface{}{
				"code": helloWorldCode,
			},
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return resultData["success"].(bool) && resultData["formattedCode"] != nil
			},
		},
		{
			name: "GoFmt with Invalid Code",
			tool: "go_fmt",
			params: map[string]interface{}{
				"code": "package main\n\nfunc main() { fmt.Println(\"missing import\") }",
			},
			setupMock: func() {
				s.mockServer.AddToolHandler("go_fmt", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					// Return error for invalid code
					response := map[string]interface{}{
						"success": false,
						"error":   "Failed to format code: missing import 'fmt'",
					}
					jsonData, _ := json.Marshal(response)
					return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(jsonData)}}}, nil
				})
			},
			resetMock: true,
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return !resultData["success"].(bool) && resultData["error"] != nil
			},
		},
		{
			name: "GoBuild with Valid Code",
			tool: "go_build",
			params: map[string]interface{}{
				"code": helloWorldCode,
			},
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return resultData["success"].(bool) && resultData["outputPath"] != nil
			},
		},
		{
			name: "GoBuild with Project Path",
			tool: "go_build",
			params: map[string]interface{}{
				"project_path": s.tempDir,
				"outputPath":   "hello-test-output",
			},
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return resultData["success"].(bool) && strings.Contains(resultData["outputPath"].(string), "hello-test-output")
			},
		},
		{
			name: "GoBuild with Invalid Code",
			tool: "go_build",
			params: map[string]interface{}{
				"code": "package main\n\nfunc main() { undefinedFunction() }",
			},
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return !resultData["success"].(bool) && resultData["stderr"] != nil
			},
		},
		{
			name: "GoRun with Valid Code",
			tool: "go_run",
			params: map[string]interface{}{
				"code": helloWorldCode,
			},
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return resultData["success"].(bool) &&
					strings.Contains(resultData["stdout"].(string), "Hello, World from Go Development MCP Server!")
			},
		},
		{
			name: "GoRun with Timeout",
			tool: "go_run",
			params: map[string]interface{}{
				"code": `package main
import (
	"fmt"
	"time"
)
func main() {
	fmt.Println("Starting infinite loop...")
	for {
		time.Sleep(1 * time.Second)
	}
}`,
				"timeoutSecs": 1,
			},
			setupMock: func() {
				s.mockServer.AddToolHandler("go_run", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					// Simulate timeout
					response := map[string]interface{}{
						"success":  false,
						"exitCode": -1,
						"stderr":   "Error: execution timed out after 1 second",
						"stdout":   "Starting infinite loop...",
						"metadata": map[string]interface{}{
							"timeoutOccurred": true,
						},
					}
					jsonData, _ := json.Marshal(response)
					return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(jsonData)}}}, nil
				})
			},
			resetMock: true,
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				metadata, ok := resultData["metadata"].(map[string]interface{})
				return !resultData["success"].(bool) && ok && metadata["timeoutOccurred"].(bool)
			},
		},
		{
			name: "GoTest with Valid Code",
			tool: "go_test",
			params: map[string]interface{}{
				"code": `package main
import "testing"
func TestHello(t *testing.T) {
	t.Log("Hello test")
}`,
				"testCode": `package main
import "testing"
func TestWorld(t *testing.T) {
	t.Log("World test")
}`,
				"verbose": true,
			},
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return resultData["success"].(bool) && strings.Contains(resultData["output"].(string), "PASS")
			},
		},
		{
			name: "GoTest with Failing Test",
			tool: "go_test",
			params: map[string]interface{}{
				"testCode": `package main
import "testing"
func TestFailing(t *testing.T) {
	t.Error("Intentionally failing test")
}`,
			},
			setupMock: func() {
				s.mockServer.AddToolHandler("go_test", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					// Simulate failing test
					response := map[string]interface{}{
						"success":  false,
						"message":  "Tests failed",
						"output":   "--- FAIL: TestFailing (0.00s)\n    main_test.go:4: Intentionally failing test",
						"exitCode": 1,
					}
					jsonData, _ := json.Marshal(response)
					return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(jsonData)}}}, nil
				})
			},
			resetMock: true,
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return !resultData["success"].(bool) && resultData["output"] != nil
			},
		},
		{
			name: "GoMod Tidy",
			tool: "go_mod",
			params: map[string]interface{}{
				"command":      "tidy",
				"project_path": s.tempDir,
			},
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return resultData["success"].(bool)
			},
		},
		{
			name: "GoMod with Invalid Command",
			tool: "go_mod",
			params: map[string]interface{}{
				"command":      "invalid_command",
				"project_path": s.tempDir,
			},
			setupMock: func() {
				s.mockServer.AddToolHandler("go_mod", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					// Simulate error for invalid command
					response := map[string]interface{}{
						"success": false,
						"message": "Error: unknown command 'invalid_command'",
						"stderr":  "go: unknown subcommand \"invalid_command\"\nRun 'go help mod' for usage.",
					}
					jsonData, _ := json.Marshal(response)
					return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(jsonData)}}}, nil
				})
			},
			resetMock: true,
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return !resultData["success"].(bool) && strings.Contains(resultData["stderr"].(string), "unknown subcommand")
			},
		},
		{
			name: "GoAnalyze with Warnings",
			tool: "go_analyze",
			params: map[string]interface{}{
				"code": `package main
import "fmt"
func main() {
	var unusedVar int
	fmt.Println("Hello")
}`,
				"vet": true,
			},
			setupMock: func() {
				s.mockServer.AddToolHandler("go_analyze", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					// Return analysis with warnings
					response := map[string]interface{}{
						"success":   true,
						"message":   "Analysis completed with warnings",
						"vetResult": "main.go:4:2: unusedVar declared but not used",
					}
					jsonData, _ := json.Marshal(response)
					return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(jsonData)}}}, nil
				})
			},
			resetMock: true,
			validateResult: func(result *mcp.CallToolResult) bool {
				textContent := result.Content[0].(mcp.TextContent)
				var resultData map[string]interface{}
				err := json.Unmarshal([]byte(textContent.Text), &resultData)
				if err != nil {
					return false
				}
				return resultData["success"].(bool) && strings.Contains(resultData["vetResult"].(string), "unused")
			},
		},
		{
			name: "Server Error",
			tool: "go_fmt",
			params: map[string]interface{}{
				"code": "package main",
			},
			setupMock: func() {
				s.mockServer.AddToolHandler("go_fmt", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					// Simulate server error
					return nil, fmt.Errorf("internal server error: service unavailable")
				})
			},
			resetMock:     true,
			expectedError: true,
		},
	}

	// Run each test case as a subtest
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Clear previous requests
			s.mockServer.ClearReceivedRequests()

			// Set up custom mock handler if provided
			if tc.setupMock != nil {
				tc.setupMock()
			}

			// Reset to default handler after test if needed
			if tc.resetMock {
				defer func() {
					switch tc.tool {
					case "go_fmt":
						s.mockServer.AddToolHandler("go_fmt", mock.DefaultGoFmtHandler)
					case "go_build":
						s.mockServer.AddToolHandler("go_build", mock.DefaultGoBuildHandler)
					case "go_run":
						s.mockServer.AddToolHandler("go_run", mock.DefaultGoRunHandler)
					case "go_test":
						s.mockServer.AddToolHandler("go_test", mock.DefaultGoTestHandler)
					case "go_mod":
						s.mockServer.AddToolHandler("go_mod", mock.DefaultGoModHandler)
					case "go_analyze":
						s.mockServer.AddToolHandler("go_analyze", mock.DefaultGoAnalyzeHandler)
					}
				}()
			}

			// Invoke the tool and validate result
			result, err := s.invokeTool(tc.tool, tc.params)

			if tc.expectedError {
				s.Error(err, "Expected an error but got none")
				return
			}

			s.NoError(err, "Expected no error")
			s.Require().NotNil(result, "Result should not be nil")

			// Additional validation if provided
			if tc.validateResult != nil {
				s.True(tc.validateResult(result), "Result validation failed")
			}
		})
	}
}
