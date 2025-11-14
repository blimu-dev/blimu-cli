package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// CLIConfig represents the CLI configuration
type CLIConfig struct {
	CurrentEnvironment string                 `yaml:"current_environment,omitempty"`
	Environments       map[string]Environment `yaml:"environments,omitempty"`
	DefaultAPIURL      string                 `yaml:"default_api_url,omitempty"`
}

// Environment represents a single environment configuration
type Environment struct {
	Name      string `yaml:"name"`
	APIKey    string `yaml:"api_key,omitempty"` // Keep for backward compatibility
	APIURL    string `yaml:"api_url,omitempty"`
	ID        string `yaml:"id,omitempty"`         // Environment ID from the API
	LookupKey string `yaml:"lookup_key,omitempty"` // Optional lookup key for the environment

	// New OAuth fields
	AccessToken  string     `yaml:"access_token,omitempty"`
	RefreshToken string     `yaml:"refresh_token,omitempty"`
	ExpiresAt    *time.Time `yaml:"expires_at,omitempty"`
	TokenType    string     `yaml:"token_type,omitempty"`
}

// GetCLIConfigPath returns the path to the CLI configuration file
func GetCLIConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".blimu")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.yml"), nil
}

// LoadCLIConfig loads CLI configuration from file and environment variables
func LoadCLIConfig() (*CLIConfig, error) {
	config := &CLIConfig{
		Environments:  make(map[string]Environment),
		DefaultAPIURL: "https://api.blimu.dev", // Runtime API for OAuth, platform API determined at runtime
	}

	// Try to load from config file
	configPath, err := GetCLIConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CLI config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse CLI config file: %w", err)
		}
	}

	// Override with environment variables if present
	if apiKey := os.Getenv("BLIMU_SECRET_KEY"); apiKey != "" {
		envName := "default"
		if config.CurrentEnvironment == "" {
			config.CurrentEnvironment = envName
		}

		env := config.Environments[envName]
		env.Name = envName
		env.APIKey = apiKey

		if apiURL := os.Getenv("BLIMU_API_URL"); apiURL != "" {
			env.APIURL = apiURL
		} else if env.APIURL == "" {
			env.APIURL = config.DefaultAPIURL
		}

		config.Environments[envName] = env
	}

	// Set default current environment if none set
	if config.CurrentEnvironment == "" && len(config.Environments) > 0 {
		for name := range config.Environments {
			config.CurrentEnvironment = name
			break
		}
	}

	return config, nil
}

// SaveCLIConfig saves the CLI configuration to file
func (c *CLIConfig) Save() error {
	configPath, err := GetCLIConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal CLI config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write CLI config file: %w", err)
	}

	return nil
}

// GetCurrentEnvironment returns the current environment configuration
func (c *CLIConfig) GetCurrentEnvironment() (*Environment, error) {
	if c.CurrentEnvironment == "" {
		return nil, fmt.Errorf("no current environment set")
	}

	env, exists := c.Environments[c.CurrentEnvironment]
	if !exists {
		return nil, fmt.Errorf("current environment '%s' not found", c.CurrentEnvironment)
	}

	return &env, nil
}

// SetCurrentEnvironment sets the current environment
func (c *CLIConfig) SetCurrentEnvironment(name string) error {
	if _, exists := c.Environments[name]; !exists {
		return fmt.Errorf("environment '%s' not found", name)
	}

	c.CurrentEnvironment = name
	return c.Save()
}

// AddEnvironment adds or updates an environment
func (c *CLIConfig) AddEnvironment(name string, env Environment) error {
	if c.Environments == nil {
		c.Environments = make(map[string]Environment)
	}

	env.Name = name
	c.Environments[name] = env

	// Set as current if it's the first environment
	if c.CurrentEnvironment == "" {
		c.CurrentEnvironment = name
	}

	return c.Save()
}

// RemoveEnvironment removes an environment
func (c *CLIConfig) RemoveEnvironment(name string) error {
	if _, exists := c.Environments[name]; !exists {
		return fmt.Errorf("environment '%s' not found", name)
	}

	delete(c.Environments, name)

	// If we removed the current environment, switch to another one
	if c.CurrentEnvironment == name {
		c.CurrentEnvironment = ""
		for envName := range c.Environments {
			c.CurrentEnvironment = envName
			break
		}
	}

	return c.Save()
}

// ListEnvironments returns all configured environments
func (c *CLIConfig) ListEnvironments() map[string]Environment {
	return c.Environments
}

// GetAPIClient returns the API configuration for the current environment
func (c *CLIConfig) GetAPIClient() (apiURL, apiKey string, err error) {
	env, err := c.GetCurrentEnvironment()
	if err != nil {
		return "", "", err
	}

	apiURL = env.APIURL
	if apiURL == "" {
		apiURL = c.DefaultAPIURL
	}

	if env.APIKey == "" {
		return "", "", fmt.Errorf("no API key configured for environment '%s'", c.CurrentEnvironment)
	}

	return apiURL, env.APIKey, nil
}

// NeedsTokenRefresh checks if the OAuth token needs refresh
func (e *Environment) NeedsTokenRefresh() bool {
	if e.AccessToken == "" || e.ExpiresAt == nil {
		return false
	}
	// Refresh if token expires in the next 5 minutes
	return time.Until(*e.ExpiresAt) < 5*time.Minute
}

// IsOAuthAuthenticated checks if environment uses OAuth authentication
func (e *Environment) IsOAuthAuthenticated() bool {
	return e.AccessToken != "" && e.TokenType == "Bearer"
}
