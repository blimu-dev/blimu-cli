package blimu_platform

import (
	"context"
	"fmt"
	"net/url"
)

// ApiKeysService handles API Keys related operations
type ApiKeysService struct {
	client *Client
}

// ListWithContext GET /v1/workspace/{workspaceId}/api-keys
func (s *ApiKeysService) ListWithContext(ctx context.Context, workspaceId string) (ApiKeyListDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/api-keys", workspaceId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero ApiKeyListDtoOutput
		return zero, err
	}
	var result ApiKeyListDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ApiKeyListDtoOutput
		return zero, err
	}

	return result, nil
}

// List GET /v1/workspace/{workspaceId}/api-keys
//
// This is a convenience method that calls ListWithContext with context.Background().
func (s *ApiKeysService) List(workspaceId string) (ApiKeyListDtoOutput, error) {
	return s.ListWithContext(context.Background(), workspaceId)
}

// CreateWithContext POST /v1/workspace/{workspaceId}/api-keys
func (s *ApiKeysService) CreateWithContext(ctx context.Context, workspaceId string, body ApiKeyCreateDto) (ApiKeyDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/api-keys", workspaceId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "POST", path, queryValues, body, nil)
	if err != nil {
		var zero ApiKeyDtoOutput
		return zero, err
	}
	var result ApiKeyDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ApiKeyDtoOutput
		return zero, err
	}

	return result, nil
}

// Create POST /v1/workspace/{workspaceId}/api-keys
//
// This is a convenience method that calls CreateWithContext with context.Background().
func (s *ApiKeysService) Create(workspaceId string, body ApiKeyCreateDto) (ApiKeyDtoOutput, error) {
	return s.CreateWithContext(context.Background(), workspaceId, body)
}

// DeleteWithContext DELETE /v1/workspace/{workspaceId}/api-keys/{id}
func (s *ApiKeysService) DeleteWithContext(ctx context.Context, workspaceId string, id string) (interface{}, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/api-keys/%v", workspaceId, id)
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

// Delete DELETE /v1/workspace/{workspaceId}/api-keys/{id}
//
// This is a convenience method that calls DeleteWithContext with context.Background().
func (s *ApiKeysService) Delete(workspaceId string, id string) (interface{}, error) {
	return s.DeleteWithContext(context.Background(), workspaceId, id)
}

// GetWithContext GET /v1/workspace/{workspaceId}/api-keys/{id}
func (s *ApiKeysService) GetWithContext(ctx context.Context, workspaceId string, id string) (ApiKeyDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/api-keys/%v", workspaceId, id)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero ApiKeyDtoOutput
		return zero, err
	}
	var result ApiKeyDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero ApiKeyDtoOutput
		return zero, err
	}

	return result, nil
}

// Get GET /v1/workspace/{workspaceId}/api-keys/{id}
//
// This is a convenience method that calls GetWithContext with context.Background().
func (s *ApiKeysService) Get(workspaceId string, id string) (ApiKeyDtoOutput, error) {
	return s.GetWithContext(context.Background(), workspaceId, id)
}
