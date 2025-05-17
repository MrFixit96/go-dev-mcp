package server

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CreateLoggingMiddleware creates a middleware that logs tool execution time
func CreateLoggingMiddleware() server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			toolName := req.Params.Name
			log.Printf("Executing tool: %s", toolName)
			
			start := time.Now()
			result, err := next(ctx, req)
			duration := time.Since(start)
			
			if err != nil {
				log.Printf("Tool %s failed after %v: %v", toolName, duration, err)
			} else {
				log.Printf("Tool %s completed in %v", toolName, duration)
			}
			
			return result, err
		}
	}
}

// CreateAuthMiddleware creates a middleware that checks if the tool is allowed
func CreateAuthMiddleware(allowedTools map[string]bool) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			toolName := req.Params.Name
			
			if !allowedTools[toolName] {
				log.Printf("Unauthorized tool execution attempt: %s", toolName)
				return mcp.NewToolResultError("Unauthorized tool execution"), nil
			}
			
			return next(ctx, req)
		}
	}
}