package config

import (
	"encoding/json"
	"fmt"
)

// ConfigPayload represents the complete configuration payload to send to the API
type ConfigPayload struct {
	Resources    map[string]ResourceConfig    `json:"resources"`
	Entitlements map[string]EntitlementConfig `json:"entitlements"`
	Features     map[string]FeatureConfig     `json:"features"`
	Plans        map[string]PlanConfig        `json:"plans"`
	Version      string                       `json:"version"`
}

// MergeToJSON converts the BlimuConfig to a JSON payload for API submission
func (config *BlimuConfig) MergeToJSON() ([]byte, error) {
	payload := ConfigPayload{
		Resources:    config.Resources,
		Entitlements: config.Entitlements,
		Features:     config.Features,
		Plans:        config.Plans,
		Version:      "1.0", // Config schema version
	}

	// Ensure empty maps are not nil for JSON serialization
	if payload.Resources == nil {
		payload.Resources = make(map[string]ResourceConfig)
	}
	if payload.Entitlements == nil {
		payload.Entitlements = make(map[string]EntitlementConfig)
	}
	if payload.Features == nil {
		payload.Features = make(map[string]FeatureConfig)
	}
	if payload.Plans == nil {
		payload.Plans = make(map[string]PlanConfig)
	}

	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return jsonData, nil
}

// ValidateForAPI performs basic client-side validation before sending to API
func (config *BlimuConfig) ValidateForAPI() error {
	// Basic validation - resources must exist
	if len(config.Resources) == 0 {
		return fmt.Errorf("at least one resource must be defined")
	}

	// Validate resource names are not empty
	for resourceName := range config.Resources {
		if resourceName == "" {
			return fmt.Errorf("resource names cannot be empty")
		}
	}

	// Validate entitlement names follow resource:action format
	for entitlementName := range config.Entitlements {
		if !isValidEntitlementName(entitlementName) {
			return fmt.Errorf("entitlement '%s' must follow format 'resource:action'", entitlementName)
		}
	}

	return nil
}

func isValidEntitlementName(name string) bool {
	// Basic validation - must contain exactly one colon
	colonCount := 0
	for _, char := range name {
		if char == ':' {
			colonCount++
		}
	}
	return colonCount == 1 && len(name) > 2
}

// MergeOpenAPISpecs merges a base OpenAPI spec with a custom resource spec
// The custom spec paths and components are added to the base spec
func MergeOpenAPISpecs(baseSpec, customSpec map[string]interface{}) (map[string]interface{}, error) {
	// Create a deep copy of the base spec to avoid modifying the original
	mergedSpec := make(map[string]interface{})
	for k, v := range baseSpec {
		mergedSpec[k] = v
	}

	// Merge paths
	if customPaths, ok := customSpec["paths"].(map[string]interface{}); ok {
		if basePaths, ok := mergedSpec["paths"].(map[string]interface{}); ok {
			// Merge custom paths into base paths
			for path, pathSpec := range customPaths {
				basePaths[path] = pathSpec
			}
		} else {
			// If base spec has no paths, use custom paths
			mergedSpec["paths"] = customPaths
		}
	}

	// Merge components (schemas, security schemes, etc.)
	if customComponents, ok := customSpec["components"].(map[string]interface{}); ok {
		if baseComponents, ok := mergedSpec["components"].(map[string]interface{}); ok {
			// Merge each component type
			for componentType, customComponentSpecs := range customComponents {
				if customSpecs, ok := customComponentSpecs.(map[string]interface{}); ok {
					if baseSpecs, ok := baseComponents[componentType].(map[string]interface{}); ok {
						// Merge custom component specs into base component specs
						for specName, specDef := range customSpecs {
							baseSpecs[specName] = specDef
						}
					} else {
						// If base doesn't have this component type, add it
						baseComponents[componentType] = customSpecs
					}
				}
			}
		} else {
			// If base spec has no components, use custom components
			mergedSpec["components"] = customComponents
		}
	}

	// Merge tags (if present in custom spec)
	if customTags, ok := customSpec["tags"].([]interface{}); ok {
		if baseTags, ok := mergedSpec["tags"].([]interface{}); ok {
			// Append custom tags to base tags (avoiding duplicates)
			existingTags := make(map[string]bool)
			for _, tag := range baseTags {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					if name, ok := tagMap["name"].(string); ok {
						existingTags[name] = true
					}
				}
			}

			for _, tag := range customTags {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					if name, ok := tagMap["name"].(string); ok {
						if !existingTags[name] {
							baseTags = append(baseTags, tag)
						}
					}
				}
			}
			mergedSpec["tags"] = baseTags
		} else {
			// If base spec has no tags, use custom tags
			mergedSpec["tags"] = customTags
		}
	}

	return mergedSpec, nil
}
