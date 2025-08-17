package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var envCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current environment",
	Long:  `Show the currently active environment`,
	RunE:  runEnvCurrent,
}

func runEnvCurrent(cmd *cobra.Command, args []string) error {
	cliConfig, currentEnv, err := getCurrentEnvironmentInfo()
	if err != nil {
		fmt.Println("No current environment set.")
		fmt.Println("Use 'blimu env create <name>' to create an environment.")
		return nil
	}

	// Get SDK client
	sdk, err := getSDKClient()
	if err != nil {
		return err
	}

	// Fetch environment details from API if we have an ID
	if currentEnv.ID != "" {
		apiEnv, err := sdk.Environments.Get(currentEnv.ID)
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
			fmt.Printf("  API URL: %s\n", func() string {
				if currentEnv.APIURL != "" {
					return currentEnv.APIURL
				}
				return cliConfig.DefaultAPIURL
			}())

			return nil
		}
	}

	// Fallback: show local configuration only
	fmt.Printf("Current environment: %s\n", cliConfig.CurrentEnvironment)
	fmt.Printf("  API URL: %s\n", func() string {
		if currentEnv.APIURL != "" {
			return currentEnv.APIURL
		}
		return cliConfig.DefaultAPIURL
	}())

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
