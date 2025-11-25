package auth

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/blimu-dev/blimu-cli/internal/oauth"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	platform "github.com/blimu-dev/blimu-platform-go"
	"github.com/spf13/cobra"
)

// LoginCommand represents the login command
type LoginCommand struct {
	APIURL string
}

// NewLoginCmd creates the login command
func NewLoginCmd() *cobra.Command {
	cmd := &LoginCommand{}

	cobraCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Blimu using OAuth",
		Long:  "Start the OAuth authentication flow to log in to your Blimu account",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.Run(cobraCmd)
		},
	}

	cobraCmd.Flags().StringVar(&cmd.APIURL, "api-url", "", "Platform API URL for OAuth (defaults to https://platform-api.blimu.dev)")

	return cobraCmd
}

// Run executes the login command
func (c *LoginCommand) Run(cmd *cobra.Command) error {
	// Load CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	// Check if dev mode is enabled
	devMode, _ := cmd.Flags().GetBool("dev")

	// Use platform API OAuth endpoints (which proxy to Clerk internally)
	platformURL := "https://platform-api.blimu.dev"
	if devMode {
		platformURL = "http://localhost:3010"
	} else if c.APIURL != "" {
		platformURL = c.APIURL
	}

	fmt.Printf("🔐 Starting OAuth authentication via platform API...\n")

	// Create callback server
	server, err := oauth.NewCallbackServer()
	if err != nil {
		return fmt.Errorf("failed to create callback server: %w", err)
	}

	// Start callback server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}
	defer server.Shutdown(context.Background())

	// Show callback server info
	fmt.Printf("📡 Callback server started on port %d\n", server.GetPort())
	if server.GetPort() != 8080 {
		fmt.Printf("⚠️  Using alternative port %d (8080 was busy)\n", server.GetPort())
		fmt.Printf("   Make sure %s is configured in your OAuth app\n", server.GetRedirectURI())
	}

	// Generate PKCE challenge
	pkce, err := oauth.GeneratePKCEChallenge()
	if err != nil {
		return fmt.Errorf("failed to generate PKCE challenge: %w", err)
	}

	// Generate state parameter
	state, err := oauth.GenerateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	// Create OAuth client using platform API endpoints (which proxy to Clerk)
	oauthConfig := oauth.Config{
		ClientID:    "blimu_cli", // Platform API OAuth client ID
		AuthURL:     fmt.Sprintf("%s/oauth/authorize", platformURL),
		TokenURL:    fmt.Sprintf("%s/oauth/token", platformURL),
		RedirectURI: server.GetRedirectURI(),
		Scopes: []string{
			"openid",
			"profile",
			"email",
		},
	}

	oauthClient := oauth.NewClient(oauthConfig)

	// Generate authorization URL
	authURL := oauthClient.GetAuthorizationURL(state, pkce.Challenge)

	// Open browser
	fmt.Printf("🌐 Opening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open automatically, visit: %s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("⚠️  Failed to open browser automatically: %v\n", err)
		fmt.Printf("Please manually visit the URL above.\n\n")
	}

	fmt.Printf("⏳ Waiting for authentication callback...\n")

	// Wait for callback
	result, err := server.WaitForCallback(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive callback: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("authentication failed: %s", result.Error)
	}

	if result.State != state {
		return fmt.Errorf("invalid state parameter")
	}

	fmt.Printf("✅ Received authorization callback\n")

	// Exchange code for tokens
	fmt.Printf("🔄 Exchanging authorization code for tokens...\n")

	tokenResp, err := oauthClient.ExchangeCodeForTokens(ctx, result.Code, pkce.Verifier)
	if err != nil {
		return fmt.Errorf("failed to exchange code for tokens: %w", err)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Platform API URL is already determined above
	// No need to redetermine it here since we're using platform API throughout

	// Create initial environment config
	envConfig := config.Environment{
		APIURL:       platformURL,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    &expiresAt,
		TokenType:    "Bearer",
	}

	// Try to fetch workspace and environment information using the new token
	fmt.Printf("🔍 Fetching workspace and environment information...\n")
	if workspaceID, environmentID, err := fetchUserWorkspaceAndEnvironment(tokenResp.AccessToken, platformURL); err != nil {
		fmt.Printf("⚠️  Could not fetch workspace/environment information: %v\n", err)
		return fmt.Errorf("failed to fetch workspace/environment information: %w", err)
	} else {
		if workspaceID != "" {
			envConfig.WorkspaceID = workspaceID
			fmt.Printf("📋 Found workspace ID: %s\n", workspaceID)
		} else {
			return fmt.Errorf("failed to fetch workspace information: %w", err)
		}

		if environmentID != "" {
			envConfig.ID = environmentID
			fmt.Printf("📋 Found environment ID: %s\n", environmentID)
		} else {
			return fmt.Errorf("failed to fetch environment information: %w", err)
		}
	}

	if err := cliConfig.AddEnvironment(envConfig); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	fmt.Printf("✅ OAuth authentication successful!\n")
	fmt.Printf("   Environment: %s\n", envConfig.ID)
	fmt.Printf("   Platform API: %s\n", platformURL)
	if envConfig.WorkspaceID != "" {
		fmt.Printf("   Workspace ID: %s\n", envConfig.WorkspaceID)
	}
	if envConfig.ID != "" {
		fmt.Printf("   Environment ID: %s\n", envConfig.ID)
	}
	fmt.Printf("   Token expires: %s\n", expiresAt.Format(time.RFC3339))

	// Show available environments
	fmt.Printf("\n🌍 Fetching your available environments...\n")
	if environments, err := shared.FetchUserEnvironments(devMode); err != nil {
		fmt.Printf("⚠️  Could not fetch environments: %v\n", err)
	} else if len(environments) > 1 {
		fmt.Printf("\nYou have access to %d environments:\n", len(environments))
		shared.DisplayEnvironments(environments)
		fmt.Printf("\nUse 'blimu env switch' to switch between environments.\n")
	} else if len(environments) == 1 {
		fmt.Printf("You have access to 1 environment: %s\n", environments[0].Name)
	}

	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// fetchUserWorkspaceAndEnvironment attempts to fetch the user's workspace and environment IDs using the access token
func fetchUserWorkspaceAndEnvironment(accessToken, platformURL string) (workspaceID, environmentID string, err error) {
	// Create a temporary platform client with the new access token
	client := platform.NewClient(
		platform.WithBaseURL(platformURL),
		platform.WithBearer(accessToken),
	)

	// Get user's active resources
	activeResources, err := client.Me.GetActiveResources()
	if err != nil {
		return "", "", fmt.Errorf("failed to get active resources: %w", err)
	}

	fmt.Printf("🔍 Found %d active resources\n", len(activeResources))

	if len(activeResources) == 0 {
		return "", "", fmt.Errorf("no active resources found for user")
	}

	// Look for workspace and environment resources in the active resources
	for i, resource := range activeResources {
		fmt.Printf("   Resource %d: Role=%s, Inherited=%t\n", i+1, resource.Role, resource.Inherited)

		if resourceData, ok := resource.Resource.(map[string]interface{}); ok {
			// Print the resource data for debugging
			fmt.Printf("   Resource data: %+v\n", resourceData)

			// Check resource type
			if resourceType, exists := resourceData["type"]; exists {
				fmt.Printf("   Resource type: %v\n", resourceType)

				// Look for environment
				if resourceType == "environment" && environmentID == "" {
					if id, exists := resourceData["id"]; exists {
						if idStr, ok := id.(string); ok {
							environmentID = idStr
							fmt.Printf("   ✅ Found environment ID: %s\n", environmentID)
						}
					}
					// Also check for workspace ID in environment resource
					if wsID, exists := resourceData["workspaceId"]; exists {
						if idStr, ok := wsID.(string); ok {
							workspaceID = idStr
							fmt.Printf("   ✅ Found workspace ID from environment: %s\n", workspaceID)
						}
					}
				}

				// Look for workspace
				if resourceType == "workspace" && workspaceID == "" {
					if id, exists := resourceData["id"]; exists {
						if idStr, ok := id.(string); ok {
							workspaceID = idStr
							fmt.Printf("   ✅ Found workspace ID: %s\n", workspaceID)
						}
					}
				}
			}

		}
	}

	// Return what we found, even if incomplete
	fmt.Printf("🔍 Final results: workspaceID='%s', environmentID='%s'\n", workspaceID, environmentID)

	if workspaceID == "" && environmentID == "" {
		return "", "", fmt.Errorf("no workspace or environment found in active resources")
	}

	return workspaceID, environmentID, nil
}
