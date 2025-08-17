package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// BlimuConfig represents the complete .blimu configuration
type BlimuConfig struct {
	Resources    map[string]ResourceConfig    `yaml:"-"`
	Entitlements map[string]EntitlementConfig `yaml:"-"`
	Features     map[string]FeatureConfig     `yaml:"-"`
	Plans        map[string]PlanConfig        `yaml:"-"`
	SDKConfig    *SDKConfig                   `yaml:"-"`
}

// ResourceConfig represents a single resource configuration
type ResourceConfig struct {
	Roles            []string                `yaml:"roles,omitempty" json:"roles,omitempty"`
	RolesInheritance map[string][]string     `yaml:"roles_inheritance,omitempty" json:"roles_inheritance,omitempty"`
	Parents          map[string]ParentConfig `yaml:"parents,omitempty" json:"parents,omitempty"`
}

// ParentConfig represents parent resource configuration
type ParentConfig struct {
	Required bool `yaml:"required" json:"required"`
}

// EntitlementConfig represents an entitlement configuration
type EntitlementConfig struct {
	Roles []string `yaml:"roles,omitempty" json:"roles,omitempty"`
	Plans []string `yaml:"plans,omitempty" json:"plans,omitempty"`
}

// FeatureConfig represents a feature configuration
type FeatureConfig struct {
	Plans          []string `yaml:"plans,omitempty" json:"plans,omitempty"`
	DefaultEnabled bool     `yaml:"default_enabled,omitempty" json:"default_enabled,omitempty"`
	Entitlements   []string `yaml:"entitlements,omitempty" json:"entitlements,omitempty"`
}

// PlanConfig represents a plan configuration
type PlanConfig struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
}

// SDKConfig represents SDK generation configuration
type SDKConfig struct {
	Name    string      `yaml:"name,omitempty"`
	BaseURL string      `yaml:"baseURL,omitempty"`
	Clients []SDKClient `yaml:"clients,omitempty"`
}

// SDKClient represents configuration for a single client SDK
type SDKClient struct {
	Type              string   `yaml:"type"`
	OutDir            string   `yaml:"outDir"`
	PackageName       string   `yaml:"packageName"`
	ModuleName        string   `yaml:"moduleName,omitempty"`
	Name              string   `yaml:"name"`
	IncludeTags       []string `yaml:"includeTags,omitempty"`
	ExcludeTags       []string `yaml:"excludeTags,omitempty"`
	IncludeQueryKeys  bool     `yaml:"includeQueryKeys,omitempty"`
	OperationIDParser string   `yaml:"operationIdParser,omitempty"`
	PostGenCommand    string   `yaml:"postGenCommand,omitempty"`
}

// Legacy CLIConfig - now replaced by enhanced version in cli_config.go
// Keeping for backward compatibility during transition

// LoadBlimuConfig loads all .blimu configuration files
func LoadBlimuConfig(dir string) (*BlimuConfig, error) {
	blimuDir := filepath.Join(dir, ".blimu")
	config := &BlimuConfig{}

	// Load resources.yml
	if err := loadResourcesConfig(blimuDir, config); err != nil {
		return nil, err
	}

	// Load entitlements.yml (optional)
	if err := loadEntitlementsConfig(blimuDir, config); err != nil {
		return nil, err
	}

	// Load features.yml (optional)
	if err := loadFeaturesConfig(blimuDir, config); err != nil {
		return nil, err
	}

	// Load plans.yml (optional)
	if err := loadPlansConfig(blimuDir, config); err != nil {
		return nil, err
	}

	// Load config.yml (optional)
	if err := loadSDKConfig(blimuDir, config); err != nil {
		return nil, err
	}

	return config, nil
}

func loadResourcesConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "resources.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read resources.yml: %w", err)
	}

	if err := yaml.Unmarshal(data, &config.Resources); err != nil {
		return fmt.Errorf("failed to parse resources.yml: %w", err)
	}

	return nil
}

func loadEntitlementsConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "entitlements.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config.Entitlements = make(map[string]EntitlementConfig)
			return nil
		}
		return fmt.Errorf("failed to read entitlements.yml: %w", err)
	}

	if err := yaml.Unmarshal(data, &config.Entitlements); err != nil {
		return fmt.Errorf("failed to parse entitlements.yml: %w", err)
	}

	return nil
}

func loadFeaturesConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "features.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config.Features = make(map[string]FeatureConfig)
			return nil
		}
		return fmt.Errorf("failed to read features.yml: %w", err)
	}

	if err := yaml.Unmarshal(data, &config.Features); err != nil {
		return fmt.Errorf("failed to parse features.yml: %w", err)
	}

	return nil
}

func loadPlansConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "plans.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config.Plans = make(map[string]PlanConfig)
			return nil
		}
		return fmt.Errorf("failed to read plans.yml: %w", err)
	}

	if err := yaml.Unmarshal(data, &config.Plans); err != nil {
		return fmt.Errorf("failed to parse plans.yml: %w", err)
	}

	return nil
}

func loadSDKConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No config.yml file, that's okay
			return nil
		}
		return fmt.Errorf("failed to read config.yml: %w", err)
	}

	var sdkConfig SDKConfig
	if err := yaml.Unmarshal(data, &sdkConfig); err != nil {
		return fmt.Errorf("failed to parse config.yml: %w", err)
	}

	config.SDKConfig = &sdkConfig
	return nil
}

// SaveBlimuConfig saves all .blimu configuration files
func SaveBlimuConfig(dir string, config *BlimuConfig) error {
	blimuDir := filepath.Join(dir, ".blimu")
	if err := os.MkdirAll(blimuDir, 0755); err != nil {
		return fmt.Errorf("failed to create .blimu directory: %w", err)
	}

	// Save resources.yml
	if err := saveResourcesConfig(blimuDir, config); err != nil {
		return err
	}

	// Save entitlements.yml if not empty
	if len(config.Entitlements) > 0 {
		if err := saveEntitlementsConfig(blimuDir, config); err != nil {
			return err
		}
	}

	// Save features.yml if not empty
	if len(config.Features) > 0 {
		if err := saveFeaturesConfig(blimuDir, config); err != nil {
			return err
		}
	}

	// Save plans.yml if not empty
	if len(config.Plans) > 0 {
		if err := savePlansConfig(blimuDir, config); err != nil {
			return err
		}
	}

	// Save config.yml if SDKConfig exists
	if config.SDKConfig != nil {
		if err := saveSDKConfig(blimuDir, config); err != nil {
			return err
		}
	}

	return nil
}

func saveResourcesConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "resources.yml")
	data, err := yaml.Marshal(config.Resources)
	if err != nil {
		return fmt.Errorf("failed to marshal resources config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write resources.yml: %w", err)
	}

	return nil
}

func saveEntitlementsConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "entitlements.yml")
	data, err := yaml.Marshal(config.Entitlements)
	if err != nil {
		return fmt.Errorf("failed to marshal entitlements config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write entitlements.yml: %w", err)
	}

	return nil
}

func saveFeaturesConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "features.yml")
	data, err := yaml.Marshal(config.Features)
	if err != nil {
		return fmt.Errorf("failed to marshal features config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write features.yml: %w", err)
	}

	return nil
}

func savePlansConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "plans.yml")
	data, err := yaml.Marshal(config.Plans)
	if err != nil {
		return fmt.Errorf("failed to marshal plans config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write plans.yml: %w", err)
	}

	return nil
}

func saveSDKConfig(blimuDir string, config *BlimuConfig) error {
	configPath := filepath.Join(blimuDir, "config.yml")
	data, err := yaml.Marshal(config.SDKConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal SDK config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config.yml: %w", err)
	}

	return nil
}

// FindBlimuConfig searches for .blimu directory in current and parent directories
func FindBlimuConfig(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, ".blimu", "resources.yml")
		if _, err := os.Stat(configPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf(".blimu directory not found")
}

// LoadLegacyCLIConfig - legacy function, use LoadCLIConfig from cli_config.go instead

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
