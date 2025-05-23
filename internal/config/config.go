package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config represents the server configuration
type Config struct {
	Version        string         `json:"version"`
	LogLevel       string         `json:"logLevel"`
	SandboxType    string         `json:"sandboxType"`
	ResourceLimits ResourceLimits `json:"resourceLimits"`
	NLProcessing   NLProcessing   `json:"nlProcessing"`
}

// ResourceLimits defines resource constraints for the execution environment
type ResourceLimits struct {
	CPULimit    int `json:"cpuLimit"`
	MemoryLimit int `json:"memoryLimit"` // in MB
	TimeoutSecs int `json:"timeoutSecs"`
}

// NLProcessing contains settings for natural language processing features
type NLProcessing struct {
	EnableFuzzyMatching bool    `json:"enableFuzzyMatching"`
	MatchThreshold      float64 `json:"matchThreshold"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version:     "1.0.0",
		LogLevel:    "info",
		SandboxType: "process",
		ResourceLimits: ResourceLimits{
			CPULimit:    2,
			MemoryLimit: 512, // 512 MB
			TimeoutSecs: 30,
		},
		NLProcessing: NLProcessing{
			EnableFuzzyMatching: true,
			MatchThreshold:      0.4,
		},
	}
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := DefaultConfig()
		err = save(config, configPath)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	// Read config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse config
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// getConfigPath returns the path to the config file
func getConfigPath() (string, error) {
	var configDir string

	// Get config directory based on OS
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch os.Getenv("GOOS") {
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "go-dev-mcp")
	case "darwin":
		configDir = filepath.Join(homeDir, "Library", "Application Support", "go-dev-mcp")
	default: // linux and others
		configDir = filepath.Join(homeDir, ".config", "go-dev-mcp")
	}

	// Create config directory if it doesn't exist
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// save saves the configuration to disk
func save(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
