package env

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// NewCurrentCmd creates the current command
func NewCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current environment",
		Long:  `Show the currently active environment`,
		RunE:  runCurrent,
	}
}

func runCurrent(cmd *cobra.Command, args []string) error {
	// Get current environment info
	cliConfig, currentEnv, err := shared.GetCurrentEnvironmentInfo()
	if err != nil {
		fmt.Println("No current environment set.")
		fmt.Println("Use 'blimu env create <name>' to create an environment.")
		return nil
	}

	// Get SDK client
	client, err := shared.GetSDKClient()
	if err != nil {
		return fmt.Errorf("failed to get API client: %w", err)
	}

	// Fetch environment details from API if we have an ID
	if currentEnv.ID != "" {
		apiEnv, err := client.Environments.Get(currentEnv.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch environment details from API: %v\n\n", err)
			// Fall back to showing local config only
		} else {
			// Show API data (most up-to-date)
			fmt.Printf("Current environment: %s\n", cliConfig.CurrentEnvironment)
			fmt.Printf("  Name: %s\n", apiEnv.Name)
			fmt.Printf("  ID: %s\n", apiEnv.Id)
			if apiEnv.LookupKey != nil && *apiEnv.LookupKey != "" {
				fmt.Printf("  Lookup Key: %s\n", *apiEnv.LookupKey)
			}
			fmt.Printf("  Workspace ID: %s\n", apiEnv.WorkspaceId)
			fmt.Printf("  Created: %s\n", apiEnv.CreatedAt)
			fmt.Printf("  Updated: %s\n", apiEnv.UpdatedAt)

			// Show local configuration
			fmt.Printf("\nLocal configuration:\n")
			apiURL := currentEnv.APIURL
			if apiURL == "" {
				apiURL = cliConfig.DefaultAPIURL
			}
			fmt.Printf("  API URL: %s\n", apiURL)

			return nil
		}
	}

	// Fallback: show local configuration only
	fmt.Printf("Current environment: %s\n", cliConfig.CurrentEnvironment)
	apiURL := currentEnv.APIURL
	if apiURL == "" {
		apiURL = cliConfig.DefaultAPIURL
	}
	fmt.Printf("  API URL: %s\n", apiURL)

	if currentEnv.LookupKey != "" {
		fmt.Printf("  Lookup Key: %s\n", currentEnv.LookupKey)
	}

	if currentEnv.ID != "" {
		fmt.Printf("  Environment ID: %s\n", currentEnv.ID)
	} else {
		fmt.Printf("  Warning: No environment ID stored locally. Run 'blimu env list' to sync with API.\n")
	}

	return nil
}
