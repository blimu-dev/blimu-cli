package resources

import (
	"fmt"
	"strings"

	blimu "github.com/blimu-dev/blimu-cli/internal/sdk"
	"github.com/blimu-dev/blimu-cli/pkg/shared"
	"github.com/spf13/cobra"
)

// CreateCommand represents the create resource command
type CreateCommand struct {
	ResourceType  string
	ResourceID    string
	Parent        string
	WorkspaceID   string
	EnvironmentID string
}

// NewCreateCmd creates the create command
func NewCreateCmd() *cobra.Command {
	cmd := &CreateCommand{}

	cobraCmd := &cobra.Command{
		Use:   "create <resource-type> <resource-id>",
		Short: "Create a resource",
		Long: `Create a resource in your Blimu environment.

Example:
  blimu resources create organization org123
  blimu resources create workspace ws456 --parent organization:org123`,
		Args: cobra.ExactArgs(2),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			cmd.ResourceType = args[0]
			cmd.ResourceID = args[1]
			return cmd.Run()
		},
	}

	cobraCmd.Flags().StringVar(&cmd.Parent, "parent", "", "Parent resource in format 'type:id'")
	cobraCmd.Flags().StringVar(&cmd.WorkspaceID, "workspace-id", "", "Workspace ID (uses current environment's workspace if available)")
	cobraCmd.Flags().StringVar(&cmd.EnvironmentID, "environment-id", "", "Environment ID (uses current environment ID if available)")

	return cobraCmd
}

// Run executes the create resource command
func (c *CreateCommand) Run() error {
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
		return fmt.Errorf("environment-id is required for resource creation. Either:\n" +
			"  1. Provide --environment-id flag\n" +
			"  2. Configure your current environment with an ID using 'blimu env create --workspace-id <workspace-id> <env-name>'")
	}

	if c.WorkspaceID == "" {
		return fmt.Errorf("workspace-id is required for resource creation. Provide --workspace-id flag.\n" +
			"Use 'blimu workspaces list' to find your workspace ID (when available)")
	}

	fmt.Printf("🔧 Creating resource '%s:%s' in workspace '%s', environment '%s'...\n",
		c.ResourceType, c.ResourceID, c.WorkspaceID, c.EnvironmentID)

	// Get SDK client
	client, err := shared.GetSDKClient()
	if err != nil {
		return err
	}

	// Prepare resource body
	body := blimu.ResourceCreateDto{
		Id:      c.ResourceID,
		Type:    c.ResourceType,
		Name:    c.ResourceID, // Use ID as name by default
		Parents: []map[string]interface{}{},
	}

	// Handle parent resource if specified
	if c.Parent != "" {
		parts := strings.SplitN(c.Parent, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid parent format. Use 'type:id' format")
		}
		parentType, parentID := parts[0], parts[1]

		body.Parents = []map[string]interface{}{
			{
				"id":   parentID,
				"type": parentType,
			},
		}
	}

	// Create the resource
	result, err := client.Resources.Create(c.WorkspaceID, c.EnvironmentID, body)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	fmt.Println("✅ Resource created successfully!")
	fmt.Printf("   Type: %s\n", result.Type)
	fmt.Printf("   ID: %s\n", result.Id)
	fmt.Printf("   Name: %s\n", result.Name)
	if len(body.Parents) > 0 {
		fmt.Printf("   Parent: %s:%s\n", body.Parents[0]["type"], body.Parents[0]["id"])
	}
	fmt.Printf("   Workspace: %s\n", c.WorkspaceID)
	fmt.Printf("   Environment: %s\n", c.EnvironmentID)

	return nil
}
