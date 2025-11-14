package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	platform "github.com/blimu-dev/blimu-platform-go"
	"github.com/spf13/cobra"
)

// GenerateCommand represents the generate command
type GenerateCommand struct {
	WorkspaceID   string
	EnvironmentID string
	Directory     string
	OutputDir     string
	SDKType       string
	PackageName   string
	ClientName    string
}

// NewGenerateCmd creates the generate command
func NewGenerateCmd() *cobra.Command {
	cmd := &GenerateCommand{}

	cobraCmd := &cobra.Command{
		Use:   "generate [directory]",
		Short: "Generate SDK from .blimu configuration",
		Long: `Generate a custom SDK based on your .blimu configuration files.

This command generates an OpenAPI specification from your Blimu configuration
and can be used to create type-safe SDKs for your resources.`,
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

	cobraCmd.Flags().StringVar(&cmd.WorkspaceID, "workspace-id", "", "Workspace ID (required)")
	cobraCmd.Flags().StringVar(&cmd.EnvironmentID, "environment-id", "", "Environment ID (required)")
	cobraCmd.Flags().StringVar(&cmd.OutputDir, "output", "./generated", "Output directory for generated files")
	cobraCmd.Flags().StringVar(&cmd.SDKType, "type", "typescript", "SDK type (typescript, go, python)")
	cobraCmd.Flags().StringVar(&cmd.PackageName, "package", "blimu-client", "Package name for generated SDK")
	cobraCmd.Flags().StringVar(&cmd.ClientName, "client", "BlimuClient", "Client class name")

	return cobraCmd
}

func (c *GenerateCommand) Run() error {
	if c.WorkspaceID == "" || c.EnvironmentID == "" {
		return fmt.Errorf("workspace-id and environment-id are required for SDK generation")
	}

	// Load Blimu configuration
	blimuConfig, err := config.LoadBlimuConfig(c.Directory)
	if err != nil {
		return fmt.Errorf("failed to load .blimu configuration: %w", err)
	}

	fmt.Printf("🔧 Generating %s SDK from configuration in %s...\n", c.SDKType, c.Directory)

	// Get auth client
	authClient, err := shared.GetAuthClient()
	if err != nil {
		return fmt.Errorf("authentication required for SDK generation. Run 'blimu auth login' first: %w", err)
	}

	// Get platform SDK client
	sdk := authClient.GetPlatformSDK()
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

	// Build the SDK generation request
	request := platform.DefinitionGenerateSdkRequestDto{
		Resources:    make(map[string]interface{}),
		Entitlements: make(map[string]interface{}),
		Features:     make(map[string]interface{}),
		Plans:        make(map[string]interface{}),
		Namespace:    getString(configMap, "namespace"),
		Version:      getString(configMap, "version"),
		SdkOptions: map[string]interface{}{
			"type":         c.SDKType,
			"package_name": c.PackageName,
			"client_name":  c.ClientName,
		},
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

	// Generate OpenAPI spec
	response, err := sdk.Definitions.GetOpenApi(c.WorkspaceID, c.EnvironmentID, request)
	if err != nil {
		return fmt.Errorf("failed to generate SDK: %w", err)
	}

	if !response.Success {
		fmt.Printf("❌ SDK generation failed with %d error(s):\n\n", len(response.Errors))

		for i, errorData := range response.Errors {
			message := getStringFromMap(errorData, "message")
			resource := getStringFromMap(errorData, "resource")
			field := getStringFromMap(errorData, "field")

			fmt.Printf("%d. %s\n", i+1, message)
			if resource != "" {
				fmt.Printf("   Resource: %s\n", resource)
			}
			if field != "" {
				fmt.Printf("   Field: %s\n", field)
			}
			fmt.Printf("\n")
		}

		return fmt.Errorf("SDK generation failed")
	}

	// Create output directory
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write OpenAPI spec to file
	specFile := filepath.Join(c.OutputDir, "openapi.json")
	specJSON, err := json.MarshalIndent(response.Spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	if err := os.WriteFile(specFile, specJSON, 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec: %w", err)
	}

	fmt.Printf("✅ SDK generation completed successfully!\n")
	fmt.Printf("📄 OpenAPI specification: %s\n", specFile)
	fmt.Printf("🔧 SDK Type: %s\n", c.SDKType)
	fmt.Printf("📦 Package: %s\n", c.PackageName)
	fmt.Printf("🏗️  Client: %s\n", c.ClientName)

	fmt.Printf("\n💡 Use the generated OpenAPI spec with your preferred SDK generator:\n")
	switch c.SDKType {
	case "typescript":
		fmt.Printf("   npx @openapitools/openapi-generator-cli generate -i %s -g typescript-axios -o %s\n", specFile, c.OutputDir)
	case "go":
		fmt.Printf("   openapi-generator generate -i %s -g go -o %s\n", specFile, c.OutputDir)
	case "python":
		fmt.Printf("   openapi-generator generate -i %s -g python -o %s\n", specFile, c.OutputDir)
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

// getStringFromMap safely extracts a string value from a map[string]interface{}
func getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
