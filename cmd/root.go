package cmd

import (
	"fmt"
	"os"

	"github.com/blimu-dev/blimu-cli/cmd/auth"
	"github.com/blimu-dev/blimu-cli/cmd/check"
	"github.com/blimu-dev/blimu-cli/cmd/definitions"
	"github.com/blimu-dev/blimu-cli/cmd/env"
	"github.com/blimu-dev/blimu-cli/cmd/generate"
	initcmd "github.com/blimu-dev/blimu-cli/cmd/initcmd"
	"github.com/blimu-dev/blimu-cli/cmd/pull"
	"github.com/blimu-dev/blimu-cli/cmd/push"

	"github.com/blimu-dev/blimu-cli/cmd/resources"
	"github.com/blimu-dev/blimu-cli/cmd/roles"
	"github.com/blimu-dev/blimu-cli/cmd/validate"
	"github.com/spf13/cobra"
)

var cfgFile string
var devMode bool

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

// GetDevMode returns whether dev mode is enabled
func GetDevMode() bool {
	return devMode
}

// GetDevAPIURL returns the development API URL
func GetDevAPIURL() string {
	return "http://localhost:3010"
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&devMode, "dev", false, "Use development mode (localhost:3010)")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Register commands using factory pattern
	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(env.NewEnvCmd())
	rootCmd.AddCommand(resources.NewResourcesCmd())
	rootCmd.AddCommand(roles.NewRolesCmd())
	rootCmd.AddCommand(validate.NewValidateCmd())
	rootCmd.AddCommand(generate.NewGenerateCmd())
	rootCmd.AddCommand(initcmd.NewInitCmd())
	rootCmd.AddCommand(check.NewCheckCmd())
	rootCmd.AddCommand(definitions.NewDefinitionsCmd())
	rootCmd.AddCommand(push.NewPushCmd())
	rootCmd.AddCommand(pull.NewPullCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
