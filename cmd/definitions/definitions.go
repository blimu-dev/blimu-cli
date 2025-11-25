package definitions

import (
	"github.com/spf13/cobra"
)

// NewDefinitionsCmd creates the definitions command group
func NewDefinitionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "definitions",
		Short: "Definition management commands",
		Long:  `Commands for managing Blimu definitions (resources, entitlements, features, plans)`,
	}

	cmd.AddCommand(NewUpdateCmd())

	return cmd
}
