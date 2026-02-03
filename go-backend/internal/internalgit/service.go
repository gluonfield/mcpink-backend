package internalgit

import (
	"context"
	"fmt"
	"time"

	"github.com/augustdev/autoclip/internal/storage/pg/generated/internalrepos"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	client      *Client
	db          *pgxpool.Pool
	webhookURL  string
	repoQueries internalrepos.Querier
	userQueries users.Querier
}

type ServiceConfig struct {
	Client     *Client
	DB         *pgxpool.Pool
	WebhookURL string
}

func NewService(cfg ServiceConfig) *Service {
	return &Service{
		client:      cfg.Client,
		db:          cfg.DB,
		webhookURL:  cfg.WebhookURL,
		repoQueries: internalrepos.New(cfg.DB),
		userQueries: users.New(cfg.DB),
	}
}

func (s *Service) Client() *Client {
	return s.client
}

// GetOrCreateGiteaUser ensures a Gitea user exists for the given MCP user
func (s *Service) GetOrCreateGiteaUser(ctx context.Context, userID string) (string, error) {
	// Check if user already has a Gitea username
	user, err := s.userQueries.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user.GiteaUsername != nil && *user.GiteaUsername != "" {
		return *user.GiteaUsername, nil
	}

	// Create new Gitea user
	giteaUsername := fmt.Sprintf("%s-%s", s.client.config.UserPrefix, userID)
	// Use a placeholder email - Gitea requires it but we don't need it
	giteaEmail := fmt.Sprintf("%s@users.ml.ink", giteaUsername)

	_, err = s.client.GetOrCreateUser(ctx, giteaUsername, giteaEmail)
	if err != nil {
		return "", fmt.Errorf("failed to create gitea user: %w", err)
	}

	// Store gitea username in database
	_, err = s.userQueries.SetGiteaUsername(ctx, users.SetGiteaUsernameParams{
		ID:             userID,
		GiteaUsername:  &giteaUsername,
	})
	if err != nil {
		return "", fmt.Errorf("failed to store gitea username: %w", err)
	}

	return giteaUsername, nil
}

// CreateRepo creates a new internal git repository
func (s *Service) CreateRepo(ctx context.Context, userID, repoName, description string, private bool) (*CreateRepoResult, error) {
	// Get or create Gitea user
	giteaUsername, err := s.GetOrCreateGiteaUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create the repo
	repo, err := s.client.CreateRepo(ctx, giteaUsername, &CreateRepoRequest{
		Name:          repoName,
		Description:   description,
		Private:       private,
		AutoInit:      true,
		DefaultBranch: "main",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create repo: %w", err)
	}

	// Create webhook for the repo
	_, err = s.client.CreateRepoWebhook(ctx, giteaUsername, repoName, &CreateWebhookRequest{
		Type: "gitea",
		Config: WebhookConfig{
			URL:         s.webhookURL,
			ContentType: "json",
			Secret:      s.client.config.WebhookSecret,
		},
		Events: []string{"push"},
		Active: true,
	})
	if err != nil {
		// Log but don't fail - webhook can be added later
		fmt.Printf("warning: failed to create webhook for %s/%s: %v\n", giteaUsername, repoName, err)
	}

	// Store repo in database
	_, err = s.repoQueries.CreateInternalRepo(ctx, internalrepos.CreateInternalRepoParams{
		UserID:   userID,
		Provider: "gitea",
		RepoID:   repo.ID,
		FullName: repo.FullName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store repo in database: %w", err)
	}

	// Get a push token
	token, err := s.createPushToken(ctx, giteaUsername, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to create push token: %w", err)
	}

	repoPath := s.client.GetRepoPath(giteaUsername, repoName)
	gitRemote := s.client.GetHTTPSCloneURL(giteaUsername, repoName, token)

	return &CreateRepoResult{
		Repo:      repoPath,
		GitRemote: gitRemote,
		Message:   "Push your code, then call create_app to deploy",
	}, nil
}

// GetPushToken gets a fresh push token for a repository
func (s *Service) GetPushToken(ctx context.Context, userID, repoFullName string) (*GetPushTokenResult, error) {
	// Verify repo belongs to user
	repo, err := s.repoQueries.GetInternalRepoByFullName(ctx, repoFullName)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	if repo.UserID != userID {
		return nil, fmt.Errorf("unauthorized: repo belongs to another user")
	}

	// Get user's gitea username
	user, err := s.userQueries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user.GiteaUsername == nil || *user.GiteaUsername == "" {
		return nil, fmt.Errorf("user has no gitea account")
	}

	giteaUsername := *user.GiteaUsername

	// Extract repo name from full_name (format: "username/reponame")
	repoName := ""
	for i := len(repoFullName) - 1; i >= 0; i-- {
		if repoFullName[i] == '/' {
			repoName = repoFullName[i+1:]
			break
		}
	}
	if repoName == "" {
		repoName = repoFullName
	}

	// Create new token
	token, err := s.createPushToken(ctx, giteaUsername, repoName)
	if err != nil {
		return nil, err
	}

	gitRemote := s.client.GetHTTPSCloneURL(giteaUsername, repoName, token)
	expiresAt := time.Now().Add(DefaultTokenDuration).Format(time.RFC3339)

	return &GetPushTokenResult{
		GitRemote: gitRemote,
		ExpiresAt: expiresAt,
	}, nil
}

// createPushToken creates a new access token for pushing
func (s *Service) createPushToken(ctx context.Context, giteaUsername, repoName string) (string, error) {
	// Token name includes repo for easier management
	tokenName := fmt.Sprintf("push-%s-%d", repoName, time.Now().Unix())

	// Delete any existing tokens with similar names to avoid clutter
	// (tokens accumulate over time otherwise)
	_ = s.client.DeleteAccessTokenByName(ctx, giteaUsername, tokenName)

	token, err := s.client.CreateAccessToken(ctx, giteaUsername, &CreateAccessTokenRequest{
		Name: tokenName,
		Scopes: []string{
			"write:repository",
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create access token: %w", err)
	}

	return token.SHA1, nil
}

// DeleteRepo deletes an internal git repository
func (s *Service) DeleteRepo(ctx context.Context, userID, repoFullName string) error {
	// Verify repo belongs to user
	repo, err := s.repoQueries.GetInternalRepoByFullName(ctx, repoFullName)
	if err != nil {
		return fmt.Errorf("repo not found: %w", err)
	}
	if repo.UserID != userID {
		return fmt.Errorf("unauthorized: repo belongs to another user")
	}

	// Get user's gitea username
	user, err := s.userQueries.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user.GiteaUsername == nil {
		return fmt.Errorf("user has no gitea account")
	}

	// Extract repo name
	repoName := ""
	for i := len(repoFullName) - 1; i >= 0; i-- {
		if repoFullName[i] == '/' {
			repoName = repoFullName[i+1:]
			break
		}
	}
	if repoName == "" {
		repoName = repoFullName
	}

	// Delete from Gitea
	if err := s.client.DeleteRepo(ctx, *user.GiteaUsername, repoName); err != nil {
		return fmt.Errorf("failed to delete repo from gitea: %w", err)
	}

	// Delete from database
	if err := s.repoQueries.DeleteInternalRepoByFullName(ctx, repoFullName); err != nil {
		return fmt.Errorf("failed to delete repo from database: %w", err)
	}

	return nil
}

// GetRepoByFullName retrieves an internal repo by its full name
func (s *Service) GetRepoByFullName(ctx context.Context, fullName string) (internalrepos.InternalRepo, error) {
	return s.repoQueries.GetInternalRepoByFullName(ctx, fullName)
}

// GetSSHCloneURL returns the SSH clone URL for a repo
func (s *Service) GetSSHCloneURL(giteaUsername, repoName string) string {
	return s.client.GetSSHCloneURL(giteaUsername, repoName)
}
