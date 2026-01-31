package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/augustdev/autoclip/internal/github_oauth"
	"github.com/augustdev/autoclip/internal/storage/pg"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/apikeys"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/users"
)

type Service struct {
	config      Config
	db          *pg.DB
	usersQ      users.Querier
	apiKeysQ    apikeys.Querier
	githubOAuth *github_oauth.OAuthService
}

type Session struct {
	Token  string
	UserID string
	User   *users.User
}

type APIKeyResult struct {
	ID        string
	Name      string
	Prefix    string
	FullKey   string
	CreatedAt time.Time
}

func NewService(
	config Config,
	db *pg.DB,
	usersQ users.Querier,
	apiKeysQ apikeys.Querier,
	githubOAuth *github_oauth.OAuthService,
) *Service {
	return &Service{
		config:      config,
		db:          db,
		usersQ:      usersQ,
		apiKeysQ:    apiKeysQ,
		githubOAuth: githubOAuth,
	}
}

func (s *Service) HandleOAuthCallback(ctx context.Context, code string) (*Session, error) {
	tokenResp, err := s.githubOAuth.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	ghUser, err := s.githubOAuth.GetUser(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get github user: %w", err)
	}

	encryptedToken, err := s.encryptToken(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Parse scopes from GitHub response
	newScopes := parseScopes(tokenResp.Scope)

	user, err := s.usersQ.GetUserByGitHubID(ctx, ghUser.ID)
	if err != nil {
		// New user - just store the scopes we got
		user, err = s.usersQ.CreateUser(ctx, users.CreateUserParams{
			GithubID:       ghUser.ID,
			GithubUsername: ghUser.Login,
			GithubToken:    encryptedToken,
			AvatarUrl:      pgtype.Text{String: ghUser.AvatarURL, Valid: ghUser.AvatarURL != ""},
			GithubScopes:   newScopes,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Existing user - apply downgrade protection
		finalToken := encryptedToken
		finalScopes := newScopes

		hasRepoScopeOld := contains(user.GithubScopes, "repo")
		hasRepoScopeNew := contains(newScopes, "repo")

		// Downgrade attempt: old has repo, new doesn't
		if hasRepoScopeOld && !hasRepoScopeNew {
			// Check if old token is still valid
			oldTokenPlain, decryptErr := s.DecryptToken(user.GithubToken)
			if decryptErr == nil && s.githubOAuth.IsTokenValid(ctx, oldTokenPlain) {
				// Old token still works - keep it
				finalToken = user.GithubToken
				finalScopes = user.GithubScopes
			}
			// Otherwise accept the new weak token
		}

		user, err = s.usersQ.UpdateGitHubToken(ctx, users.UpdateGitHubTokenParams{
			ID:             user.ID,
			GithubToken:    finalToken,
			GithubUsername: ghUser.Login,
			AvatarUrl:      pgtype.Text{String: ghUser.AvatarURL, Valid: ghUser.AvatarURL != ""},
			GithubScopes:   finalScopes,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate jwt: %w", err)
	}

	return &Session{
		Token:  token,
		UserID: user.ID,
		User:   &user,
	}, nil
}

func (s *Service) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
		return "", fmt.Errorf("invalid subject claim type")
	}

	return "", fmt.Errorf("invalid token")
}

func (s *Service) GenerateAPIKey(ctx context.Context, userID string, name string) (*APIKeyResult, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	fullKey := fmt.Sprintf("dk_live_%x", keyBytes)

	keyHash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	prefix := fullKey[:16]

	apiKey, err := s.apiKeysQ.CreateAPIKey(ctx, apikeys.CreateAPIKeyParams{
		UserID:    userID,
		Name:      name,
		KeyHash:   string(keyHash),
		KeyPrefix: prefix,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}

	return &APIKeyResult{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Prefix:    apiKey.KeyPrefix,
		FullKey:   fullKey,
		CreatedAt: apiKey.CreatedAt.Time,
	}, nil
}

func (s *Service) ValidateAPIKey(ctx context.Context, key string) (string, error) {
	if len(key) < 16 {
		return "", fmt.Errorf("invalid api key format")
	}

	prefix := key[:16]

	apiKey, err := s.apiKeysQ.GetAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		return "", fmt.Errorf("api key not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(key)); err != nil {
		return "", fmt.Errorf("invalid api key")
	}

	_ = s.apiKeysQ.UpdateAPIKeyLastUsed(ctx, apiKey.ID)

	return apiKey.UserID, nil
}

func (s *Service) RevokeAPIKey(ctx context.Context, userID string, keyID string) error {
	return s.apiKeysQ.RevokeAPIKey(ctx, apikeys.RevokeAPIKeyParams{
		ID:     keyID,
		UserID: userID,
	})
}

func (s *Service) GetUserByID(ctx context.Context, userID string) (*users.User, error) {
	user, err := s.usersQ.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (s *Service) ListAPIKeys(ctx context.Context, userID string) ([]apikeys.ListAPIKeysByUserIDRow, error) {
	return s.apiKeysQ.ListAPIKeysByUserID(ctx, userID)
}

func (s *Service) SetGitHubAppInstallation(ctx context.Context, userID string, installationID int64) error {
	_, err := s.usersQ.SetGitHubAppInstallation(ctx, users.SetGitHubAppInstallationParams{
		ID:                      userID,
		GithubAppInstallationID: pgtype.Int8{Int64: installationID, Valid: true},
	})
	return err
}

func (s *Service) ClearGitHubAppInstallation(ctx context.Context, userID string) error {
	_, err := s.usersQ.ClearGitHubAppInstallation(ctx, userID)
	return err
}

func (s *Service) EncryptToken(token string) (string, error) {
	return s.encryptToken(token)
}

func (s *Service) generateJWT(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.config.JWTExpiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) encryptToken(token string) (string, error) {
	key := sha256.Sum256([]byte(s.config.APIKeyEncryptionKey))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *Service) DecryptToken(encrypted string) (string, error) {
	key := sha256.Sum256([]byte(s.config.APIKeyEncryptionKey))

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func parseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{}
	}
	return strings.FieldsFunc(scopeStr, func(r rune) bool {
		return r == ' ' || r == ','
	})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
