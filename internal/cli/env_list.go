package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/blimu-dev/blimu-cli/pkg/api"
	"github.com/spf13/cobra"
)

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List environments from API",
	Long:  `List all environments from the API and show which one is currently active locally`,
	RunE:  runEnvList,
}

func runEnvList(cmd *cobra.Command, args []string) error {
	cliConfig, currentEnv, err := getCurrentEnvironmentInfo()
	if err != nil {
		fmt.Println("No current environment configured.")
		fmt.Println("Use 'blimu env create <name>' to create an environment.")
		return nil
	}

	// Get auth client for API operations
	authClient, err := getAuthClient()
	if err != nil {
		return err
	}

	apiClient := api.NewClient(authClient)

	// Fetch environments from API
	apiEnvironments, err := apiClient.ListEnvironments()
	if err != nil {
		return fmt.Errorf("failed to fetch environments from API: %w", err)
	}

	if len(apiEnvironments.Data) == 0 {
		fmt.Println("No environments found in your workspace.")
		fmt.Println("Create environments via the Blimu dashboard or API.")
		return nil
	}

	fmt.Printf("Available environments:\n\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tLOOKUP KEY\tWORKSPACE ID\tCREATED\tCURRENT")

	for _, env := range apiEnvironments.Data {
		current := ""
		// Check if this environment ID matches the current one in config
		if currentEnv.ID == env.Id {
			current = "✓"
		}

		name := env.Name
		lookupKey := "-"
		if env.LookupKey != nil {
			lookupKey = *env.LookupKey
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			env.Id, name, lookupKey, env.WorkspaceId, env.CreatedAt, current)
	}

	w.Flush()

	// Show local configuration info
	fmt.Printf("\nLocal configuration:\n")
	fmt.Printf("  Current environment: %s\n", cliConfig.CurrentEnvironment)
	fmt.Printf("  API URL: %s\n", func() string {
		if currentEnv.APIURL != "" {
			return currentEnv.APIURL
		}
		return cliConfig.DefaultAPIURL
	}())

	return nil
}
