package auth

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/blimu-dev/blimu-cli/internal/oauth"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/spf13/cobra"
)

// LoginCommand represents the login command
type LoginCommand struct {
	Environment string
	APIURL      string
}

// NewLoginCmd creates the login command
func NewLoginCmd() *cobra.Command {
	cmd := &LoginCommand{}

	cobraCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Blimu using OAuth",
		Long:  "Start the OAuth authentication flow to log in to your Blimu account",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.Run()
		},
	}

	cobraCmd.Flags().StringVar(&cmd.Environment, "environment", "env_blimu_platform", "Environment to authenticate with")
	cobraCmd.Flags().StringVar(&cmd.APIURL, "api-url", "", "Runtime API URL for OAuth (defaults to https://api.blimu.dev)")

	return cobraCmd
}

// Run executes the login command
func (c *LoginCommand) Run() error {
	// Load CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	// Determine runtime API URL for OAuth authentication
	runtimeURL := c.APIURL
	if runtimeURL == "" {
		runtimeURL = "https://api.blimu.dev" // Always use runtime API for OAuth
	}
	// Override if user explicitly set platform URL
	if runtimeURL == "https://platform-api.blimu.dev" {
		runtimeURL = "https://api.blimu.dev"
	}

	fmt.Printf("🔐 Starting OAuth authentication with %s...\n", runtimeURL)

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

	// Create OAuth client using runtime API
	oauthConfig := oauth.Config{
		ClientID:    "blimu_cli",
		AuthURL:     fmt.Sprintf("%s/v1/%s/oauth/authorize", runtimeURL, c.Environment),
		TokenURL:    fmt.Sprintf("%s/v1/%s/oauth/token", runtimeURL, c.Environment),
		RedirectURI: server.GetRedirectURI(),
		Scopes: []string{
			"workspace:read",
			"workspace:manage",
			"environment:read",
			"environment:manage",
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

	// Exchange code for tokens
	fmt.Printf("🔄 Exchanging authorization code for tokens...\n")

	tokenResp, err := oauthClient.ExchangeCodeForTokens(ctx, result.Code, pkce.Verifier)
	if err != nil {
		return fmt.Errorf("failed to exchange code for tokens: %w", err)
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Save tokens to config
	envConfig := config.Environment{
		Name:         c.Environment,
		APIURL:       "https://platform-api.blimu.dev", // Store platform URL for operations
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    &expiresAt,
		TokenType:    "Bearer",
	}

	if err := cliConfig.AddEnvironment(c.Environment, envConfig); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	fmt.Printf("✅ Authentication successful!\n")
	fmt.Printf("   Environment: %s\n", c.Environment)
	fmt.Printf("   OAuth via: %s\n", runtimeURL)
	fmt.Printf("   Operations via: https://platform-api.blimu.dev\n")
	fmt.Printf("   Token expires: %s\n", expiresAt.Format(time.RFC3339))

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
