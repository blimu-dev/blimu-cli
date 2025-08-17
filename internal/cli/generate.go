package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blimu-dev/blimu-cli/pkg/api"
	"github.com/blimu-dev/blimu-cli/pkg/auth"
	"github.com/blimu-dev/blimu-cli/pkg/blimu"
	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/blimu-dev/sdk-gen/pkg/generator"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate custom SDKs based on your .blimu configuration",
	Long: `Generate custom SDKs that include methods for your defined resources.
The SDKs will include CRUD operations for each resource defined in your .blimu/resources.yml file.

SDK configuration is defined in .blimu/config.yml. You can specify multiple clients
with different types (typescript, go) and output locations.

For example, if you have an 'organization' resource, you'll get:
- client.Organization.create()
- client.Organization.list()
- client.Organization.get()
- client.Organization.update()
- client.Organization.delete()`,
	RunE: runGenerate,
}

var (
	generateForce bool
)

func init() {
	generateCmd.Flags().BoolVarP(&generateForce, "force", "f", false, "Force generation even if output directories exist")
}

func runGenerate(cmd *cobra.Command, args []string) error {
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

	// Check if SDK configuration exists
	if blimuConfig.SDKConfig == nil || len(blimuConfig.SDKConfig.Clients) == 0 {
		return fmt.Errorf("no SDK configuration found in .blimu/config.yml. Please add SDK client configurations")
	}

	// Validate configuration
	result := blimu.ValidateConfig(blimuConfig)
	if !result.Valid {
		fmt.Printf("❌ Configuration has %d error(s). Run 'blimu validate' for details.\n", len(result.Errors))
		return fmt.Errorf("configuration is invalid")
	}

	// Check if output directories exist
	if !generateForce {
		for _, client := range blimuConfig.SDKConfig.Clients {
			if _, err := os.Stat(client.OutDir); err == nil {
				return fmt.Errorf("output directory %s already exists. Use --force to overwrite", client.OutDir)
			}
		}
	}

	// Load CLI config for API details
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	apiURL, apiKey, err := cliConfig.GetAPIClient()
	if err != nil {
		return fmt.Errorf("failed to get API client config: %w", err)
	}

	fmt.Printf("🔐 Authenticating with %s...\n", apiURL)

	// Create authenticated client and test connection
	authClient := auth.NewClient(apiURL, apiKey)
	if err := authClient.ValidateAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("✅ Authentication successful!")

	// Convert config to JSON
	configJSON, err := blimuConfig.MergeToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert config to JSON: %w", err)
	}

	// Create API client
	apiClient := api.NewClient(authClient)

	// Fetch base Blimu API OpenAPI spec (once for all clients)
	fmt.Println("📥 Fetching base Blimu API specification...")
	baseSpec, err := apiClient.FetchOpenAPISpec()
	if err != nil {
		return fmt.Errorf("failed to fetch base OpenAPI spec: %w", err)
	}

	fmt.Printf("🚀 Generating %d SDK(s)...\n", len(blimuConfig.SDKConfig.Clients))

	// Generate SDK for each client configuration
	for i, client := range blimuConfig.SDKConfig.Clients {
		fmt.Printf("\n[%d/%d] Generating %s SDK...\n", i+1, len(blimuConfig.SDKConfig.Clients), client.Type)
		fmt.Printf("   Output: %s\n", client.OutDir)
		fmt.Printf("   Package: %s\n", client.PackageName)
		fmt.Printf("   Client: %s\n", client.Name)

		// Request SDK generation via API
		sdkOptions := api.SDKGenerationOptions{
			Type:        client.Type,
			PackageName: client.PackageName,
			ClientName:  client.Name,
		}

		response, err := apiClient.GenerateSDK(configJSON, sdkOptions)
		if err != nil {
			return fmt.Errorf("failed to generate SDK via API for client %s: %w", client.Type, err)
		}

		if !response.Success {
			fmt.Printf("❌ SDK generation failed with %d error(s):\n\n", len(response.Errors))
			for i, err := range response.Errors {
				fmt.Printf("%d. %s.%s: %s\n", i+1, err.Resource, err.Field, err.Message)
			}
			return fmt.Errorf("SDK generation failed for client %s", client.Type)
		}

		// Merge base spec with custom resource spec
		fmt.Println("🔄 Merging base API with custom resource specification...")
		mergedSpec, err := config.MergeOpenAPISpecs(baseSpec, response.Spec)
		if err != nil {
			return fmt.Errorf("failed to merge OpenAPI specs: %w", err)
		}

		// Generate SDK locally
		err = generateSDKFromSpec(mergedSpec, client.OutDir, client.PackageName, client.Name, client.Type)
		if err != nil {
			return fmt.Errorf("failed to generate SDK from spec for client %s: %w", client.Type, err)
		}

		fmt.Printf("✅ %s SDK generated successfully!\n", client.Type)
	}

	fmt.Println("\n🎉 All SDKs generated successfully!")
	fmt.Printf("\n📝 Next steps:\n")

	// Show instructions for each generated SDK
	for _, client := range blimuConfig.SDKConfig.Clients {
		fmt.Printf("\n%s SDK (%s):\n", client.Type, client.OutDir)
		fmt.Printf("  1. cd %s\n", client.OutDir)

		if client.Type == "typescript" {
			fmt.Printf("  2. npm install\n")
			fmt.Printf("  3. Import and use your client:\n")
			fmt.Printf("     import { %s } from './%s';\n", client.Name, client.PackageName)
			fmt.Printf("     const client = new %s({ baseURL: '%s', apiKey: 'your-key' });\n",
				client.Name, getBaseURL(blimuConfig.SDKConfig, apiURL))
		} else if client.Type == "go" {
			fmt.Printf("  2. go mod tidy\n")
			fmt.Printf("  3. Import and use your client:\n")
			fmt.Printf("     import \"%s\"\n", client.PackageName)
			fmt.Printf("     client := %s.NewClient(\"%s\", \"your-api-key\")\n",
				client.PackageName, getBaseURL(blimuConfig.SDKConfig, apiURL))
		}
	}

	// Show available resources
	fmt.Printf("\n🎯 Available resources in your SDKs:\n")
	for resourceName := range blimuConfig.Resources {
		resourceTitle := capitalizeFirst(resourceName)
		resourceTag := fmt.Sprintf("Resources.%s", resourceTitle)
		fmt.Printf("  • client.%s.create(), .list(), .get(), .update(), .delete()\n", resourceTag)
	}

	return nil
}

func generateSDKFromSpec(spec map[string]interface{}, outputDir, packageName, clientName, sdkType string) error {
	// Create temporary directory for the spec
	tempDir, err := os.MkdirTemp("", "blimu-spec-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Save spec to temporary file
	specPath := filepath.Join(tempDir, "custom-spec.json")
	specJSON, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	if err := os.WriteFile(specPath, specJSON, 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	// Generate SDK using sdk-gen
	sdkOptions := generator.GenerateSDKOptions{
		Spec:        specPath,
		Type:        sdkType,
		OutDir:      outputDir,
		PackageName: packageName,
		Name:        clientName,
	}

	if err := generator.GenerateSDK(sdkOptions); err != nil {
		return fmt.Errorf("failed to generate SDK: %w", err)
	}

	return nil
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:] // Simple ASCII uppercase conversion
}

func getBaseURL(sdkConfig *config.SDKConfig, apiURL string) string {
	if sdkConfig.BaseURL != "" {
		return sdkConfig.BaseURL
	}
	return apiURL
}
