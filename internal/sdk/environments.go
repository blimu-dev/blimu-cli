package blimu_platform

import (
	"context"
	"fmt"
	"net/url"
)

// EnvironmentsService handles Environments related operations
type EnvironmentsService struct {
	client *Client
}

// ListWithContext GET /v1/workspace/{workspaceId}/environments
// List environments
func (s *EnvironmentsService) ListWithContext(ctx context.Context, workspaceId string, query *EnvironmentsListQuery) (EnvironmentListDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments", workspaceId)
	// Convert query parameters
	var queryValues url.Values
	if query != nil {
		queryValues = query.ToValues()
	}
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero EnvironmentListDtoOutput
		return zero, err
	}
	var result EnvironmentListDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero EnvironmentListDtoOutput
		return zero, err
	}

	return result, nil
}

// List GET /v1/workspace/{workspaceId}/environments
// List environments
//
// This is a convenience method that calls ListWithContext with context.Background().
func (s *EnvironmentsService) List(workspaceId string, query *EnvironmentsListQuery) (EnvironmentListDtoOutput, error) {
	return s.ListWithContext(context.Background(), workspaceId, query)
}

// CreateWithContext POST /v1/workspace/{workspaceId}/environments
// Create a new environment
func (s *EnvironmentsService) CreateWithContext(ctx context.Context, workspaceId string, body EnvironmentCreateDto) (EnvironmentDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments", workspaceId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "POST", path, queryValues, body, nil)
	if err != nil {
		var zero EnvironmentDtoOutput
		return zero, err
	}
	var result EnvironmentDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero EnvironmentDtoOutput
		return zero, err
	}

	return result, nil
}

// Create POST /v1/workspace/{workspaceId}/environments
// Create a new environment
//
// This is a convenience method that calls CreateWithContext with context.Background().
func (s *EnvironmentsService) Create(workspaceId string, body EnvironmentCreateDto) (EnvironmentDtoOutput, error) {
	return s.CreateWithContext(context.Background(), workspaceId, body)
}

// DeleteWithContext DELETE /v1/workspace/{workspaceId}/environments/{environmentId}
// Delete an environment
func (s *EnvironmentsService) DeleteWithContext(ctx context.Context, workspaceId string, environmentId string) (interface{}, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v", workspaceId, environmentId)
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

// Delete DELETE /v1/workspace/{workspaceId}/environments/{environmentId}
// Delete an environment
//
// This is a convenience method that calls DeleteWithContext with context.Background().
func (s *EnvironmentsService) Delete(workspaceId string, environmentId string) (interface{}, error) {
	return s.DeleteWithContext(context.Background(), workspaceId, environmentId)
}

// ReadWithContext GET /v1/workspace/{workspaceId}/environments/{environmentId}
// Read an environment by ID
func (s *EnvironmentsService) ReadWithContext(ctx context.Context, workspaceId string, environmentId string) (EnvironmentWithDefinitionDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v", workspaceId, environmentId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero EnvironmentWithDefinitionDtoOutput
		return zero, err
	}
	var result EnvironmentWithDefinitionDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero EnvironmentWithDefinitionDtoOutput
		return zero, err
	}

	return result, nil
}

// Read GET /v1/workspace/{workspaceId}/environments/{environmentId}
// Read an environment by ID
//
// This is a convenience method that calls ReadWithContext with context.Background().
func (s *EnvironmentsService) Read(workspaceId string, environmentId string) (EnvironmentWithDefinitionDtoOutput, error) {
	return s.ReadWithContext(context.Background(), workspaceId, environmentId)
}

// UpdateWithContext PUT /v1/workspace/{workspaceId}/environments/{environmentId}
// Update an environment
func (s *EnvironmentsService) UpdateWithContext(ctx context.Context, workspaceId string, environmentId string, body EnvironmentUpdateDto) (EnvironmentDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v", workspaceId, environmentId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "PUT", path, queryValues, body, nil)
	if err != nil {
		var zero EnvironmentDtoOutput
		return zero, err
	}
	var result EnvironmentDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero EnvironmentDtoOutput
		return zero, err
	}

	return result, nil
}

// Update PUT /v1/workspace/{workspaceId}/environments/{environmentId}
// Update an environment
//
// This is a convenience method that calls UpdateWithContext with context.Background().
func (s *EnvironmentsService) Update(workspaceId string, environmentId string, body EnvironmentUpdateDto) (EnvironmentDtoOutput, error) {
	return s.UpdateWithContext(context.Background(), workspaceId, environmentId, body)
}
