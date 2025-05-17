package server

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MrFixit96/go-dev-mcp/internal/config"
	"github.com/MrFixit96/go-dev-mcp/internal/tools"
)

// FuzzyMatchMiddleware creates a middleware that applies fuzzy matching to tool names
// It attempts to match natural language tool requests to the correct tool
func FuzzyMatchMiddleware(cfg *config.Config, s *server.MCPServer) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Only apply fuzzy matching if it's enabled in config
			if !cfg.NLProcessing.EnableFuzzyMatching {
				return next(ctx, req)
			}

			// Check if the requested tool exists directly
			if _, err := s.GetTool(req.Params.Name); err == nil {
				// Tool exists, proceed normally
				return next(ctx, req)
			}

			// Tool doesn't exist - try fuzzy matching
			match := tools.MatchToolName(req.Params.Name)
			if match != nil && match.Score >= cfg.NLProcessing.MatchThreshold {
				log.Printf("Fuzzy matching: '%s' matched to tool '%s' (score: %.2f, reason: %s)",
					req.Params.Name, match.ToolName, match.Score, match.Reason)
				
				// Create a new request with the matched tool name
				newReq := req
				newReq.Params.Name = match.ToolName
				
				// Add metadata to the response
				ctx = context.WithValue(ctx, "fuzzy_match_info", map[string]interface{}{
					"original_request": req.Params.Name,
					"matched_tool":     match.ToolName,
					"score":            match.Score,
					"reason":           match.Reason,
				})
				
				return next(ctx, newReq)
			}
			
			// No match found, continue with original request (which will likely fail)
			log.Printf("No fuzzy match found for tool '%s'", req.Params.Name)
			return next(ctx, req)
		}
	}
}

// ApplyFuzzyMatching adds fuzzy matching middleware to an MCP server
func ApplyFuzzyMatching(s *server.MCPServer, cfg *config.Config) {
	middleware := FuzzyMatchMiddleware(cfg, s)
	s.Use(middleware)
}