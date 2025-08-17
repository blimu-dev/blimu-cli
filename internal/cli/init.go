package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new .blimu configuration and environment",
	Long: `Initialize a new .blimu configuration directory with a basic resources.yml file.
This creates a starting point for defining your resource structure and optionally
sets up a new environment for connecting to the Blimu API.`,
	RunE: runInit,
}

var (
	initForce     bool
	initEnvName   string
	initAPIKey    string
	initAPIURL    string
	initEnvID     string
	initLookupKey string
)

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force initialization even if .blimu directory already exists")
	initCmd.Flags().StringVar(&initEnvName, "env-name", "", "Environment name to create (optional)")
	initCmd.Flags().StringVar(&initAPIKey, "api-key", "", "API key for the environment (optional)")
	initCmd.Flags().StringVar(&initAPIURL, "api-url", "", "API URL for the environment (defaults to https://api.blimu.dev)")
	initCmd.Flags().StringVar(&initLookupKey, "lookup-key", "", "Environment lookup key (optional)")
	initCmd.Flags().StringVar(&initEnvID, "env-id", "", "Environment ID from the API (optional)")
}

func runInit(cmd *cobra.Command, args []string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	blimuDir := filepath.Join(currentDir, ".blimu")

	// Check if .blimu directory already exists
	if _, err := os.Stat(blimuDir); err == nil && !initForce {
		return fmt.Errorf(".blimu directory already exists. Use --force to overwrite")
	}

	// Create the base configuration
	baseConfig := &config.BlimuConfig{
		Resources: map[string]config.ResourceConfig{
			"organization": {
				Roles: []string{"admin", "editor", "viewer"},
			},
			"projects": {
				Roles: []string{"admin", "editor", "viewer"},
				RolesInheritance: map[string][]string{
					"editor": {"organization->admin"},
					"viewer": {"organization->editor"},
				},
				Parents: map[string]config.ParentConfig{
					"organization": {Required: true},
				},
			},
		},
		Plans: map[string]config.PlanConfig{
			"starter": {
				Name:        "Starter Plan",
				Description: "Perfect for getting started",
			},
			"pro": {
				Name:        "Pro Plan",
				Description: "For growing teams",
			},
		},
		Entitlements: map[string]config.EntitlementConfig{
			"organization:create_project": {
				Roles: []string{"admin"},
				Plans: []string{"pro"},
			},
			"projects:delete": {
				Roles: []string{"admin"},
			},
		},
		Features: map[string]config.FeatureConfig{
			"project_management": {
				Plans:          []string{"pro"},
				DefaultEnabled: false,
				Entitlements:   []string{"organization:create_project", "projects:delete"},
			},
		},
		SDKConfig: &config.SDKConfig{
			Name:    "My Project SDKs",
			BaseURL: "https://api.blimu.dev",
			Clients: []config.SDKClient{
				{
					Type:           "typescript",
					OutDir:         "./sdk-ts",
					PackageName:    "my-project-sdk",
					Name:           "MyProjectClient",
					PostGenCommand: "npx prettier --write .",
				},
				{
					Type:           "go",
					OutDir:         "./sdk-go",
					PackageName:    "github.com/myorg/my-project-sdk-go",
					ModuleName:     "github.com/myorg/my-project-sdk-go",
					Name:           "MyProjectClient",
					PostGenCommand: "goimports -w .",
				},
			},
		},
	}

	// Save the configuration
	if err := config.SaveBlimuConfig(currentDir, baseConfig); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Initialized .blimu configuration in %s\n", blimuDir)
	fmt.Println("\n📁 Created files:")
	fmt.Println("  • resources.yml - Define your resources and roles")
	fmt.Println("  • entitlements.yml - Define permissions and access control")
	fmt.Println("  • features.yml - Define features with plan scoping")
	fmt.Println("  • plans.yml - Define your billing plans")
	fmt.Println("  • config.yml - Define SDK generation settings")

	// Handle environment setup if requested
	if initEnvName != "" {
		if err := setupEnvironment(); err != nil {
			fmt.Printf("⚠️  Configuration created but environment setup failed: %v\n", err)
			fmt.Println("You can set up the environment later using 'blimu env create'")
		}
	}

	fmt.Println("\n📝 Next steps:")
	fmt.Println("  1. Edit the .blimu/*.yml files to match your needs")
	if initEnvName == "" {
		fmt.Println("  2. Set up an environment: 'blimu env create <name> --api-key <key>'")
		fmt.Println("  3. Run 'blimu validate' to check your configuration")
		fmt.Println("  4. Run 'blimu auth push' to deploy your configuration")
		fmt.Println("  5. Run 'blimu generate' to create your custom SDKs")
	} else {
		fmt.Println("  2. Run 'blimu validate' to check your configuration")
		fmt.Println("  3. Run 'blimu auth push' to deploy your configuration")
		fmt.Println("  4. Run 'blimu generate' to create your custom SDKs")
	}
	fmt.Println("\n🔗 Configuration relationships:")
	fmt.Println("  • Entitlements reference resources and plans")
	fmt.Println("  • Features reference plans and entitlements")
	fmt.Println("  • Resources define the foundation for everything")
	fmt.Println("  • Config.yml defines which SDKs to generate and where")

	return nil
}

func setupEnvironment() error {
	// Get API key from flag or environment variable
	apiKey := initAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("BLIMU_SECRET_KEY")
	}

	if apiKey == "" {
		return fmt.Errorf("API key is required. Provide it via --api-key flag or BLIMU_SECRET_KEY environment variable")
	}

	// Load or create CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		// If config doesn't exist, create a new one
		cliConfig = &config.CLIConfig{
			Environments:  make(map[string]config.Environment),
			DefaultAPIURL: "https://api.blimu.dev",
		}
	}

	// Create environment
	env := config.Environment{
		Name:      initEnvName,
		APIKey:    apiKey,
		APIURL:    initAPIURL,
		ID:        initEnvID,
		LookupKey: initLookupKey,
	}

	if err := cliConfig.AddEnvironment(initEnvName, env); err != nil {
		return fmt.Errorf("failed to add environment: %w", err)
	}

	fmt.Printf("✅ Created environment '%s'\n", initEnvName)
	if cliConfig.CurrentEnvironment == initEnvName {
		fmt.Printf("   Set as current environment\n")
	}

	return nil
}
