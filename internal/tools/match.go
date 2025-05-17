package tools

import (
	"strings"
	"unicode"
)

// MatchScore represents the result of a fuzzy match
type MatchScore struct {
	ToolName string  // Name of the matching tool
	Score    float64 // Score from 0.0 to 1.0 (higher is better)
	Reason   string  // Why this match was selected
}

// MatchToolName finds the best tool match for a natural language query
// Returns empty string if no good match is found
func MatchToolName(query string) *MatchScore {
	if query == "" {
		return nil
	}

	// Normalize query
	query = strings.TrimSpace(strings.ToLower(query))
	
	// Track best match
	bestMatch := &MatchScore{
		ToolName: "",
		Score:    0.0,
		Reason:   "",
	}

	// Check each tool and its metadata
	for toolName, metadata := range ToolNLMetadata {
		// 1. Check exact match with tool name
		if strings.ToLower(toolName) == query {
			return &MatchScore{
				ToolName: toolName,
				Score:    1.0,
				Reason:   "Exact tool name match",
			}
		}
		
		// 2. Check aliases (exact matches)
		for _, alias := range metadata.Aliases {
			if strings.ToLower(alias) == query {
				return &MatchScore{
					ToolName: toolName,
					Score:    0.95,
					Reason:   "Exact alias match: " + alias,
				}
			}
		}
		
		// 3. Try substring matching against tool name
		if strings.Contains(strings.ToLower(toolName), query) {
			score := 0.85
			if score > bestMatch.Score {
				bestMatch.ToolName = toolName
				bestMatch.Score = score
				bestMatch.Reason = "Tool name contains query"
			}
		}
		
		// 4. Check aliases with substring matching
		for _, alias := range metadata.Aliases {
			if strings.Contains(strings.ToLower(alias), query) {
				score := 0.8
				if score > bestMatch.Score {
					bestMatch.ToolName = toolName
					bestMatch.Score = score
					bestMatch.Reason = "Alias contains query: " + alias
				}
			}
		}
		
		// 5. Check if query contains tool name
		if strings.Contains(query, strings.ToLower(toolName)) {
			score := 0.75
			if score > bestMatch.Score {
				bestMatch.ToolName = toolName
				bestMatch.Score = score
				bestMatch.Reason = "Query contains tool name"
			}
		}
		
		// 6. Check if query contains any aliases
		for _, alias := range metadata.Aliases {
			if strings.Contains(query, strings.ToLower(alias)) {
				score := 0.7
				if score > bestMatch.Score {
					bestMatch.ToolName = toolName
					bestMatch.Score = score
					bestMatch.Reason = "Query contains alias: " + alias
				}
			}
		}
		
		// 7. Check examples with similarity scoring
		for _, example := range metadata.Examples {
			similarity := calculateSimilarity(query, strings.ToLower(example))
			if similarity > bestMatch.Score {
				bestMatch.ToolName = toolName
				bestMatch.Score = similarity
				bestMatch.Reason = "Similar to example: " + example
			}
		}
		
		// 8. Last resort: token overlap with tool name and description
		tokens1 := tokenize(query)
		tokens2 := tokenize(strings.ToLower(toolName))
		overlap := calculateTokenOverlap(tokens1, tokens2)
		if overlap > 0.3 && overlap > bestMatch.Score {
			bestMatch.ToolName = toolName
			bestMatch.Score = overlap * 0.6 // Scale down token matches
			bestMatch.Reason = "Token overlap with tool name"
		}
	}
	
	// Only return a match if we're reasonably confident
	if bestMatch.Score > 0.4 {
		return bestMatch
	}
	
	return nil
}

// calculateSimilarity scores how similar two strings are (0.0 to 1.0)
func calculateSimilarity(s1, s2 string) float64 {
	// Special case for exact match
	if s1 == s2 {
		return 0.9
	}
	
	// Special case for substring
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		return 0.8
	}
	
	// Check token overlap
	tokens1 := tokenize(s1)
	tokens2 := tokenize(s2)
	
	return calculateTokenOverlap(tokens1, tokens2) * 0.7
}

// tokenize splits a string into tokens (words)
func tokenize(s string) []string {
	// Split on whitespace and remove empty strings
	tokens := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
	
	// Convert to lowercase
	for i, token := range tokens {
		tokens[i] = strings.ToLower(token)
	}
	
	return tokens
}

// calculateTokenOverlap determines how many tokens overlap between two sets
func calculateTokenOverlap(tokens1, tokens2 []string) float64 {
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}
	
	// Create a map for tokens1
	tokenMap := make(map[string]bool)
	for _, t1 := range tokens1 {
		if len(t1) > 2 { // Only consider meaningful words (>2 chars)
			tokenMap[t1] = true
		}
	}
	
	// Count matches in tokens2
	matches := 0
	for _, t2 := range tokens2 {
		if len(t2) > 2 && tokenMap[t2] {
			matches++
		}
	}
	
	// Average overlap ratio
	return float64(matches) / float64((len(tokens1)+len(tokens2))/2)
}