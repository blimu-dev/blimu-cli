package cli

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/spf13/cobra"
)

var envSwitchCmd = &cobra.Command{
	Use:   "switch <environment-name>",
	Short: "Switch to a different environment",
	Long:  `Switch the current active environment to the specified environment`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvSwitch,
}

func runEnvSwitch(cmd *cobra.Command, args []string) error {
	envName := args[0]

	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	if err := cliConfig.SetCurrentEnvironment(envName); err != nil {
		return err
	}

	fmt.Printf("✅ Switched to environment '%s'\n", envName)
	return nil
}
