package cli

import (
	"fmt"
	"strings"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/spf13/cobra"
)

var envDeleteCmd = &cobra.Command{
	Use:   "delete <environment-name>",
	Short: "Delete an environment",
	Long:  `Delete an environment configuration`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvDelete,
}

func runEnvDelete(cmd *cobra.Command, args []string) error {
	envName := args[0]

	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete environment '%s'? (y/N): ", envName)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := cliConfig.RemoveEnvironment(envName); err != nil {
		return err
	}

	fmt.Printf("✅ Deleted environment '%s'\n", envName)

	// If we switched to a different current environment, mention it
	if cliConfig.CurrentEnvironment != "" && cliConfig.CurrentEnvironment != envName {
		fmt.Printf("   Current environment is now '%s'\n", cliConfig.CurrentEnvironment)
	}

	return nil
}
