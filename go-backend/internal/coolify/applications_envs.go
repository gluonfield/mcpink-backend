package coolify

import (
	"context"
	"fmt"
)

// EnvVar represents an environment variable for an application.
type EnvVar struct {
	ID            int     `json:"id,omitempty"`
	UUID          string  `json:"uuid,omitempty"`
	ApplicationID int     `json:"application_id,omitempty"`
	Key           string  `json:"key"`
	Value         string  `json:"value"`
	IsBuildTime   bool    `json:"is_build_time,omitempty"`
	IsShownOnce   bool    `json:"is_shown_once,omitempty"`
	IsLiteral     bool    `json:"is_literal,omitempty"`
	IsMultiline   bool    `json:"is_multiline,omitempty"`
	IsPreview     bool    `json:"is_preview,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

// CreateEnvRequest represents the request body for creating an environment variable.
// See: https://coolify.io/docs/api-reference/api/operations/create-env-by-application-uuid
type CreateEnvRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime *bool  `json:"is_build_time,omitempty"`
	IsLiteral   *bool  `json:"is_literal,omitempty"`
	IsMultiline *bool  `json:"is_multiline,omitempty"`
	IsPreview   *bool  `json:"is_preview,omitempty"`
}

// UpdateEnvRequest represents the request body for updating an environment variable.
// See: https://coolify.io/docs/api-reference/api/operations/update-env-by-application-uuid
type UpdateEnvRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime *bool  `json:"is_build_time,omitempty"`
	IsLiteral   *bool  `json:"is_literal,omitempty"`
	IsMultiline *bool  `json:"is_multiline,omitempty"`
	IsPreview   *bool  `json:"is_preview,omitempty"`
}

// BulkUpdateEnvsRequest represents the request body for bulk updating environment variables.
// See: https://coolify.io/docs/api-reference/api/operations/update-envs-by-application-uuid
type BulkUpdateEnvsRequest struct {
	Data []CreateEnvRequest `json:"data"`
}

// CreateEnvResponse represents the response from creating an environment variable.
type CreateEnvResponse struct {
	UUID string `json:"uuid"`
}

// ListEnvs retrieves all environment variables for an application.
// See: https://coolify.io/docs/api-reference/api/operations/list-envs-by-application-uuid
func (s *ApplicationsService) ListEnvs(ctx context.Context, uuid string) ([]EnvVar, error) {
	if uuid == "" {
		return nil, fmt.Errorf("coolify: uuid is required")
	}

	var envs []EnvVar
	if err := s.client.do(ctx, "GET", "/api/v1/applications/"+uuid+"/envs", nil, nil, &envs); err != nil {
		return nil, fmt.Errorf("failed to list environment variables: %w", err)
	}
	return envs, nil
}

// CreateEnv creates a new environment variable for an application.
// See: https://coolify.io/docs/api-reference/api/operations/create-env-by-application-uuid
func (s *ApplicationsService) CreateEnv(ctx context.Context, uuid string, req *CreateEnvRequest) (*CreateEnvResponse, error) {
	if uuid == "" {
		return nil, fmt.Errorf("coolify: uuid is required")
	}
	if req == nil {
		return nil, fmt.Errorf("coolify: request is required")
	}
	if req.Key == "" {
		return nil, fmt.Errorf("coolify: key is required")
	}

	var resp CreateEnvResponse
	if err := s.client.do(ctx, "POST", "/api/v1/applications/"+uuid+"/envs", nil, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create environment variable: %w", err)
	}
	return &resp, nil
}

// UpdateEnv updates an existing environment variable for an application.
// See: https://coolify.io/docs/api-reference/api/operations/update-env-by-application-uuid
func (s *ApplicationsService) UpdateEnv(ctx context.Context, uuid string, req *UpdateEnvRequest) error {
	if uuid == "" {
		return fmt.Errorf("coolify: uuid is required")
	}
	if req == nil {
		return fmt.Errorf("coolify: request is required")
	}
	if req.Key == "" {
		return fmt.Errorf("coolify: key is required")
	}

	if err := s.client.do(ctx, "PATCH", "/api/v1/applications/"+uuid+"/envs", nil, req, nil); err != nil {
		return fmt.Errorf("failed to update environment variable: %w", err)
	}
	return nil
}

// BulkUpdateEnvs updates multiple environment variables for an application at once.
// See: https://coolify.io/docs/api-reference/api/operations/update-envs-by-application-uuid
func (s *ApplicationsService) BulkUpdateEnvs(ctx context.Context, uuid string, req *BulkUpdateEnvsRequest) error {
	if uuid == "" {
		return fmt.Errorf("coolify: uuid is required")
	}
	if req == nil {
		return fmt.Errorf("coolify: request is required")
	}

	if err := s.client.do(ctx, "PATCH", "/api/v1/applications/"+uuid+"/envs/bulk", nil, req, nil); err != nil {
		return fmt.Errorf("failed to bulk update environment variables: %w", err)
	}
	return nil
}

// DeleteEnvRequest represents the request body for deleting an environment variable.
type DeleteEnvRequest struct {
	Key string `json:"key"`
}

// DeleteEnv deletes an environment variable from an application.
// See: https://coolify.io/docs/api-reference/api/operations/delete-env-by-application-uuid
func (s *ApplicationsService) DeleteEnv(ctx context.Context, uuid string, req *DeleteEnvRequest) error {
	if uuid == "" {
		return fmt.Errorf("coolify: uuid is required")
	}
	if req == nil {
		return fmt.Errorf("coolify: request is required")
	}
	if req.Key == "" {
		return fmt.Errorf("coolify: key is required")
	}

	if err := s.client.do(ctx, "DELETE", "/api/v1/applications/"+uuid+"/envs", nil, req, nil); err != nil {
		return fmt.Errorf("failed to delete environment variable: %w", err)
	}
	return nil
}
