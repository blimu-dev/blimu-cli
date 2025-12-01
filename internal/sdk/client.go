// Package blimu_platform provides a Go SDK for Blimu Platform
package blimu_platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ClientOption configures the client
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithHeaders sets default headers for all requests
func WithHeaders(headers map[string]string) ClientOption {
	return func(c *Client) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithApiKey sets the API key for authentication
func WithApiKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithBearer sets the bearer token for authentication
func WithBearer(token string) ClientOption {
	return func(c *Client) {
		c.bearer = token
	}
}

// Client is the main client for the Blimu Platform API
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	apiKey     string
	bearer     string

	// Services

	ApiKeys      *ApiKeysService
	Definitions  *DefinitionsService
	Environments *EnvironmentsService
	Me           *MeService
	Resources    *ResourcesService
	Users        *UsersService
}

// NewClient creates a new client with the given options
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL:    "https://app-api.blimu.dev",
		httpClient: http.DefaultClient,
		headers:    make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize services

	c.ApiKeys = &ApiKeysService{client: c}
	c.Definitions = &DefinitionsService{client: c}
	c.Environments = &EnvironmentsService{client: c}
	c.Me = &MeService{client: c}
	c.Resources = &ResourcesService{client: c}
	c.Users = &UsersService{client: c}

	return c
}

// request makes an HTTP request
func (c *Client) request(ctx context.Context, method, path string, query url.Values, body interface{}, headers map[string]string) (*http.Response, error) {
	// Build URL
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set content type for JSON bodies
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set authentication headers
	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}
	if c.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearer)
	}

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// decodeResponse decodes an HTTP response into the given interface
func (c *Client) decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if v == nil {
		return nil
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	// For non-JSON responses, read as string if the target is a string pointer
	if strPtr, ok := v.(*string); ok {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		*strPtr = string(body)
		return nil
	}

	return fmt.Errorf("unsupported content type: %s", contentType)
}

// APIError represents an API error response
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
