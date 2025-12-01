package pull

import (
	"fmt"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// PullCommand represents the pull command
type PullCommand struct {
	WorkspaceID   string
	EnvironmentID string
	Directory     string
}

// NewPullCmd creates the pull command
func NewPullCmd() *cobra.Command {
	cmd := &PullCommand{}

	cobraCmd := &cobra.Command{
		Use:   "pull [directory]",
		Short: "Pull definitions from the cloud to local files",
		Long: `Pull environment definitions from the cloud and save them to local .blimu definition files.
This will overwrite existing local definition files if they exist.

The following files will be created/updated:
  - resources.yml (always)
  - entitlements.yml (if not empty)
  - features.yml (if not empty)
  - plans.yml (if not empty)

Examples:
  # Pull definitions to current directory
  blimu pull --workspace-id ws_123 --environment-id env_456

  # Pull definitions to specific directory
  blimu pull /path/to/project --workspace-id ws_123 --environment-id env_456`,
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

func (c *PullCommand) Run(cmd *cobra.Command) error {
	fmt.Printf("🔧 Starting pull command in directory: %s\n", c.Directory)

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
		return fmt.Errorf("environment-id is required for pull. Either:\n" +
			"  1. Provide --environment-id flag\n" +
			"  2. Configure your current environment with an ID using 'blimu env create --workspace-id <workspace-id> <env-name>'")
	}

	if c.WorkspaceID == "" {
		return fmt.Errorf("workspace-id is required for pull. Provide --workspace-id flag.\n" +
			"Use 'blimu workspaces list' to find your workspace ID (when available)")
	}

	fmt.Printf("📥 Pulling definitions from cloud...\n")

	// Check if dev mode is enabled
	devMode, _ := cmd.Flags().GetBool("dev")

	// Get auth client
	authClient, err := shared.GetAuthClientWithDevMode(devMode)
	if err != nil {
		return fmt.Errorf("authentication required for pull. Run 'blimu auth login' first: %w", err)
	}

	// Get platform SDK client
	sdk := authClient.GetAppSDK()
	if sdk == nil {
		return fmt.Errorf("platform SDK not available")
	}

	// Get definitions from the cloud
	definitions, err := sdk.Definitions.Get(c.WorkspaceID, c.EnvironmentID)
	if err != nil {
		return fmt.Errorf("failed to pull definitions: %w", err)
	}

	// Convert platform SDK response to BlimuConfig
	blimuConfig := &config.BlimuConfig{
		Resources:    convertToResourceConfig(definitions.Resources),
		Entitlements: convertToEntitlementConfig(definitions.Entitlements),
		Features:     convertToFeatureConfig(definitions.Features),
		Plans:        convertToPlanConfig(definitions.Plans),
	}

	// Save to local files
	if err := config.SaveBlimuConfig(c.Directory, blimuConfig); err != nil {
		return fmt.Errorf("failed to save definitions to local files: %w", err)
	}

	fmt.Printf("✅ Definitions pulled successfully!\n")
	fmt.Printf("  📋 Workspace: %s\n", c.WorkspaceID)
	fmt.Printf("  🌍 Environment: %s\n", c.EnvironmentID)
	fmt.Printf("  📁 Directory: %s/.blimu\n", c.Directory)

	return nil
}

// convertToResourceConfig converts map[string]interface{} to ResourceConfig map
func convertToResourceConfig(data map[string]interface{}) map[string]config.ResourceConfig {
	result := make(map[string]config.ResourceConfig)
	for k, v := range data {
		if vMap, ok := v.(map[string]interface{}); ok {
			resourceConfig := config.ResourceConfig{}
			if roles, ok := vMap["roles"].([]interface{}); ok {
				resourceConfig.Roles = make([]string, len(roles))
				for i, role := range roles {
					if roleStr, ok := role.(string); ok {
						resourceConfig.Roles[i] = roleStr
					}
				}
			}
			if rolesInheritance, ok := vMap["roles_inheritance"].(map[string]interface{}); ok {
				resourceConfig.RolesInheritance = make(map[string][]string)
				for role, inheritances := range rolesInheritance {
					if inheritancesArr, ok := inheritances.([]interface{}); ok {
						resourceConfig.RolesInheritance[role] = make([]string, len(inheritancesArr))
						for i, inh := range inheritancesArr {
							if inhStr, ok := inh.(string); ok {
								resourceConfig.RolesInheritance[role][i] = inhStr
							}
						}
					}
				}
			}
			if parents, ok := vMap["parents"].(map[string]interface{}); ok {
				resourceConfig.Parents = make(map[string]config.ParentConfig)
				for parentName, parentData := range parents {
					if parentMap, ok := parentData.(map[string]interface{}); ok {
						resourceConfig.Parents[parentName] = config.ParentConfig{
							Required: getBool(parentMap, "required"),
						}
					}
				}
			}
			result[k] = resourceConfig
		}
	}
	return result
}

// convertToEntitlementConfig converts map[string]interface{} to EntitlementConfig map
func convertToEntitlementConfig(data map[string]interface{}) map[string]config.EntitlementConfig {
	result := make(map[string]config.EntitlementConfig)
	for k, v := range data {
		if vMap, ok := v.(map[string]interface{}); ok {
			entitlementConfig := config.EntitlementConfig{}
			if roles, ok := vMap["roles"].([]interface{}); ok {
				entitlementConfig.Roles = make([]string, len(roles))
				for i, role := range roles {
					if roleStr, ok := role.(string); ok {
						entitlementConfig.Roles[i] = roleStr
					}
				}
			}
			if plans, ok := vMap["plans"].([]interface{}); ok {
				entitlementConfig.Plans = make([]string, len(plans))
				for i, plan := range plans {
					if planStr, ok := plan.(string); ok {
						entitlementConfig.Plans[i] = planStr
					}
				}
			}
			result[k] = entitlementConfig
		}
	}
	return result
}

// convertToFeatureConfig converts map[string]interface{} to FeatureConfig map
func convertToFeatureConfig(data map[string]interface{}) map[string]config.FeatureConfig {
	result := make(map[string]config.FeatureConfig)
	for k, v := range data {
		if vMap, ok := v.(map[string]interface{}); ok {
			featureConfig := config.FeatureConfig{}
			if plans, ok := vMap["plans"].([]interface{}); ok {
				featureConfig.Plans = make([]string, len(plans))
				for i, plan := range plans {
					if planStr, ok := plan.(string); ok {
						featureConfig.Plans[i] = planStr
					}
				}
			}
			if defaultEnabled, ok := vMap["default_enabled"].(bool); ok {
				featureConfig.DefaultEnabled = defaultEnabled
			}
			if entitlements, ok := vMap["entitlements"].([]interface{}); ok {
				featureConfig.Entitlements = make([]string, len(entitlements))
				for i, ent := range entitlements {
					if entStr, ok := ent.(string); ok {
						featureConfig.Entitlements[i] = entStr
					}
				}
			}
			result[k] = featureConfig
		}
	}
	return result
}

// convertToPlanConfig converts map[string]interface{} to PlanConfig map
func convertToPlanConfig(data map[string]interface{}) map[string]config.PlanConfig {
	result := make(map[string]config.PlanConfig)
	for k, v := range data {
		if vMap, ok := v.(map[string]interface{}); ok {
			planConfig := config.PlanConfig{}
			if name, ok := vMap["name"].(string); ok {
				planConfig.Name = name
			}
			if description, ok := vMap["description"].(string); ok {
				planConfig.Description = description
			} else if summary, ok := vMap["summary"].(string); ok {
				planConfig.Description = summary
			}
			result[k] = planConfig
		}
	}
	return result
}

// getBool safely extracts a boolean value from a map[string]interface{}
func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}
