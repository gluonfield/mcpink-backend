package authz

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseTokenValidator struct {
	authClient *auth.Client
}

func NewFirebaseTokenValidator(ctx context.Context, serviceAccountJSON string) (*FirebaseTokenValidator, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}

	return &FirebaseTokenValidator{
		authClient: authClient,
	}, nil
}

func (f *FirebaseTokenValidator) AuthClient() *auth.Client {
	return f.authClient
}

func (f *FirebaseTokenValidator) ValidateToken(tokenString string) (string, []string, error) {
	ctx := context.Background()

	token, err := f.authClient.VerifyIDToken(ctx, tokenString)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if token.UID == "" {
		return "", nil, fmt.Errorf("%w: no user ID found in Firebase token", ErrInvalidToken)
	}

	var roles []string
	if token.Claims != nil {
		if rolesInterface, ok := token.Claims["roles"]; ok {
			if rolesArray, ok := rolesInterface.([]interface{}); ok {
				for _, r := range rolesArray {
					if roleStr, ok := r.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		} else if roleInterface, ok := token.Claims["role"]; ok {
			if roleStr, ok := roleInterface.(string); ok {
				roles = []string{roleStr}
			}
		}
	}

	return token.UID, roles, nil
}
