package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ErrorType defines categories of errors that can occur during Go operations
type ErrorType string

const (
	ErrorTypeCompilation ErrorType = "compilation"
	ErrorTypeExecution   ErrorType = "execution"
	ErrorTypeSystem      ErrorType = "system"
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeTimeout     ErrorType = "timeout"
	ErrorTypeUnknown     ErrorType = "unknown"
)

// ErrorDetail represents a structured error with context
type ErrorDetail struct {
	Type        ErrorType `json:"type"`
	Message     string    `json:"message"`
	File        string    `json:"file,omitempty"`
	Line        int       `json:"line,omitempty"`
	Column      int       `json:"column,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// AppendSuggestion adds a suggestion to the error detail
func (e *ErrorDetail) AppendSuggestion(suggestion string) {
	e.Suggestions = append(e.Suggestions, suggestion)
}

// ErrorResponse represents a comprehensive error response
type ErrorResponse struct {
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	ErrorDetails []ErrorDetail `json:"errorDetails"`
	Timestamp    time.Time     `json:"timestamp"`
	Duration     string        `json:"duration,omitempty"`
	ExitCode     int           `json:"exitCode,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, errorType ErrorType, details string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now(),
		ErrorDetails: []ErrorDetail{
			{
				Type:    errorType,
				Message: details,
			},
		},
	}
}

// AddErrorDetail adds an error detail to the response
func (r *ErrorResponse) AddErrorDetail(detail ErrorDetail) {
	r.ErrorDetails = append(r.ErrorDetails, detail)
}

// SetExitCode sets the exit code for the error response
func (r *ErrorResponse) SetExitCode(code int) {
	r.ExitCode = code
}

// SetDuration sets the duration for the error response
func (r *ErrorResponse) SetDuration(duration time.Duration) {
	r.Duration = duration.String()
}

// ToJSON converts the error response to JSON
func (r *ErrorResponse) ToJSON() string {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(`{"success":false,"message":"Error marshaling response: %s"}`, err)
	}
	return string(jsonBytes)
}

// ParseGoErrors parses Go compiler error output into structured error details
func ParseGoErrors(stderr string) []ErrorDetail {
	if stderr == "" {
		return nil
	}

	var details []ErrorDetail
	
	// Split error message by lines
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Simple parsing of Go errors which typically follow the format:
		// file.go:line:column: error description
		parts := strings.SplitN(line, ":", 4)
		if len(parts) >= 3 {
			// Try to extract file, line, and column
			file := parts[0]
			lineNum := 0
			colNum := 0
			
			fmt.Sscanf(parts[1], "%d", &lineNum)
			if len(parts) > 2 {
				fmt.Sscanf(parts[2], "%d", &colNum)
			}
			
			errorMsg := ""
			if len(parts) > 3 {
				errorMsg = strings.TrimSpace(parts[3])
			} else if len(parts) > 2 {
				errorMsg = strings.TrimSpace(parts[2])
			}
			
			detail := ErrorDetail{
				Type:    ErrorTypeCompilation,
				Message: errorMsg,
				File:    file,
				Line:    lineNum,
			}
			
			if colNum > 0 {
				detail.Column = colNum
			}
			
			// Add suggestions based on common Go errors
			if strings.Contains(errorMsg, "undefined:") || strings.Contains(errorMsg, "undeclared name:") {
				detail.AppendSuggestion("Check if you have imported the necessary package")
				detail.AppendSuggestion("Verify that the variable or function name is spelled correctly")
			} else if strings.Contains(errorMsg, "syntax error:") {
				detail.AppendSuggestion("Check for missing braces, parentheses, or semicolons")
				detail.AppendSuggestion("Verify that syntax is correct according to Go language specification")
			}
			
			details = append(details, detail)
		} else {
			// For error messages that don't match the expected format
			details = append(details, ErrorDetail{
				Type:    ErrorTypeUnknown,
				Message: line,
			})
		}
	}
	
	return details
}