package cli

import (
	"fmt"
	"path/filepath"

	blimu "github.com/blimu-dev/blimu-go"

	"github.com/blimu-dev/blimu-cli/pkg/auth"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Commands for managing authentication with Blimu API`,
}

var testAuthCmd = &cobra.Command{
	Use:   "test",
	Short: "Test authentication with Blimu API",
	Long: `Test your authentication credentials with the Blimu API.
Requires BLIMU_SECRET_KEY environment variable to be set.`,
	RunE: runTestAuth,
}

var pushCmd = &cobra.Command{
	Use:   "push [directory]",
	Short: "Push .blimu configuration to Blimu API",
	Long: `Push your local .blimu configuration (resources, entitlements, features, plans) 
to the Blimu API. This will update your environment's authorization definitions.

The command will:
1. Load and validate your local .blimu configuration
2. Convert it to the API format
3. Push it to your Blimu environment

By default, uses the current environment. Use --env to specify a different environment.`,
	RunE: runPushAuth,
	Args: cobra.MaximumNArgs(1),
}

var (
	pushEnvName string
)

func init() {
	pushCmd.Flags().StringVar(&pushEnvName, "env", "", "Environment to push to (uses current environment if not specified)")

	authCmd.AddCommand(testAuthCmd)
	authCmd.AddCommand(pushCmd)
}

func runTestAuth(cmd *cobra.Command, args []string) error {
	// Load CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	apiURL, apiKey, err := cliConfig.GetAPIClient()
	if err != nil {
		return fmt.Errorf("failed to get API client config: %w", err)
	}

	currentEnv, _ := cliConfig.GetCurrentEnvironment()
	envName := cliConfig.CurrentEnvironment
	if currentEnv != nil && currentEnv.Name != "" {
		envName = currentEnv.Name
	}

	fmt.Printf("🔐 Testing authentication for environment '%s' with %s...\n", envName, apiURL)

	// Create authenticated client
	client := auth.NewClient(apiURL, apiKey)

	// Test authentication
	if err := client.ValidateAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("✅ Authentication successful!")
	fmt.Printf("   Environment: %s\n", envName)
	fmt.Printf("   API URL: %s\n", apiURL)
	fmt.Printf("   API Key: %s...%s\n",
		apiKey[:8],
		apiKey[len(apiKey)-4:])

	return nil
}

func runPushAuth(cmd *cobra.Command, args []string) error {
	// Load CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	// Determine which environment to use
	var targetEnvName string
	var targetEnv *config.Environment

	if pushEnvName != "" {
		// Use specified environment
		targetEnvName = pushEnvName
		env, exists := cliConfig.Environments[pushEnvName]
		if !exists {
			return fmt.Errorf("environment '%s' not found", pushEnvName)
		}
		targetEnv = &env
	} else {
		// Use current environment
		targetEnvName = cliConfig.CurrentEnvironment
		env, err := cliConfig.GetCurrentEnvironment()
		if err != nil {
			return fmt.Errorf("failed to get current environment: %w", err)
		}
		targetEnv = env
	}

	// Get API client config for the target environment
	apiURL := targetEnv.APIURL
	if apiURL == "" {
		apiURL = cliConfig.DefaultAPIURL
	}
	apiKey := targetEnv.APIKey

	if apiKey == "" {
		return fmt.Errorf("no API key configured for environment '%s'", targetEnvName)
	}

	// Determine directory to search for .blimu config
	searchDir := "."
	if len(args) > 0 {
		searchDir = args[0]
	}

	// Convert to absolute path
	absSearchDir, err := filepath.Abs(searchDir)
	if err != nil {
		return fmt.Errorf("failed to resolve directory path: %w", err)
	}

	fmt.Printf("🔍 Searching for .blimu configuration in %s...\n", absSearchDir)

	// Find .blimu config directory
	configDir, err := config.FindBlimuConfig(absSearchDir)
	if err != nil {
		return fmt.Errorf("failed to find .blimu configuration: %w", err)
	}

	fmt.Printf("📁 Found .blimu configuration in %s\n", configDir)

	// Load .blimu config
	blimuConfig, err := config.LoadBlimuConfig(configDir)
	if err != nil {
		return fmt.Errorf("failed to load .blimu configuration: %w", err)
	}

	fmt.Printf("✅ Loaded configuration with:\n")
	fmt.Printf("   - Resources: %d\n", len(blimuConfig.Resources))
	fmt.Printf("   - Entitlements: %d\n", len(blimuConfig.Entitlements))
	fmt.Printf("   - Features: %d\n", len(blimuConfig.Features))
	fmt.Printf("   - Plans: %d\n", len(blimuConfig.Plans))

	// Convert to API format
	definition, err := convertConfigToDefinition(blimuConfig)
	if err != nil {
		return fmt.Errorf("failed to convert configuration: %w", err)
	}

	// Create Blimu client
	client := blimu.NewClient(
		blimu.WithBaseURL(apiURL),
		blimu.WithApiKeyAuth(apiKey),
	)

	fmt.Printf("🚀 Pushing configuration for environment '%s' to %s...\n", targetEnvName, apiURL)

	// Push to API
	result, err := client.Definitions.Update(definition)
	if err != nil {
		return fmt.Errorf("failed to push configuration: %w", err)
	}

	fmt.Println("✅ Configuration pushed successfully!")
	fmt.Printf("   - Resources: %v\n", len(result.Resources.Roles))
	fmt.Printf("   - Entitlements: %v\n", len(result.Entitlements.Role))
	fmt.Printf("   - Features: %s\n", result.Features.Name)
	fmt.Printf("   - Plans: %s\n", result.Plans.Name)

	return nil
}

// convertConfigToDefinition converts blimucli config format to blimu-go Definition format
func convertConfigToDefinition(blimuConfig *config.BlimuConfig) (blimu.Definition, error) {
	definition := blimu.Definition{}

	// Convert Resources
	if len(blimuConfig.Resources) > 0 {
		// For now, we'll take the first resource as an example
		// In a real implementation, you'd need to handle multiple resources
		for resourceName, resourceConfig := range blimuConfig.Resources {
			definition.Resources = blimu.DefinitionResources{
				Roles: resourceConfig.Roles,
				Parents: blimu.DefinitionParents{
					Required: false, // Default value, you might need to adjust based on your config
				},
				RolesInheritance: blimu.DefinitionRolesInheritance{},
				Fields: blimu.DefinitionFields{
					Required: true,
					Type:     blimu.DefinitionType{},
				},
				Tenant: true, // Default value, adjust as needed
			}
			// For now, just use the first resource
			_ = resourceName
			break
		}
	}

	// Convert Entitlements
	if len(blimuConfig.Entitlements) > 0 {
		// Take first entitlement as example
		for _, entitlementConfig := range blimuConfig.Entitlements {
			definition.Entitlements = blimu.DefinitionEntitlements{
				Role:  entitlementConfig.Roles,
				Plan:  "", // You might need to join plans or handle differently
				Limit: blimu.DefinitionLimit{},
			}
			if len(entitlementConfig.Plans) > 0 {
				definition.Entitlements.Plan = entitlementConfig.Plans[0]
			}
			break
		}
	}

	// Convert Features
	if len(blimuConfig.Features) > 0 {
		// Take first feature as example
		for featureName, featureConfig := range blimuConfig.Features {
			definition.Features = blimu.DefinitionFeatures{
				Name:         featureName,
				Summary:      fmt.Sprintf("Feature: %s", featureName),
				Entitlements: featureConfig.Entitlements,
			}
			break
		}
	}

	// Convert Plans
	if len(blimuConfig.Plans) > 0 {
		// Take first plan as example
		for planName, planConfig := range blimuConfig.Plans {
			definition.Plans = blimu.DefinitionPlans{
				Name:    planConfig.Name,
				Summary: planConfig.Description,
			}
			_ = planName
			break
		}
	}

	return definition, nil
}
