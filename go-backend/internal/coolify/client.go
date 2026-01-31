// Package coolify provides a Go SDK for the Coolify API.
// See https://coolify.io/docs/api-reference for the full API documentation.
package coolify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Config holds the configuration for the Coolify client.
type Config struct {
	// BaseURL is the base URL of the Coolify instance (e.g., "https://coolify.example.com")
	BaseURL string
	// Token is the API token for authentication (Bearer token)
	Token string
	// Timeout is the HTTP client timeout (default: 30s)
	Timeout time.Duration
}

// Client is the Coolify API client.
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    *url.URL

	// Services
	Applications *ApplicationsService
}

// NewClient creates a new Coolify API client.
func NewClient(config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("coolify: BaseURL is required")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("coolify: Token is required")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("coolify: invalid BaseURL: %w", err)
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	c := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}

	// Initialize services
	c.Applications = &ApplicationsService{client: c}

	return c, nil
}

// Error represents an API error response.
type Error struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("coolify: %s (status %d)", e.Message, e.StatusCode)
	}
	return fmt.Sprintf("coolify: request failed with status %d: %s", e.StatusCode, e.Body)
}

// request performs an HTTP request to the Coolify API.
func (c *Client) request(ctx context.Context, method, path string, query url.Values, body any) (*http.Response, error) {
	// Build URL
	u := *c.baseURL
	u.Path = path
	if query != nil {
		u.RawQuery = query.Encode()
	}

	// Prepare body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("coolify: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("coolify: failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.config.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coolify: request failed: %w", err)
	}

	return resp, nil
}

// do performs an HTTP request and decodes the response into result.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, result any) error {
	resp, err := c.request(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("coolify: failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &Error{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}

		// Try to extract error message from JSON
		var errResp struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			if errResp.Message != "" {
				apiErr.Message = errResp.Message
			} else if errResp.Error != "" {
				apiErr.Message = errResp.Error
			}
		}

		return apiErr
	}

	// Decode response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("coolify: failed to decode response: %w", err)
		}
	}

	return nil
}
