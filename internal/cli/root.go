package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "blimu",
	Short: "Blimu CLI - Generate custom SDKs and manage your Blimu configuration",
	Long: `Blimu CLI is a command-line tool for working with Blimu configurations.
It allows you to:
- Initialize new .blimu configurations
- Validate your resource configurations  
- Generate custom SDKs based on your resources
- Authenticate with Blimu API`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(resourcesCmd)
	rootCmd.AddCommand(rolesCmd)
	rootCmd.AddCommand(checkCmd)
}
