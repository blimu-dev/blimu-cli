package validate

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/api"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// ValidateCommand represents the validate command
type ValidateCommand struct {
	WorkspaceID   string
	EnvironmentID string
	Directory     string
}

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	cmd := &ValidateCommand{}

	cobraCmd := &cobra.Command{
		Use:   "validate [directory]",
		Short: "Validate .blimu configuration",
		Long: `Validate your local .blimu configuration files for syntax and semantic errors.

This command validates your Blimu configuration against the platform API and reports any issues.
For full validation, provide workspace and environment IDs.`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Directory = args[0]
			} else {
				cmd.Directory = "."
			}
			return cmd.Run()
		},
		Args: cobra.MaximumNArgs(1),
	}

	cobraCmd.Flags().StringVar(&cmd.WorkspaceID, "workspace-id", "", "Workspace ID for platform validation")
	cobraCmd.Flags().StringVar(&cmd.EnvironmentID, "environment-id", "", "Environment ID for platform validation")

	return cobraCmd
}

func (c *ValidateCommand) Run() error {
	// Load Blimu configuration
	blimuConfig, err := config.LoadBlimuConfig(c.Directory)
	if err != nil {
		return fmt.Errorf("failed to load .blimu configuration: %w", err)
	}

	fmt.Printf("📋 Validating Blimu configuration in %s...\n", c.Directory)

	// Convert config to JSON for validation
	configJSON, err := blimuConfig.MergeToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Get auth client for API validation
	authClient, err := shared.GetAuthClient()
	if err != nil {
		fmt.Printf("⚠️  No authentication configured. Performing local validation only.\n")
		fmt.Printf("Use 'blimu auth login' to enable platform validation.\n\n")
		return c.performLocalValidation(blimuConfig)
	}

	// Create API client
	apiClient := api.NewClient(authClient)

	// Validate via platform API
	result, err := apiClient.ValidateConfig(configJSON, c.WorkspaceID, c.EnvironmentID)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Display results
	if result.Valid {
		fmt.Printf("✅ Configuration is valid!\n")

		if len(result.Spec) > 0 {
			fmt.Printf("\n📊 Generated OpenAPI specification with %d paths\n", len(result.Spec))
		}
	} else {
		fmt.Printf("❌ Configuration has %d error(s):\n\n", len(result.Errors))

		for i, err := range result.Errors {
			fmt.Printf("%d. %s\n", i+1, err.Message)
			if err.Resource != "" {
				fmt.Printf("   Resource: %s\n", err.Resource)
			}
			if err.Field != "" {
				fmt.Printf("   Field: %s\n", err.Field)
			}
			fmt.Printf("\n")
		}

		return fmt.Errorf("configuration validation failed")
	}

	return nil
}

func (c *ValidateCommand) performLocalValidation(blimuConfig *config.BlimuConfig) error {
	fmt.Printf("🔍 Performing local validation...\n\n")

	// Basic structure validation
	errors := []string{}

	if len(blimuConfig.Resources) == 0 {
		errors = append(errors, "No resources defined")
	}

	// Validate resources have required fields
	for resourceName, resource := range blimuConfig.Resources {
		if len(resource.Roles) == 0 {
			errors = append(errors, fmt.Sprintf("Resource '%s' has no roles defined", resourceName))
		}
	}

	// Validate entitlements reference valid resources
	for entitlementName := range blimuConfig.Entitlements {
		// Basic format check: should be "resource:action"
		if !contains(entitlementName, ":") {
			errors = append(errors, fmt.Sprintf("Entitlement '%s' should follow 'resource:action' format", entitlementName))
		}
	}

	if len(errors) > 0 {
		fmt.Printf("❌ Found %d local validation error(s):\n\n", len(errors))
		for i, err := range errors {
			fmt.Printf("%d. %s\n", i+1, err)
		}
		fmt.Printf("\n💡 For complete validation, use platform API with --workspace-id and --environment-id\n")
		return fmt.Errorf("local validation failed")
	}

	fmt.Printf("✅ Local validation passed!\n")
	fmt.Printf("💡 For complete validation, use platform API with --workspace-id and --environment-id\n")

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
