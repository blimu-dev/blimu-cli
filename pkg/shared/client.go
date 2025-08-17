package shared

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	blimu "github.com/blimu-dev/blimu-go"
)

// GetSDKClient returns a configured SDK client using the current environment
func GetSDKClient() (*blimu.Client, error) {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load CLI config: %w", err)
	}

	currentEnv, err := cliConfig.GetCurrentEnvironment()
	if err != nil {
		return nil, fmt.Errorf("no current environment configured. Please configure an environment first")
	}

	// Determine API URL
	apiURL := currentEnv.APIURL
	if apiURL == "" {
		apiURL = cliConfig.DefaultAPIURL
	}

	// Create SDK client
	client := blimu.NewClient(
		blimu.WithBaseURL(apiURL),
		blimu.WithApiKeyAuth(currentEnv.APIKey),
	)
	return client, nil
}

// GetCurrentEnvironmentInfo returns the current environment configuration and metadata
func GetCurrentEnvironmentInfo() (*config.CLIConfig, *config.Environment, error) {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load CLI config: %w", err)
	}

	currentEnv, err := cliConfig.GetCurrentEnvironment()
	if err != nil {
		return nil, nil, fmt.Errorf("no current environment configured. Please configure an environment first")
	}

	return cliConfig, currentEnv, nil
}
