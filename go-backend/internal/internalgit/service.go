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

// CreateRepo creates a new internal git repository under user's GitHub username.
// Idempotent: if repo already exists for this user, returns credentials.
func (s *Service) CreateRepo(ctx context.Context, userID, repoName, description string, private bool) (*CreateRepoResult, error) {
	// Get user's GitHub username
	user, err := s.userQueries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	githubUsername := user.GithubUsername
	if githubUsername == "" {
		return nil, fmt.Errorf("user has no github username")
	}

	fullName := fmt.Sprintf("%s/%s", githubUsername, repoName)

	// Check if repo already exists in our database
	existingRepo, err := s.repoQueries.GetInternalRepoByFullName(ctx, fullName)
	if err == nil {
		// Repo exists - verify ownership
		if existingRepo.UserID != userID {
			return nil, fmt.Errorf("repo belongs to another user")
		}
		// Ensure deploy key exists (idempotent)
		if err := s.client.AddDeployKey(ctx, githubUsername, repoName, "coolify-deploy", s.client.config.DeployPublicKey); err != nil {
			if !isAlreadyExistsError(err) {
				return nil, fmt.Errorf("failed to add deploy key: %w", err)
			}
		}
		repoPath := s.client.GetRepoPath(githubUsername, repoName)
		gitRemote := s.client.GetHTTPSCloneURL(githubUsername, repoName, s.client.config.AdminToken)
		return &CreateRepoResult{
			Repo:      repoPath,
			GitRemote: gitRemote,
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			Message:   "Repo already exists. Push your code, then call create_app to deploy",
		}, nil
	}

	// Ensure Gitea user exists (matches GitHub username)
	email := fmt.Sprintf("%s@users.ml.ink", githubUsername)
	if err := s.client.EnsureUser(ctx, githubUsername, email); err != nil {
		return nil, fmt.Errorf("failed to ensure gitea user: %w", err)
	}

	// Create the repo under user's account (or get existing)
	repo, err := s.client.CreateRepoForUser(ctx, githubUsername, repoName, description, private)
	if err != nil {
		// Try to get existing repo from Gitea
		repo, err = s.client.GetRepo(ctx, githubUsername, repoName)
		if err != nil {
			return nil, fmt.Errorf("failed to create repo: %w", err)
		}
	}

	// Add deploy key for Coolify SSH access (ignore "already exists" errors)
	if err := s.client.AddDeployKey(ctx, githubUsername, repoName, "coolify-deploy", s.client.config.DeployPublicKey); err != nil {
		// Ignore if key already exists
		if !isAlreadyExistsError(err) {
			return nil, fmt.Errorf("failed to add deploy key: %w", err)
		}
	}

	// Add coolify-deploy bot as read-only collaborator (workaround for Coolify deploy-key bug)
	if err := s.client.AddCollaborator(ctx, githubUsername, repoName, DeployBotUsername); err != nil {
		if !isAlreadyExistsError(err) {
			return nil, fmt.Errorf("failed to add deploy bot collaborator: %w", err)
		}
	}

	// Create webhook for the repo (ignore "already exists" errors)
	_, err = s.client.CreateRepoWebhook(ctx, githubUsername, repoName, s.webhookURL, s.client.config.WebhookSecret)
	if err != nil && !isAlreadyExistsError(err) {
		fmt.Printf("warning: failed to create webhook for %s/%s: %v\n", githubUsername, repoName, err)
	}

	// Store repo in database (upsert)
	_, err = s.repoQueries.CreateInternalRepo(ctx, internalrepos.CreateInternalRepoParams{
		UserID:   userID,
		Provider: "gitea",
		RepoID:   repo.ID,
		FullName: repo.FullName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store repo in database: %w", err)
	}

	repoPath := s.client.GetRepoPath(githubUsername, repoName)
	gitRemote := s.client.GetHTTPSCloneURL(githubUsername, repoName, s.client.config.AdminToken)

	return &CreateRepoResult{
		Repo:      repoPath,
		GitRemote: gitRemote,
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
		Message:   "Push your code, then call create_app to deploy",
	}, nil
}

func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return contains(s, "already exist") || contains(s, "already exists") || contains(s, "has already been taken")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetPushToken returns git remote with admin token
func (s *Service) GetPushToken(ctx context.Context, userID, repoFullName string) (*GetPushTokenResult, error) {
	// Verify repo belongs to user
	repo, err := s.repoQueries.GetInternalRepoByFullName(ctx, repoFullName)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	if repo.UserID != userID {
		return nil, fmt.Errorf("unauthorized: repo belongs to another user")
	}

	// Extract owner and repo name from full_name (format: "owner/reponame")
	owner, repoName := splitFullName(repoFullName)
	if owner == "" || repoName == "" {
		return nil, fmt.Errorf("invalid repo full name: %s", repoFullName)
	}

	gitRemote := s.client.GetHTTPSCloneURL(owner, repoName, s.client.config.AdminToken)

	return &GetPushTokenResult{
		GitRemote: gitRemote,
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
	}, nil
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

	owner, repoName := splitFullName(repoFullName)

	// Delete from Gitea
	if err := s.client.DeleteRepo(ctx, owner, repoName); err != nil {
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
func (s *Service) GetSSHCloneURL(owner, repoName string) string {
	return s.client.GetSSHCloneURL(owner, repoName)
}

func splitFullName(fullName string) (owner, repo string) {
	for i := 0; i < len(fullName); i++ {
		if fullName[i] == '/' {
			return fullName[:i], fullName[i+1:]
		}
	}
	return "", ""
}
