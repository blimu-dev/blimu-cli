package cli

import (
	"fmt"
	"os"

	"github.com/blimu-dev/blimu-cli/pkg/api"
	"github.com/blimu-dev/blimu-cli/pkg/auth"
	"github.com/blimu-dev/blimu-cli/pkg/blimu"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your .blimu configuration",
	Long: `Validate your .blimu configuration files.
This checks for:
- Valid resource definitions
- Correct role inheritance syntax
- Valid parent relationships
- No circular dependencies
- Cross-file reference validation
- Optional server-side validation via API`,
	RunE: runValidate,
}

var (
	validateRemote bool
)

func init() {
	validateCmd.Flags().BoolVarP(&validateRemote, "remote", "r", false, "Also run server-side validation via Blimu API")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Find .blimu configuration
	configDir, err := config.FindBlimuConfig(".")
	if err != nil {
		return fmt.Errorf("no .blimu configuration found. Run 'blimu init' to create one")
	}

	fmt.Printf("📁 Found .blimu configuration in: %s\n", configDir)

	// Load configuration
	blimuConfig, err := config.LoadBlimuConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Always do local validation first
	fmt.Println("🔍 Running local validation...")
	result := blimu.ValidateConfig(blimuConfig)

	if !result.Valid {
		fmt.Printf("❌ Local validation failed with %d error(s):\n\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("%d. %s\n", i+1, err.Error())
		}
		if !validateRemote {
			os.Exit(1)
		}
		fmt.Println("\n⚠️  Continuing with remote validation despite local errors...")
	} else {
		fmt.Println("✅ Local validation passed!")
	}

	// If remote validation is requested
	if validateRemote {
		fmt.Println("\n🌐 Running server-side validation...")
		if err := runRemoteValidation(blimuConfig); err != nil {
			return fmt.Errorf("remote validation failed: %w", err)
		}
	}

	// Show local summary if validation passed
	if result.Valid {
		showConfigSummary(blimuConfig)
	}

	return nil
}

func runRemoteValidation(blimuConfig *config.BlimuConfig) error {
	// Convert config to JSON
	configJSON, err := blimuConfig.MergeToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert config to JSON: %w", err)
	}

	// Create authenticated client
	authClient, err := auth.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("authentication required for remote validation: %w", err)
	}

	// Test authentication
	if err := authClient.ValidateAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Create API client
	apiClient := api.NewClient(authClient)

	// Validate config remotely
	response, err := apiClient.ValidateConfig(configJSON)
	if err != nil {
		return fmt.Errorf("API validation request failed: %w", err)
	}

	if response.Valid {
		fmt.Println("✅ Server-side validation passed!")
		fmt.Println("🎉 Your configuration is ready for deployment!")
	} else {
		fmt.Printf("❌ Server-side validation failed with %d error(s):\n\n", len(response.Errors))
		for i, err := range response.Errors {
			fmt.Printf("%d. %s.%s: %s\n", i+1, err.Resource, err.Field, err.Message)
		}
		return fmt.Errorf("server-side validation failed")
	}

	return nil
}

func showConfigSummary(blimuConfig *config.BlimuConfig) {
	fmt.Printf("\n📊 Configuration summary:\n")

	fmt.Printf("  📁 Resources (%d):\n", len(blimuConfig.Resources))
	for resourceName, resourceConfig := range blimuConfig.Resources {
		fmt.Printf("    • %s: %d roles", resourceName, len(resourceConfig.Roles))
		if len(resourceConfig.Parents) > 0 {
			fmt.Printf(", %d parents", len(resourceConfig.Parents))
		}
		if len(resourceConfig.RolesInheritance) > 0 {
			fmt.Printf(", %d role inheritances", len(resourceConfig.RolesInheritance))
		}
		fmt.Println()
	}

	if len(blimuConfig.Plans) > 0 {
		fmt.Printf("  💰 Plans (%d):\n", len(blimuConfig.Plans))
		for planName, planConfig := range blimuConfig.Plans {
			fmt.Printf("    • %s: %s\n", planName, planConfig.Name)
		}
	}

	if len(blimuConfig.Entitlements) > 0 {
		fmt.Printf("  🔐 Entitlements (%d):\n", len(blimuConfig.Entitlements))
		for entitlementName, entitlementConfig := range blimuConfig.Entitlements {
			fmt.Printf("    • %s: %d roles", entitlementName, len(entitlementConfig.Roles))
			if len(entitlementConfig.Plans) > 0 {
				fmt.Printf(", %d plans", len(entitlementConfig.Plans))
			}
			fmt.Println()
		}
	}

	if len(blimuConfig.Features) > 0 {
		fmt.Printf("  🚀 Features (%d):\n", len(blimuConfig.Features))
		for featureName, featureConfig := range blimuConfig.Features {
			fmt.Printf("    • %s: %d plans", featureName, len(featureConfig.Plans))
			if len(featureConfig.Entitlements) > 0 {
				fmt.Printf(", %d entitlements", len(featureConfig.Entitlements))
			}
			fmt.Println()
		}
	}
}
