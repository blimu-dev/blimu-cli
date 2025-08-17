package generate

import (
	"github.com/spf13/cobra"
)

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [directory]",
		Short: "Generate SDK from .blimu configuration",
		Long:  `Generate a custom SDK based on your .blimu configuration files`,
		// TODO: Implement generation logic
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.MaximumNArgs(1),
	}
}
