package cli

import (
	"fmt"
	"sync"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	blimu "github.com/blimu-dev/blimu-go"
)

// Context represents the shared context for CLI commands
// Similar to kubectl's context pattern, this contains all the shared state
// that commands need to access
type Context struct {
	// Configuration
	CLIConfig   *config.CLIConfig
	BlimuConfig *config.BlimuConfig

	// API client - lazily initialized
	client     *blimu.Client
	clientOnce sync.Once
	clientErr  error

	// Current environment info
	currentEnvironment *config.Environment
	environmentName    string

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewContext creates a new CLI context
func NewContext() *Context {
	return &Context{}
}

// LoadCLIConfig loads the CLI configuration and sets it in the context
func (c *Context) LoadCLIConfig() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	c.CLIConfig = cliConfig
	return nil
}

// LoadBlimuConfig loads the .blimu configuration from the given directory
func (c *Context) LoadBlimuConfig(configDir string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	blimuConfig, err := config.LoadBlimuConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to load .blimu configuration: %w", err)
	}

	c.BlimuConfig = blimuConfig
	return nil
}

// GetClient returns the Blimu API client, initializing it if necessary
func (c *Context) GetClient() (*blimu.Client, error) {
	c.clientOnce.Do(func() {
		c.mu.RLock()
		defer c.mu.RUnlock()

		if c.CLIConfig == nil {
			c.clientErr = fmt.Errorf("CLI config not loaded")
			return
		}

		apiURL, apiKey, err := c.CLIConfig.GetAPIClient()
		if err != nil {
			c.clientErr = fmt.Errorf("failed to get API client config: %w", err)
			return
		}

		c.client = blimu.NewClient(
			blimu.WithBaseURL(apiURL),
			blimu.WithApiKeyAuth(apiKey),
		)
	})

	return c.client, c.clientErr
}

// GetCurrentEnvironment returns the current environment information
func (c *Context) GetCurrentEnvironment() (*config.Environment, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.CLIConfig == nil {
		return nil, "", fmt.Errorf("CLI config not loaded")
	}

	// Get current environment
	currentEnv, err := c.CLIConfig.GetCurrentEnvironment()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current environment: %w", err)
	}

	envName := c.CLIConfig.CurrentEnvironment
	if currentEnv != nil && currentEnv.Name != "" {
		envName = currentEnv.Name
	}

	return currentEnv, envName, nil
}

// SetCurrentEnvironment sets the current environment
func (c *Context) SetCurrentEnvironment(envName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.CLIConfig == nil {
		return fmt.Errorf("CLI config not loaded")
	}

	// Validate environment exists
	if _, exists := c.CLIConfig.Environments[envName]; !exists {
		return fmt.Errorf("environment '%s' not found", envName)
	}

	c.CLIConfig.CurrentEnvironment = envName
	return c.CLIConfig.Save()
}

// GetEnvironment returns a specific environment by name
func (c *Context) GetEnvironment(envName string) (*config.Environment, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.CLIConfig == nil {
		return nil, fmt.Errorf("CLI config not loaded")
	}

	env, exists := c.CLIConfig.Environments[envName]
	if !exists {
		return nil, fmt.Errorf("environment '%s' not found", envName)
	}

	return &env, nil
}

// ListEnvironments returns all available environments
func (c *Context) ListEnvironments() (map[string]config.Environment, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.CLIConfig == nil {
		return nil, fmt.Errorf("CLI config not loaded")
	}

	return c.CLIConfig.Environments, nil
}

// GetAPIConfig returns the API URL and key for the current environment
func (c *Context) GetAPIConfig() (string, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.CLIConfig == nil {
		return "", "", fmt.Errorf("CLI config not loaded")
	}

	return c.CLIConfig.GetAPIClient()
}

// GetAPIConfigForEnvironment returns the API URL and key for a specific environment
func (c *Context) GetAPIConfigForEnvironment(envName string) (string, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.CLIConfig == nil {
		return "", "", fmt.Errorf("CLI config not loaded")
	}

	env, exists := c.CLIConfig.Environments[envName]
	if !exists {
		return "", "", fmt.Errorf("environment '%s' not found", envName)
	}

	apiURL := env.APIURL
	if apiURL == "" {
		apiURL = c.CLIConfig.DefaultAPIURL
	}

	return apiURL, env.APIKey, nil
}

// IsEnvironmentSet returns true if a current environment is configured
func (c *Context) IsEnvironmentSet() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.CLIConfig != nil && c.CLIConfig.CurrentEnvironment != ""
}

// ValidateEnvironment validates that the current environment is properly configured
func (c *Context) ValidateEnvironment() error {
	if !c.IsEnvironmentSet() {
		return fmt.Errorf("no environment set. Use 'blimu env switch' to set an environment")
	}

	_, _, err := c.GetCurrentEnvironment()
	return err
}

// PrintEnvironmentInfo prints information about the current environment
func (c *Context) PrintEnvironmentInfo() error {
	currentEnv, envName, err := c.GetCurrentEnvironment()
	if err != nil {
		return err
	}

	fmt.Printf("🌍 Current Environment: %s\n", envName)
	if currentEnv != nil {
		if currentEnv.Name != "" {
			fmt.Printf("   Name: %s\n", currentEnv.Name)
		}
		if currentEnv.APIURL != "" {
			fmt.Printf("   API URL: %s\n", currentEnv.APIURL)
		}
	}

	return nil
}

// GetEnvironmentSummary returns a summary of the current environment for use in command output
func (c *Context) GetEnvironmentSummary() (string, error) {
	if !c.IsEnvironmentSet() {
		return "no environment set", nil
	}

	_, envName, err := c.GetCurrentEnvironment()
	if err != nil {
		return "", err
	}

	return envName, nil
}

// ValidateAndGetClient is a convenience method that validates environment and returns client
func (c *Context) ValidateAndGetClient() (*blimu.Client, string, error) {
	if err := c.ValidateEnvironment(); err != nil {
		return nil, "", err
	}

	client, err := c.GetClient()
	if err != nil {
		return nil, "", err
	}

	_, envName, err := c.GetCurrentEnvironment()
	if err != nil {
		return nil, "", err
	}

	return client, envName, nil
}
