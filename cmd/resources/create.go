package resources

import (
	"fmt"
	"strings"

	"github.com/blimu-dev/blimu-cli/pkg/shared"
	blimu "github.com/blimu-dev/blimu-platform-go"
	"github.com/spf13/cobra"
)

// CreateCommand represents the create resource command
type CreateCommand struct {
	ResourceType string
	ResourceID   string
	Parent       string
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

	return cobraCmd
}

// Run executes the create resource command
func (c *CreateCommand) Run() error {
	// Get current environment info
	cliConfig, currentEnv, err := shared.GetCurrentEnvironmentInfo()
	if err != nil {
		return err
	}

	envName := cliConfig.CurrentEnvironment
	if currentEnv != nil && currentEnv.Name != "" {
		envName = currentEnv.Name
	}

	fmt.Printf("🔧 Creating resource '%s:%s' in environment '%s'...\n", c.ResourceType, c.ResourceID, envName)

	// Get SDK client
	client, err := shared.GetSDKClient()
	if err != nil {
		return err
	}

	// Prepare resource body
	body := blimu.ResourceUpdateBody{
		ExtraFields: blimu.ResourceUpdateBodyExtraFields{},
		Parents:     []blimu.ResourceUpdateBodyParentsItem{},
	}

	// Handle parent resource if specified
	if c.Parent != "" {
		parts := strings.SplitN(c.Parent, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid parent format. Use 'type:id' format")
		}
		parentType, parentID := parts[0], parts[1]

		body.Parents = []blimu.ResourceUpdateBodyParentsItem{
			{
				Id:   parentID,
				Type: parentType,
			},
		}
	}

	// Create the resource
	result, err := client.Resources.Create(c.ResourceType, body)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	fmt.Println("✅ Resource created successfully!")
	fmt.Printf("   Type: %s\n", result.Type)
	fmt.Printf("   ID: %s\n", result.Id)

	return nil
}
