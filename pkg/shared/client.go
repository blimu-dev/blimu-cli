package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/blimu-dev/blimu-cli/internal/oauth"
	"github.com/blimu-dev/blimu-cli/pkg/auth"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	platform "github.com/blimu-dev/blimu-platform-go"
	// runtime "github.com/blimu-dev/blimu-go" // Will be used for token refresh
)

// GetSDKClient returns a configured platform SDK client using the current environment
func GetSDKClient() (*platform.Client, error) {
	return GetSDKClientWithDevMode(false)
}

// GetSDKClientWithDevMode returns a configured platform SDK client with optional dev mode
func GetSDKClientWithDevMode(devMode bool) (*platform.Client, error) {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load CLI config: %w", err)
	}

	currentEnv, err := cliConfig.GetCurrentEnvironment()
	if err != nil {
		return nil, fmt.Errorf("no current environment configured. Please configure an environment first")
	}

	// Determine platform API URL
	platformURL := "https://platform-api.blimu.dev"
	if devMode {
		platformURL = "http://localhost:3010"
	} else if currentEnv.APIURL != "" && currentEnv.APIURL != "https://api.blimu.dev" {
		// If user has custom platform URL configured
		platformURL = currentEnv.APIURL
	}

	// Check if we have Clerk OAuth tokens
	if currentEnv.IsOAuthAuthenticated() {
		// Check if token needs refresh
		if currentEnv.NeedsTokenRefresh() && currentEnv.RefreshToken != "" {
			// TODO: Implement Clerk token refresh
			fmt.Printf("⚠️  Token refresh not yet implemented for Clerk OAuth.\n")
			fmt.Printf("Please run 'blimu auth login' to re-authenticate\n")
			return nil, fmt.Errorf("token expired, please re-authenticate")
		}

		// Use Clerk JWT token with platform SDK
		client := platform.NewClient(
			platform.WithBaseURL(platformURL),
			platform.WithBearer(currentEnv.AccessToken),
		)
		return client, nil
	}

	return nil, fmt.Errorf("no valid authentication found. Please run 'blimu auth login' to authenticate")
}

// GetAuthClient returns a configured auth client using the current environment
func GetAuthClient() (*auth.Client, error) {
	return GetAuthClientWithDevMode(false)
}

// GetAuthClientWithDevMode returns a configured auth client with optional dev mode
func GetAuthClientWithDevMode(devMode bool) (*auth.Client, error) {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load CLI config: %w", err)
	}

	currentEnv, err := cliConfig.GetCurrentEnvironment()
	if err != nil {
		return nil, fmt.Errorf("no current environment configured. Please configure an environment first")
	}

	// Determine platform API URL
	platformURL := "https://platform-api.blimu.dev"
	if devMode {
		platformURL = "http://localhost:3010"
	} else if currentEnv.APIURL != "" && currentEnv.APIURL != "https://api.blimu.dev" {
		// If user has custom platform URL configured
		platformURL = currentEnv.APIURL
	}

	// Check if we have Clerk OAuth tokens
	if currentEnv.IsOAuthAuthenticated() {
		// Check if token needs refresh
		if currentEnv.NeedsTokenRefresh() && currentEnv.RefreshToken != "" {
			// TODO: Implement Clerk token refresh
			fmt.Printf("⚠️  Token refresh not yet implemented for Clerk OAuth.\n")
			fmt.Printf("Please run 'blimu auth login' to re-authenticate\n")
			return nil, fmt.Errorf("token expired, please re-authenticate")
		}

		// Create client with Clerk token for platform operations
		return auth.NewClientWithClerkToken(platformURL, currentEnv.AccessToken), nil
	}

	return nil, fmt.Errorf("no valid authentication found. Please run 'blimu auth login' to authenticate")
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

// refreshTokens handles OAuth token refresh
func refreshTokens(cliConfig *config.CLIConfig, env *config.Environment, apiURL string) error {
	oauthConfig := oauth.Config{
		ClientID: "blimu_cli",
		TokenURL: fmt.Sprintf("%s/v1/%s/oauth/token", apiURL, env.ID),
	}

	oauthClient := oauth.NewClient(oauthConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tokenResp, err := oauthClient.RefreshToken(ctx, env.RefreshToken)
	if err != nil {
		return err
	}

	// Update environment with new tokens
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	env.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		env.RefreshToken = tokenResp.RefreshToken
	}
	env.ExpiresAt = &expiresAt

	return cliConfig.AddEnvironment(*env)
}
