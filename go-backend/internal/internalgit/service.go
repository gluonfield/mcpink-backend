package internalgit

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"

	"github.com/augustdev/autoclip/internal/storage/pg"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/gittokens"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/internalrepos"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/users"
	"github.com/jackc/pgx/v5/pgtype"
)

const tokenPrefix = "mlg_"

type Service struct {
	config      Config
	repoQueries internalrepos.Querier
	tokenQ      gittokens.Querier
	userQueries users.Querier
}

func NewService(config Config, db *pg.DB) (*Service, error) {
	if config.PublicGitURL == "" {
		return nil, fmt.Errorf("internalgit: PublicGitURL is required")
	}

	return &Service{
		config:      config,
		repoQueries: internalrepos.New(db.Pool),
		tokenQ:      gittokens.New(db.Pool),
		userQueries: users.New(db.Pool),
	}, nil
}

// CreateRepo creates a new internal git repository record and returns a push token.
// Idempotent: if repo already exists for this user, returns a new token.
func (s *Service) CreateRepo(ctx context.Context, userID, repoName, description string, private bool) (*CreateRepoResult, error) {
	user, err := s.userQueries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	gitUsername, err := s.ResolveUsername(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("resolve username: %w", err)
	}

	fullName := fmt.Sprintf("%s/%s", gitUsername, repoName)

	existingRepo, err := s.repoQueries.GetInternalRepoByFullName(ctx, fullName)
	if err == nil {
		if existingRepo.UserID != userID {
			return nil, fmt.Errorf("repo belongs to another user")
		}
		rawToken, err := s.createToken(ctx, userID, &existingRepo.ID, nil)
		if err != nil {
			return nil, fmt.Errorf("create token: %w", err)
		}
		return &CreateRepoResult{
			Repo:      s.repoPath(gitUsername, repoName),
			GitRemote: s.cloneURL(gitUsername, repoName, rawToken),
			ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
			Message:   "Repo already exists. Push your code, then call create_app to deploy",
		}, nil
	}

	// Create DB record (bare repo is created on first push by git-server)
	barePath := fmt.Sprintf("%s/%s.git", gitUsername, repoName)
	_, err = s.repoQueries.CreateInternalRepo(ctx, internalrepos.CreateInternalRepoParams{
		UserID:   userID,
		Name:     repoName,
		CloneUrl: s.cloneURLWithoutAuth(gitUsername, repoName),
		Provider: "internal",
		FullName: fullName,
		BarePath: &barePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store repo in database: %w", err)
	}

	// Re-fetch to get the repo ID for token creation
	repo, err := s.repoQueries.GetInternalRepoByFullName(ctx, fullName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created repo: %w", err)
	}

	rawToken, err := s.createToken(ctx, userID, &repo.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &CreateRepoResult{
		Repo:      s.repoPath(gitUsername, repoName),
		GitRemote: s.cloneURL(gitUsername, repoName, rawToken),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
		Message:   "Push your code, then call create_app to deploy",
	}, nil
}

// GetPushToken returns git remote with a per-repo scoped token.
func (s *Service) GetPushToken(ctx context.Context, userID, repoFullName string) (*GetPushTokenResult, error) {
	repo, err := s.repoQueries.GetInternalRepoByFullName(ctx, repoFullName)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	if repo.UserID != userID {
		return nil, fmt.Errorf("unauthorized: repo belongs to another user")
	}

	owner, repoName := splitFullName(repoFullName)
	if owner == "" || repoName == "" {
		return nil, fmt.Errorf("invalid repo full name: %s", repoFullName)
	}

	rawToken, err := s.createToken(ctx, userID, &repo.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}

	return &GetPushTokenResult{
		GitRemote: s.cloneURL(owner, repoName, rawToken),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339),
	}, nil
}

// DeleteRepo deletes an internal git repository from the database.
// Bare repo cleanup on disk should be handled separately.
func (s *Service) DeleteRepo(ctx context.Context, userID, repoFullName string) error {
	repo, err := s.repoQueries.GetInternalRepoByFullName(ctx, repoFullName)
	if err != nil {
		return fmt.Errorf("repo not found: %w", err)
	}
	if repo.UserID != userID {
		return fmt.Errorf("unauthorized: repo belongs to another user")
	}

	repoID := repo.ID
	if err := s.tokenQ.RevokeTokensByRepoID(ctx, &repoID); err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	if err := s.repoQueries.DeleteInternalRepoByFullName(ctx, repoFullName); err != nil {
		return fmt.Errorf("failed to delete repo from database: %w", err)
	}

	return nil
}

// GetRepoByFullName retrieves an internal repo by its full name.
func (s *Service) GetRepoByFullName(ctx context.Context, fullName string) (internalrepos.InternalRepo, error) {
	return s.repoQueries.GetInternalRepoByFullName(ctx, fullName)
}

// createToken generates a new random token, stores its hash, and returns the raw token.
func (s *Service) createToken(ctx context.Context, userID string, repoID *string, expiresAt *time.Time) (string, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	rawToken := tokenPrefix + base64.RawURLEncoding.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(rawToken))
	hashStr := hex.EncodeToString(hash[:])

	prefix := rawToken[:8]

	var expiresAtPg pgtype.Timestamptz
	if expiresAt != nil {
		expiresAtPg = pgtype.Timestamptz{Time: *expiresAt, Valid: true}
	}

	_, err := s.tokenQ.CreateToken(ctx, gittokens.CreateTokenParams{
		TokenHash:   hashStr,
		TokenPrefix: prefix,
		UserID:      userID,
		RepoID:      repoID,
		Scopes:      []string{"push", "pull"},
		ExpiresAt:   expiresAtPg,
	})
	if err != nil {
		return "", fmt.Errorf("store token: %w", err)
	}

	return rawToken, nil
}

// cloneURL builds https://x-git-token:{token}@git.ml.ink/{owner}/{repo}.git
func (s *Service) cloneURL(owner, repoName, token string) string {
	u, _ := url.Parse(s.config.PublicGitURL)
	u.User = url.UserPassword("x-git-token", token)
	u.Path = fmt.Sprintf("/%s/%s.git", owner, repoName)
	return u.String()
}

// cloneURLWithoutAuth builds https://git.ml.ink/{owner}/{repo}.git
func (s *Service) cloneURLWithoutAuth(owner, repoName string) string {
	u, _ := url.Parse(s.config.PublicGitURL)
	u.Path = fmt.Sprintf("/%s/%s.git", owner, repoName)
	return u.String()
}

// repoPath returns "ml.ink/{owner}/{repo}" (strips "git." prefix).
func (s *Service) repoPath(owner, repoName string) string {
	u, _ := url.Parse(s.config.PublicGitURL)
	host := u.Hostname()
	if len(host) > 4 && host[:4] == "git." {
		host = host[4:]
	}
	return fmt.Sprintf("%s/%s/%s", host, owner, repoName)
}

func splitFullName(fullName string) (owner, repo string) {
	for i := 0; i < len(fullName); i++ {
		if fullName[i] == '/' {
			return fullName[:i], fullName[i+1:]
		}
	}
	return "", ""
}
