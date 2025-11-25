package generate

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blimu-dev/blimu-cli/pkg/api"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	sdkconfig "github.com/blimu-dev/sdk-gen/pkg/config"
	"github.com/blimu-dev/sdk-gen/pkg/generator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed sdk-baseconfig.yml
var embeddedBaseConfig []byte

// GenerateCommand represents the generate command
type GenerateCommand struct {
	WorkspaceID   string
	EnvironmentID string
	Directory     string
}

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	cmd := &GenerateCommand{}

	cobraCmd := &cobra.Command{
		Use:   "generate [directory]",
		Short: "Generate SDKs from .blimu configuration for multiple languages",
		Long: `Generate custom SDKs based on your .blimu configuration files for multiple languages.

This command generates an OpenAPI specification from your Blimu configuration and then uses
a '.blimu/sdk.yml' file in the directory to determine which languages to generate and their options.

The '.blimu/sdk.yml' file must be present in the directory and defines the client configurations
for different languages (TypeScript, Go, Python, etc.).

Examples:
  # Generate SDKs for all languages defined in .blimu/sdk.yml (in current directory)
  blimu generate --workspace-id ws_123 --environment-id env_456

  # Generate SDKs using .blimu/sdk.yml from specific directory
  blimu generate /path/to/project --workspace-id ws_123 --environment-id env_456`,
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

func (c *GenerateCommand) Run(cmd *cobra.Command) error {
	fmt.Printf("🔧 Starting generate command in directory: %s\n", c.Directory)

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
		return fmt.Errorf("environment-id is required for SDK generation. Either:\n" +
			"  1. Provide --environment-id flag\n" +
			"  2. Configure your current environment with an ID using 'blimu env create --workspace-id <workspace-id> <env-name>'")
	}

	if c.WorkspaceID == "" {
		return fmt.Errorf("workspace-id is required for SDK generation. Provide --workspace-id flag.\n" +
			"Use 'blimu workspaces list' to find your workspace ID (when available)")
	}

	fmt.Printf("🔧 Generating SDK from database definitions...\n")

	// Check if dev mode is enabled
	devMode, _ := cmd.Flags().GetBool("dev")

	// Get auth client
	authClient, err := shared.GetAuthClientWithDevMode(devMode)
	if err != nil {
		return fmt.Errorf("authentication required for SDK generation. Run 'blimu auth login' first: %w", err)
	}

	// Get API client for direct HTTP calls
	apiClient := api.NewClient(authClient)

	// Generate OpenAPI spec from database (using GET endpoint)
	response, err := apiClient.GetOpenAPIFromDb(c.WorkspaceID, c.EnvironmentID)
	if err != nil {
		return fmt.Errorf("failed to generate OpenAPI spec: %w", err)
	}

	if !response.Success {
		fmt.Printf("❌ OpenAPI spec generation failed with %d error(s):\n\n", len(response.Errors))

		for i, errorData := range response.Errors {
			fmt.Printf("%d. %s\n", i+1, errorData.Message)
			if errorData.Resource != "" {
				fmt.Printf("   Resource: %s\n", errorData.Resource)
			}
			if errorData.Field != "" {
				fmt.Printf("   Field: %s\n", errorData.Field)
			}
			fmt.Printf("\n")
		}

		return fmt.Errorf("OpenAPI spec generation failed")
	}

	// Create temporary OpenAPI spec file for sdk-gen
	tempDir, err := os.MkdirTemp("", "blimu-openapi-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory

	specFile := filepath.Join(tempDir, "openapi.json")
	specJSON, err := json.MarshalIndent(response.Spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	if err := os.WriteFile(specFile, specJSON, 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec: %w", err)
	}

	fmt.Printf("📄 Generated OpenAPI specification\n")

	// Look for sdk.yml in the directory
	sdkConfigPath := filepath.Join(c.Directory, ".blimu", "sdk.yml")
	fmt.Printf("🔍 Looking for SDK config at: %s\n", sdkConfigPath)
	if _, statErr := os.Stat(sdkConfigPath); statErr == nil {
		// sdk.yml exists, use it for multi-language generation
		fmt.Printf("✅ Found SDK config, using multi-language generation\n")
		err = c.generateWithConfigFile(specFile, sdkConfigPath)
	} else {
		fmt.Printf("❌ SDK config not found: %v\n", statErr)
		return fmt.Errorf("no .blimu/sdk.yml found in %s", c.Directory)
	}

	if err != nil {
		return fmt.Errorf("failed to generate SDK: %w", err)
	}

	return nil
}

// generateWithConfigFile generates SDKs for multiple languages using an existing config file with custom OpenAPI spec
func (c *GenerateCommand) generateWithConfigFile(specFile, configPath string) error {
	fmt.Printf("🔧 Loading SDK config from: %s\n", configPath)

	// Read the config file content
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read SDK config file: %w", err)
	}

	// Parse the YAML content
	var configMap map[string]interface{}
	if err := yaml.Unmarshal(configData, &configMap); err != nil {
		return fmt.Errorf("failed to parse SDK config: %w", err)
	}

	// Get the directory containing the original config file
	configDir := filepath.Dir(configPath)
	fmt.Printf("📁 Config file directory: %s\n", configDir)

	// Load base config from embedded file
	baseConfig, err := loadBaseConfig()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not load base config: %v\n", err)
		fmt.Printf("   Continuing without base config merge...\n")
		baseConfig = make(map[string]interface{})
	} else {
		fmt.Printf("✅ Loaded base config\n")
	}

	// Merge base config with client-specific configs
	if clients, ok := configMap["clients"].([]interface{}); ok {
		fmt.Printf("📋 Found %d clients in config\n", len(clients))
		for i, clientInterface := range clients {
			if client, ok := clientInterface.(map[string]interface{}); ok {
				clientType := ""
				if t, ok := client["type"].(string); ok {
					clientType = t
				}

				if clientType == "" {
					return fmt.Errorf("clients[%d] missing required field 'type'", i)
				}

				// Find and merge base config for this client type
				mergedClient := mergeClientConfig(baseConfig, clientType, client, configDir)
				clients[i] = mergedClient

				if outDir, exists := mergedClient["outDir"]; exists {
					if outDirStr, ok := outDir.(string); ok {
						fmt.Printf("📁 %s client: %s\n", clientType, outDirStr)
					}
				}
			}
		}
	}

	// Add spec field if missing
	if _, exists := configMap["spec"]; !exists {
		configMap["spec"] = "placeholder"
	}

	// Marshal back to YAML
	resolvedConfigData, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal resolved config: %w", err)
	}

	// Create a temporary config file with resolved paths
	tempDir, err := os.MkdirTemp("", "blimu-sdk-config-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempConfigPath := filepath.Join(tempDir, "sdk.yml")
	if err := os.WriteFile(tempConfigPath, resolvedConfigData, 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	// Load the config with resolved paths
	cfg, err := sdkconfig.Load(tempConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load SDK config: %w", err)
	}

	// Replace the spec with our custom generated one
	cfg.Spec = specFile

	fmt.Printf("🔧 Generating SDKs for %d language(s)...\n", len(cfg.Clients))

	// Use sdk-gen service to generate from the modified config
	service := generator.NewService()
	err = service.GenerateFromConfig(cfg, "")
	if err != nil {
		return err
	}

	fmt.Printf("✅ Multi-language SDKs generated successfully!\n")
	for _, client := range cfg.Clients {
		fmt.Printf("  📁 %s: %s\n", client.Type, client.OutDir)
		fmt.Printf("  📦 Package: %s\n", client.PackageName)
		fmt.Printf("  🏗️  Client: %s\n", client.Name)
		fmt.Printf("\n")
	}

	return nil
}

// loadBaseConfig loads the base SDK configuration from the embedded sdk-baseconfig.yml file
func loadBaseConfig() (map[string]interface{}, error) {
	var baseConfig map[string]interface{}
	if err := yaml.Unmarshal(embeddedBaseConfig, &baseConfig); err != nil {
		return nil, fmt.Errorf("failed to parse embedded base config: %w", err)
	}

	return baseConfig, nil
}

// mergeClientConfig merges base configuration with client-specific configuration
// The client-specific config takes precedence over base config
func mergeClientConfig(baseConfig map[string]interface{}, clientType string, clientConfig map[string]interface{}, configDir string) map[string]interface{} {
	// Start with a copy of the client config (user's config takes precedence)
	merged := make(map[string]interface{})
	for k, v := range clientConfig {
		merged[k] = v
	}

	// Find base config for this client type
	// Try exact match (e.g., "typescript", "go", "typescript-types")
	var baseClientConfig map[string]interface{}
	if base, ok := baseConfig[clientType]; ok {
		if baseMap, ok := base.(map[string]interface{}); ok {
			baseClientConfig = baseMap
		}
	}

	// Merge base config into merged config (only if not already set in client config)
	if baseClientConfig != nil {
		for key, baseValue := range baseClientConfig {
			// Special handling for typeAugmentation - always merge (deep merge)
			if key == "typeAugmentation" {
				if baseAug, ok := baseValue.(map[string]interface{}); ok {
					if userAug, exists := merged[key]; exists {
						// Both exist - merge them (user takes precedence)
						if userAugMap, ok := userAug.(map[string]interface{}); ok {
							mergedAug := make(map[string]interface{})
							// Start with base values
							for k, v := range baseAug {
								mergedAug[k] = v
							}
							// Override with user values
							for k, v := range userAugMap {
								mergedAug[k] = v
							}
							merged[key] = mergedAug
						} else {
							merged[key] = baseAug
						}
					} else {
						// Only base exists
						merged[key] = baseAug
					}
				}
				continue
			}

			// Skip if client config already has this key (user override)
			if _, exists := merged[key]; exists {
				continue
			}

			// Special handling for postCommand - convert array to array format
			if key == "postCommand" {
				if baseArray, ok := baseValue.([]interface{}); ok {
					// Convert to string array format expected by sdk-gen
					postCmdArray := make([]string, len(baseArray))
					for i, item := range baseArray {
						if str, ok := item.(string); ok {
							postCmdArray[i] = str
						}
					}
					merged[key] = postCmdArray
				}
			} else {
				merged[key] = baseValue
			}
		}
	}

	// Handle outDir - required field, provide default if missing
	if outDir, exists := merged["outDir"]; !exists || outDir == "" {
		// Try to derive from typeAugmentation.outputFileName if available
		if typeAug, ok := merged["typeAugmentation"]; ok {
			if typeAugMap, ok := typeAug.(map[string]interface{}); ok {
				if outputFileName, ok := typeAugMap["outputFileName"].(string); ok && outputFileName != "" {
					// Use directory of outputFileName as outDir
					outDirPath := filepath.Dir(outputFileName)
					if outDirPath == "." || outDirPath == "" {
						// Default to config directory
						merged["outDir"] = configDir
					} else {
						// Resolve relative to config directory
						if filepath.IsAbs(outDirPath) {
							merged["outDir"] = outDirPath
						} else {
							merged["outDir"] = filepath.Join(configDir, outDirPath)
						}
					}
				} else {
					// Default to config directory
					merged["outDir"] = configDir
				}
			} else {
				// Default to config directory
				merged["outDir"] = configDir
			}
		} else {
			// Default to config directory
			merged["outDir"] = configDir
		}
	} else {
		// Resolve outDir path relative to config file location if it's relative
		if outDirStr, ok := outDir.(string); ok {
			if outDirStr == "" {
				merged["outDir"] = configDir
			} else if !filepath.IsAbs(outDirStr) {
				resolvedPath := filepath.Join(configDir, outDirStr)
				merged["outDir"] = resolvedPath
			}
		}
	}

	return merged
}
