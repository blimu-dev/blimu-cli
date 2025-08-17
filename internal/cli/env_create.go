package cli

import (
	"fmt"
	"os"

	"github.com/blimu-dev/blimu-cli/pkg/auth"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	blimu "github.com/blimu-dev/blimu-go"
	"github.com/spf13/cobra"
)

var envCreateCmd = &cobra.Command{
	Use:   "create <environment-name> [lookup-key]",
	Short: "Create a new environment",
	Long: `Create a new environment configuration.

If this is your first environment, you'll need to provide API credentials.
For subsequent environments, credentials from the current environment will be reused.

Examples:
  # First environment - provide credentials
  blimu env create production --api-key sk_prod_... --api-url https://api.blimu.dev
  
  # Additional environments - reuse current credentials  
  blimu env create staging staging-key`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runEnvCreate,
}

func init() {
	envCreateCmd.Flags().String("api-key", "", "API key for the environment")
	envCreateCmd.Flags().String("api-url", "", "API URL for the environment (defaults to https://api.blimu.dev)")
}

func runEnvCreate(cmd *cobra.Command, args []string) error {
	envName := args[0]
	lookupKey := ""
	if len(args) > 1 {
		lookupKey = args[1]
	}

	// Load CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	// Determine API credentials
	var apiKey, apiURL string

	// Check if we have a current environment to reuse credentials from
	if cliConfig.CurrentEnvironment != "" {
		currentEnv, err := cliConfig.GetCurrentEnvironment()
		if err == nil && currentEnv.APIKey != "" {
			// Reuse credentials from current environment
			apiKey = currentEnv.APIKey
			apiURL = currentEnv.APIURL
		}
	}

	// If no credentials from current environment, get from flags/env vars
	if apiKey == "" {
		flagAPIKey, _ := cmd.Flags().GetString("api-key")
		flagAPIURL, _ := cmd.Flags().GetString("api-url")

		if flagAPIKey != "" {
			apiKey = flagAPIKey
		} else {
			apiKey = os.Getenv("BLIMU_SECRET_KEY")
		}

		if flagAPIURL != "" {
			apiURL = flagAPIURL
		} else if apiURL == "" {
			apiURL = cliConfig.DefaultAPIURL
		}
	}

	if apiKey == "" {
		return fmt.Errorf("API key is required. Provide it via --api-key flag or BLIMU_SECRET_KEY environment variable")
	}

	// Create SDK client
	authClient := auth.NewClient(apiURL, apiKey)
	sdk := authClient.GetSDK()

	// Create environment via API
	createRequest := blimu.EnvironmentCreateDto{
		Name:      envName,
		LookupKey: lookupKey,
	}

	createdEnv, err := sdk.Environments.Create(createRequest)
	if err != nil {
		return fmt.Errorf("failed to create environment via API: %w", err)
	}

	// Add the created environment to local CLI config
	env := config.Environment{
		Name:      envName,
		APIKey:    apiKey,        // Use the API key we determined
		APIURL:    apiURL,        // Use the API URL we determined
		ID:        createdEnv.Id, // Store the API-generated ID
		LookupKey: lookupKey,     // Store the lookup key
	}

	if err := cliConfig.AddEnvironment(envName, env); err != nil {
		return fmt.Errorf("failed to add environment to local config: %w", err)
	}

	fmt.Printf("✅ Created environment '%s' (ID: %s)\n", envName, createdEnv.Id)
	if lookupKey != "" {
		fmt.Printf("   Lookup key: %s\n", lookupKey)
	}
	fmt.Printf("   Workspace ID: %s\n", createdEnv.WorkspaceId)

	// If this is the first environment, mention it's now current
	if cliConfig.CurrentEnvironment == envName {
		fmt.Printf("   Set as current environment\n")
	}

	return nil
}
