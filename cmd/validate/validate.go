package validate

import (
	"github.com/spf13/cobra"
)

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [directory]",
		Short: "Validate .blimu configuration",
		Long:  `Validate your local .blimu configuration files for syntax and semantic errors`,
		// TODO: Implement validation logic
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.MaximumNArgs(1),
	}
}
