package coolify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type SourcesService struct {
	client *Client
}

type CreateGitHubAppSourceRequest struct {
	Name           string `json:"name"`
	Organization   string `json:"organization,omitempty"`
	APIUrl         string `json:"api_url"`
	HTMLUrl        string `json:"html_url"`
	CustomUser     string `json:"custom_user"`
	CustomPort     int    `json:"custom_port"`
	AppID          int64  `json:"app_id"`
	InstallationID int64  `json:"installation_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	WebhookSecret  string `json:"webhook_secret"`
	PrivateKeyUUID string `json:"private_key_uuid"`
	IsSystemWide   bool   `json:"is_system_wide"`
}

type GitHubAppSource struct {
	ID   int64  `json:"id"`
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func (s *SourcesService) ListGitHubApps(ctx context.Context) ([]GitHubAppSource, error) {
	var result []GitHubAppSource
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/github-apps", nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SourcesService) CreateGitHubApp(ctx context.Context, req *CreateGitHubAppSourceRequest) (*GitHubAppSource, error) {
	var result GitHubAppSource
	err := s.client.do(ctx, http.MethodPost, "/api/v1/github-apps", nil, req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *SourcesService) DeleteGitHubApp(ctx context.Context, uuid string) error {
	return s.client.do(ctx, http.MethodDelete, "/api/v1/github-apps/"+uuid, nil, nil, nil)
}

type UpdateGitHubAppSourceRequest struct {
	InstallationID int64 `json:"installation_id"`
}

func (s *SourcesService) GetGitHubApp(ctx context.Context, uuid string) (*GitHubAppSource, error) {
	var result GitHubAppSource
	err := s.client.do(ctx, http.MethodGet, "/api/v1/github-apps/"+uuid, nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *SourcesService) UpdateGitHubApp(ctx context.Context, uuid string, req *UpdateGitHubAppSourceRequest) error {
	// Coolify PATCH requires integer ID, not UUID. Some Coolify instances do not support GET by UUID,
	// so we fall back to listing and matching by UUID.
	source, err := s.GetGitHubApp(ctx, uuid)
	if err != nil {
		var apiErr *Error
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
			return fmt.Errorf("failed to get source: %w", err)
		}

		sources, listErr := s.ListGitHubApps(ctx)
		if listErr != nil {
			return fmt.Errorf("failed to list sources after uuid lookup failed: %w (original: %v)", listErr, err)
		}

		for i := range sources {
			if sources[i].UUID == uuid {
				source = &sources[i]
				break
			}
		}
		if source == nil {
			return fmt.Errorf("github app source not found in coolify: %s", uuid)
		}
	}

	return s.client.do(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/github-apps/%d", source.ID), nil, req, nil)
}
