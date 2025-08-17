package auth

import (
	"fmt"
	"os"

	blimu "github.com/blimu-dev/blimu-go"
)

// Client represents an authenticated Blimu client wrapping the SDK
type Client struct {
	sdk     *blimu.Client
	baseURL string
	apiKey  string
}

// NewClient creates a new authenticated client using the blimu-go SDK
func NewClient(baseURL, apiKey string) *Client {
	sdk := blimu.NewClient(
		blimu.WithBaseURL(baseURL),
		blimu.WithApiKeyAuth(apiKey),
	)

	return &Client{
		sdk:     sdk,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// NewClientFromEnv creates a new client from environment variables
func NewClientFromEnv() (*Client, error) {
	apiKey := os.Getenv("BLIMU_SECRET_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("BLIMU_SECRET_KEY environment variable is required")
	}

	baseURL := os.Getenv("BLIMU_API_URL")
	if baseURL == "" {
		baseURL = "https://api.blimu.dev"
	}

	return NewClient(baseURL, apiKey), nil
}

// GetSDK returns the underlying blimu-go SDK client
func (c *Client) GetSDK() *blimu.Client {
	return c.sdk
}

// GetBaseURL returns the base URL used by this client
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetAPIKey returns the API key used by this client
func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// ValidateAuth validates the authentication by making a test request
// This could use any lightweight endpoint to verify the API key works
func (c *Client) ValidateAuth() error {
	// Try to get current definitions as a way to validate auth
	// This is a read-only operation that should work with valid credentials
	_, err := c.sdk.Definitions.Get()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}
