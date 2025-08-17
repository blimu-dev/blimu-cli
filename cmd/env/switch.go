package env

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/config"
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
		Use:   "switch <environment-name>",
		Short: "Switch to a different environment",
		Long:  `Switch the current active environment to the specified environment`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			cmd.EnvName = args[0]
			return cmd.Run()
		},
	}

	return cobraCmd
}

// Run executes the switch environment command
func (c *SwitchCommand) Run() error {
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	if err := cliConfig.SetCurrentEnvironment(c.EnvName); err != nil {
		return err
	}

	fmt.Printf("✅ Switched to environment '%s'\n", c.EnvName)
	return nil
}
