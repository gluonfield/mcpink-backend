package mcp_oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	config    Config
	jwtSecret string
}

func NewService(config Config, jwtSecret string) *Service {
	return &Service{
		config:    config,
		jwtSecret: jwtSecret,
	}
}

// AuthCodeClaims contains the data embedded in the JWT auth code.
// No database table needed - the JWT is self-contained.
type AuthCodeClaims struct {
	jwt.RegisteredClaims
	UserID        string `json:"uid"`
	ClientID      string `json:"cid"`
	RedirectURI   string `json:"ruri"`
	CodeChallenge string `json:"cc"`
	APIKeyID      string `json:"kid"`
	APIKey        string `json:"key"` // The actual dk_live_* key
}

// GenerateAuthCode creates a signed JWT containing the auth code data.
// The API key is embedded directly - when the client exchanges the code,
// we just verify the JWT and return the embedded key.
func (s *Service) GenerateAuthCode(userID, clientID, redirectURI, codeChallenge, apiKeyID, apiKey string) (string, error) {
	now := time.Now()
	claims := AuthCodeClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
		UserID:        userID,
		ClientID:      clientID,
		RedirectURI:   redirectURI,
		CodeChallenge: codeChallenge,
		APIKeyID:      apiKeyID,
		APIKey:        apiKey,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateAuthCode validates the JWT auth code and returns the claims.
func (s *Service) ValidateAuthCode(code string) (*AuthCodeClaims, error) {
	token, err := jwt.ParseWithClaims(code, &AuthCodeClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid auth code: %w", err)
	}

	claims, ok := token.Claims.(*AuthCodeClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid auth code claims")
	}

	return claims, nil
}

// VerifyPKCE verifies that the code_verifier matches the code_challenge.
// Uses S256 method: BASE64URL(SHA256(verifier)) == challenge
func VerifyPKCE(verifier, challenge string) bool {
	hash := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])
	return computed == challenge
}
