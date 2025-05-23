package server

import (
	"context"
	"log"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MrFixit96/go-dev-mcp/internal/config"
)

// FuzzyMatchKey is a type for the context key to avoid string collisions
type FuzzyMatchKey string

// FuzzyMatchInfoKey is the context key for fuzzy match information
const FuzzyMatchInfoKey FuzzyMatchKey = "fuzzy_match_info"

// simpleToolMatch provides basic tool name matching functionality
// Returns the matched tool name and a confidence score between 0.0 and 1.0
func simpleToolMatch(input string) (string, float64) {
	// Convert to lowercase for case-insensitive matching
	input = strings.ToLower(input)

	// Define tool mappings with their aliases
	toolMappings := map[string][]string{
		"go_run":   {"run", "execute", "go run", "golang run"},
		"go_build": {"build", "compile", "go build", "golang build"},
		"go_fmt":   {"fmt", "format", "go fmt", "golang fmt", "format code"},
		"go_test":  {"test", "testing", "go test", "golang test", "run tests"},
	}

	bestMatch := ""
	highestScore := 0.0

	// Check for exact matches first (highest priority)
	for toolName := range toolMappings {
		if strings.ToLower(toolName) == input {
			return toolName, 1.0
		}
	}

	// Check for partial matches
	for toolName, aliases := range toolMappings {
		// Check if input contains the tool name
		if strings.Contains(input, strings.ToLower(toolName)) {
			score := 0.9 // High confidence for containing the exact tool name
			if score > highestScore {
				highestScore = score
				bestMatch = toolName
			}
		}

		// Check aliases
		for _, alias := range aliases {
			if input == alias {
				score := 0.8 // Good confidence for exact alias match
				if score > highestScore {
					highestScore = score
					bestMatch = toolName
				}
			} else if strings.Contains(input, alias) {
				score := 0.7 // Decent confidence for partial alias match
				if score > highestScore {
					highestScore = score
					bestMatch = toolName
				}
			}
		}
	}

	return bestMatch, highestScore
}

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

			// Try simple matching for common tool names
			toolName := req.Params.Name
			matchedTool, score := simpleToolMatch(toolName)

			if matchedTool != "" && score >= cfg.NLProcessing.MatchThreshold {
				log.Printf("Tool matching: '%s' matched to tool '%s' (score: %.2f)",
					toolName, matchedTool, score)

				// Create a new request with the matched tool name
				newReq := req
				newReq.Params.Name = matchedTool
				// Add metadata to the response
				ctx = context.WithValue(ctx, FuzzyMatchInfoKey, map[string]interface{}{
					"original_request": req.Params.Name,
					"matched_tool":     matchedTool,
					"score":            score,
					"reason":           "Simple prefix/suffix matching",
				})

				return next(ctx, newReq)
			}

			// No match found or below threshold, continue with original request
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
