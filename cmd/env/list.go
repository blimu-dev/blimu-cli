package env

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// ListCommand represents the list environments command
type ListCommand struct {
	WorkspaceID string
}

// NewListCmd creates the list command
func NewListCmd() *cobra.Command {
	cmd := &ListCommand{}

	cobraCmd := &cobra.Command{
		Use:   "list",
		Short: "List environments from API",
		Long:  `List all environments from the API and show which one is currently active locally`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.Run()
		},
	}

	cobraCmd.Flags().StringVar(&cmd.WorkspaceID, "workspace-id", "", "Workspace ID (required for platform API)")

	return cobraCmd
}

func (c *ListCommand) Run() error {
	cliConfig, currentEnv, err := shared.GetCurrentEnvironmentInfo()
	if err != nil {
		fmt.Println("No current environment configured.")
		fmt.Println("Use 'blimu env create <name>' to create an environment.")
		return nil
	}

	// Check if workspace ID is provided
	if c.WorkspaceID == "" {
		fmt.Printf("⚠️  Workspace ID is required for listing environments.\n")
		fmt.Printf("Use --workspace-id flag or run 'blimu workspaces list' to find your workspace ID.\n")
		fmt.Printf("Showing local environments only:\n\n")

		// Show local environments
		if len(cliConfig.Environments) == 0 {
			fmt.Println("No local environments configured.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCURRENT\tAUTH\tAPI URL")

		for name, env := range cliConfig.Environments {
			current := ""
			if name == cliConfig.CurrentEnvironment {
				current = "*"
			}

			authType := "None"
			if env.IsOAuthAuthenticated() {
				authType = "OAuth"
			}

			apiURL := env.APIURL
			if apiURL == "" {
				apiURL = cliConfig.DefaultAPIURL
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, current, authType, apiURL)
		}

		w.Flush()
		return nil
	}

	// Get platform SDK client
	client, err := shared.GetSDKClient()
	if err != nil {
		return fmt.Errorf("failed to get API client: %w", err)
	}

	// Fetch environments from platform API
	apiEnvironments, err := client.Environments.List(c.WorkspaceID, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch environments from API: %w", err)
	}

	if len(apiEnvironments.Data) == 0 {
		fmt.Printf("No environments found in workspace %s.\n", c.WorkspaceID)
		fmt.Println("Create environments via the Blimu dashboard or 'blimu env create'.")
		return nil
	}

	// Display environments in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tID\tLOOKUP KEY\tWORKSPACE ID\tCREATED")

	for _, envData := range apiEnvironments.Data {
		// Extract fields from map[string]interface{}
		name := getStringFromMap(envData, "name")
		id := getStringFromMap(envData, "id")
		lookupKey := getStringFromMap(envData, "lookupKey")
		workspaceId := getStringFromMap(envData, "workspaceId")
		createdAt := getStringFromMap(envData, "createdAt")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			name,
			id,
			lookupKey,
			workspaceId,
			createdAt,
		)
	}

	w.Flush()

	// Show current local environment
	fmt.Printf("\nCurrent local environment: %s\n", cliConfig.CurrentEnvironment)
	if currentEnv != nil && currentEnv.ID != "" {
		fmt.Printf("Local environment ID: %s\n", currentEnv.ID)
	}

	return nil
}

// getStringFromMap safely extracts a string value from a map[string]interface{}
func getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
