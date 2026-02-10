package auth

import "time"

type Config struct {
	JWTSecret           string
	JWTExpiry           time.Duration
	APIKeyEncryptionKey string
	FrontendURL         string
}
