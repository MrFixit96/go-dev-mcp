// package mock provides a configurable mock server for testing MCP functionality.
package mock

import (
	"time"
)

// ResponseScenario defines different response scenarios for mock testing
type ResponseScenario string

const (
	// ScenarioSuccess represents a successful response
	ScenarioSuccess ResponseScenario = "success"
	// ScenarioFailure represents a failure response
	ScenarioFailure ResponseScenario = "failure"
	// ScenarioTimeout represents a timeout response
	ScenarioTimeout ResponseScenario = "timeout"
	// ScenarioNetworkError represents a network error
	ScenarioNetworkError ResponseScenario = "network_error"
	// ScenarioMalformedRequest represents an invalid request format
	ScenarioMalformedRequest ResponseScenario = "malformed_request"
	// ScenarioServerError represents a server error (500)
	ScenarioServerError ResponseScenario = "server_error"
)

// ToolConfig defines configuration for a specific tool
type ToolConfig struct {
	// Scenario defines the response scenario to use
	Scenario ResponseScenario
	// Delay adds artificial delay before responding (for timeout testing)
	Delay time.Duration
	// CustomResponse allows overriding the default response for this tool
	CustomResponse map[string]interface{}
	// ErrorMessage provides a custom error message when using failure scenarios
	ErrorMessage string
	// StatusCode allows overriding the HTTP status code
	StatusCode int
}

// ServerConfig provides configuration for the mock server
type ServerConfig struct {
	// DefaultScenario is the fallback scenario for tools without specific config
	DefaultScenario ResponseScenario
	// DefaultDelay is the default delay for all responses
	DefaultDelay time.Duration
	// ToolConfigs maps tool names to their configurations
	ToolConfigs map[string]ToolConfig
	// ValidateJSONRPC determines whether to validate JSON-RPC 2.0 compliance
	ValidateJSONRPC bool
	// AllowUnregisteredTools determines behavior for tools without handlers
	AllowUnregisteredTools bool
	// DefaultStatusCode is the default HTTP status code for responses
	DefaultStatusCode int
}

// NewDefaultConfig creates a new ServerConfig with default values
func NewDefaultConfig() *ServerConfig {
	return &ServerConfig{
		DefaultScenario:        ScenarioSuccess,
		DefaultDelay:           0 * time.Millisecond,
		ToolConfigs:            make(map[string]ToolConfig),
		ValidateJSONRPC:        true,
		AllowUnregisteredTools: false,
		DefaultStatusCode:      200,
	}
}

// NewFailureConfig creates a new ServerConfig with all tools set to fail
func NewFailureConfig() *ServerConfig {
	return &ServerConfig{
		DefaultScenario:        ScenarioFailure,
		DefaultDelay:           0 * time.Millisecond,
		ToolConfigs:            make(map[string]ToolConfig),
		ValidateJSONRPC:        true,
		AllowUnregisteredTools: false,
		DefaultStatusCode:      200, // Even failures use 200 with error content in MCP
	}
}

// NewTimeoutConfig creates a new ServerConfig with all tools set to timeout
func NewTimeoutConfig(delay time.Duration) *ServerConfig {
	return &ServerConfig{
		DefaultScenario:        ScenarioTimeout,
		DefaultDelay:           delay,
		ToolConfigs:            make(map[string]ToolConfig),
		ValidateJSONRPC:        true,
		AllowUnregisteredTools: false,
		DefaultStatusCode:      200,
	}
}

// SetToolConfig sets configuration for a specific tool
func (c *ServerConfig) SetToolConfig(toolName string, config ToolConfig) {
	c.ToolConfigs[toolName] = config
}

// GetToolConfig gets configuration for a specific tool, falling back to defaults
func (c *ServerConfig) GetToolConfig(toolName string) ToolConfig {
	if config, exists := c.ToolConfigs[toolName]; exists {
		return config
	}
	// Return default config for this tool
	return ToolConfig{
		Scenario:   c.DefaultScenario,
		Delay:      c.DefaultDelay,
		StatusCode: c.DefaultStatusCode,
	}
}
