// Package mock provides mock implementations of MCP server components for testing
package mock

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/mock"
)

// ToolExecutor is a mock implementation of a tool executor
type ToolExecutor struct {
	mock.Mock
}

// ExecuteTool mocks the execution of an MCP tool
func (m *ToolExecutor) ExecuteTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := m.Called(ctx, req)
	result := args.Get(0)
	err := args.Error(1)

	if result == nil {
		return nil, err
	}

	return result.(*mcp.CallToolResult), err
}

// SetupResponse configures the mock to return a specific response for a given tool
func (m *ToolExecutor) SetupResponse(toolName string, input map[string]interface{}, result *mcp.CallToolResult, err error) *mock.Call {
	// Use a matcher function that checks for the tool name
	return m.On("ExecuteTool", mock.Anything, mock.MatchedBy(func(r mcp.CallToolRequest) bool {
		return r.Params.Name == toolName
	})).Return(result, err)
}

// SetupSuccessResponse configures the mock to return a success response
func (m *ToolExecutor) SetupSuccessResponse(toolName string, responseText string) *mock.Call {
	// Use nil for the input map to indicate we don't care about the input
	var anyInput map[string]interface{}
	return m.SetupResponse(toolName, anyInput, mcp.NewToolResultText(responseText), nil)
}

// SetupErrorResponse configures the mock to return an error response
func (m *ToolExecutor) SetupErrorResponse(toolName string, errorMessage string) *mock.Call {
	// Use nil for the input map to indicate we don't care about the input
	var anyInput map[string]interface{}
	return m.SetupResponse(toolName, anyInput, mcp.NewToolResultError(errorMessage), nil)
}

// ExecutionStrategy is a mock implementation of an execution strategy
type ExecutionStrategy struct {
	mock.Mock
}

// Execute mocks the execution of a command
func (m *ExecutionStrategy) Execute(ctx context.Context, input interface{}, args []string) (interface{}, error) {
	mockArgs := m.Called(ctx, input, args)
	return mockArgs.Get(0), mockArgs.Error(1)
}
