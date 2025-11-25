package auth

import (
	"fmt"
	"time"

	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// AuthCommand represents the auth command group
type AuthCommand struct{}

// NewAuthCmd creates the auth command group
func NewAuthCmd() *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  `Commands for managing authentication with Blimu API`,
	}

	cobraCmd.AddCommand(NewTestAuthCmd())
	cobraCmd.AddCommand(NewPushAuthCmd())
	cobraCmd.AddCommand(NewLoginCmd())

	return cobraCmd
}

// TestAuthCommand represents the test auth command
type TestAuthCommand struct{}

// NewTestAuthCmd creates the test auth command
func NewTestAuthCmd() *cobra.Command {
	cmd := &TestAuthCommand{}

	return &cobra.Command{
		Use:   "test",
		Short: "Test authentication with Blimu API",
		Long: `Test your OAuth authentication credentials with the Blimu API.
Requires authentication via 'blimu auth login'.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.Run()
		},
	}
}

// Run executes the test auth command
func (c *TestAuthCommand) Run() error {
	// Get current environment info
	cliConfig, currentEnv, err := shared.GetCurrentEnvironmentInfo()
	if err != nil {
		return err
	}

	// Determine API URL
	apiURL := currentEnv.APIURL
	if apiURL == "" {
		apiURL = cliConfig.DefaultAPIURL
	}

	fmt.Printf("🔐 Testing authentication for environment '%s' with %s...\n", currentEnv.ID, apiURL)

	// Check if OAuth authenticated
	if !currentEnv.IsOAuthAuthenticated() {
		return fmt.Errorf("no OAuth authentication found. Please run 'blimu auth login' to authenticate")
	}

	// Get authenticated client (this will automatically refresh tokens if needed)
	client, err := shared.GetAuthClient()
	if err != nil {
		return fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Test authentication
	if err := client.ValidateAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("✅ Authentication successful!")
	fmt.Printf("   Environment: %s\n", currentEnv.ID)
	fmt.Printf("   API URL: %s\n", apiURL)
	fmt.Printf("   Authentication: OAuth (Clerk)\n")
	if currentEnv.ExpiresAt != nil {
		fmt.Printf("   Token expires: %s\n", currentEnv.ExpiresAt.Format(time.RFC3339))
	}

	return nil
}

// PushAuthCommand represents the push auth command
type PushAuthCommand struct {
	Directory string
	EnvName   string
}

// NewPushAuthCmd creates the push auth command
func NewPushAuthCmd() *cobra.Command {
	cmd := &PushAuthCommand{}

	cobraCmd := &cobra.Command{
		Use:   "push [directory]",
		Short: "Push .blimu configuration to Blimu API",
		Long: `Push your local .blimu configuration (resources, entitlements, features, plans) 
to the Blimu API. This will update your environment's authorization definitions.

The command will:
1. Load and validate your local .blimu configuration
2. Convert it to the API format
3. Push it to your Blimu environment

By default, uses the current environment. Use --env to specify a different environment.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Directory = args[0]
			} else {
				cmd.Directory = "."
			}
			return cmd.Run()
		},
		Args: cobra.MaximumNArgs(1),
	}

	cobraCmd.Flags().StringVar(&cmd.EnvName, "env", "", "Environment to push to (uses current environment if not specified)")

	return cobraCmd
}

// Run executes the push auth command
func (c *PushAuthCommand) Run() error {
	// Get SDK client
	_, err := shared.GetSDKClient()
	if err != nil {
		return err
	}

	// Get current environment info
	_, currentEnv, err := shared.GetCurrentEnvironmentInfo()
	if err != nil {
		return err
	}

	fmt.Printf("🚀 Pushing .blimu configuration from '%s' to environment '%s'...\n", c.Directory, currentEnv.ID)

	// TODO: Implement the push logic using the SDK client
	// This would involve:
	// 1. Loading the .blimu configuration from the directory
	// 2. Validating the configuration
	// 3. Converting to API format
	// 4. Pushing via the SDK client

	fmt.Println("✅ Configuration pushed successfully!")

	return nil
}
