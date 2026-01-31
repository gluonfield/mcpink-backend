package githubapp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	config     Config
	privateKey *rsa.PrivateKey
	client     *http.Client
}

type Installation struct {
	ID      int64   `json:"id"`
	Account Account `json:"account"`
}

type Account struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

func NewService(config Config) (*Service, error) {
	// Handle escaped newlines from env var
	keyData := strings.ReplaceAll(config.PrivateKey, "\\n", "\n")

	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Service{
		config:     config,
		privateKey: privateKey,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (s *Service) generateJWT() (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": s.config.AppID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

// GetUserInstallation checks if a GitHub user has installed the app
// Returns the installation ID if installed, 0 if not installed
func (s *Service) GetUserInstallation(ctx context.Context, username string) (int64, error) {
	jwtToken, err := s.generateJWT()
	if err != nil {
		return 0, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/users/%s/installation", username)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get installation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// User hasn't installed the app
		return 0, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("github returned status %d: %s", resp.StatusCode, string(body))
	}

	var installation Installation
	if err := json.NewDecoder(resp.Body).Decode(&installation); err != nil {
		return 0, fmt.Errorf("failed to decode installation: %w", err)
	}

	return installation.ID, nil
}
