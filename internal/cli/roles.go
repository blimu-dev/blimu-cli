package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	blimu "github.com/blimu-dev/blimu-go"
	"github.com/spf13/cobra"
)

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "User role management commands",
	Long:  `Commands for managing user roles in your Blimu environment`,
}

var rolesCreateCmd = &cobra.Command{
	Use:   "create <user-id> <role> <resource-type> <resource-id>",
	Short: "Assign a role to a user on a resource",
	Long: `Assign a role to a user on a specific resource.

Example:
  blimu roles create user123 admin organization org456
  blimu roles create user789 editor workspace ws123`,
	Args: cobra.ExactArgs(4),
	RunE: runRolesCreate,
}

var rolesBulkCmd = &cobra.Command{
	Use:   "bulk <csv-file>",
	Short: "Bulk assign roles from CSV file",
	Long: `Bulk assign user roles from a CSV file.

The CSV file should have the following columns:
- user_id: User ID
- role: Role name
- resource_type: Resource type
- resource_id: Resource ID

Example CSV:
user_id,role,resource_type,resource_id
user123,admin,organization,org456
user789,editor,workspace,ws123
user456,viewer,project,proj789

The command processes role assignments in batches to avoid payload size limits.
Use --batch-size to control the number of assignments processed per batch (maximum 1000).

For better error handling:
- Use --continue-on-error to process all batches even if some fail
- Use --skip-existing to avoid conflicts with existing role assignments (when API supports it)`,
	Args: cobra.ExactArgs(1),
	RunE: runRolesBulk,
}

var (
	rolesBatchSize       int
	rolesContinueOnError bool
	rolesSkipExisting    bool
)

func init() {
	rolesBulkCmd.Flags().IntVar(&rolesBatchSize, "batch-size", 1000, "Number of role assignments to process in each batch (max 1000)")
	rolesBulkCmd.Flags().BoolVar(&rolesContinueOnError, "continue-on-error", false, "Continue processing remaining batches even if some batches fail")
	rolesBulkCmd.Flags().BoolVar(&rolesSkipExisting, "skip-existing", false, "Skip role assignments that already exist (requires API support)")

	rolesCmd.AddCommand(rolesCreateCmd)
	rolesCmd.AddCommand(rolesBulkCmd)
}

func runRolesCreate(cmd *cobra.Command, args []string) error {
	userID := args[0]
	role := args[1]
	resourceType := args[2]
	resourceID := args[3]

	// Get shared context
	ctx := GetContext()

	// Validate environment is set
	if err := ctx.ValidateEnvironment(); err != nil {
		return err
	}

	// Get current environment info
	_, envName, err := ctx.GetCurrentEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get current environment: %w", err)
	}

	fmt.Printf("👤 Assigning role '%s' to user '%s' on resource '%s:%s' in environment '%s'...\n",
		role, userID, resourceType, resourceID, envName)

	// Get API client from context
	client, err := ctx.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get API client: %w", err)
	}

	// Prepare role assignment body
	body := blimu.UserRoleCreateBodyDto{
		Role:         role,
		ResourceType: resourceType,
		ResourceId:   resourceID,
	}

	// Assign the role
	result, err := client.UserRoles.Create(userID, body)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	fmt.Println("✅ Role assigned successfully!")
	fmt.Printf("   User: %s\n", result.UserId)
	fmt.Printf("   Role: %s\n", result.Role)
	fmt.Printf("   Resource: %s:%s\n", result.ResourceType, result.ResourceId)
	fmt.Printf("   Environment: %s\n", result.EnvironmentId)

	return nil
}

func runRolesBulk(cmd *cobra.Command, args []string) error {
	csvFile := args[0]

	// Get shared context
	ctx := GetContext()

	// Validate environment is set
	if err := ctx.ValidateEnvironment(); err != nil {
		return err
	}

	// Get current environment info
	_, envName, err := ctx.GetCurrentEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get current environment: %w", err)
	}

	fmt.Printf("📥 Loading user roles from %s...\n", csvFile)

	// Parse CSV file
	userRoles, err := parseUserRolesCSV(csvFile)
	if err != nil {
		return fmt.Errorf("failed to parse CSV file: %w", err)
	}

	fmt.Printf("📊 Found %d user roles to assign in environment '%s'\n", len(userRoles), envName)

	// Warn about unsupported flags
	if rolesSkipExisting {
		fmt.Printf("⚠️  --skip-existing flag is not yet supported by the API. Flag will be ignored.\n")
	}

	// Get API client from context
	client, err := ctx.GetClient()
	if err != nil {
		return fmt.Errorf("failed to get API client: %w", err)
	}

	// Process in batches to avoid payload limits
	if rolesBatchSize <= 0 {
		rolesBatchSize = 1000 // Default batch size
	} else if rolesBatchSize > 1000 {
		fmt.Printf("⚠️  Batch size %d exceeds maximum of 1000. Using 1000 instead.\n", rolesBatchSize)
		rolesBatchSize = 1000
	}
	var totalSuccessful, totalFailed, totalProcessed int
	var allCreated []blimu.BulkUserRoleResultDtoOutputCreatedItem
	var allErrors []blimu.BulkUserRoleResultDtoOutputErrorsItem

	for i := 0; i < len(userRoles); i += rolesBatchSize {
		end := i + rolesBatchSize
		if end > len(userRoles) {
			end = len(userRoles)
		}

		batch := userRoles[i:end]
		batchNum := (i / rolesBatchSize) + 1
		totalBatches := (len(userRoles) + rolesBatchSize - 1) / rolesBatchSize

		fmt.Printf("🔄 Processing batch %d/%d (%d role assignments)...\n", batchNum, totalBatches, len(batch))

		// Prepare bulk create request for this batch
		bulkBody := blimu.BulkUserRoleCreateBodyDto{
			UserRoles: batch,
		}

		// Execute bulk create for this batch
		result, err := client.UserRoles.BulkCreate(bulkBody)
		if err != nil {
			if rolesContinueOnError {
				fmt.Printf("   ❌ Batch %d failed: %v\n", batchNum, err)
				fmt.Printf("   ⏭️  Continuing with remaining batches...\n")
				continue
			}
			return fmt.Errorf("failed to bulk assign roles in batch %d: %w", batchNum, err)
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
	if totalFailed > 0 && !rolesContinueOnError {
		fmt.Printf("\n💡 Tip: Use --continue-on-error to process all batches even if some fail\n")
		fmt.Printf("💡 Tip: Use --skip-existing to avoid duplicate errors on retry (when API supports it)\n")
	}

	if len(allCreated) > 0 {
		fmt.Printf("\n✅ Successfully assigned %d roles:\n", len(allCreated))
		// Show first 10 and summarize if more
		displayLimit := 10
		for i, created := range allCreated {
			if i < displayLimit {
				fmt.Printf("   • User %s -> %s on %s:%s\n",
					created.UserId, created.Role, created.ResourceType, created.ResourceId)
			} else {
				fmt.Printf("   ... and %d more role assignments\n", len(allCreated)-displayLimit)
				break
			}
		}
	}

	if len(allErrors) > 0 {
		fmt.Printf("\n❌ Failed to assign %d roles:\n", len(allErrors))
		// Show first 10 errors and summarize if more
		displayLimit := 10
		for i, errorItem := range allErrors {
			if i < displayLimit {
				fmt.Printf("   • User %s -> %s on %s:%s - %s\n",
					errorItem.UserRole.UserId, errorItem.UserRole.Role,
					errorItem.UserRole.ResourceType, errorItem.UserRole.ResourceId,
					errorItem.Error)
			} else {
				fmt.Printf("   ... and %d more errors\n", len(allErrors)-displayLimit)
				break
			}
		}
	}

	return nil
}

func parseUserRolesCSV(filename string) ([]blimu.BulkUserRoleCreateBodyDtoUserRolesItem, error) {
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
	requiredColumns := []string{"user_id", "role", "resource_type", "resource_id"}
	for _, col := range requiredColumns {
		if _, exists := columnMap[col]; !exists {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	var userRoles []blimu.BulkUserRoleCreateBodyDtoUserRolesItem

	// Parse data rows
	for i, record := range records[1:] {
		if len(record) < len(requiredColumns) {
			return nil, fmt.Errorf("row %d: insufficient columns", i+2)
		}

		userRole := blimu.BulkUserRoleCreateBodyDtoUserRolesItem{
			UserId:       record[columnMap["user_id"]],
			Role:         record[columnMap["role"]],
			ResourceType: record[columnMap["resource_type"]],
			ResourceId:   record[columnMap["resource_id"]],
		}

		userRoles = append(userRoles, userRole)
	}

	return userRoles, nil
}
