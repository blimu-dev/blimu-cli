package blimu_platform

import (
	"context"
	"fmt"
	"net/url"
)

// ResourcesService handles Resources related operations
type ResourcesService struct {
	client *Client
}

// ListWithContext GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources
// List resources for an environment
func (s *ResourcesService) ListWithContext(ctx context.Context, workspaceId string, environmentId string, query *ResourcesListQuery) (ResourceListResponseDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/resources", workspaceId, environmentId)
	// Convert query parameters
	var queryValues url.Values
	if query != nil {
		queryValues = query.ToValues()
	}
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero ResourceListResponseDtoOutput
		return zero, err
	}
	var result ResourceListResponseDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ResourceListResponseDtoOutput
		return zero, err
	}

	return result, nil
}

// List GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources
// List resources for an environment
//
// This is a convenience method that calls ListWithContext with context.Background().
func (s *ResourcesService) List(workspaceId string, environmentId string, query *ResourcesListQuery) (ResourceListResponseDtoOutput, error) {
	return s.ListWithContext(context.Background(), workspaceId, environmentId, query)
}

// CreateWithContext POST /v1/workspaces/{workspaceId}/environments/{environmentId}/resources
// Create a new resource
func (s *ResourcesService) CreateWithContext(ctx context.Context, workspaceId string, environmentId string, body ResourceCreateDto) (ResourceDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/resources", workspaceId, environmentId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "POST", path, queryValues, body, nil)
	if err != nil {
		var zero ResourceDtoOutput
		return zero, err
	}
	var result ResourceDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ResourceDtoOutput
		return zero, err
	}

	return result, nil
}

// Create POST /v1/workspaces/{workspaceId}/environments/{environmentId}/resources
// Create a new resource
//
// This is a convenience method that calls CreateWithContext with context.Background().
func (s *ResourcesService) Create(workspaceId string, environmentId string, body ResourceCreateDto) (ResourceDtoOutput, error) {
	return s.CreateWithContext(context.Background(), workspaceId, environmentId, body)
}

// DeleteWithContext DELETE /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}
// Delete a resource
func (s *ResourcesService) DeleteWithContext(ctx context.Context, workspaceId string, environmentId string, resourceType string, resourceId string) (interface{}, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/resources/%v/%v", workspaceId, environmentId, resourceType, resourceId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "DELETE", path, queryValues, nil, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}

	if err := s.client.decodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Delete DELETE /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}
// Delete a resource
//
// This is a convenience method that calls DeleteWithContext with context.Background().
func (s *ResourcesService) Delete(workspaceId string, environmentId string, resourceType string, resourceId string) (interface{}, error) {
	return s.DeleteWithContext(context.Background(), workspaceId, environmentId, resourceType, resourceId)
}

// GetWithContext GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}
// Get a specific resource
func (s *ResourcesService) GetWithContext(ctx context.Context, workspaceId string, environmentId string, resourceType string, resourceId string) (ResourceDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/resources/%v/%v", workspaceId, environmentId, resourceType, resourceId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero ResourceDtoOutput
		return zero, err
	}
	var result ResourceDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ResourceDtoOutput
		return zero, err
	}

	return result, nil
}

// Get GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}
// Get a specific resource
//
// This is a convenience method that calls GetWithContext with context.Background().
func (s *ResourcesService) Get(workspaceId string, environmentId string, resourceType string, resourceId string) (ResourceDtoOutput, error) {
	return s.GetWithContext(context.Background(), workspaceId, environmentId, resourceType, resourceId)
}

// UpdateWithContext PUT /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}
// Update a resource
func (s *ResourcesService) UpdateWithContext(ctx context.Context, workspaceId string, environmentId string, resourceType string, resourceId string, body ResourceUpdateDto) (ResourceDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/resources/%v/%v", workspaceId, environmentId, resourceType, resourceId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "PUT", path, queryValues, body, nil)
	if err != nil {
		var zero ResourceDtoOutput
		return zero, err
	}
	var result ResourceDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ResourceDtoOutput
		return zero, err
	}

	return result, nil
}

// Update PUT /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}
// Update a resource
//
// This is a convenience method that calls UpdateWithContext with context.Background().
func (s *ResourcesService) Update(workspaceId string, environmentId string, resourceType string, resourceId string, body ResourceUpdateDto) (ResourceDtoOutput, error) {
	return s.UpdateWithContext(context.Background(), workspaceId, environmentId, resourceType, resourceId, body)
}

// ListChildrenWithContext GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}/children
// List children resources for a specific resource
func (s *ResourcesService) ListChildrenWithContext(ctx context.Context, workspaceId string, environmentId string, resourceType string, resourceId string, query *ResourcesListChildrenQuery) (ResourceListResponseDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/resources/%v/%v/children", workspaceId, environmentId, resourceType, resourceId)
	// Convert query parameters
	var queryValues url.Values
	if query != nil {
		queryValues = query.ToValues()
	}
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero ResourceListResponseDtoOutput
		return zero, err
	}
	var result ResourceListResponseDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ResourceListResponseDtoOutput
		return zero, err
	}

	return result, nil
}

// ListChildren GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}/children
// List children resources for a specific resource
//
// This is a convenience method that calls ListChildrenWithContext with context.Background().
func (s *ResourcesService) ListChildren(workspaceId string, environmentId string, resourceType string, resourceId string, query *ResourcesListChildrenQuery) (ResourceListResponseDtoOutput, error) {
	return s.ListChildrenWithContext(context.Background(), workspaceId, environmentId, resourceType, resourceId, query)
}

// GetResourceUsersWithContext GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}/users
// Get users with roles on a resource
func (s *ResourcesService) GetResourceUsersWithContext(ctx context.Context, workspaceId string, environmentId string, resourceType string, resourceId string, query *ResourcesGetResourceUsersQuery) (ResourceUserListResponseDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/resources/%v/%v/users", workspaceId, environmentId, resourceType, resourceId)
	// Convert query parameters
	var queryValues url.Values
	if query != nil {
		queryValues = query.ToValues()
	}
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero ResourceUserListResponseDtoOutput
		return zero, err
	}
	var result ResourceUserListResponseDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ResourceUserListResponseDtoOutput
		return zero, err
	}

	return result, nil
}

// GetResourceUsers GET /v1/workspaces/{workspaceId}/environments/{environmentId}/resources/{resourceType}/{resourceId}/users
// Get users with roles on a resource
//
// This is a convenience method that calls GetResourceUsersWithContext with context.Background().
func (s *ResourcesService) GetResourceUsers(workspaceId string, environmentId string, resourceType string, resourceId string, query *ResourcesGetResourceUsersQuery) (ResourceUserListResponseDtoOutput, error) {
	return s.GetResourceUsersWithContext(context.Background(), workspaceId, environmentId, resourceType, resourceId, query)
}
