package shared

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// EnvironmentInfo represents an environment with its metadata
type EnvironmentInfo struct {
	ID          string
	Name        string
	WorkspaceID string
	IsLocal     bool
	IsActive    bool
}

// FetchUserEnvironments fetches environments available to the user
func FetchUserEnvironments(devMode bool) ([]EnvironmentInfo, error) {
	var environments []EnvironmentInfo

	// Get current CLI config
	cliConfig, _, err := GetCurrentEnvironmentInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get current environment info: %w", err)
	}

	currentEnvName := cliConfig.CurrentEnvironment

	// Add local environments
	for name, env := range cliConfig.Environments {
		environments = append(environments, EnvironmentInfo{
			ID:          env.ID,
			Name:        name,
			WorkspaceID: env.WorkspaceID,
			IsLocal:     true,
			IsActive:    name == currentEnvName,
		})
	}

	// Try to fetch remote environments if we have authentication
	if len(cliConfig.Environments) > 0 {
		// Get the current environment to use for API calls
		currentEnv := cliConfig.Environments[currentEnvName]
		if currentEnv.IsOAuthAuthenticated() {
			remoteEnvs, err := fetchRemoteEnvironments(devMode)
			if err != nil {
				fmt.Printf("⚠️  Could not fetch remote environments: %v\n", err)
			} else {
				// Add remote environments that aren't already local
				for _, remoteEnv := range remoteEnvs {
					// Check if this environment is already in local config
					found := false
					for _, localEnv := range environments {
						if localEnv.ID == remoteEnv.ID {
							found = true
							break
						}
					}
					if !found {
						environments = append(environments, remoteEnv)
					}
				}
			}
		}
	}

	return environments, nil
}

// fetchRemoteEnvironments fetches environments from user's effective resources
func fetchRemoteEnvironments(devMode bool) ([]EnvironmentInfo, error) {
	client, err := GetSDKClientWithDevMode(devMode)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}

	// Get user's active resources (effective permissions)
	userAccess, err := client.Me.GetAccess()
	if err != nil {
		return nil, fmt.Errorf("failed to get user's active resources: %w", err)
	}

	fmt.Printf("🔍 Found %d workspaces for user\n", len(userAccess.Workspaces))

	var environments []EnvironmentInfo

	// Parse environments from workspaces
	for i, workspaceData := range userAccess.Workspaces {
		workspaceID := getStringFromMap(workspaceData, "id")
		workspaceName := getStringFromMap(workspaceData, "name")

		fmt.Printf("   Workspace %d: id=%s, name=%s\n", i+1, workspaceID, workspaceName)

		// Extract environments from this workspace
		envsRaw, exists := workspaceData["environments"]
		if !exists {
			fmt.Printf("      No environments found in workspace\n")
			continue
		}

		envsArray, ok := envsRaw.([]interface{})
		if !ok {
			fmt.Printf("      ⚠️  Environments field is not an array\n")
			continue
		}

		for j, envRaw := range envsArray {
			envData, ok := envRaw.(map[string]interface{})
			if !ok {
				fmt.Printf("      ⚠️  Environment %d: invalid format\n", j+1)
				continue
			}

			envID := getStringFromMap(envData, "id")
			envName := getStringFromMap(envData, "name")
			envType := getStringFromMap(envData, "type")

			// Verify this is an environment resource
			if envType != "environment" {
				fmt.Printf("      ⚠️  Environment %d: invalid type '%s'\n", j+1, envType)
				continue
			}

			fmt.Printf("      ✅ Found environment: id=%s, name=%s, workspaceId=%s\n", envID, envName, workspaceID)

			// If no name is provided, use the ID as name
			if envName == "" {
				envName = envID
			}

			if envID != "" {
				environments = append(environments, EnvironmentInfo{
					ID:          envID,
					Name:        envName,
					WorkspaceID: workspaceID,
					IsLocal:     false,
					IsActive:    false,
				})
			}
		}
	}

	return environments, nil
}

// DisplayEnvironments shows environments in a formatted table
func DisplayEnvironments(environments []EnvironmentInfo) {
	if len(environments) == 0 {
		fmt.Println("No environments found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tNAME\tID\tWORKSPACE ID\tSOURCE\tACTIVE")

	for i, env := range environments {
		active := ""
		if env.IsActive {
			active = "*"
		}

		source := "Remote"
		if env.IsLocal {
			source = "Local"
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			i+1,
			env.Name,
			env.ID,
			env.WorkspaceID,
			source,
			active,
		)
	}

	w.Flush()
}

// PromptEnvironmentSelection prompts user to select an environment
func PromptEnvironmentSelection(environments []EnvironmentInfo) (*EnvironmentInfo, error) {
	if len(environments) == 0 {
		return nil, fmt.Errorf("no environments available")
	}

	fmt.Println("\nAvailable environments:")
	DisplayEnvironments(environments)

	fmt.Printf("\nSelect an environment (1-%d): ", len(environments))
	var input string
	fmt.Scanln(&input)

	// Parse selection
	selection, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || selection < 1 || selection > len(environments) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	return &environments[selection-1], nil
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
