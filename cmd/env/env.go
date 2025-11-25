package env

import (
	"github.com/spf13/cobra"
)

// NewEnvCmd creates the env command group
func NewEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Environment management commands",
		Long:  `Commands for managing Blimu environments`,
	}

	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewSwitchCmd())
	cmd.AddCommand(NewCurrentCmd())

	return cmd
}
