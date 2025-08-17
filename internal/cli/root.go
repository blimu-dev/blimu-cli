package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global context instance - similar to kubectl's approach
var globalContext *Context

// GetContext returns the global CLI context, initializing it if necessary
func GetContext() *Context {
	if globalContext == nil {
		globalContext = NewContext()
	}
	return globalContext
}

var rootCmd = &cobra.Command{
	Use:   "blimu",
	Short: "Blimu CLI - Generate custom SDKs and manage your Blimu configuration",
	Long: `Blimu CLI is a command-line tool for working with Blimu configurations.
It allows you to:
- Initialize new .blimu configurations
- Validate your resource configurations  
- Generate custom SDKs based on your resources
- Authenticate with Blimu API`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the global context
		ctx := GetContext()

		// Load CLI configuration for all commands
		if err := ctx.LoadCLIConfig(); err != nil {
			return fmt.Errorf("failed to load CLI configuration: %w", err)
		}

		return nil
	},
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
