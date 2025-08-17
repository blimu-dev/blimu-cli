package resources

import (
	"github.com/spf13/cobra"
)

// NewResourcesCmd creates the resources command group
func NewResourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Resource management commands",
		Long:  `Commands for managing resources in your Blimu environment`,
	}

	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewBulkCmd())

	return cmd
}
