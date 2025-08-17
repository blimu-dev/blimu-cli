package check

import (
	"github.com/spf13/cobra"
)

// NewCheckCmd creates the check command
func NewCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <user-id> <entitlement> <resource-id>",
		Short: "Check user entitlement",
		Long:  `Check if a user has a specific entitlement for a resource`,
		// TODO: Implement check logic
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(3),
	}
}
