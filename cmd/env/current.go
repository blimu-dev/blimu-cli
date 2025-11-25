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

	// Show local configuration
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
	}

	if currentEnv.WorkspaceID != "" {
		fmt.Printf("  Workspace ID: %s\n", currentEnv.WorkspaceID)
	}

	// Show authentication status
	if currentEnv.IsOAuthAuthenticated() {
		fmt.Printf("  Authentication: OAuth")
		if currentEnv.ExpiresAt != nil {
			fmt.Printf(" (expires: %s)", currentEnv.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("\n")
	} else {
		fmt.Printf("  Authentication: None (run 'blimu auth login')\n")
	}

	// Note about platform API requirements
	if currentEnv.ID != "" {
		fmt.Printf("\n⚠️  Note: Platform API environment details require workspace ID.\n")
		fmt.Printf("Use 'blimu env create --workspace-id <id>' for full API integration.\n")
	}

	return nil
}
