package internalgit

import (
	"context"
	"fmt"
	"net/http"
)

// GetUser retrieves a user by username
func (c *Client) GetUser(ctx context.Context, username string) (*GiteaUser, error) {
	var user GiteaUser
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/users/%s", username), nil, nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new Gitea user (requires admin token)
func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*GiteaUser, error) {
	var user GiteaUser
	err := c.do(ctx, http.MethodPost, "/api/v1/admin/users", nil, req, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetOrCreateUser gets or creates a Gitea user
func (c *Client) GetOrCreateUser(ctx context.Context, username, email string) (*GiteaUser, error) {
	user, err := c.GetUser(ctx, username)
	if err == nil {
		return user, nil
	}

	// Check if it's a 404, create user if so
	if apiErr, ok := err.(*Error); ok && apiErr.StatusCode == 404 {
		return c.CreateUser(ctx, &CreateUserRequest{
			Username:           username,
			Email:              email,
			Password:           GeneratePassword(),
			MustChangePassword: false,
			Visibility:         "private",
		})
	}

	return nil, err
}

// CreateRepo creates a new repository for a user
func (c *Client) CreateRepo(ctx context.Context, username string, req *CreateRepoRequest) (*GiteaRepo, error) {
	var repo GiteaRepo
	err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%s/repos", username), nil, req, &repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetRepo retrieves a repository
func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*GiteaRepo, error) {
	var r GiteaRepo
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo), nil, nil, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteRepo deletes a repository
func (c *Client) DeleteRepo(ctx context.Context, owner, repo string) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo), nil, nil, nil)
}

// CreateRepoWebhook creates a webhook for a repository
func (c *Client) CreateRepoWebhook(ctx context.Context, owner, repo string, req *CreateWebhookRequest) (*GiteaWebhook, error) {
	var webhook GiteaWebhook
	err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/repos/%s/%s/hooks", owner, repo), nil, req, &webhook)
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// CreateAccessToken creates an access token for a user (requires admin privileges)
func (c *Client) CreateAccessToken(ctx context.Context, username string, req *CreateAccessTokenRequest) (*GiteaAccessToken, error) {
	var token GiteaAccessToken
	err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/users/%s/tokens", username), nil, req, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// ListAccessTokens lists all access tokens for a user
func (c *Client) ListAccessTokens(ctx context.Context, username string) ([]GiteaAccessToken, error) {
	var tokens []GiteaAccessToken
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/users/%s/tokens", username), nil, nil, &tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// DeleteAccessToken deletes an access token
func (c *Client) DeleteAccessToken(ctx context.Context, username string, tokenID int64) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/users/%s/tokens/%d", username, tokenID), nil, nil, nil)
}

// DeleteAccessTokenByName deletes all tokens with a given name
func (c *Client) DeleteAccessTokenByName(ctx context.Context, username, tokenName string) error {
	tokens, err := c.ListAccessTokens(ctx, username)
	if err != nil {
		return err
	}

	for _, t := range tokens {
		if t.Name == tokenName {
			if err := c.DeleteAccessToken(ctx, username, t.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
