package internalgit

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

type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    *url.URL
}

func NewClient(config Config) (*Client, error) {
	if !config.IsEnabled() {
		return nil, fmt.Errorf("internalgit: not enabled or missing required config")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("internalgit: invalid BaseURL: %w", err)
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: baseURL,
	}, nil
}

func (c *Client) Config() Config {
	return c.config
}

type Error struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("internalgit: %s (status %d)", e.Message, e.StatusCode)
	}
	return fmt.Sprintf("internalgit: request failed with status %d: %s", e.StatusCode, e.Body)
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, body any) (*http.Response, error) {
	u := *c.baseURL
	u.Path = path
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("internalgit: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("internalgit: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.config.AdminToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("internalgit: request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, result any) error {
	resp, err := c.request(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("internalgit: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &Error{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}

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

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("internalgit: failed to decode response: %w", err)
		}
	}

	return nil
}

// requestWithUserToken makes a request using a user's token instead of admin token
func (c *Client) requestWithUserToken(ctx context.Context, method, path string, query url.Values, body any, token string) (*http.Response, error) {
	u := *c.baseURL
	u.Path = path
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("internalgit: failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("internalgit: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) doWithUserToken(ctx context.Context, method, path string, query url.Values, body, result any, token string) error {
	resp, err := c.requestWithUserToken(ctx, method, path, query, body, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("internalgit: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &Error{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}

		var errResp struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Message != "" {
			apiErr.Message = errResp.Message
		}

		return apiErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("internalgit: failed to decode response: %w", err)
		}
	}

	return nil
}

// GetHTTPSCloneURL returns the HTTPS clone URL with embedded token
func (c *Client) GetHTTPSCloneURL(username, repoName, token string) string {
	u := *c.baseURL
	u.User = url.UserPassword(username, token)
	u.Path = fmt.Sprintf("/%s/%s.git", username, repoName)
	return u.String()
}

// GetSSHCloneURL returns the SSH clone URL
func (c *Client) GetSSHCloneURL(username, repoName string) string {
	return fmt.Sprintf("%s:%s/%s.git", c.config.SSHURL, username, repoName)
}

// GetRepoPath returns the repo path in format "ml.ink/{username}/{repo}"
func (c *Client) GetRepoPath(username, repoName string) string {
	host := c.baseURL.Host
	// Remove port if present
	if idx := len(host) - 1; idx > 0 {
		for i := len(host) - 1; i >= 0; i-- {
			if host[i] == ':' {
				host = host[:i]
				break
			}
		}
	}
	// Remove "git." prefix if present for cleaner paths
	if len(host) > 4 && host[:4] == "git." {
		host = host[4:]
	}
	return fmt.Sprintf("%s/%s/%s", host, username, repoName)
}

// GeneratePassword generates a random password for new users
func GeneratePassword() string {
	// Generate a random 32-char password
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
