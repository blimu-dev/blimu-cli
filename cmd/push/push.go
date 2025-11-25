package push

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	platform "github.com/blimu-dev/blimu-cli/internal/sdk"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// PushCommand represents the push command
type PushCommand struct {
	WorkspaceID   string
	EnvironmentID string
	Directory     string
}

// NewPushCmd creates the push command
func NewPushCmd() *cobra.Command {
	cmd := &PushCommand{}

	cobraCmd := &cobra.Command{
		Use:   "push [directory]",
		Short: "Push local definitions to the cloud",
		Long: `Push your local .blimu definition files (resources.yml, entitlements.yml, features.yml, plans.yml) 
to the cloud. Only files that exist and are non-empty will be pushed. Missing files will be ignored,
and existing definitions in the database will be preserved for those fields.

Examples:
  # Push definitions using current directory .blimu config
  blimu push --workspace-id ws_123 --environment-id env_456

  # Push definitions from specific directory
  blimu push /path/to/project --workspace-id ws_123 --environment-id env_456`,
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

func (c *PushCommand) Run(cmd *cobra.Command) error {
	fmt.Printf("🔧 Starting push command in directory: %s\n", c.Directory)

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
		return fmt.Errorf("environment-id is required for push. Either:\n" +
			"  1. Provide --environment-id flag\n" +
			"  2. Configure your current environment with an ID using 'blimu env create --workspace-id <workspace-id> <env-name>'")
	}

	if c.WorkspaceID == "" {
		return fmt.Errorf("workspace-id is required for push. Provide --workspace-id flag.\n" +
			"Use 'blimu workspaces list' to find your workspace ID (when available)")
	}

	// Load definitions files (only those that exist and are non-empty)
	blimuDir := filepath.Join(c.Directory, ".blimu")
	request := platform.DefinitionUpdateDto{
		Resources:    make(map[string]interface{}),
		Entitlements: make(map[string]interface{}),
		Features:     make(map[string]interface{}),
		Plans:        make(map[string]interface{}),
	}

	// Load resources.yml (required)
	resourcesPath := filepath.Join(blimuDir, "resources.yml")
	loaded, err := loadDefinitionFile(resourcesPath, "resources")
	if err != nil {
		return fmt.Errorf("failed to load resources.yml: %w", err)
	}
	if len(loaded) == 0 {
		return fmt.Errorf("resources.yml is required and cannot be empty")
	}
	request.Resources = loaded
	fmt.Printf("✅ Loaded resources.yml\n")

	// Load entitlements.yml (optional)
	entitlementsPath := filepath.Join(blimuDir, "entitlements.yml")
	loaded, err = loadDefinitionFile(entitlementsPath, "entitlements")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load entitlements.yml: %w", err)
		}
		fmt.Printf("⏭️  Skipping entitlements.yml (file not found)\n")
	} else if len(loaded) > 0 {
		request.Entitlements = loaded
		fmt.Printf("✅ Loaded entitlements.yml\n")
	}

	// Load features.yml (optional)
	featuresPath := filepath.Join(blimuDir, "features.yml")
	loaded, err = loadDefinitionFile(featuresPath, "features")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load features.yml: %w", err)
		}
		fmt.Printf("⏭️  Skipping features.yml (file not found)\n")
	} else if len(loaded) > 0 {
		request.Features = loaded
		fmt.Printf("✅ Loaded features.yml\n")
	}

	// Load plans.yml (optional)
	plansPath := filepath.Join(blimuDir, "plans.yml")
	loaded, err = loadDefinitionFile(plansPath, "plans")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load plans.yml: %w", err)
		}
		fmt.Printf("⏭️  Skipping plans.yml (file not found)\n")
	} else if len(loaded) > 0 {
		request.Plans = loaded
		fmt.Printf("✅ Loaded plans.yml\n")
	}

	fmt.Printf("📤 Pushing definitions to cloud...\n")

	// Check if dev mode is enabled
	devMode, _ := cmd.Flags().GetBool("dev")

	// Get auth client
	authClient, err := shared.GetAuthClientWithDevMode(devMode)
	if err != nil {
		return fmt.Errorf("authentication required for push. Run 'blimu auth login' first: %w", err)
	}

	// Get platform SDK client
	sdk := authClient.GetPlatformSDK()
	if sdk == nil {
		return fmt.Errorf("platform SDK not available")
	}

	// Update definitions in the cloud (partial update - only provided fields will be updated)
	_, err = sdk.Definitions.Update(c.WorkspaceID, c.EnvironmentID, request)
	if err != nil {
		return fmt.Errorf("failed to push definitions: %w", err)
	}

	fmt.Printf("✅ Definitions pushed successfully!\n")
	fmt.Printf("  📋 Workspace: %s\n", c.WorkspaceID)
	fmt.Printf("  🌍 Environment: %s\n", c.EnvironmentID)

	return nil
}

// loadDefinitionFile loads a YAML definition file and parses it into a map
func loadDefinitionFile(filePath, fileType string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Check if file is empty or only whitespace
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil, fmt.Errorf("file is empty")
	}

	// Parse YAML
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", fileType, err)
	}

	// Extract the root key (e.g., "resources", "entitlements", etc.)
	if rootValue, ok := yamlData[fileType]; ok {
		if rootMap, ok := rootValue.(map[string]interface{}); ok {
			return rootMap, nil
		}
	}

	// If no root key, use the entire config
	return yamlData, nil
}
