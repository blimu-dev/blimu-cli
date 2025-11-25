package blimu_platform

import (
	"context"
	"fmt"
	"net/url"
)

// DefinitionsService handles Definitions related operations
type DefinitionsService struct {
	client *Client
}

// GetWithContext GET /v1/workspace/{workspaceId}/environments/{environmentId}/definitions
// Get environment definitions
func (s *DefinitionsService) GetWithContext(ctx context.Context, workspaceId string, environmentId string) (DefinitionDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v/definitions", workspaceId, environmentId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero DefinitionDtoOutput
		return zero, err
	}
	var result DefinitionDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero DefinitionDtoOutput
		return zero, err
	}

	return result, nil
}

// Get GET /v1/workspace/{workspaceId}/environments/{environmentId}/definitions
// Get environment definitions
//
// This is a convenience method that calls GetWithContext with context.Background().
func (s *DefinitionsService) Get(workspaceId string, environmentId string) (DefinitionDtoOutput, error) {
	return s.GetWithContext(context.Background(), workspaceId, environmentId)
}

// UpdateWithContext PUT /v1/workspace/{workspaceId}/environments/{environmentId}/definitions
// Update environment definitions
func (s *DefinitionsService) UpdateWithContext(ctx context.Context, workspaceId string, environmentId string, body DefinitionUpdateDto) (interface{}, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v/definitions", workspaceId, environmentId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "PUT", path, queryValues, body, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}

	if err := s.client.decodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Update PUT /v1/workspace/{workspaceId}/environments/{environmentId}/definitions
// Update environment definitions
//
// This is a convenience method that calls UpdateWithContext with context.Background().
func (s *DefinitionsService) Update(workspaceId string, environmentId string, body DefinitionUpdateDto) (interface{}, error) {
	return s.UpdateWithContext(context.Background(), workspaceId, environmentId, body)
}

// GetOpenApiWithContext GET /v1/workspace/{workspaceId}/environments/{environmentId}/definitions/openapi
// Generate OpenAPI spec from database definitions
//
// Generates a custom OpenAPI specification from the environment's stored definitions in the database. The generated spec can be used to create type-safe SDKs.
func (s *DefinitionsService) GetOpenApiWithContext(ctx context.Context, workspaceId string, environmentId string) (DefinitionGenerateSdkResponseDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v/definitions/openapi", workspaceId, environmentId)
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero DefinitionGenerateSdkResponseDtoOutput
		return zero, err
	}
	var result DefinitionGenerateSdkResponseDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero DefinitionGenerateSdkResponseDtoOutput
		return zero, err
	}

	return result, nil
}

// GetOpenApi GET /v1/workspace/{workspaceId}/environments/{environmentId}/definitions/openapi
// Generate OpenAPI spec from database definitions
//
// Generates a custom OpenAPI specification from the environment's stored definitions in the database. The generated spec can be used to create type-safe SDKs.
//
// This is a convenience method that calls GetOpenApiWithContext with context.Background().
func (s *DefinitionsService) GetOpenApi(workspaceId string, environmentId string) (DefinitionGenerateSdkResponseDtoOutput, error) {
	return s.GetOpenApiWithContext(context.Background(), workspaceId, environmentId)
}

// CreateOpenApiWithContext POST /v1/workspace/{workspaceId}/environments/{environmentId}/definitions/openapi
// Generate custom openapi spec based on the environment definitions
//
// Validates configuration and generates a custom OpenAPI specification tailored to the user's resource definitions. The generated spec can be used to create type-safe SDKs.
func (s *DefinitionsService) CreateOpenApiWithContext(ctx context.Context, workspaceId string, environmentId string, body DefinitionGenerateSdkRequestDto) (DefinitionGenerateSdkResponseDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v/definitions/openapi", workspaceId, environmentId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "POST", path, queryValues, body, nil)
	if err != nil {
		var zero DefinitionGenerateSdkResponseDtoOutput
		return zero, err
	}
	var result DefinitionGenerateSdkResponseDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero DefinitionGenerateSdkResponseDtoOutput
		return zero, err
	}

	return result, nil
}

// CreateOpenApi POST /v1/workspace/{workspaceId}/environments/{environmentId}/definitions/openapi
// Generate custom openapi spec based on the environment definitions
//
// Validates configuration and generates a custom OpenAPI specification tailored to the user's resource definitions. The generated spec can be used to create type-safe SDKs.
//
// This is a convenience method that calls CreateOpenApiWithContext with context.Background().
func (s *DefinitionsService) CreateOpenApi(workspaceId string, environmentId string, body DefinitionGenerateSdkRequestDto) (DefinitionGenerateSdkResponseDtoOutput, error) {
	return s.CreateOpenApiWithContext(context.Background(), workspaceId, environmentId, body)
}

// ValidateWithContext POST /v1/workspace/{workspaceId}/environments/{environmentId}/definitions/validate
// Validate Blimu configuration
//
// Validates a complete Blimu configuration including resources, entitlements, features, and plans. Returns validation errors and optionally generates an OpenAPI spec if valid.
func (s *DefinitionsService) ValidateWithContext(ctx context.Context, workspaceId string, environmentId string, body DefinitionValidateRequestDto) (DefinitionValidateResponseDtoOutput, error) {
	// Build path with parameters
	path := fmt.Sprintf("/v1/workspace/%v/environments/%v/definitions/validate", workspaceId, environmentId)
	var queryValues url.Values
	// Make request with body
	resp, err := s.client.request(ctx, "POST", path, queryValues, body, nil)
	if err != nil {
		var zero DefinitionValidateResponseDtoOutput
		return zero, err
	}
	var result DefinitionValidateResponseDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero DefinitionValidateResponseDtoOutput
		return zero, err
	}

	return result, nil
}

// Validate POST /v1/workspace/{workspaceId}/environments/{environmentId}/definitions/validate
// Validate Blimu configuration
//
// Validates a complete Blimu configuration including resources, entitlements, features, and plans. Returns validation errors and optionally generates an OpenAPI spec if valid.
//
// This is a convenience method that calls ValidateWithContext with context.Background().
func (s *DefinitionsService) Validate(workspaceId string, environmentId string, body DefinitionValidateRequestDto) (DefinitionValidateResponseDtoOutput, error) {
	return s.ValidateWithContext(context.Background(), workspaceId, environmentId, body)
}
