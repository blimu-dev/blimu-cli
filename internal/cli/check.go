package cli

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	blimu "github.com/blimu-dev/blimu-go"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check <user-id> <entitlement> <resource-type> <resource-id>",
	Short: "Check if a user has a specific entitlement on a resource",
	Long: `Check if a user has a specific entitlement on a resource in your Blimu environment.

This command verifies whether a user has the necessary permissions to perform
a specific action on a resource based on your configured entitlements.

Example:
  blimu check user123 organization:create_workspace organization org456
  blimu check user789 workspace:delete workspace ws123`,
	Args: cobra.ExactArgs(4),
	RunE: runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	userID := args[0]
	entitlement := args[1]
	resourceType := args[2]
	resourceID := args[3]

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

	fmt.Printf("🔍 Checking entitlement '%s' for user '%s' on resource '%s:%s' in environment '%s'...\n",
		entitlement, userID, resourceType, resourceID, envName)

	// Create Blimu client
	client := blimu.NewClient(
		blimu.WithBaseURL(apiURL),
		blimu.WithApiKeyAuth(apiKey),
	)

	// Prepare entitlement check body
	body := blimu.EntitlementCheckBodyDto{
		UserId:      userID,
		Entitlement: entitlement,
		ResourceId:  resourceID,
	}

	// Check entitlement
	result, err := client.Authorization.CheckEntitlement(body)
	if err != nil {
		return fmt.Errorf("failed to check entitlement: %w", err)
	}

	// Display results
	fmt.Printf("\n📋 Entitlement Check Results:\n")
	fmt.Printf("   User: %s\n", userID)
	fmt.Printf("   Entitlement: %s\n", entitlement)
	fmt.Printf("   Resource: %s:%s\n", resourceType, resourceID)
	fmt.Printf("   Environment: %s\n", envName)

	if result.Allowed {
		fmt.Printf("   Status: ✅ ALLOWED\n")
		if result.Reason != "" {
			fmt.Printf("   Reason: %s\n", result.Reason)
		}
	} else {
		fmt.Printf("   Status: ❌ DENIED\n")
		if result.Reason != "" {
			fmt.Printf("   Reason: %s\n", result.Reason)
		}
		if len(result.RequiredRoles) > 0 {
			fmt.Printf("   Required roles:\n")
			for _, role := range result.RequiredRoles {
				fmt.Printf("     - %s\n", role)
			}
		}
		if len(result.UserRoles) > 0 {
			fmt.Printf("   User's current roles:\n")
			for _, role := range result.UserRoles {
				fmt.Printf("     - %s\n", role)
			}
		}
	}

	return nil
}
