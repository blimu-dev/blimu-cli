package blimu_platform

import (
	"fmt"
	"net/url"
)

// Generated models from OpenAPI specification

// ApiKeyCreateDto
type ApiKeyCreateDto struct {
	EnvironmentId string `json:"environmentId"`
	Name          string `json:"name"`
}

// ApiKeyDtoOutput
type ApiKeyDtoOutput struct {
	CreatedAt   string `json:"createdAt"`
	Id          string `json:"id"`
	IsActive    bool   `json:"isActive"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	UpdatedAt   string `json:"updatedAt"`
	WorkspaceId string `json:"workspaceId"`
}

// ApiKeyListDtoOutput
type ApiKeyListDtoOutput struct {
	Data  []map[string]interface{} `json:"data"`
	Total float64                  `json:"total"`
}

// DefinitionDtoOutput
type DefinitionDtoOutput struct {
	Entitlements map[string]interface{} `json:"entitlements"`
	Features     map[string]interface{} `json:"features"`
	Plans        map[string]interface{} `json:"plans"`
	Resources    map[string]interface{} `json:"resources"`
}

// DefinitionGenerateSdkRequestDto
type DefinitionGenerateSdkRequestDto struct {
	Entitlements map[string]interface{} `json:"entitlements"`
	Features     map[string]interface{} `json:"features"`
	Plans        map[string]interface{} `json:"plans"`
	Resources    map[string]interface{} `json:"resources"`
	SdkOptions   map[string]interface{} `json:"sdk_options"`
	Version      string                 `json:"version"`
}

// DefinitionGenerateSdkResponseDtoOutput
type DefinitionGenerateSdkResponseDtoOutput struct {
	Errors  []map[string]interface{} `json:"errors"`
	Spec    map[string]interface{}   `json:"spec"`
	Success bool                     `json:"success"`
}

// DefinitionUpdateDto
type DefinitionUpdateDto struct {
	Entitlements map[string]interface{} `json:"entitlements"`
	Features     map[string]interface{} `json:"features"`
	Plans        map[string]interface{} `json:"plans"`
	Resources    map[string]interface{} `json:"resources"`
}

// DefinitionValidateRequestDto
type DefinitionValidateRequestDto struct {
	Entitlements map[string]interface{} `json:"entitlements"`
	Features     map[string]interface{} `json:"features"`
	Plans        map[string]interface{} `json:"plans"`
	Resources    map[string]interface{} `json:"resources"`
	Version      string                 `json:"version"`
}

// DefinitionValidateResponseDtoOutput
type DefinitionValidateResponseDtoOutput struct {
	Errors []map[string]interface{} `json:"errors"`
	Spec   map[string]interface{}   `json:"spec"`
	Valid  bool                     `json:"valid"`
}

// EnvironmentCreateDto
type EnvironmentCreateDto struct {
	LookupKey string `json:"lookupKey"`
	Name      string `json:"name"`
}

// EnvironmentDtoOutput
type EnvironmentDtoOutput struct {
	CreatedAt   string  `json:"createdAt"`
	Id          string  `json:"id"`
	LookupKey   *string `json:"lookupKey"`
	Name        string  `json:"name"`
	UpdatedAt   string  `json:"updatedAt"`
	WorkspaceId string  `json:"workspaceId"`
}

// EnvironmentListDtoOutput
type EnvironmentListDtoOutput struct {
	Data []map[string]interface{} `json:"data"`
	Meta map[string]interface{}   `json:"meta"`
}

// EnvironmentUpdateDto
type EnvironmentUpdateDto struct {
	LookupKey string `json:"lookupKey"`
	Name      string `json:"name"`
}

// EnvironmentWithDefinitionDtoOutput
type EnvironmentWithDefinitionDtoOutput struct {
	CreatedAt   string                  `json:"createdAt"`
	Definition  *map[string]interface{} `json:"definition"`
	Id          string                  `json:"id"`
	LookupKey   *string                 `json:"lookupKey"`
	Name        string                  `json:"name"`
	UpdatedAt   string                  `json:"updatedAt"`
	WorkspaceId string                  `json:"workspaceId"`
}

// ResourceCreateDto
type ResourceCreateDto struct {
	Id      string                   `json:"id"`
	Name    string                   `json:"name"`
	Parents []map[string]interface{} `json:"parents"`
	Type    string                   `json:"type"`
}

// ResourceDtoOutput
type ResourceDtoOutput struct {
	CreatedAt string                   `json:"createdAt"`
	Id        string                   `json:"id"`
	Name      *string                  `json:"name"`
	Parents   []map[string]interface{} `json:"parents"`
	Type      string                   `json:"type"`
}

// ResourceListResponseDtoOutput
type ResourceListResponseDtoOutput struct {
	Items []map[string]interface{} `json:"items"`
	Limit float64                  `json:"limit"`
	Page  float64                  `json:"page"`
	Total float64                  `json:"total"`
}

// ResourceUpdateDto
type ResourceUpdateDto struct {
	Name    string                   `json:"name"`
	Parents []map[string]interface{} `json:"parents"`
}

// ResourceUserListResponseDtoOutput
type ResourceUserListResponseDtoOutput struct {
	Items []map[string]interface{} `json:"items"`
	Limit float64                  `json:"limit"`
	Page  float64                  `json:"page"`
	Total float64                  `json:"total"`
}

// UserAccessDtoOutput
type UserAccessDtoOutput struct {
	Roles      map[string]interface{}   `json:"roles"`
	Workspaces []map[string]interface{} `json:"workspaces"`
}

// UserDtoOutput
type UserDtoOutput struct {
	AvatarUrl     *string `json:"avatarUrl"`
	CreatedAt     string  `json:"createdAt"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"emailVerified"`
	FirstName     *string `json:"firstName"`
	Id            string  `json:"id"`
	LastLoginAt   *string `json:"lastLoginAt"`
	LastName      *string `json:"lastName"`
	UpdatedAt     string  `json:"updatedAt"`
}

// UserListResponseDtoOutput
type UserListResponseDtoOutput struct {
	Items []map[string]interface{} `json:"items"`
	Limit float64                  `json:"limit"`
	Page  float64                  `json:"page"`
	Total float64                  `json:"total"`
}

// UserResourceDtoOutput
type UserResourceDtoOutput struct {
	Inherited    bool     `json:"inherited"`
	Name         string   `json:"name"`
	ParentIds    []string `json:"parentIds"`
	ResourceId   string   `json:"resourceId"`
	ResourceType string   `json:"resourceType"`
	Role         string   `json:"role"`
}

// Query parameter structs for operations

// EnvironmentsListQuery represents query parameters for Environments.List
type EnvironmentsListQuery struct {
	EnvironmentId *string `json:"environmentId"`
	Limit         *int64  `json:"limit"`
	Page          *int64  `json:"page"`
	Search        *string `json:"search"`
}

// ToValues converts the query struct to url.Values
func (q *EnvironmentsListQuery) ToValues() url.Values {
	if q == nil {
		return nil
	}

	values := make(url.Values)
	// Handle optional environmentId parameter
	if q.EnvironmentId != nil {
		values.Set("environmentId", fmt.Sprintf("%v", *q.EnvironmentId))
	}
	// Handle optional limit parameter
	if q.Limit != nil {
		values.Set("limit", fmt.Sprintf("%v", *q.Limit))
	}
	// Handle optional page parameter
	if q.Page != nil {
		values.Set("page", fmt.Sprintf("%v", *q.Page))
	}
	// Handle optional search parameter
	if q.Search != nil {
		values.Set("search", fmt.Sprintf("%v", *q.Search))
	}

	return values
}

// ResourcesListQuery represents query parameters for Resources.List
type ResourcesListQuery struct {
	Limit  *float64 `json:"limit"`
	Page   *float64 `json:"page"`
	Parent *string  `json:"parent"`
	Search *string  `json:"search"`
	Type   string   `json:"type"`
}

// ToValues converts the query struct to url.Values
func (q *ResourcesListQuery) ToValues() url.Values {
	if q == nil {
		return nil
	}

	values := make(url.Values)
	// Handle optional limit parameter
	if q.Limit != nil {
		values.Set("limit", fmt.Sprintf("%v", *q.Limit))
	}
	// Handle optional page parameter
	if q.Page != nil {
		values.Set("page", fmt.Sprintf("%v", *q.Page))
	}
	// Handle optional parent parameter
	if q.Parent != nil {
		values.Set("parent", fmt.Sprintf("%v", *q.Parent))
	}
	// Handle optional search parameter
	if q.Search != nil {
		values.Set("search", fmt.Sprintf("%v", *q.Search))
	}
	// Handle type parameter
	if fmt.Sprintf("%v", q.Type) != "" {
		values.Set("type", fmt.Sprintf("%v", q.Type))
	}

	return values
}

// ResourcesListChildrenQuery represents query parameters for Resources.ListChildren
type ResourcesListChildrenQuery struct {
	Limit  *float64 `json:"limit"`
	Page   *float64 `json:"page"`
	Parent *string  `json:"parent"`
	Search *string  `json:"search"`
	Type   string   `json:"type"`
}

// ToValues converts the query struct to url.Values
func (q *ResourcesListChildrenQuery) ToValues() url.Values {
	if q == nil {
		return nil
	}

	values := make(url.Values)
	// Handle optional limit parameter
	if q.Limit != nil {
		values.Set("limit", fmt.Sprintf("%v", *q.Limit))
	}
	// Handle optional page parameter
	if q.Page != nil {
		values.Set("page", fmt.Sprintf("%v", *q.Page))
	}
	// Handle optional parent parameter
	if q.Parent != nil {
		values.Set("parent", fmt.Sprintf("%v", *q.Parent))
	}
	// Handle optional search parameter
	if q.Search != nil {
		values.Set("search", fmt.Sprintf("%v", *q.Search))
	}
	// Handle type parameter
	if fmt.Sprintf("%v", q.Type) != "" {
		values.Set("type", fmt.Sprintf("%v", q.Type))
	}

	return values
}

// ResourcesGetResourceUsersQuery represents query parameters for Resources.GetResourceUsers
type ResourcesGetResourceUsersQuery struct {
	Limit *float64 `json:"limit"`
	Page  *float64 `json:"page"`
}

// ToValues converts the query struct to url.Values
func (q *ResourcesGetResourceUsersQuery) ToValues() url.Values {
	if q == nil {
		return nil
	}

	values := make(url.Values)
	// Handle optional limit parameter
	if q.Limit != nil {
		values.Set("limit", fmt.Sprintf("%v", *q.Limit))
	}
	// Handle optional page parameter
	if q.Page != nil {
		values.Set("page", fmt.Sprintf("%v", *q.Page))
	}

	return values
}

// UsersListQuery represents query parameters for Users.List
type UsersListQuery struct {
	Limit  *float64 `json:"limit"`
	Page   *float64 `json:"page"`
	Search *string  `json:"search"`
}

// ToValues converts the query struct to url.Values
func (q *UsersListQuery) ToValues() url.Values {
	if q == nil {
		return nil
	}

	values := make(url.Values)
	// Handle optional limit parameter
	if q.Limit != nil {
		values.Set("limit", fmt.Sprintf("%v", *q.Limit))
	}
	// Handle optional page parameter
	if q.Page != nil {
		values.Set("page", fmt.Sprintf("%v", *q.Page))
	}
	// Handle optional search parameter
	if q.Search != nil {
		values.Set("search", fmt.Sprintf("%v", *q.Search))
	}

	return values
}
