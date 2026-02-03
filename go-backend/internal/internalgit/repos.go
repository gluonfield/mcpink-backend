package internalgit

import (
	"context"
	"fmt"
	"time"

	"code.gitea.io/sdk/gitea"
)

// EnsureUser creates a Gitea user if it doesn't exist
func (c *Client) EnsureUser(ctx context.Context, username, email string) error {
	// Check if user exists
	_, _, err := c.api.GetUserInfo(username)
	if err == nil {
		return nil // user exists
	}

	// Create user with random password (we won't use it, admin token handles everything)
	mustChange := false
	visibility := gitea.VisibleTypePrivate
	_, _, err = c.api.AdminCreateUser(gitea.CreateUserOption{
		Username:           username,
		Email:              email,
		Password:           generateRandomPassword(),
		MustChangePassword: &mustChange,
		Visibility:         &visibility,
	})
	if err != nil {
		return fmt.Errorf("failed to create gitea user: %w", err)
	}
	return nil
}

// CreateRepoForUser creates a repository under a specific user (requires admin)
func (c *Client) CreateRepoForUser(ctx context.Context, username, repoName, description string, private bool) (*gitea.Repository, error) {
	defaultBranch := "main"
	repo, _, err := c.api.AdminCreateRepo(username, gitea.CreateRepoOption{
		Name:          repoName,
		Description:   description,
		Private:       private,
		AutoInit:      false, // Don't auto-init so user can push directly
		DefaultBranch: defaultBranch,
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// GetRepo retrieves a repository
func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*gitea.Repository, error) {
	r, _, err := c.api.GetRepo(owner, repo)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// DeleteRepo deletes a repository
func (c *Client) DeleteRepo(ctx context.Context, owner, repo string) error {
	_, err := c.api.DeleteRepo(owner, repo)
	return err
}

// AddDeployKey adds a deploy key to a repository for SSH access
func (c *Client) AddDeployKey(ctx context.Context, owner, repo, title, publicKey string) error {
	_, _, err := c.api.CreateDeployKey(owner, repo, gitea.CreateKeyOption{
		Title:    title,
		Key:      publicKey,
		ReadOnly: true,
	})
	return err
}

// CreateRepoWebhook creates a webhook for a repository
func (c *Client) CreateRepoWebhook(ctx context.Context, owner, repo, webhookURL, secret string) (*gitea.Hook, error) {
	hook, _, err := c.api.CreateRepoHook(owner, repo, gitea.CreateHookOption{
		Type: gitea.HookTypeGitea,
		Config: map[string]string{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       secret,
		},
		Events: []string{"push"},
		Active: true,
	})
	if err != nil {
		return nil, err
	}
	return hook, nil
}

// AddCollaborator adds a user as collaborator to a repository with read-only access
func (c *Client) AddCollaborator(ctx context.Context, owner, repo, collaborator string) error {
	perm := gitea.AccessModeRead
	_, err := c.api.AddCollaborator(owner, repo, collaborator, gitea.AddCollaboratorOption{
		Permission: &perm,
	})
	return err
}

// generateRandomPassword generates a random password (not stored, just for user creation)
func generateRandomPassword() string {
	return fmt.Sprintf("Rnd%dPass!", time.Now().UnixNano()%1000000000)
}
