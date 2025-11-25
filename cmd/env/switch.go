package env

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// SwitchCommand represents the switch environment command
type SwitchCommand struct {
	EnvName string
}

// NewSwitchCmd creates the switch command
func NewSwitchCmd() *cobra.Command {
	cmd := &SwitchCommand{}

	cobraCmd := &cobra.Command{
		Use:   "switch [environment-name]",
		Short: "Switch to a different environment",
		Long: `Switch the current active environment to the specified environment.

If no environment name is provided, you'll be prompted to select from available environments.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.EnvName = args[0]
			}
			// Check if dev mode is enabled
			devMode, _ := cobraCmd.Flags().GetBool("dev")
			return cmd.Run(devMode)
		},
	}

	return cobraCmd
}

// Run executes the switch environment command
func (c *SwitchCommand) Run(devMode bool) error {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	var targetEnvName string

	// If no environment name provided, show selection
	if c.EnvName == "" {
		fmt.Println("🔍 Fetching available environments...")

		environments, err := shared.FetchUserEnvironments(devMode)
		if err != nil {
			return fmt.Errorf("failed to fetch environments: %w", err)
		}

		if len(environments) == 0 {
			fmt.Println("No environments found. Use 'blimu auth login' to authenticate first.")
			return nil
		}

		selectedEnv, err := shared.PromptEnvironmentSelection(environments)
		if err != nil {
			return fmt.Errorf("failed to select environment: %w", err)
		}

		targetEnvName = selectedEnv.Name

		// If it's a remote environment not in local config, we need to add it
		if !selectedEnv.IsLocal {
			fmt.Printf("📥 Adding remote environment '%s' to local configuration...\n", selectedEnv.Name)

			// Create environment config (will need authentication later)
			envConfig := config.Environment{
				ID:          selectedEnv.ID,
				WorkspaceID: selectedEnv.WorkspaceID,
				APIURL:      "", // Will be set when user authenticates
			}

			if err := cliConfig.AddEnvironment(envConfig); err != nil {
				return fmt.Errorf("failed to add environment to local config: %w", err)
			}
		}
	} else {
		targetEnvName = c.EnvName
	}

	if err := cliConfig.SetCurrentEnvironment(targetEnvName); err != nil {
		return err
	}

	fmt.Printf("✅ Switched to environment '%s'\n", targetEnvName)

	// Show current environment info
	if env, exists := cliConfig.Environments[targetEnvName]; exists {
		if env.ID != "" {
			fmt.Printf("   Environment ID: %s\n", env.ID)
		}
		if env.WorkspaceID != "" {
			fmt.Printf("   Workspace ID: %s\n", env.WorkspaceID)
		}
		if !env.IsOAuthAuthenticated() {
			fmt.Printf("   ⚠️  Authentication required. Run 'blimu auth login' to authenticate.\n")
		}
	}

	return nil
}
