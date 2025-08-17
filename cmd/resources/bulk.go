package resources

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/blimu-dev/blimu-cli/pkg/shared"
	blimu "github.com/blimu-dev/blimu-go"
	"github.com/spf13/cobra"
)

// BulkCommand represents the bulk create resources command
type BulkCommand struct {
	CSVFile         string
	BatchSize       int
	ContinueOnError bool
	SkipExisting    bool
}

// NewBulkCmd creates the bulk command
func NewBulkCmd() *cobra.Command {
	cmd := &BulkCommand{}

	cobraCmd := &cobra.Command{
		Use:   "bulk <csv-file>",
		Short: "Bulk create resources from CSV file",
		Long: `Bulk create resources from a CSV file.

The CSV file should have the following columns:
- type: Resource type
- id: Resource ID
- parent_type: Parent resource type (optional)
- parent_id: Parent resource ID (optional)

Example CSV:
type,id,parent_type,parent_id
organization,org123,,
workspace,ws456,organization,org123
project,proj789,workspace,ws456

The command processes resources in batches to avoid payload size limits.
Use --batch-size to control the number of resources processed per batch (maximum 1000).

For better error handling:
- Use --continue-on-error to process all batches even if some fail
- Use --skip-existing to avoid conflicts with existing resources (when API supports it)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			cmd.CSVFile = args[0]
			return cmd.Run()
		},
	}

	cobraCmd.Flags().IntVar(&cmd.BatchSize, "batch-size", 1000, "Number of resources to process in each batch (max 1000)")
	cobraCmd.Flags().BoolVar(&cmd.ContinueOnError, "continue-on-error", false, "Continue processing remaining batches even if some batches fail")
	cobraCmd.Flags().BoolVar(&cmd.SkipExisting, "skip-existing", false, "Skip resources that already exist (requires API support)")

	return cobraCmd
}

// Run executes the bulk create command
func (c *BulkCommand) Run() error {
	// Get current environment info
	cliConfig, currentEnv, err := shared.GetCurrentEnvironmentInfo()
	if err != nil {
		return err
	}

	envName := cliConfig.CurrentEnvironment
	if currentEnv != nil && currentEnv.Name != "" {
		envName = currentEnv.Name
	}

	fmt.Printf("📥 Loading resources from %s...\n", c.CSVFile)

	// Parse CSV file
	resources, err := c.parseResourcesCSV()
	if err != nil {
		return fmt.Errorf("failed to parse CSV file: %w", err)
	}

	fmt.Printf("📊 Found %d resources to create in environment '%s'\n", len(resources), envName)

	// Warn about unsupported flags
	if c.SkipExisting {
		fmt.Printf("⚠️  --skip-existing flag is not yet supported by the API. Flag will be ignored.\n")
	}

	// Get SDK client
	client, err := shared.GetSDKClient()
	if err != nil {
		return err
	}

	// Process in batches to avoid payload limits
	if c.BatchSize <= 0 {
		c.BatchSize = 1000 // Default batch size
	} else if c.BatchSize > 1000 {
		fmt.Printf("⚠️  Batch size %d exceeds maximum of 1000. Using 1000 instead.\n", c.BatchSize)
		c.BatchSize = 1000
	}

	return c.processBatches(client, resources)
}

// processBatches processes resources in batches
func (c *BulkCommand) processBatches(client *blimu.Client, resources []Resource) error {
	var totalSuccessful, totalFailed, totalProcessed int
	var allCreated []blimu.BulkResourceResultDtoOutputCreatedItem
	var allErrors []blimu.BulkResourceResultDtoOutputErrorsItem

	for i := 0; i < len(resources); i += c.BatchSize {
		end := i + c.BatchSize
		if end > len(resources) {
			end = len(resources)
		}

		batch := resources[i:end]
		batchNum := (i / c.BatchSize) + 1
		totalBatches := (len(resources) + c.BatchSize - 1) / c.BatchSize

		fmt.Printf("\n📦 Processing batch %d/%d (%d resources)...\n", batchNum, totalBatches, len(batch))

		// Convert batch to API format
		batchBody := c.convertBatchToAPIFormat(batch)

		// Create batch
		result, err := client.Resources.BulkCreate(blimu.BulkResourceCreateBodyDto{
			Resources: batchBody,
		})
		if err != nil {
			if c.ContinueOnError {
				fmt.Printf("❌ Batch %d failed: %v\n", batchNum, err)
				totalFailed += len(batch)
				continue
			}
			return fmt.Errorf("batch %d failed: %w", batchNum, err)
		}

		// Process results
		created := len(result.Created)
		errors := len(result.Errors)

		totalSuccessful += created
		totalFailed += errors
		totalProcessed += len(batch)

		allCreated = append(allCreated, result.Created...)
		allErrors = append(allErrors, result.Errors...)

		fmt.Printf("✅ Batch %d completed: %d created, %d errors\n", batchNum, created, errors)

		// Show errors for this batch
		if errors > 0 {
			fmt.Printf("   Errors in batch %d:\n", batchNum)
			for _, err := range result.Errors {
				fmt.Printf("   - %s:%s: %s\n", err.Resource.Type, err.Resource.Id, err.Error)
			}
		}
	}

	// Summary
	fmt.Printf("\n📊 Bulk creation completed!\n")
	fmt.Printf("   Total processed: %d\n", totalProcessed)
	fmt.Printf("   Successfully created: %d\n", totalSuccessful)
	fmt.Printf("   Failed: %d\n", totalFailed)

	if totalFailed > 0 {
		fmt.Printf("\n❌ Failed resources:\n")
		for _, err := range allErrors {
			fmt.Printf("   - %s:%s: %s\n", err.Resource.Type, err.Resource.Id, err.Error)
		}
	}

	return nil
}

// convertBatchToAPIFormat converts a batch of resources to API format
func (c *BulkCommand) convertBatchToAPIFormat(resources []Resource) []blimu.BulkResourceCreateBodyDtoResourcesItem {
	batchBody := make([]blimu.BulkResourceCreateBodyDtoResourcesItem, len(resources))
	for j, resource := range resources {
		batchBody[j] = blimu.BulkResourceCreateBodyDtoResourcesItem{
			Id:          resource.ID,
			Type:        resource.Type,
			ExtraFields: blimu.BulkResourceCreateBodyDtoResourcesItemExtraFields{},
			Parents:     []blimu.BulkResourceCreateBodyDtoResourcesItemParentsItem{},
		}

		// Add parent if specified
		if resource.ParentType != "" && resource.ParentID != "" {
			batchBody[j].Parents = append(batchBody[j].Parents, blimu.BulkResourceCreateBodyDtoResourcesItemParentsItem{
				Id:   resource.ParentID,
				Type: resource.ParentType,
			})
		}
	}
	return batchBody
}

// parseResourcesCSV parses the CSV file containing resources
func (c *BulkCommand) parseResourcesCSV() ([]Resource, error) {
	file, err := os.Open(c.CSVFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Validate header
	header := records[0]
	expectedHeaders := []string{"type", "id", "parent_type", "parent_id"}
	if len(header) < 2 {
		return nil, fmt.Errorf("CSV must have at least 'type' and 'id' columns")
	}

	headerSet := make(map[string]bool)
	for _, h := range header {
		headerSet[h] = true
	}

	for _, expectedHeader := range expectedHeaders {
		if !headerSet[expectedHeader] {
			return nil, fmt.Errorf("CSV must have '%s' column", expectedHeader)
		}
	}

	// Parse records
	var resources []Resource
	for i, record := range records[1:] {
		if len(record) < 2 {
			return nil, fmt.Errorf("row %d: must have at least type and id", i+2)
		}

		resource := Resource{
			Type: record[0],
			ID:   record[1],
		}

		// Optional parent fields
		if len(record) > 2 {
			resource.ParentType = record[2]
		}
		if len(record) > 3 {
			resource.ParentID = record[3]
		}

		// Validate that if parent_type is provided, parent_id is also provided
		if resource.ParentType != "" && resource.ParentID == "" {
			return nil, fmt.Errorf("row %d: parent_type provided but parent_id is missing", i+2)
		}
		if resource.ParentID != "" && resource.ParentType == "" {
			return nil, fmt.Errorf("row %d: parent_id provided but parent_type is missing", i+2)
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// Resource represents a resource from CSV
type Resource struct {
	Type       string
	ID         string
	ParentType string
	ParentID   string
}
