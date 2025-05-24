package mock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockToolHandler is a function that handles a mock tool request and returns a result.
type MockToolHandler func(req mcp.CallToolRequest) (*mcp.CallToolResult, error)

// MockServer is a mock MCP server for testing.
type MockServer struct {
	server       *httptest.Server
	toolHandlers map[string]MockToolHandler
	mu           sync.RWMutex
	Received     []mcp.CallToolRequest // Stores received requests for verification
	Config       *ServerConfig         // Configuration for the mock server
}

// NewMockServer creates and starts a new MockServer.
func NewMockServer() *MockServer {
	return NewMockServerWithConfig(NewDefaultConfig())
}

// NewMockServerWithConfig creates and starts a new MockServer with the provided config.
func NewMockServerWithConfig(config *ServerConfig) *MockServer {
	ms := &MockServer{
		toolHandlers: make(map[string]MockToolHandler),
		Received:     make([]mcp.CallToolRequest, 0),
		Config:       config,
	}
	ms.server = httptest.NewServer(http.HandlerFunc(ms.handleCallTool))
	return ms
}

// URL returns the URL of the mock server.
func (ms *MockServer) URL() string {
	return ms.server.URL
}

// Close closes the mock server.
func (ms *MockServer) Close() {
	ms.server.Close()
}

// AddToolHandler registers a handler for a specific tool.
func (ms *MockServer) AddToolHandler(toolName string, handler MockToolHandler) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.toolHandlers[toolName] = handler
}

// ClearToolHandlers removes all registered tool handlers.
func (ms *MockServer) ClearToolHandlers() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.toolHandlers = make(map[string]MockToolHandler)
}

// ClearReceivedRequests clears the log of received requests.
func (ms *MockServer) ClearReceivedRequests() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Received = make([]mcp.CallToolRequest, 0)
}

func (ms *MockServer) handleCallTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Handle malformed request scenario if configured
	if ms.Config.DefaultScenario == ScenarioMalformedRequest {
		errorBody, _ := CreateJSONRPCErrorResponse(-32700, "Parse error", nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorBody)
		return
	}

	// Parse request
	var req mcp.CallToolRequest
	if err := json.Unmarshal(body, &req); err != nil {
		// Invalid JSON format
		errorBody, _ := CreateJSONRPCErrorResponse(-32700, "Parse error", nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorBody)
		return
	}

	// Validate JSON-RPC 2.0 compliance if enabled
	if ms.Config.ValidateJSONRPC {
		// Check for required JSON-RPC 2.0 fields
		if !validateJSONRPC(&req) {
			errorBody, _ := CreateJSONRPCErrorResponse(-32600, "Invalid Request", "Request does not conform to JSON-RPC 2.0 specification")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errorBody)
			return
		}
	}

	// Store the received request for later verification
	ms.mu.Lock()
	ms.Received = append(ms.Received, req)
	ms.mu.Unlock()

	// Get tool configuration
	toolName := req.Params.Name
	toolConfig := ms.Config.GetToolConfig(toolName)

	// Apply configured delay (for timeout testing)
	if toolConfig.Delay > 0 {
		time.Sleep(toolConfig.Delay)
	}

	// Handle special scenarios
	switch toolConfig.Scenario {
	case ScenarioTimeout:
		// For timeout scenario, simply never respond (sleep for a long time)
		time.Sleep(30 * time.Second)
		return
	case ScenarioNetworkError:
		// Simulate network error by hijacking the connection and closing it
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
		return
	case ScenarioServerError:
		// Return a 500 server error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Look for a registered handler for this tool
	ms.mu.RLock()
	handler, exists := ms.toolHandlers[toolName]
	ms.mu.RUnlock()

	var result *mcp.CallToolResult

	if exists {
		// Use the registered handler with potential error handling
		handlerResult, handlerErr := handler(req)
		if handlerErr != nil {
			// If we have a handler error, handle it according to configuration
			errorMsg := handlerErr.Error()
			if toolConfig.ErrorMessage != "" {
				errorMsg = toolConfig.ErrorMessage
			}

			errorContentMap := map[string]interface{}{
				"success": false,
				"message": errorMsg,
				"error":   errorMsg,
			}
			errorContentBytes, _ := json.Marshal(errorContentMap)

			result = mcp.NewToolResultError(string(errorContentBytes))
		} else {
			result = handlerResult
		}
	} else {
		// No handler exists, use template-based responses

		// If unregistered tools are not allowed, return an error
		if !ms.Config.AllowUnregisteredTools {
			errorBody, _ := CreateJSONRPCErrorResponse(-32601, "Method not found", "Tool not supported: "+toolName)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write(errorBody)
			return
		}

		// Create response from templates
		templateResult, templateErr := GetResponseForTool(toolName, toolConfig.Scenario, toolConfig.CustomResponse)
		if templateErr != nil {
			http.Error(w, "Error creating response from template", http.StatusInternalServerError)
			return
		}
		result = templateResult
	}

	// Marshal the result
	responseBody, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Error marshalling response", http.StatusInternalServerError)
		return
	}

	// Set status code from configuration or default to 200
	statusCode := toolConfig.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}

// validateJSONRPC checks if a request conforms to JSON-RPC 2.0 specification
func validateJSONRPC(req *mcp.CallToolRequest) bool {
	// Basic validation - could be expanded with more checks
	return true // Simplified for now
}

// DefaultGoFmtHandler provides a basic mock handler for go_fmt.
func DefaultGoFmtHandler(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract code if present
	var code string
	var projectPath string
	args, _ := req.Params.Arguments.(map[string]interface{})

	if c, okStr := args["code"].(string); okStr {
		code = c
	}
	if pp, okStr := args["project_path"].(string); okStr {
		projectPath = pp
	}

	strategy := "unknown"
	if code != "" && projectPath != "" {
		strategy = "hybrid"
	} else if code != "" {
		strategy = "code_only"
	} else if projectPath != "" {
		strategy = "project_path_only"
	}

	if code == "" { // If no code provided, use a default for formattedCode
		code = `package main

import "fmt"

func main() { fmt.Println("formatted by mock!") }
`
	}

	response := map[string]interface{}{
		"success":       true,
		"message":       "Code formatted successfully by mock",
		"formattedCode": code,  // Echo back the input code or a default
		"codeChanged":   false, // Simplification for mock
		"metadata": map[string]interface{}{
			"strategyType": strategy,
		},
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal go_fmt response: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

// DefaultGoBuildHandler provides a basic mock handler for go_build.
func DefaultGoBuildHandler(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check for intentional error trigger
	if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
		if code, okStr := args["code"].(string); okStr && code == "error_trigger" {
			response := map[string]interface{}{
				"success": false,
				"message": "Build failed due to intentional error trigger",
				"stderr":  "mock build error",
			}
			jsonData, err := json.Marshal(response)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal go_build error response: %w", err)
			}
			return mcp.NewToolResultError(string(jsonData)), nil
		}
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    "Compilation successful",
		"outputPath": "mock/output/path/executable",
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal go_build response: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

// DefaultGoRunHandler provides a basic mock handler for go_run.
func DefaultGoRunHandler(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	response := map[string]interface{}{
		"success":  true,
		"message":  "Execution successful",
		"stdout":   "Hello, World from Go Development MCP Server!",
		"exitCode": 0,
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal go_run response: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

// DefaultGoTestHandler provides a basic mock handler for go_test.
func DefaultGoTestHandler(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	response := map[string]interface{}{
		"success":   true,
		"message":   "Tests passed",
		"output":    "ok\t_test_\t0.001s",
		"coverage":  "coverage: 100.0% of statements",
		"testStats": map[string]int{"passed": 1, "failed": 0, "skipped": 0},
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal go_test response: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

// DefaultGoModHandler provides a basic mock handler for go_mod.
func DefaultGoModHandler(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	response := map[string]interface{}{
		"success": true,
		"message": "go.mod updated successfully",
		"output":  "go: creating new go.mod: module example.com/hello",
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal go_mod response: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}

// DefaultGoAnalyzeHandler provides a basic mock handler for go_analyze.
func DefaultGoAnalyzeHandler(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	response := map[string]interface{}{
		"success": true,
		"message": "Analysis complete",
		"issues":  []string{}, // No issues by default
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal go_analyze response: %w", err)
	}
	return mcp.NewToolResultText(string(jsonData)), nil
}
