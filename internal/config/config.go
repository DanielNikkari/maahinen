package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Agent  AgentConfig  `yaml:"agent"`
	UI     UIConfig     `yaml:"ui"`
	Ollama OllamaConfig `yaml:"ollama"`
}

// AgentConfig contains agent-related configuration
type AgentConfig struct {
	SystemPrompt string `yaml:"system_prompt"`
	AutoConfirm  bool   `yaml:"auto_confirm"`
}

// UIConfig contains UI-related configuration
type UIConfig struct {
	SpinnerStyle string `yaml:"spinner_style"`
	ShowToolPanel bool  `yaml:"show_tool_panel"`
}

// OllamaConfig contains Ollama-related configuration
type OllamaConfig struct {
	BaseURL      string `yaml:"base_url"`
	DefaultModel string `yaml:"default_model"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			SystemPrompt: `You are Maahinen, a helpful coding assistant. You help users with programming tasks,
answer questions about code, and assist with debugging. Be concise and practical.
You should aim to take action, for example, when user asks you for example write code
you should utilize your tools to complete the user request. If a tool call fails, first try to fix the issue by recalling the tool with
corrected arguments. In case you run into a problem, try to iterate on the issue before returning a final response. However, if the issue
is not fixed in reasonable amount of tries, let the user know there is an issue.`,
			AutoConfirm: false,
		},
		UI: UIConfig{
			SpinnerStyle:  "dots",
			ShowToolPanel: true,
		},
		Ollama: OllamaConfig{
			BaseURL:      "http://localhost:11434",
			DefaultModel: "qwen2.5-coder:7b",
		},
	}
}

// Load loads configuration from a file
// If the file doesn't exist, it returns the default configuration
func Load(configPath string) (*Config, error) {
	// If config path is not provided, try default locations
	if configPath == "" {
		configPath = findConfigFile()
	}

	// If no config file exists, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to a file
func (c *Config) Save(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// findConfigFile looks for a config file in default locations
func findConfigFile() string {
	// Try these locations in order:
	// 1. ./config.yaml (current directory)
	// 2. ~/.maahinen/config.yaml (user home directory)
	// 3. ~/.config/maahinen/config.yaml (XDG config directory)

	locations := []string{
		"config.yaml",
	}

	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations,
			filepath.Join(home, ".maahinen", "config.yaml"),
			filepath.Join(home, ".config", "maahinen", "config.yaml"),
		)
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Default to current directory
	return "config.yaml"
}

// GetConfigPath returns the path where config should be saved
func GetConfigPath() string {
	// Check if config exists in current directory first
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	// Otherwise use user's home directory
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".maahinen", "config.yaml")
	}

	// Fallback to current directory
	return "config.yaml"
}
