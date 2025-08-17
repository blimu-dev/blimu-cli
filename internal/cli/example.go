package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Example command demonstrating shared context pattern",
	Long: `This command demonstrates how the shared context pattern works.
It shows how easy it is to access configuration, API client, and environment info
without manually loading them in each command.`,
	RunE: runExample,
}

func init() {
	// This would be added to root.go in a real implementation
	// rootCmd.AddCommand(exampleCmd)
}

func runExample(cmd *cobra.Command, args []string) error {
	// Get shared context - this is the key pattern
	ctx := GetContext()

	fmt.Println("🚀 Example: Using Shared Context Pattern")
	fmt.Println("========================================")

	// 1. Easy environment validation and client access
	client, envName, err := ctx.ValidateAndGetClient()
	if err != nil {
		return fmt.Errorf("context validation failed: %w", err)
	}

	fmt.Printf("✅ Environment: %s\n", envName)
	fmt.Printf("✅ API Client: %T (ready to use)\n", client)

	// 2. Easy access to configuration
	if ctx.CLIConfig != nil {
		fmt.Printf("✅ CLI Config loaded with %d environments\n", len(ctx.CLIConfig.Environments))
	}

	// 3. Easy access to .blimu configuration if loaded
	if ctx.BlimuConfig != nil {
		fmt.Printf("✅ Blimu Config loaded with:\n")
		fmt.Printf("   - Resources: %d\n", len(ctx.BlimuConfig.Resources))
		fmt.Printf("   - Entitlements: %d\n", len(ctx.BlimuConfig.Entitlements))
		fmt.Printf("   - Features: %d\n", len(ctx.BlimuConfig.Features))
		fmt.Printf("   - Plans: %d\n", len(ctx.BlimuConfig.Plans))
	} else {
		fmt.Println("ℹ️  No .blimu configuration loaded (use 'blimu auth push' to load)")
	}

	// 4. Easy API operations
	fmt.Println("\n🔍 Testing API connectivity...")

	// Example: Get environment info from API
	if envName != "no environment set" {
		// This would work if we had the environment ID
		fmt.Println("✅ API client ready for operations")
		fmt.Println("   (Environment details would be fetched here)")
	}

	fmt.Println("\n✨ Benefits of this pattern:")
	fmt.Println("   - No manual config loading in each command")
	fmt.Println("   - Shared API client with lazy initialization")
	fmt.Println("   - Consistent environment validation")
	fmt.Println("   - Thread-safe access to shared state")
	fmt.Println("   - Easy to add new commands that need the same context")

	return nil
}
