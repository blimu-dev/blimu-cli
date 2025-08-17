package roles

import (
	"github.com/spf13/cobra"
)

// NewRolesCmd creates the roles command group
func NewRolesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roles",
		Short: "Role management commands",
		Long:  `Commands for managing user roles in your Blimu environment`,
	}

	// TODO: Add subcommands
	// cmd.AddCommand(NewAssignRoleCmd())
	// cmd.AddCommand(NewRemoveRoleCmd())

	return cmd
}
