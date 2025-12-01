package definitions

import (
	"encoding/json"
	"fmt"

	platform "github.com/blimu-dev/blimu-cli/internal/sdk"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// UpdateCommand represents the definitions update command
type UpdateCommand struct {
	WorkspaceID   string
	EnvironmentID string
	Directory     string
}

// NewUpdateCmd creates the definitions update command
func NewUpdateCmd() *cobra.Command {
	cmd := &UpdateCommand{}

	cobraCmd := &cobra.Command{
		Use:   "update [directory]",
		Short: "Update definitions in the cloud from local .blimu configuration",
		Long: `Update your environment's definitions (resources, entitlements, features, plans) 
in the cloud by reading your local .blimu configuration files.

This command will:
1. Load and validate your local .blimu configuration
2. Convert it to the API format
3. Push it to your Blimu environment

Examples:
  # Update definitions using current directory .blimu config
  blimu definitions update --workspace-id ws_123 --environment-id env_456

  # Update definitions from specific directory
  blimu definitions update /path/to/project --workspace-id ws_123 --environment-id env_456`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cmd.Directory = args[0]
			} else {
				cmd.Directory = "."
			}
			return cmd.Run(cobraCmd)
		},
		Args: cobra.MaximumNArgs(1),
	}

	cobraCmd.Flags().StringVar(&cmd.WorkspaceID, "workspace-id", "", "Workspace ID (uses current environment's workspace if available)")
	cobraCmd.Flags().StringVar(&cmd.EnvironmentID, "environment-id", "", "Environment ID (uses current environment ID if available)")

	return cobraCmd
}

func (c *UpdateCommand) Run(cmd *cobra.Command) error {
	fmt.Printf("🔧 Starting definitions update from directory: %s\n", c.Directory)

	// Get current environment info to auto-populate missing IDs
	_, currentEnv, err := shared.GetCurrentEnvironmentInfo()
	if err != nil {
		return fmt.Errorf("failed to get current environment info: %w", err)
	}

	// Auto-populate environment ID from current environment if not provided
	if c.EnvironmentID == "" && currentEnv.ID != "" {
		c.EnvironmentID = currentEnv.ID
		fmt.Printf("📋 Using environment ID from current environment: %s\n", c.EnvironmentID)
	}

	// Auto-populate workspace ID from current environment if not provided
	if c.WorkspaceID == "" && currentEnv.WorkspaceID != "" {
		c.WorkspaceID = currentEnv.WorkspaceID
		fmt.Printf("📋 Using workspace ID from current environment: %s\n", c.WorkspaceID)
	}

	// Check required parameters
	if c.EnvironmentID == "" {
		return fmt.Errorf("environment-id is required for definitions update. Either:\n" +
			"  1. Provide --environment-id flag\n" +
			"  2. Configure your current environment with an ID using 'blimu env create --workspace-id <workspace-id> <env-name>'")
	}

	if c.WorkspaceID == "" {
		return fmt.Errorf("workspace-id is required for definitions update. Provide --workspace-id flag.\n" +
			"Use 'blimu workspaces list' to find your workspace ID (when available)")
	}

	// Load Blimu configuration
	blimuConfig, err := config.LoadBlimuConfig(c.Directory)
	if err != nil {
		return fmt.Errorf("failed to load .blimu configuration: %w", err)
	}

	fmt.Printf("🔧 Updating definitions from configuration in %s...\n", c.Directory)

	// Check if dev mode is enabled
	devMode, _ := cmd.Flags().GetBool("dev")

	// Get auth client
	authClient, err := shared.GetAuthClientWithDevMode(devMode)
	if err != nil {
		return fmt.Errorf("authentication required for definitions update. Run 'blimu auth login' first: %w", err)
	}

	// Get platform SDK client
	sdk := authClient.GetAppSDK()
	if sdk == nil {
		return fmt.Errorf("platform SDK not available")
	}

	// Convert config to request format
	configJSON, err := blimuConfig.MergeToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize configuration: %w", err)
	}

	// Parse config for request
	var configMap map[string]interface{}
	if err := json.Unmarshal(configJSON, &configMap); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Build the definitions update request
	request := platform.DefinitionUpdateDto{
		Resources:    make(map[string]interface{}),
		Entitlements: make(map[string]interface{}),
		Features:     make(map[string]interface{}),
		Plans:        make(map[string]interface{}),
	}

	// Copy data from config
	if resources, ok := configMap["resources"].(map[string]interface{}); ok {
		request.Resources = resources
	}
	if entitlements, ok := configMap["entitlements"].(map[string]interface{}); ok {
		request.Entitlements = entitlements
	}
	if features, ok := configMap["features"].(map[string]interface{}); ok {
		request.Features = features
	}
	if plans, ok := configMap["plans"].(map[string]interface{}); ok {
		request.Plans = plans
	}

	fmt.Printf("📤 Pushing definitions to cloud...\n")

	// Update definitions in the cloud
	_, err = sdk.Definitions.Update(c.WorkspaceID, c.EnvironmentID, request)
	if err != nil {
		return fmt.Errorf("failed to update definitions: %w", err)
	}

	fmt.Printf("✅ Definitions updated successfully!\n")
	fmt.Printf("  📋 Workspace: %s\n", c.WorkspaceID)
	fmt.Printf("  🌍 Environment: %s\n", c.EnvironmentID)
	if version := getString(configMap, "version"); version != "" {
		fmt.Printf("  🏷️  Version: %s\n", version)
	}

	return nil
}

// getString safely extracts a string value from a map[string]interface{}
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
