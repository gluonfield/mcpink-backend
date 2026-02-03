package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) handleCreateRepo(ctx context.Context, req *mcp.CallToolRequest, input CreateRepoInput) (*mcp.CallToolResult, CreateRepoOutput, error) {
	user := UserFromContext(ctx)
	if user == nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "not authenticated"}}}, CreateRepoOutput{}, nil
	}

	if input.Name == "" {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "name is required"}}}, CreateRepoOutput{}, nil
	}

	source := input.Source
	if source == "" {
		source = "private"
	}

	switch source {
	case "private":
		return s.createPrivateRepo(ctx, user.ID, input)
	case "github":
		return s.createGitHubRepoUnified(ctx, user.ID, input)
	default:
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "source must be 'private' or 'github'"}}}, CreateRepoOutput{}, nil
	}
}

func (s *Server) createPrivateRepo(ctx context.Context, userID string, input CreateRepoInput) (*mcp.CallToolResult, CreateRepoOutput, error) {
	if s.internalGitSvc == nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "internal git not configured"}}}, CreateRepoOutput{}, nil
	}

	private := true
	if input.Private != nil {
		private = *input.Private
	}

	result, err := s.internalGitSvc.CreateRepo(ctx, userID, input.Name, input.Description, private)
	if err != nil {
		s.logger.Error("failed to create internal repo", "error", err, "name", input.Name)
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to create repository: %v", err)}}}, CreateRepoOutput{}, nil
	}

	return nil, CreateRepoOutput{
		Repo:      result.Repo,
		GitRemote: result.GitRemote,
		Message:   result.Message,
	}, nil
}

func (s *Server) createGitHubRepoUnified(ctx context.Context, userID string, input CreateRepoInput) (*mcp.CallToolResult, CreateRepoOutput, error) {
	// For GitHub repos, redirect to the existing create_github_repo tool logic
	// This requires OAuth repo scope
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "For GitHub repositories, use source='private' (recommended) or use the existing create_github_repo tool directly."}}}, CreateRepoOutput{}, nil
}

func (s *Server) handleGetPushToken(ctx context.Context, req *mcp.CallToolRequest, input GetPushTokenInput) (*mcp.CallToolResult, GetPushTokenOutput, error) {
	user := UserFromContext(ctx)
	if user == nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "not authenticated"}}}, GetPushTokenOutput{}, nil
	}

	if input.Repo == "" {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "repo is required"}}}, GetPushTokenOutput{}, nil
	}

	// Determine source from repo path
	if strings.HasPrefix(input.Repo, "github.com/") {
		return s.getGitHubPushTokenUnified(ctx, user.ID, input.Repo)
	}

	if strings.HasPrefix(input.Repo, "ml.ink/") {
		return s.getPrivatePushToken(ctx, user.ID, input.Repo)
	}

	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "repo must start with 'github.com/' or 'ml.ink/'"}}}, GetPushTokenOutput{}, nil
}

func (s *Server) getPrivatePushToken(ctx context.Context, userID, repoPath string) (*mcp.CallToolResult, GetPushTokenOutput, error) {
	if s.internalGitSvc == nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "internal git not configured"}}}, GetPushTokenOutput{}, nil
	}

	// Extract full_name from path (ml.ink/username/repo -> username/repo)
	fullName := strings.TrimPrefix(repoPath, "ml.ink/")

	result, err := s.internalGitSvc.GetPushToken(ctx, userID, fullName)
	if err != nil {
		s.logger.Error("failed to get push token", "error", err, "repo", fullName)
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get push token: %v", err)}}}, GetPushTokenOutput{}, nil
	}

	return nil, GetPushTokenOutput{
		GitRemote: result.GitRemote,
		ExpiresAt: result.ExpiresAt,
	}, nil
}

func (s *Server) getGitHubPushTokenUnified(ctx context.Context, userID, repoPath string) (*mcp.CallToolResult, GetPushTokenOutput, error) {
	// Get GitHub credentials
	creds, err := s.authService.GetGitHubCredsByUserID(ctx, userID)
	if err != nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "GitHub not connected"}}}, GetPushTokenOutput{}, nil
	}

	if creds.GithubAppInstallationID == nil {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "GitHub App not installed"}}}, GetPushTokenOutput{}, nil
	}

	// Extract repo from github.com/owner/repo format
	repo := strings.TrimPrefix(repoPath, "github.com/")
	parts := strings.Split(repo, "/")
	repoName := repo
	if len(parts) == 2 {
		repoName = parts[1]
	}

	installationToken, err := s.githubAppService.CreateInstallationToken(ctx, *creds.GithubAppInstallationID, []string{repoName})
	if err != nil {
		s.logger.Error("failed to get GitHub push token", "error", err, "repo", repo)
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get push token: %v", err)}}}, GetPushTokenOutput{}, nil
	}

	return nil, GetPushTokenOutput{
		GitRemote: fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", installationToken.Token, repo),
		ExpiresAt: installationToken.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
