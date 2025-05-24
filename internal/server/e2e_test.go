package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mock_testing "github.com/MrFixit96/go-dev-mcp/internal/testing/mock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"
)

const helloWorldCode = `package main

import "fmt"

func main() {
    // A simple hello world program
    fmt.Println("Hello, World from Go Development MCP Server!")
}
`

// E2ETestSuite is a suite for end-to-end tests using a mock MCP server.
type E2ETestSuite struct {
	suite.Suite
	mockServer *mock_testing.MockServer
	tempDir    string
}

// SetupSuite sets up the test suite.
func (s *E2ETestSuite) SetupSuite() {
	s.mockServer = mock_testing.NewMockServer()
	// Register default handlers for all tools
	s.mockServer.AddToolHandler("go_fmt", mock_testing.DefaultGoFmtHandler)
	s.mockServer.AddToolHandler("go_build", mock_testing.DefaultGoBuildHandler)
	s.mockServer.AddToolHandler("go_run", mock_testing.DefaultGoRunHandler)
	s.mockServer.AddToolHandler("go_test", mock_testing.DefaultGoTestHandler)
	s.mockServer.AddToolHandler("go_mod", mock_testing.DefaultGoModHandler)
	s.mockServer.AddToolHandler("go_analyze", mock_testing.DefaultGoAnalyzeHandler)

	// Create a temporary directory for the test project
	var err error
	s.tempDir, err = os.MkdirTemp("", "e2e-test-project-")
	s.Require().NoError(err, "Failed to create temp directory")

	// Create a simple Go project in the temp directory
	mainGoPath := filepath.Join(s.tempDir, "main.go")
	err = os.WriteFile(mainGoPath, []byte(helloWorldCode), 0644)
	s.Require().NoError(err, "Failed to write main.go")

	// Initialize go.mod
	// In a real e2e test against the actual server, this would be done by the go_mod tool.
	// For mock server tests, we often prepare the state if not testing the tool itself directly.
	// However, the PowerShell script does call `go mod init`, so we can simulate that part
	// or assume the `go_mod` tool will be called if needed by a test case.
}

// TearDownSuite tears down the test suite.
func (s *E2ETestSuite) TearDownSuite() {
	s.mockServer.Close()
	err := os.RemoveAll(s.tempDir)
	s.NoError(err, "Failed to remove temp directory")
}

// BeforeTest runs before each test.
func (s *E2ETestSuite) BeforeTest(suiteName, testName string) {
	s.mockServer.ClearReceivedRequests() // Clear requests before each test
}

// Helper function to invoke a tool via HTTP against the mock server
func (s *E2ETestSuite) invokeTool(toolName string, params map[string]interface{}) (*mcp.CallToolResult, error) {
	reqPayload := mcp.CallToolRequest{
		Params: struct {
			Name      string      `json:"name"`
			Arguments interface{} `json:"arguments,omitempty"`
			Meta      *mcp.Meta   `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: params,
		},
	}

	jsonBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.mockServer.URL()+"/calltool", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	var toolResult mcp.CallToolResult
	if err := json.NewDecoder(resp.Body).Decode(&toolResult); err != nil {
		return nil, fmt.Errorf("failed to decode tool result: %w", err)
	}
	return &toolResult, nil
}

func (s *E2ETestSuite) TestGoFmt_CodeOnly() {
	params := map[string]interface{}{
		"code": helloWorldCode,
	}
	result, err := s.invokeTool("go_fmt", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	var fmtResultData struct {
		Success       bool   `json:"success"`
		FormattedCode string `json:"formattedCode"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &fmtResultData)
	s.Require().NoError(err, "Failed to unmarshal go_fmt result")

	s.True(fmtResultData.Success, "go_fmt should succeed")
	s.NotEmpty(fmtResultData.FormattedCode, "Formatted code should not be empty")
	s.Contains(fmtResultData.FormattedCode, "Hello, World from Go Development MCP Server!", "Formatted code should contain original content")
}

func (s *E2ETestSuite) TestGoFmt_ProjectPath() {
	params := map[string]interface{}{
		"project_path": s.tempDir,
	}
	result, err := s.invokeTool("go_fmt", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	// The default mock handler for go_fmt returns formattedCode.
	// A more sophisticated mock might return a list of formatted files for project_path.
	var fmtResultData struct {
		Success       bool   `json:"success"`
		FormattedCode string `json:"formattedCode"` // Assuming mock returns this for simplicity
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &fmtResultData)
	s.Require().NoError(err, "Failed to unmarshal go_fmt result")

	s.True(fmtResultData.Success, "go_fmt with project_path should succeed")
	s.NotEmpty(fmtResultData.FormattedCode, "Formatted code should not be empty for project_path mock")
}

func (s *E2ETestSuite) TestGoFmt_Hybrid() {
	params := map[string]interface{}{
		"code":         helloWorldCode,
		"project_path": s.tempDir,
	}
	result, err := s.invokeTool("go_fmt", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	var fmtResultData struct {
		Success       bool   `json:"success"`
		FormattedCode string `json:"formattedCode"`
		Metadata      struct {
			StrategyType string `json:"strategyType"`
		} `json:"metadata"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &fmtResultData)
	s.Require().NoError(err, "Failed to unmarshal go_fmt hybrid result")

	s.True(fmtResultData.Success, "go_fmt hybrid should succeed")
	s.NotEmpty(fmtResultData.FormattedCode, "Formatted code should not be empty for hybrid")
	s.Equal("hybrid", fmtResultData.Metadata.StrategyType, "Strategy type should be hybrid")
}

func (s *E2ETestSuite) TestGoMod_InitProject() {
	params := map[string]interface{}{
		"command":      "init",
		"modulePath":   "example.com/e2etest",
		"project_path": s.tempDir, // Mock handler might not use this, but good to include
	}
	result, err := s.invokeTool("go_mod", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	var modResultData struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Output  string `json:"output"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &modResultData)
	s.Require().NoError(err, "Failed to unmarshal go_mod result")

	s.True(modResultData.Success, "go_mod init should succeed")
	s.Contains(modResultData.Output, "go: creating new go.mod: module example.com/e2etest", "Output should indicate module creation")

	// Verify that the mock server received the request with correct params
	s.Require().Len(s.mockServer.Received, 1, "Mock server should have received one request")
	receivedReq := s.mockServer.Received[0]
	s.Equal("go_mod", receivedReq.Params.Name)
	args, ok := receivedReq.Params.Arguments.(map[string]interface{})
	s.Require().True(ok, "Arguments should be a map")
	s.Equal("init", args["command"])
	s.Equal("example.com/e2etest", args["modulePath"])
}

func (s *E2ETestSuite) TestGoAnalyze_CodeNoIssues() {
	params := map[string]interface{}{
		"code": "package main\nfunc main() { println(\"valid code\") }",
	}
	result, err := s.invokeTool("go_analyze", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	var analyzeResultData struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Issues  []string `json:"issues"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &analyzeResultData)
	s.Require().NoError(err, "Failed to unmarshal go_analyze result")

	s.True(analyzeResultData.Success, "go_analyze should succeed for valid code")
	s.Empty(analyzeResultData.Issues, "Issues should be empty for valid code with default handler")
}

func (s *E2ETestSuite) TestGoAnalyze_CodeWithIssues() {
	expectedIssues := []string{"mock issue 1", "mock issue 2"}
	// Override the default analyze handler to return issues
	s.mockServer.AddToolHandler("go_analyze", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		response := map[string]interface{}{
			"success": true,
			"message": "Analysis complete, issues found",
			"issues":  expectedIssues,
		}
		jsonData, _ := json.Marshal(response)
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(jsonData)}}}, nil
	})
	// Restore default handler after this test
	defer s.mockServer.AddToolHandler("go_analyze", mock_testing.DefaultGoAnalyzeHandler)
	params := map[string]interface{}{
		"code": "package main\nfunc main() { var x int }", // Code content doesn't strictly matter here
	}
	result, err := s.invokeTool("go_analyze", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	var analyzeResultData struct {
		Success bool     `json:"success"`
		Issues  []string `json:"issues"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &analyzeResultData)
	s.Require().NoError(err, "Failed to unmarshal go_analyze result with issues")

	s.True(analyzeResultData.Success, "go_analyze should succeed even if issues are found")
	s.Equal(expectedIssues, analyzeResultData.Issues, "Should return the predefined mock issues")
}

func (s *E2ETestSuite) TestGoBuild_ProjectPath() {
	params := map[string]interface{}{
		"project_path": s.tempDir,
	}
	result, err := s.invokeTool("go_build", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	var buildResultData struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		OutputPath string `json:"outputPath"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &buildResultData)
	s.Require().NoError(err, "Failed to unmarshal go_build result")

	s.True(buildResultData.Success, "go_build should succeed")
	s.NotEmpty(buildResultData.OutputPath, "Output path should not be empty")
	// In a real test against the server, we'd check if the executable exists.
	// The mock server just returns a path.
}

func (s *E2ETestSuite) TestGoRun_ProjectPath() {
	params := map[string]interface{}{
		"project_path": s.tempDir,
	}
	result, err := s.invokeTool("go_run", params)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	var runResultData struct {
		Success  bool   `json:"success"`
		Stdout   string `json:"stdout"`
		ExitCode int    `json:"exitCode"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &runResultData)
	s.Require().NoError(err, "Failed to unmarshal go_run result")

	s.True(runResultData.Success, "go_run should succeed")
	s.Equal(0, runResultData.ExitCode, "Exit code should be 0")
	s.Contains(runResultData.Stdout, "Hello, World from Go Development MCP Server!", "Stdout should contain expected output")
}

func (s *E2ETestSuite) TestGoBuild_ErrorHandling() { // Override the default build handler to return an error
	s.mockServer.AddToolHandler("go_build", func(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Simulate a build error
		response := map[string]interface{}{
			"success": false,
			"message": "Build failed: mock error",
			"stderr":  "compiler error: something went wrong",
		}
		jsonData, _ := json.Marshal(response)
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Text: string(jsonData)}}}, nil
	})
	// Restore default handler after this test
	defer s.mockServer.AddToolHandler("go_build", mock_testing.DefaultGoBuildHandler)

	params := map[string]interface{}{
		"code": "package main\n\nfunc main() { fmt.Println(HelloWorld) }", // Intentional error
	}
	result, err := s.invokeTool("go_build", params)
	s.Require().NoError(err) // HTTP call itself should succeed
	s.Require().NotNil(result)

	var buildResultData struct {
		Success bool   `json:"success"`
		Stderr  string `json:"stderr"`
	}
	s.Require().IsType(mcp.TextContent{}, result.Content[0])
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &buildResultData)
	s.Require().NoError(err, "Failed to unmarshal go_build error result")

	s.False(buildResultData.Success, "go_build should fail for invalid code")
	s.NotEmpty(buildResultData.Stderr, "Stderr should contain error details")
	s.Contains(buildResultData.Stderr, "compiler error", "Stderr should indicate a compiler error")

	// Verify that the mock server received the request
	s.Require().Len(s.mockServer.Received, 1, "Mock server should have received one request")
	receivedReq := s.mockServer.Received[0]
	s.Equal("go_build", receivedReq.Params.Name)
	args, ok := receivedReq.Params.Arguments.(map[string]interface{})
	s.Require().True(ok, "Arguments should be a map")
	s.Equal(params["code"], args["code"], "Received code argument should match")
}

// TestE2ETestSuite runs the E2E test suite.
func TestE2ETestSuite(t *testing.T) {
	// This check is to ensure that the tests are run with the `go test` command
	// and not as part of a regular build. This is important because these tests
	// start a mock HTTP server and interact with the file system.
	if os.Getenv("GO_TEST_MAIN_RUN") == "" && !strings.HasSuffix(os.Args[0], ".test") {
		t.Log("Skipping E2E tests when not run with 'go test'.")
		return
	}
	suite.Run(t, new(E2ETestSuite))
}

// This main function is needed to correctly run `suite.Run`.
// It's a standard pattern for testify suites.
func TestMain(m *testing.M) {
	os.Setenv("GO_TEST_MAIN_RUN", "1")
	code := m.Run()
	os.Exit(code)
}
