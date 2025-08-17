package env

import (
	"fmt"
	"strings"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/spf13/cobra"
)

// DeleteCommand represents the delete environment command
type DeleteCommand struct {
	EnvName string
}

// NewDeleteCmd creates the delete command
func NewDeleteCmd() *cobra.Command {
	cmd := &DeleteCommand{}

	cobraCmd := &cobra.Command{
		Use:   "delete <environment-name>",
		Short: "Delete an environment",
		Long:  `Delete an environment configuration`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			cmd.EnvName = args[0]
			return cmd.Run()
		},
	}

	return cobraCmd
}

// Run executes the delete environment command
func (c *DeleteCommand) Run() error {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete environment '%s'? (y/N): ", c.EnvName)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := cliConfig.RemoveEnvironment(c.EnvName); err != nil {
		return err
	}

	fmt.Printf("✅ Deleted environment '%s'\n", c.EnvName)

	// If we switched to a different current environment, mention it
	if cliConfig.CurrentEnvironment != "" && cliConfig.CurrentEnvironment != c.EnvName {
		fmt.Printf("   Current environment is now '%s'\n", cliConfig.CurrentEnvironment)
	}

	return nil
}
