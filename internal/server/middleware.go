package server

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MrFixit96/go-dev-mcp/internal/config"
	"github.com/MrFixit96/go-dev-mcp/internal/tools"
)

// FuzzyMatchKey is a type for the context key to avoid string collisions
type FuzzyMatchKey string

// FuzzyMatchInfoKey is the context key for fuzzy match information
const FuzzyMatchInfoKey FuzzyMatchKey = "fuzzy_match_info"

// FuzzyMatchMiddleware creates a middleware that applies fuzzy matching to tool names
// It attempts to match natural language tool requests to the correct tool
func FuzzyMatchMiddleware(cfg *config.Config, s *server.MCPServer) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Only apply fuzzy matching if it's enabled in config
			if !cfg.NLProcessing.EnableFuzzyMatching {
				return next(ctx, req)
			}

			// We don't have a direct way to check if a tool exists in this version,
			// so we'll just try to match all requests that come through

			// Try fuzzy matching
			match := tools.MatchToolName(req.Params.Name)
			if match != nil && match.Score >= cfg.NLProcessing.MatchThreshold {
				log.Printf("Fuzzy matching: '%s' matched to tool '%s' (score: %.2f, reason: %s)",
					req.Params.Name, match.ToolName, match.Score, match.Reason)

				// Create a new request with the matched tool name
				newReq := req
				newReq.Params.Name = match.ToolName
				// Add metadata to the response
				ctx = context.WithValue(ctx, FuzzyMatchInfoKey, map[string]interface{}{
					"original_request": req.Params.Name,
					"matched_tool":     match.ToolName,
					"score":            match.Score,
					"reason":           match.Reason,
				})

				return next(ctx, newReq)
			}

			// No match found or exact match already, continue with original request
			return next(ctx, req)
		}
	}
}

// ApplyFuzzyMatching adds fuzzy matching middleware to an MCP server
func ApplyFuzzyMatching(s *server.MCPServer, cfg *config.Config) {
	// In this version of the library, we need to apply middleware during server creation
	// We'll modify main.go to use WithToolHandlerMiddleware instead of calling s.Use()
	log.Println("Fuzzy matching middleware is ready to use with WithToolHandlerMiddleware")
}
