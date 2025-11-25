package blimu_platform

import (
	"context"
	"net/url"
)

// MeService handles Me related operations
type MeService struct {
	client *Client
}

// GetAccessWithContext GET /v1/me/access
// Get active resources for current user
func (s *MeService) GetAccessWithContext(ctx context.Context) (UserAccessDtoOutput, error) {
	path := "/v1/me/access"
	var queryValues url.Values
	// Make request
	resp, err := s.client.request(ctx, "GET", path, queryValues, nil, nil)
	if err != nil {
		var zero UserAccessDtoOutput
		return zero, err
	}
	var result UserAccessDtoOutput

	if err := s.client.decodeResponse(resp, &result); err != nil {
		var zero UserAccessDtoOutput
		return zero, err
	}

	return result, nil
}

// GetAccess GET /v1/me/access
// Get active resources for current user
//
// This is a convenience method that calls GetAccessWithContext with context.Background().
func (s *MeService) GetAccess() (UserAccessDtoOutput, error) {
	return s.GetAccessWithContext(context.Background())
}
