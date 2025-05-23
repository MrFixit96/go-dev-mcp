package tools

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
)

// InputSource identifies the source type of input for Go tools
type InputSource int

const (
	// SourceUnknown indicates no recognizable input source
	SourceUnknown InputSource = iota
	// SourceCode indicates the input comes from provided code
	SourceCode
	// SourceProjectPath indicates the input is a Go project directory
	SourceProjectPath
	// SourceHybrid indicates both code and project path are provided
	SourceHybrid
)

// InputContext holds information about the input to be processed
type InputContext struct {
	Source      InputSource
	Code        string
	ProjectPath string
	MainFile    string
	TestCode    string
}

// ResolveInput determines whether the request contains code or a project path
func ResolveInput(req mcp.CallToolRequest) (InputContext, error) {
	ctx := InputContext{Source: SourceUnknown}	// Extract code if provided
	if code, ok := req.GetArguments()["code"].(string); ok && code != "" {
		ctx.Code = code
		ctx.Source = SourceCode
	}

	// Extract project_path if provided
	if path, ok := req.GetArguments()["project_path"].(string); ok && path != "" {
		ctx.ProjectPath = path
		// Validate path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return ctx, fmt.Errorf("project path does not exist: %s", path)
		}

		// If both code and project_path are provided, use hybrid source
		if ctx.Code != "" {
			ctx.Source = SourceHybrid
		} else {
			ctx.Source = SourceProjectPath
		}
	}
	// Extract test code if provided
	if testCode, ok := req.GetArguments()["testCode"].(string); ok && testCode != "" {
		ctx.TestCode = testCode
	}

	// Validate input
	if ctx.Source == SourceUnknown {
		return ctx, fmt.Errorf("at least one of 'code' or 'project_path' must be provided")
	}

	// Set default main file
	if mainFile, ok := req.GetArguments()["mainFile"].(string); ok && mainFile != "" {
		ctx.MainFile = mainFile
	} else {
		ctx.MainFile = "main.go"
	}

	return ctx, nil
}
