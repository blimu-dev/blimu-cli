package blimu

// ResourceDefinition represents a resource definition from .blimu config
type ResourceDefinition struct {
	Name             string
	Roles            []string
	RolesInheritance map[string][]string
	Parents          map[string]ParentDefinition
}

// ParentDefinition represents a parent resource relationship
type ParentDefinition struct {
	Name     string
	Required bool
}

// Operation represents a CRUD operation on a resource
type Operation struct {
	Type        OperationType
	Resource    string
	Method      string
	Path        string
	OperationID string
	Summary     string
	Tags        []string
}

// OperationType represents the type of operation
type OperationType string

const (
	OperationCreate OperationType = "create"
	OperationRead   OperationType = "read"
	OperationUpdate OperationType = "update"
	OperationDelete OperationType = "delete"
	OperationList   OperationType = "list"
)

// GenerateOptions represents options for SDK generation
type GenerateOptions struct {
	OutputDir     string
	PackageName   string
	ClientName    string
	SDKType       string
	BlimuAPIURL   string
	ResourcesOnly bool
}
