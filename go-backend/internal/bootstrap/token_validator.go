package bootstrap

import (
	"context"
	"errors"
	"log/slog"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/augustdev/autoclip/internal/authz"
	"go.uber.org/fx"
)

type TokenValidatorResult struct {
	fx.Out

	TokenValidator authz.TokenValidator
	FirebaseAuth   *firebaseauth.Client `optional:"true"`
}

type TokenValidatorConfig struct {
	ValidatorType string
	JWTJWKSURL    string
}

func NewTokenValidator(cfg TokenValidatorConfig, firebaseCfg FirebaseConfig, logger *slog.Logger) (TokenValidatorResult, error) {
	switch cfg.ValidatorType {
	case "firebase":
		if firebaseCfg.ServiceAccountJSON == "" {
			return TokenValidatorResult{}, errors.New("firebase service account JSON is required (set FIREBASE_SERVICEACCOUNTJSON env var)")
		}
		validator, err := authz.NewFirebaseTokenValidator(context.Background(), firebaseCfg.ServiceAccountJSON)
		if err != nil {
			logger.Error("Failed to create Firebase token validator", slog.Any("error", err))
			return TokenValidatorResult{}, err
		}
		return TokenValidatorResult{
			TokenValidator: validator,
			FirebaseAuth:   validator.AuthClient(),
		}, nil

	case "jwk":
		tokenValidator, err := authz.NewTokenValidator(cfg.JWTJWKSURL)
		if err != nil {
			logger.Error("Failed to create JWT token validator", slog.Any("error", err))
			return TokenValidatorResult{}, err
		}
		return TokenValidatorResult{
			TokenValidator: tokenValidator,
		}, nil

	default:
		logger.Error("Invalid validator type", "validator_type", cfg.ValidatorType)
		return TokenValidatorResult{}, errors.New("validator type must be 'firebase', 'jwk', or 'test'")
	}
}
