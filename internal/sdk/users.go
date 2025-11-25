package blimu_platform

import (
	"context"
	"fmt"
	"net/url"
)

// UsersService handles Users related operations
type UsersService struct {
	client *Client
}

// ListWithContext GET /v1/workspaces/{workspaceId}/environments/{environmentId}/users
// List users for an environment
func (s *UsersService) ListWithContext(ctx context.Context, workspaceId string, environmentId string, query *UsersListQuery) (UserListResponseDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/users", workspaceId, environmentId)
	// Convert query parameters
	var queryValues url.Values
	if query != nil {
		queryValues = query.ToValues()
	}
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero UserListResponseDtoOutput
		return zero, err
	}
	var result UserListResponseDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero UserListResponseDtoOutput
		return zero, err
	}

	return result, nil
}

// List GET /v1/workspaces/{workspaceId}/environments/{environmentId}/users
// List users for an environment
//
// This is a convenience method that calls ListWithContext with context.Background().
func (s *UsersService) List(workspaceId string, environmentId string, query *UsersListQuery) (UserListResponseDtoOutput, error) {
	return s.ListWithContext(context.Background(), workspaceId, environmentId, query)
}

// GetWithContext GET /v1/workspaces/{workspaceId}/environments/{environmentId}/users/{userId}
// Get user details by ID
func (s *UsersService) GetWithContext(ctx context.Context, workspaceId string, environmentId string, userId string) (UserDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/users/%v", workspaceId, environmentId, userId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero UserDtoOutput
		return zero, err
	}
	var result UserDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero UserDtoOutput
		return zero, err
	}

	return result, nil
}

// Get GET /v1/workspaces/{workspaceId}/environments/{environmentId}/users/{userId}
// Get user details by ID
//
// This is a convenience method that calls GetWithContext with context.Background().
func (s *UsersService) Get(workspaceId string, environmentId string, userId string) (UserDtoOutput, error) {
	return s.GetWithContext(context.Background(), workspaceId, environmentId, userId)
}

// GetUserResourcesWithContext GET /v1/workspaces/{workspaceId}/environments/{environmentId}/users/{userId}/resources
// Get user resource relationships
func (s *UsersService) GetUserResourcesWithContext(ctx context.Context, workspaceId string, environmentId string, userId string) ([]UserResourceDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspaces/%v/environments/%v/users/%v/resources", workspaceId, environmentId, userId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero []UserResourceDtoOutput
		return zero, err
	}
	var result []UserResourceDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero []UserResourceDtoOutput
		return zero, err
	}

	return result, nil
}

// GetUserResources GET /v1/workspaces/{workspaceId}/environments/{environmentId}/users/{userId}/resources
// Get user resource relationships
//
// This is a convenience method that calls GetUserResourcesWithContext with context.Background().
func (s *UsersService) GetUserResources(workspaceId string, environmentId string, userId string) ([]UserResourceDtoOutput, error) {
	return s.GetUserResourcesWithContext(context.Background(), workspaceId, environmentId, userId)
}
