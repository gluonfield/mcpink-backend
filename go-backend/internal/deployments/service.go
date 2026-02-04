package deployments

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/augustdev/autoclip/internal/cloudflare"
	"github.com/augustdev/autoclip/internal/coolify"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/apps"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/dnsrecords"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/projects"
	"github.com/lithammer/shortuuid/v4"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
)

type Service struct {
	temporalClient   client.Client
	appsQ            apps.Querier
	projectsQ        projects.Querier
	dnsQ             dnsrecords.Querier
	coolifyClient    *coolify.Client
	cloudflareClient *cloudflare.Client
	logger           *slog.Logger
}

func NewService(
	temporalClient client.Client,
	appsQ apps.Querier,
	projectsQ projects.Querier,
	dnsQ dnsrecords.Querier,
	coolifyClient *coolify.Client,
	cloudflareClient *cloudflare.Client,
	logger *slog.Logger,
) *Service {
	return &Service{
		temporalClient:   temporalClient,
		appsQ:            appsQ,
		projectsQ:        projectsQ,
		dnsQ:             dnsQ,
		coolifyClient:    coolifyClient,
		cloudflareClient: cloudflareClient,
		logger:           logger,
	}
}

type CreateAppInput struct {
	UserID         string
	ProjectRef     string
	GitHubAppUUID  string
	Repo           string
	Branch         string
	Name           string
	BuildPack      string
	Port           string
	EnvVars        []EnvVar
	GitProvider    string // "github" or "gitea"
	PrivateKeyUUID string // for internal git
	SSHCloneURL    string // for internal git
	Memory         string
	CPU            string
	InstallCommand string
	BuildCommand   string
	StartCommand   string
}

type CreateAppResult struct {
	AppID      string
	Name       string
	Status     string
	Repo       string
	WorkflowID string
}

func (s *Service) CreateApp(ctx context.Context, input CreateAppInput) (*CreateAppResult, error) {
	var projectID string
	if input.ProjectRef != "" {
		project, err := s.getOrCreateProject(ctx, input.UserID, input.ProjectRef)
		if err != nil {
			return nil, err
		}
		projectID = project.ID
	} else {
		project, err := s.projectsQ.GetDefaultProject(ctx, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("default project not found for user")
		}
		projectID = project.ID
	}

	appID := shortuuid.New()
	workflowID := fmt.Sprintf("deploy-%s-%s-%s", input.UserID, input.Repo, input.Branch)

	gitProvider := input.GitProvider
	if gitProvider == "" {
		gitProvider = "github"
	}

	workflowInput := DeployWorkflowInput{
		AppID:          appID,
		UserID:         input.UserID,
		ProjectID:      projectID,
		GitHubAppUUID:  input.GitHubAppUUID,
		Repo:           input.Repo,
		Branch:         input.Branch,
		Name:           input.Name,
		BuildPack:      input.BuildPack,
		Port:           input.Port,
		EnvVars:        input.EnvVars,
		GitProvider:    gitProvider,
		PrivateKeyUUID: input.PrivateKeyUUID,
		SSHCloneURL:    input.SSHCloneURL,
		Memory:         input.Memory,
		CPU:            input.CPU,
		InstallCommand: input.InstallCommand,
		BuildCommand:   input.BuildCommand,
		StartCommand:   input.StartCommand,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "default",
	}

	run, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, DeployToCoolifyWorkflow, workflowInput)
	if err != nil {
		s.logger.Error("failed to start deploy workflow",
			"workflowID", workflowID,
			"error", err)
		return nil, fmt.Errorf("failed to start deploy workflow: %w", err)
	}

	s.logger.Info("started deploy workflow",
		"workflowID", workflowID,
		"runID", run.GetRunID())

	return &CreateAppResult{
		AppID:      appID,
		Name:       input.Name,
		Status:     string(BuildStatusQueued),
		Repo:       input.Repo,
		WorkflowID: workflowID,
	}, nil
}

func (s *Service) ListApps(ctx context.Context, userID string, limit, offset int32) ([]apps.App, error) {
	appList, err := s.appsQ.ListAppsByUserID(ctx, apps.ListAppsByUserIDParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	return appList, nil
}

func (s *Service) GetProjectByRef(ctx context.Context, userID, ref string) (*projects.Project, error) {
	return s.getOrCreateProject(ctx, userID, ref)
}

func (s *Service) getOrCreateProject(ctx context.Context, userID, ref string) (*projects.Project, error) {
	project, err := s.projectsQ.GetProjectByRef(ctx, projects.GetProjectByRefParams{
		UserID: userID,
		Ref:    ref,
	})
	if err == nil {
		return &project, nil
	}

	s.logger.Info("auto-creating project", "user_id", userID, "ref", ref)
	newProject, err := s.projectsQ.CreateProject(ctx, projects.CreateProjectParams{
		UserID: userID,
		Name:   ref,
		Ref:    ref,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return &newProject, nil
}

func (s *Service) GetAppByNameAndProject(ctx context.Context, name, projectID string) (*apps.App, error) {
	app, err := s.appsQ.GetAppByNameAndProject(ctx, apps.GetAppByNameAndProjectParams{
		Name:      &name,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %s", name)
	}
	return &app, nil
}

type GetAppByNameParams struct {
	Name    string
	Project string // "default" uses user's default project
	UserID  string
}

func (s *Service) GetAppByName(ctx context.Context, params GetAppByNameParams) (*apps.App, error) {
	project := params.Project
	if project == "" {
		project = "default"
	}

	app, err := s.appsQ.GetAppByNameAndUserProject(ctx, apps.GetAppByNameAndUserProjectParams{
		Name:   &params.Name,
		UserID: params.UserID,
		Ref:    project,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %s in project %s", params.Name, project)
	}
	return &app, nil
}

func (s *Service) RedeployApp(ctx context.Context, appID, coolifyAppUUID string) (string, error) {
	workflowID := fmt.Sprintf("redeploy-%s-%s", appID, shortuuid.New())
	return s.RedeployAppWithWorkflowID(ctx, appID, coolifyAppUUID, workflowID)
}

// RedeployFromGitHubPush starts (or reuses) a redeploy workflow triggered by a GitHub push.
//
// GitHub delivery is at-least-once, so we treat it as potentially duplicated and use a deterministic workflow ID
// derived from the commit SHA (preferred) or delivery ID (fallback).
func (s *Service) RedeployFromGitHubPush(ctx context.Context, appID, coolifyAppUUID, afterSHA, deliveryID string) (string, error) {
	key := strings.TrimSpace(afterSHA)
	if key == "" || key == "0000000000000000000000000000000000000000" {
		key = strings.TrimSpace(deliveryID)
	}
	if key == "" {
		key = shortuuid.New()
	}

	workflowID := fmt.Sprintf("redeploy-%s-%s", appID, key)
	return s.RedeployAppWithWorkflowID(ctx, appID, coolifyAppUUID, workflowID)
}

// RedeployFromInternalGitPush starts (or reuses) a redeploy workflow triggered by an internal git (Gitea) push.
func (s *Service) RedeployFromInternalGitPush(ctx context.Context, appID, coolifyAppUUID, afterSHA string) (string, error) {
	key := strings.TrimSpace(afterSHA)
	if key == "" || key == "0000000000000000000000000000000000000000" {
		key = shortuuid.New()
	}

	workflowID := fmt.Sprintf("redeploy-%s-%s", appID, key)
	return s.RedeployAppWithWorkflowID(ctx, appID, coolifyAppUUID, workflowID)
}

func (s *Service) RedeployAppWithWorkflowID(ctx context.Context, appID, coolifyAppUUID, workflowID string) (string, error) {
	if workflowID == "" {
		workflowID = fmt.Sprintf("redeploy-%s-%s", appID, shortuuid.New())
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "default",
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, RedeployToCoolifyWorkflow, RedeployWorkflowInput{
		AppID:          appID,
		CoolifyAppUUID: coolifyAppUUID,
	})
	if err != nil {
		var alreadyStarted *serviceerror.WorkflowExecutionAlreadyStarted
		if errors.As(err, &alreadyStarted) {
			s.logger.Info("redeploy workflow already started, skipping duplicate",
				"workflowID", workflowID,
				"appID", appID)
			return workflowID, nil
		}

		s.logger.Error("failed to start redeploy workflow",
			"workflowID", workflowID,
			"error", err)
		return "", fmt.Errorf("failed to start redeploy workflow: %w", err)
	}

	s.logger.Info("started redeploy workflow",
		"workflowID", workflowID,
		"runID", we.GetRunID())

	return we.GetID(), nil
}

type DeleteAppParams struct {
	Name    string
	Project string
	UserID  string
}

type DeleteAppResult struct {
	AppID      string
	Name       string
	WorkflowID string
}

func (s *Service) DeleteApp(ctx context.Context, params DeleteAppParams) (*DeleteAppResult, error) {
	project := params.Project
	if project == "" {
		project = "default"
	}

	app, err := s.appsQ.GetAppByNameAndUserProject(ctx, apps.GetAppByNameAndUserProjectParams{
		Name:   &params.Name,
		UserID: params.UserID,
		Ref:    project,
	})
	if err != nil {
		return nil, fmt.Errorf("app not found: %s in project %s", params.Name, project)
	}

	var name string
	if app.Name != nil {
		name = *app.Name
	}

	coolifyUUID := ""
	if app.CoolifyAppUuid != nil {
		coolifyUUID = *app.CoolifyAppUuid
	}

	workflowID := fmt.Sprintf("delete-app-%s", app.ID)

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "default",
	}

	input := DeleteAppWorkflowInput{
		AppID:          app.ID,
		CoolifyAppUUID: coolifyUUID,
	}

	run, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, DeleteAppWorkflow, input)
	if err != nil {
		return nil, fmt.Errorf("failed to start delete workflow: %w", err)
	}

	s.logger.Info("started delete app workflow",
		"app_id", app.ID,
		"name", name,
		"workflow_id", run.GetID())

	return &DeleteAppResult{
		AppID:      app.ID,
		Name:       name,
		WorkflowID: run.GetID(),
	}, nil
}
