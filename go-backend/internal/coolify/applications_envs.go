package coolify

import (
	"context"
	"fmt"
)

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

type CreateEnvRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime *bool  `json:"is_build_time,omitempty"`
	IsLiteral   *bool  `json:"is_literal,omitempty"`
	IsMultiline *bool  `json:"is_multiline,omitempty"`
	IsPreview   *bool  `json:"is_preview,omitempty"`
}

type UpdateEnvRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime *bool  `json:"is_build_time,omitempty"`
	IsLiteral   *bool  `json:"is_literal,omitempty"`
	IsMultiline *bool  `json:"is_multiline,omitempty"`
	IsPreview   *bool  `json:"is_preview,omitempty"`
}

type BulkUpdateEnvsRequest struct {
	Data []CreateEnvRequest `json:"data"`
}

type CreateEnvResponse struct {
	UUID string `json:"uuid"`
}

type DeleteEnvRequest struct {
	Key string `json:"key"`
}

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
