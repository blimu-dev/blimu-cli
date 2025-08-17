package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/blimu-dev/blimu-cli/pkg/config"
	blimu "github.com/blimu-dev/blimu-go"
	"github.com/spf13/cobra"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Resource management commands",
	Long:  `Commands for managing resources in your Blimu environment`,
}

var resourcesCreateCmd = &cobra.Command{
	Use:   "create <resource-type> <resource-id>",
	Short: "Create a resource",
	Long: `Create a resource in your Blimu environment.

Example:
  blimu resources create organization org123
  blimu resources create workspace ws456 --parent organization:org123`,
	Args: cobra.ExactArgs(2),
	RunE: runResourcesCreate,
}

var resourcesBulkCmd = &cobra.Command{
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
	RunE: runResourcesBulk,
}

var (
	resourceParent  string
	batchSize       int
	continueOnError bool
	skipExisting    bool
)

func init() {
	resourcesCreateCmd.Flags().StringVar(&resourceParent, "parent", "", "Parent resource in format 'type:id'")
	resourcesBulkCmd.Flags().IntVar(&batchSize, "batch-size", 1000, "Number of resources to process in each batch (max 1000)")
	resourcesBulkCmd.Flags().BoolVar(&continueOnError, "continue-on-error", false, "Continue processing remaining batches even if some batches fail")
	resourcesBulkCmd.Flags().BoolVar(&skipExisting, "skip-existing", false, "Skip resources that already exist (requires API support)")

	resourcesCmd.AddCommand(resourcesCreateCmd)
	resourcesCmd.AddCommand(resourcesBulkCmd)
}

func runResourcesCreate(cmd *cobra.Command, args []string) error {
	resourceType := args[0]
	resourceID := args[1]

	// Load CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	apiURL, apiKey, err := cliConfig.GetAPIClient()
	if err != nil {
		return fmt.Errorf("failed to get API client config: %w", err)
	}

	currentEnv, _ := cliConfig.GetCurrentEnvironment()
	envName := cliConfig.CurrentEnvironment
	if currentEnv != nil && currentEnv.Name != "" {
		envName = currentEnv.Name
	}

	fmt.Printf("🔧 Creating resource '%s:%s' in environment '%s'...\n", resourceType, resourceID, envName)

	// Create Blimu client
	client := blimu.NewClient(
		blimu.WithBaseURL(apiURL),
		blimu.WithApiKeyAuth(apiKey),
	)

	// Prepare resource body
	body := blimu.ResourceUpdateBody{
		ExtraFields: blimu.ResourceUpdateBodyExtraFields{},
		Parents:     []blimu.ResourceUpdateBodyParentsItem{},
	}

	// Handle parent resource if specified
	if resourceParent != "" {
		parts := strings.SplitN(resourceParent, ":", 2)
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
	result, err := client.Resources.Create(resourceType, body)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	fmt.Println("✅ Resource created successfully!")
	fmt.Printf("   Type: %s\n", result.Type)
	fmt.Printf("   ID: %s\n", result.Id)

	return nil
}

func runResourcesBulk(cmd *cobra.Command, args []string) error {
	csvFile := args[0]

	// Load CLI config
	cliConfig, err := config.LoadCLIConfig()
	if err != nil {
		return fmt.Errorf("failed to load CLI config: %w", err)
	}

	apiURL, apiKey, err := cliConfig.GetAPIClient()
	if err != nil {
		return fmt.Errorf("failed to get API client config: %w", err)
	}

	currentEnv, _ := cliConfig.GetCurrentEnvironment()
	envName := cliConfig.CurrentEnvironment
	if currentEnv != nil && currentEnv.Name != "" {
		envName = currentEnv.Name
	}

	fmt.Printf("📥 Loading resources from %s...\n", csvFile)

	// Parse CSV file
	resources, err := parseResourcesCSV(csvFile)
	if err != nil {
		return fmt.Errorf("failed to parse CSV file: %w", err)
	}

	fmt.Printf("📊 Found %d resources to create in environment '%s'\n", len(resources), envName)

	// Warn about unsupported flags
	if skipExisting {
		fmt.Printf("⚠️  --skip-existing flag is not yet supported by the API. Flag will be ignored.\n")
	}

	// Create Blimu client
	client := blimu.NewClient(
		blimu.WithBaseURL(apiURL),
		blimu.WithApiKeyAuth(apiKey),
	)

	// Process in batches to avoid payload limits
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	} else if batchSize > 1000 {
		fmt.Printf("⚠️  Batch size %d exceeds maximum of 1000. Using 1000 instead.\n", batchSize)
		batchSize = 1000
	}
	var totalSuccessful, totalFailed, totalProcessed int
	var allCreated []blimu.BulkResourceResultDtoOutputCreatedItem
	var allErrors []blimu.BulkResourceResultDtoOutputErrorsItem

	for i := 0; i < len(resources); i += batchSize {
		end := i + batchSize
		if end > len(resources) {
			end = len(resources)
		}

		batch := resources[i:end]
		batchNum := (i / batchSize) + 1
		totalBatches := (len(resources) + batchSize - 1) / batchSize

		fmt.Printf("🔄 Processing batch %d/%d (%d resources)...\n", batchNum, totalBatches, len(batch))

		// Prepare bulk create request for this batch
		bulkBody := blimu.BulkResourceCreateBodyDto{
			Resources: batch,
		}

		// Execute bulk create for this batch
		result, err := client.Resources.BulkCreate(bulkBody)
		if err != nil {
			if continueOnError {
				fmt.Printf("   ❌ Batch %d failed: %v\n", batchNum, err)
				fmt.Printf("   ⏭️  Continuing with remaining batches...\n")
				continue
			}
			return fmt.Errorf("failed to bulk create resources in batch %d: %w", batchNum, err)
		}

		// Accumulate results
		totalSuccessful += int(result.Summary.Successful)
		totalFailed += int(result.Summary.Failed)
		totalProcessed += int(result.Summary.Total)

		allCreated = append(allCreated, result.Created...)
		allErrors = append(allErrors, result.Errors...)

		fmt.Printf("   Batch %d: %d successful, %d failed\n", batchNum, int(result.Summary.Successful), int(result.Summary.Failed))
	}

	// Display overall results
	fmt.Printf("\n📊 Bulk operation completed:\n")
	fmt.Printf("   Successful: %d\n", totalSuccessful)
	fmt.Printf("   Failed: %d\n", totalFailed)
	fmt.Printf("   Total: %d\n", totalProcessed)

	// Show retry suggestion if there were failures
	if totalFailed > 0 && !continueOnError {
		fmt.Printf("\n💡 Tip: Use --continue-on-error to process all batches even if some fail\n")
		fmt.Printf("💡 Tip: Use --skip-existing to avoid duplicate errors on retry (when API supports it)\n")
	}

	if len(allCreated) > 0 {
		fmt.Printf("\n✅ Successfully created %d resources:\n", len(allCreated))
		// Show first 10 and summarize if more
		displayLimit := 10
		for i, created := range allCreated {
			if i < displayLimit {
				fmt.Printf("   • %s:%s\n", created.Type, created.Id)
			} else {
				fmt.Printf("   ... and %d more resources\n", len(allCreated)-displayLimit)
				break
			}
		}
	}

	if len(allErrors) > 0 {
		fmt.Printf("\n❌ Failed to create %d resources:\n", len(allErrors))
		// Show first 10 errors and summarize if more
		displayLimit := 10
		for i, errorItem := range allErrors {
			if i < displayLimit {
				fmt.Printf("   • %s:%s - %s\n", errorItem.Resource.Type, errorItem.Resource.Id, errorItem.Error)
			} else {
				fmt.Printf("   ... and %d more errors\n", len(allErrors)-displayLimit)
				break
			}
		}
	}

	return nil
}

func parseResourcesCSV(filename string) ([]blimu.BulkResourceCreateBodyDtoResourcesItem, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Parse header
	header := records[0]
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[strings.ToLower(col)] = i
	}

	// Validate required columns
	requiredColumns := []string{"type", "id"}
	for _, col := range requiredColumns {
		if _, exists := columnMap[col]; !exists {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	var resources []blimu.BulkResourceCreateBodyDtoResourcesItem

	// Parse data rows
	for i, record := range records[1:] {
		if len(record) < len(requiredColumns) {
			return nil, fmt.Errorf("row %d: insufficient columns", i+2)
		}

		resource := blimu.BulkResourceCreateBodyDtoResourcesItem{
			Type:        record[columnMap["type"]],
			Id:          record[columnMap["id"]],
			ExtraFields: blimu.BulkResourceCreateBodyDtoResourcesItemExtraFields{},
			Parents:     []blimu.BulkResourceCreateBodyDtoResourcesItemParentsItem{},
		}

		// Handle parent if specified
		if parentTypeIdx, exists := columnMap["parent_type"]; exists && parentTypeIdx < len(record) && record[parentTypeIdx] != "" {
			if parentIDIdx, exists := columnMap["parent_id"]; exists && parentIDIdx < len(record) && record[parentIDIdx] != "" {
				resource.Parents = append(resource.Parents, blimu.BulkResourceCreateBodyDtoResourcesItemParentsItem{
					Type: record[parentTypeIdx],
					Id:   record[parentIDIdx],
				})
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
