package blimu

import (
	"fmt"
	"strings"

	"github.com/blimu-dev/blimu-cli/pkg/config"
)

// ValidationError represents a validation error
type ValidationError struct {
	Resource string
	Field    string
	Message  string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s.%s: %s", e.Resource, e.Field, e.Message)
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// ValidateConfig validates a complete Blimu configuration
func ValidateConfig(config *config.BlimuConfig) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// Validate resources
	for resourceName, resourceConfig := range config.Resources {
		validateResource(resourceName, resourceConfig, config.Resources, result)
	}

	// Validate entitlements
	for entitlementName, entitlementConfig := range config.Entitlements {
		validateEntitlement(entitlementName, entitlementConfig, config, result)
	}

	// Validate features
	for featureName, featureConfig := range config.Features {
		validateFeature(featureName, featureConfig, config, result)
	}

	// Validate plans (basic validation - ensure they have names)
	for planName, planConfig := range config.Plans {
		validatePlan(planName, planConfig, result)
	}

	// Validate SDK configuration
	if config.SDKConfig != nil {
		validateSDKConfig(config.SDKConfig, result)
	}

	result.Valid = len(result.Errors) == 0
	return result
}

func validateResource(name string, resource config.ResourceConfig, allResources map[string]config.ResourceConfig, result *ValidationResult) {
	// Validate roles
	if len(resource.Roles) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Resource: name,
			Field:    "roles",
			Message:  "at least one role must be defined",
		})
	}

	// Validate role inheritance
	for role, inheritances := range resource.RolesInheritance {
		if !contains(resource.Roles, role) {
			result.Errors = append(result.Errors, ValidationError{
				Resource: name,
				Field:    "roles_inheritance",
				Message:  fmt.Sprintf("role '%s' not found in roles list", role),
			})
		}

		for _, inheritance := range inheritances {
			if err := validateInheritance(inheritance, allResources); err != nil {
				result.Errors = append(result.Errors, ValidationError{
					Resource: name,
					Field:    "roles_inheritance",
					Message:  fmt.Sprintf("invalid inheritance '%s': %s", inheritance, err),
				})
			}
		}
	}

	// Validate parents
	for parentName, parentConfig := range resource.Parents {
		if _, exists := allResources[parentName]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Resource: name,
				Field:    "parents",
				Message:  fmt.Sprintf("parent resource '%s' not found", parentName),
			})
		}

		// Check for circular dependencies - a resource cannot be its own ancestor
		if hasCircularDependency(parentName, name, allResources, map[string]bool{}) {
			result.Errors = append(result.Errors, ValidationError{
				Resource: name,
				Field:    "parents",
				Message:  fmt.Sprintf("circular dependency detected with parent '%s'", parentName),
			})
		}

		_ = parentConfig // parentConfig is used for future validations
	}
}

func validateInheritance(inheritance string, allResources map[string]config.ResourceConfig) error {
	parts := strings.Split(inheritance, "->")
	if len(parts) != 2 {
		return fmt.Errorf("inheritance must be in format 'resource->role'")
	}

	resourceName := strings.TrimSpace(parts[0])
	roleName := strings.TrimSpace(parts[1])

	resource, exists := allResources[resourceName]
	if !exists {
		return fmt.Errorf("resource '%s' not found", resourceName)
	}

	if !contains(resource.Roles, roleName) {
		return fmt.Errorf("role '%s' not found in resource '%s'", roleName, resourceName)
	}

	return nil
}

func hasCircularDependency(current, target string, allResources map[string]config.ResourceConfig, visited map[string]bool) bool {
	if visited[current] {
		return current == target
	}

	visited[current] = true

	resource, exists := allResources[current]
	if !exists {
		delete(visited, current)
		return false
	}

	for parentName := range resource.Parents {
		if parentName == target {
			delete(visited, current)
			return true
		}
		if hasCircularDependency(parentName, target, allResources, visited) {
			delete(visited, current)
			return true
		}
	}

	delete(visited, current)
	return false
}

func validateEntitlement(name string, entitlement config.EntitlementConfig, config *config.BlimuConfig, result *ValidationResult) {
	// Parse entitlement name (format: resource:action)
	parts := strings.Split(name, ":")
	if len(parts) != 2 {
		result.Errors = append(result.Errors, ValidationError{
			Resource: "entitlements",
			Field:    name,
			Message:  "entitlement name must be in format 'resource:action'",
		})
		return
	}

	resourceName := parts[0]

	// Check if the resource exists
	if _, exists := config.Resources[resourceName]; !exists {
		result.Errors = append(result.Errors, ValidationError{
			Resource: "entitlements",
			Field:    name,
			Message:  fmt.Sprintf("resource '%s' not found in resources.yml", resourceName),
		})
	}

	// Validate roles exist in the resource
	if resource, exists := config.Resources[resourceName]; exists {
		for _, role := range entitlement.Roles {
			if !contains(resource.Roles, role) {
				result.Errors = append(result.Errors, ValidationError{
					Resource: "entitlements",
					Field:    name,
					Message:  fmt.Sprintf("role '%s' not found in resource '%s'", role, resourceName),
				})
			}
		}
	}

	// Validate plans exist
	for _, plan := range entitlement.Plans {
		if _, exists := config.Plans[plan]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "entitlements",
				Field:    name,
				Message:  fmt.Sprintf("plan '%s' not found in plans.yml", plan),
			})
		}
	}
}

func validateFeature(name string, feature config.FeatureConfig, config *config.BlimuConfig, result *ValidationResult) {
	// Validate plans exist
	for _, plan := range feature.Plans {
		if _, exists := config.Plans[plan]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "features",
				Field:    name,
				Message:  fmt.Sprintf("plan '%s' not found in plans.yml", plan),
			})
		}
	}

	// Validate entitlements exist
	for _, entitlement := range feature.Entitlements {
		if _, exists := config.Entitlements[entitlement]; !exists {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "features",
				Field:    name,
				Message:  fmt.Sprintf("entitlement '%s' not found in entitlements.yml", entitlement),
			})
		}
	}
}

func validatePlan(name string, plan config.PlanConfig, result *ValidationResult) {
	// Validate plan has a name
	if strings.TrimSpace(plan.Name) == "" {
		result.Errors = append(result.Errors, ValidationError{
			Resource: "plans",
			Field:    name,
			Message:  "plan must have a name",
		})
	}

	// Validate plan has a description
	if strings.TrimSpace(plan.Description) == "" {
		result.Errors = append(result.Errors, ValidationError{
			Resource: "plans",
			Field:    name,
			Message:  "plan must have a description",
		})
	}
}

func validateSDKConfig(sdkConfig *config.SDKConfig, result *ValidationResult) {
	if len(sdkConfig.Clients) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Resource: "config",
			Field:    "clients",
			Message:  "at least one client must be defined",
		})
		return
	}

	// Validate each client
	for i, client := range sdkConfig.Clients {
		clientName := fmt.Sprintf("clients[%d]", i)

		// Validate required fields
		if strings.TrimSpace(client.Type) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "config",
				Field:    clientName + ".type",
				Message:  "client type is required",
			})
		} else {
			// Validate supported types
			supportedTypes := []string{"typescript", "go"}
			if !contains(supportedTypes, client.Type) {
				result.Errors = append(result.Errors, ValidationError{
					Resource: "config",
					Field:    clientName + ".type",
					Message:  fmt.Sprintf("unsupported client type '%s'. Supported types: %s", client.Type, strings.Join(supportedTypes, ", ")),
				})
			}
		}

		if strings.TrimSpace(client.OutDir) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "config",
				Field:    clientName + ".outDir",
				Message:  "output directory is required",
			})
		}

		if strings.TrimSpace(client.PackageName) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "config",
				Field:    clientName + ".packageName",
				Message:  "package name is required",
			})
		}

		if strings.TrimSpace(client.Name) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "config",
				Field:    clientName + ".name",
				Message:  "client name is required",
			})
		}

		// For Go clients, module name is required
		if client.Type == "go" && strings.TrimSpace(client.ModuleName) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Resource: "config",
				Field:    clientName + ".moduleName",
				Message:  "module name is required for Go clients",
			})
		}
	}

	// Check for duplicate output directories
	outDirs := make(map[string]int)
	for i, client := range sdkConfig.Clients {
		if client.OutDir != "" {
			if existingIndex, exists := outDirs[client.OutDir]; exists {
				result.Errors = append(result.Errors, ValidationError{
					Resource: "config",
					Field:    fmt.Sprintf("clients[%d].outDir", i),
					Message:  fmt.Sprintf("output directory '%s' is already used by clients[%d]", client.OutDir, existingIndex),
				})
			} else {
				outDirs[client.OutDir] = i
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
