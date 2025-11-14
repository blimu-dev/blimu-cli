package auth

import (
	"fmt"

	runtime "github.com/blimu-dev/blimu-go"
	platform "github.com/blimu-dev/blimu-platform-go"
)

// Client represents a hybrid Blimu client that uses runtime SDK for auth and platform SDK for operations
type Client struct {
	runtimeSDK  *runtime.Client  // For OAuth authentication
	platformSDK *platform.Client // For CLI operations
	baseURL     string
	token       string // JWT token from OAuth
}

// NewClientWithOAuth creates a new client for OAuth authentication using runtime SDK
func NewClientWithOAuth(runtimeBaseURL string) *Client {
	runtimeSDK := runtime.NewClient(
		runtime.WithBaseURL(runtimeBaseURL),
	)

	return &Client{
		runtimeSDK: runtimeSDK,
		baseURL:    runtimeBaseURL,
	}
}

// NewClientWithToken creates a new client with JWT token for platform operations
func NewClientWithToken(platformBaseURL, token string) *Client {
	platformSDK := platform.NewClient(
		platform.WithBaseURL(platformBaseURL),
		platform.WithBearer(token),
	)

	return &Client{
		platformSDK: platformSDK,
		baseURL:     platformBaseURL,
		token:       token,
	}
}

// NewHybridClient creates a client that can do both OAuth (runtime) and operations (platform)
func NewHybridClient(runtimeBaseURL, platformBaseURL, token string) *Client {
	runtimeSDK := runtime.NewClient(
		runtime.WithBaseURL(runtimeBaseURL),
	)

	platformSDK := platform.NewClient(
		platform.WithBaseURL(platformBaseURL),
		platform.WithBearer(token),
	)

	return &Client{
		runtimeSDK:  runtimeSDK,
		platformSDK: platformSDK,
		baseURL:     platformBaseURL, // Use platform URL as primary
		token:       token,
	}
}

// GetRuntimeSDK returns the runtime SDK for OAuth operations
func (c *Client) GetRuntimeSDK() *runtime.Client {
	return c.runtimeSDK
}

// GetPlatformSDK returns the platform SDK for CLI operations
func (c *Client) GetPlatformSDK() *platform.Client {
	return c.platformSDK
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
	if c.platformSDK == nil {
		return fmt.Errorf("no platform SDK configured - need JWT token")
	}

	// Try to get current user's active resources as a way to validate auth
	_, err := c.platformSDK.Me.GetActiveResources()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}

// Legacy methods for backward compatibility

// GetSDK returns the platform SDK (for backward compatibility)
func (c *Client) GetSDK() *platform.Client {
	return c.platformSDK
}

// GetAPIKey returns the token (for backward compatibility)
func (c *Client) GetAPIKey() string {
	return c.token
}
