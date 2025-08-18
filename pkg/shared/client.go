package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/blimu-dev/blimu-cli/internal/oauth"
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

	// Check if we have OAuth tokens
	if currentEnv.IsOAuthAuthenticated() {
		// Check if token needs refresh
		if currentEnv.NeedsTokenRefresh() && currentEnv.RefreshToken != "" {
			if err := refreshTokens(cliConfig, currentEnv, apiURL); err != nil {
				fmt.Printf("⚠️  Failed to refresh token: %v\n", err)
				fmt.Printf("Please run 'blimu auth login' to re-authenticate\n")
				return nil, err
			}
		}

		// Use Bearer token authentication
		client := blimu.NewClient(
			blimu.WithBaseURL(apiURL),
			blimu.WithHeaders(map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", currentEnv.AccessToken),
			}),
		)
		return client, nil
	}

	// Fallback to API key authentication
	if currentEnv.APIKey != "" {
		client := blimu.NewClient(
			blimu.WithBaseURL(apiURL),
			blimu.WithApiKeyAuth(currentEnv.APIKey),
		)
		return client, nil
	}

	return nil, fmt.Errorf("no valid authentication found. Please run 'blimu auth login' or configure an API key")
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
		TokenURL: fmt.Sprintf("%s/v1/%s/oauth/token", apiURL, env.Name),
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

	return cliConfig.AddEnvironment(env.Name, *env)
}
