package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/blimu-dev/blimu-cli/pkg/auth"
	blimu "github.com/blimu-dev/blimu-go"
)

// Client represents the Blimu API client
type Client struct {
	authClient *auth.Client
}

// NewClient creates a new API client
func NewClient(authClient *auth.Client) *Client {
	return &Client{
		authClient: authClient,
	}
}

// ValidateConfigRequest represents the request payload for config validation
type ValidateConfigRequest struct {
	Resources    map[string]interface{} `json:"resources"`
	Entitlements map[string]interface{} `json:"entitlements"`
	Features     map[string]interface{} `json:"features"`
	Plans        map[string]interface{} `json:"plans"`
	Version      string                 `json:"version"`
}

// ValidateConfigResponse represents the response from config validation
type ValidateConfigResponse struct {
	Valid  bool                   `json:"valid"`
	Errors []ValidationError      `json:"errors"`
	Spec   map[string]interface{} `json:"spec,omitempty"` // OpenAPI spec if valid
}

// ValidationError represents a validation error from the API
type ValidationError struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Message  string `json:"message"`
}

// ValidateConfig sends the configuration to the API for validation and spec generation
func (c *Client) ValidateConfig(configJSON []byte) (*ValidateConfigResponse, error) {
	// Parse the config JSON to build the request
	var configMap map[string]interface{}
	if err := json.Unmarshal(configJSON, &configMap); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Build the request using the SDK types
	request := blimu.DefinitionValidateRequestDto{
		Resources:    blimu.DefinitionValidateRequestDtoResources{},
		Entitlements: blimu.DefinitionValidateRequestDtoEntitlements{},
		Features:     blimu.DefinitionValidateRequestDtoFeatures{},
		Plans:        blimu.DefinitionValidateRequestDtoPlans{},
		Version:      getString(configMap, "version"),
	}

	// Use the SDK to validate config
	sdk := c.authClient.GetSDK()
	response, err := sdk.Definitions.ValidateConfig(request)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	// Convert SDK response to our response format
	result := &ValidateConfigResponse{
		Valid:  response.Valid,
		Errors: make([]ValidationError, len(response.Errors)),
		Spec:   make(map[string]interface{}), // Convert from SDK type to map
	}

	// Convert error format
	for i, sdkError := range response.Errors {
		result.Errors[i] = ValidationError{
			Resource: sdkError.Resource,
			Field:    sdkError.Field,
			Message:  sdkError.Message,
		}
	}

	return result, nil
}

// GenerateSDKRequest represents the request payload for SDK generation
type GenerateSDKRequest struct {
	Resources    map[string]interface{} `json:"resources"`
	Entitlements map[string]interface{} `json:"entitlements"`
	Features     map[string]interface{} `json:"features"`
	Plans        map[string]interface{} `json:"plans"`
	Version      string                 `json:"version"`
	SDKOptions   SDKGenerationOptions   `json:"sdk_options"`
}

// SDKGenerationOptions represents options for SDK generation
type SDKGenerationOptions struct {
	Type        string `json:"type"` // "typescript", "go", etc.
	PackageName string `json:"package_name"`
	ClientName  string `json:"client_name"`
}

// GenerateSDKResponse represents the response from SDK generation
type GenerateSDKResponse struct {
	Success bool                   `json:"success"`
	Spec    map[string]interface{} `json:"spec"` // Generated OpenAPI spec
	Errors  []ValidationError      `json:"errors,omitempty"`
}

// GenerateSDK requests SDK generation from the API
func (c *Client) GenerateSDK(configJSON []byte, options SDKGenerationOptions) (*GenerateSDKResponse, error) {
	// Parse the config JSON and add SDK options
	var request map[string]interface{}
	if err := json.Unmarshal(configJSON, &request); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Add SDK options to the request
	request["sdk_options"] = map[string]interface{}{
		"type":         options.Type,
		"package_name": options.PackageName,

		"client_name": options.ClientName,
	}

	// Make direct HTTP request to avoid SDK type conversion issues
	baseURL := c.authClient.GetBaseURL()
	requestURL := baseURL + "/v1/config/generate-sdk"

	// Marshal the request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", requestURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.authClient.GetAPIKey())

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var apiResponse struct {
		Success bool                   `json:"success"`
		Spec    map[string]interface{} `json:"spec"`
		Errors  []struct {
			Resource string `json:"resource"`
			Field    string `json:"field"`
			Message  string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to our response format
	result := &GenerateSDKResponse{
		Success: apiResponse.Success,
		Spec:    apiResponse.Spec,
		Errors:  make([]ValidationError, len(apiResponse.Errors)),
	}

	for i, apiError := range apiResponse.Errors {
		result.Errors[i] = ValidationError{
			Resource: apiError.Resource,
			Field:    apiError.Field,
			Message:  apiError.Message,
		}
	}

	return result, nil
}

// FetchOpenAPISpec fetches the base OpenAPI specification from the Blimu API
func (c *Client) FetchOpenAPISpec() (map[string]interface{}, error) {
	// Get the base URL from the auth client
	baseURL := c.authClient.GetBaseURL()

	// Construct the OpenAPI spec URL
	specURL := baseURL + "/docs/json"

	// Make HTTP request to fetch the spec
	resp, err := http.Get(specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec from %s: %w", specURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec: HTTP %d", resp.StatusCode)
	}

	// Read and parse the JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec response: %w", err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec JSON: %w", err)
	}

	return spec, nil
}

// ListEnvironments fetches environments from the API
func (c *Client) ListEnvironments() (*blimu.EnvironmentListDtoOutput, error) {
	sdk := c.authClient.GetSDK()

	// Call the API to list environments
	response, err := sdk.Environments.List(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	return &response, nil
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}
