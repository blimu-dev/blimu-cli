package auth

import (
	"fmt"

	platform "github.com/blimu-dev/blimu-cli/internal/sdk"
)

// Client represents a Blimu client that uses Clerk OAuth and platform SDK for operations
type Client struct {
	appSDK  *platform.Client // For CLI operations
	baseURL string
	token   string // JWT token from Clerk OAuth
}

// NewClientWithClerkOAuth creates a new client for Clerk OAuth authentication
func NewClientWithClerkOAuth(clerkDomain string) *Client {
	return &Client{
		baseURL: clerkDomain,
	}
}

// NewClientWithClerkToken creates a client with Clerk JWT token for platform operations
func NewClientWithClerkToken(platformBaseURL, clerkToken string) *Client {
	appSDK := platform.NewClient(
		platform.WithBaseURL(platformBaseURL),
		platform.WithBearer(clerkToken),
	)

	return &Client{
		appSDK:  appSDK,
		baseURL: platformBaseURL,
		token:   clerkToken,
	}
}

// GetClerkToken returns the Clerk JWT token
func (c *Client) GetClerkToken() string {
	return c.token
}

// GetAppSDK returns the platform SDK for CLI operations
func (c *Client) GetAppSDK() *platform.Client {
	return c.appSDK
}

// GetBaseURL returns the base URL used by this client
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetToken returns the JWT token used by this client
func (c *Client) GetToken() string {
	return c.token
}

// ValidateAuth validates the authentication by making a test request to platform API
func (c *Client) ValidateAuth() error {
	if c.appSDK == nil {
		return fmt.Errorf("no platform SDK configured - need JWT token")
	}

	// Try to get current user's active resources as a way to validate auth
	_, err := c.appSDK.Me.GetAccess()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}

// Legacy methods for backward compatibility

// GetSDK returns the platform SDK (for backward compatibility)
func (c *Client) GetSDK() *platform.Client {
	return c.appSDK
}
