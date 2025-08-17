package initcmd

import (
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a new .blimu configuration",
		Long:  `Initialize a new .blimu configuration directory with template files`,
		// TODO: Implement initialization logic
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.MaximumNArgs(1),
	}
}
