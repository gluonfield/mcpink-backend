package coolify

import (
	"context"
	"fmt"
	"net/url"
)

// ApplicationsService handles application-related API calls.
type ApplicationsService struct {
	client *Client
}

// BuildPack represents the available build pack types.
type BuildPack string

const (
	BuildPackNixpacks      BuildPack = "nixpacks"
	BuildPackStatic        BuildPack = "static"
	BuildPackDockerfile    BuildPack = "dockerfile"
	BuildPackDockerCompose BuildPack = "dockercompose"
)

// RedirectType represents the redirect configuration for www/non-www.
type RedirectType string

const (
	RedirectWWW    RedirectType = "www"
	RedirectNonWWW RedirectType = "non-www"
	RedirectBoth   RedirectType = "both"
)

// Application represents a Coolify application.
type Application struct {
	ID                               int     `json:"id,omitempty"`
	UUID                             string  `json:"uuid,omitempty"`
	Name                             string  `json:"name,omitempty"`
	Description                      *string `json:"description,omitempty"`
	FQDN                             *string `json:"fqdn,omitempty"`
	ConfigHash                       string  `json:"config_hash,omitempty"`
	Status                           string  `json:"status,omitempty"`
	RepositoryProjectID              *int    `json:"repository_project_id,omitempty"`
	SourceID                         *int    `json:"source_id,omitempty"`
	PrivateKeyID                     *int    `json:"private_key_id,omitempty"`
	DestinationType                  string  `json:"destination_type,omitempty"`
	DestinationID                    int     `json:"destination_id,omitempty"`
	EnvironmentID                    int     `json:"environment_id,omitempty"`
	CreatedAt                        string  `json:"created_at,omitempty"`
	UpdatedAt                        string  `json:"updated_at,omitempty"`
	DeletedAt                        *string `json:"deleted_at,omitempty"`

	// Git settings
	GitRepository              string  `json:"git_repository,omitempty"`
	GitBranch                  string  `json:"git_branch,omitempty"`
	GitCommitSHA               string  `json:"git_commit_sha,omitempty"`
	GitFullURL                 *string `json:"git_full_url,omitempty"`
	ManualWebhookSecretGitHub  *string `json:"manual_webhook_secret_github,omitempty"`
	ManualWebhookSecretGitLab  *string `json:"manual_webhook_secret_gitlab,omitempty"`
	ManualWebhookSecretBitbucket *string `json:"manual_webhook_secret_bitbucket,omitempty"`
	ManualWebhookSecretGitea   *string `json:"manual_webhook_secret_gitea,omitempty"`

	// Build settings
	BuildPack                        string  `json:"build_pack,omitempty"`
	StaticImage                      string  `json:"static_image,omitempty"`
	InstallCommand                   string  `json:"install_command,omitempty"`
	BuildCommand                     string  `json:"build_command,omitempty"`
	StartCommand                     string  `json:"start_command,omitempty"`
	BaseDirectory                    string  `json:"base_directory,omitempty"`
	PublishDirectory                 string  `json:"publish_directory,omitempty"`
	Dockerfile                       *string `json:"dockerfile,omitempty"`
	DockerfileLocation               string  `json:"dockerfile_location,omitempty"`
	DockerfileTargetBuild            *string `json:"dockerfile_target_build,omitempty"`
	DockerComposeLocation            string  `json:"docker_compose_location,omitempty"`
	DockerCompose                    *string `json:"docker_compose,omitempty"`
	DockerComposeRaw                 *string `json:"docker_compose_raw,omitempty"`
	DockerComposeDomains             *string `json:"docker_compose_domains,omitempty"`
	DockerComposeCustomStartCommand  *string `json:"docker_compose_custom_start_command,omitempty"`
	DockerComposeCustomBuildCommand  *string `json:"docker_compose_custom_build_command,omitempty"`

	// Docker registry settings
	DockerRegistryImageName *string `json:"docker_registry_image_name,omitempty"`
	DockerRegistryImageTag  *string `json:"docker_registry_image_tag,omitempty"`

	// Ports
	PortsExposes  string  `json:"ports_exposes,omitempty"`
	PortsMappings *string `json:"ports_mappings,omitempty"`

	// Health checks
	HealthCheckEnabled      bool    `json:"health_check_enabled,omitempty"`
	HealthCheckPath         string  `json:"health_check_path,omitempty"`
	HealthCheckPort         *string `json:"health_check_port,omitempty"`
	HealthCheckHost         *string `json:"health_check_host,omitempty"`
	HealthCheckMethod       string  `json:"health_check_method,omitempty"`
	HealthCheckReturnCode   int     `json:"health_check_return_code,omitempty"`
	HealthCheckScheme       string  `json:"health_check_scheme,omitempty"`
	HealthCheckResponseText *string `json:"health_check_response_text,omitempty"`
	HealthCheckInterval     int     `json:"health_check_interval,omitempty"`
	HealthCheckTimeout      int     `json:"health_check_timeout,omitempty"`
	HealthCheckRetries      int     `json:"health_check_retries,omitempty"`
	HealthCheckStartPeriod  int     `json:"health_check_start_period,omitempty"`
	CustomHealthcheckFound  bool    `json:"custom_healthcheck_found,omitempty"`

	// Resource limits
	LimitsMemory            string  `json:"limits_memory,omitempty"`
	LimitsMemorySwap        string  `json:"limits_memory_swap,omitempty"`
	LimitsMemorySwappiness  int     `json:"limits_memory_swappiness,omitempty"`
	LimitsMemoryReservation string  `json:"limits_memory_reservation,omitempty"`
	LimitsCPUs              string  `json:"limits_cpus,omitempty"`
	LimitsCPUSet            *string `json:"limits_cpuset,omitempty"`
	LimitsCPUShares         int     `json:"limits_cpu_shares,omitempty"`

	// Docker options
	CustomLabels           *string `json:"custom_labels,omitempty"`
	CustomDockerRunOptions string  `json:"custom_docker_run_options,omitempty"`
	CustomNetworkAliases   *string `json:"custom_network_aliases,omitempty"`
	CustomNginxConfiguration *string `json:"custom_nginx_configuration,omitempty"`

	// Deployment commands
	PostDeploymentCommand          string `json:"post_deployment_command,omitempty"`
	PostDeploymentCommandContainer string `json:"post_deployment_command_container,omitempty"`
	PreDeploymentCommand           string `json:"pre_deployment_command,omitempty"`
	PreDeploymentCommandContainer  string `json:"pre_deployment_command_container,omitempty"`

	// URL settings
	PreviewURLTemplate string  `json:"preview_url_template,omitempty"`
	Redirect           *string `json:"redirect,omitempty"`
	WatchPaths         *string `json:"watch_paths,omitempty"`

	// Swarm settings
	SwarmReplicas              int     `json:"swarm_replicas,omitempty"`
	SwarmPlacementConstraints  *string `json:"swarm_placement_constraints,omitempty"`

	// Compose settings
	ComposeParsingVersion *string `json:"compose_parsing_version,omitempty"`

	// HTTP Basic Auth
	IsHTTPBasicAuthEnabled bool    `json:"is_http_basic_auth_enabled,omitempty"`
	HTTPBasicAuthUsername  *string `json:"http_basic_auth_username,omitempty"`
	HTTPBasicAuthPassword  *string `json:"http_basic_auth_password,omitempty"`
}

// CreatePrivateGitHubAppRequest represents the request body for creating an application
// from a private GitHub repository using a GitHub App.
// See: https://coolify.io/docs/api-reference/api/operations/create-private-github-app-application
type CreatePrivateGitHubAppRequest struct {
	// Required fields
	ProjectUUID     string    `json:"project_uuid"`
	ServerUUID      string    `json:"server_uuid"`
	EnvironmentName string    `json:"environment_name,omitempty"` // Required: one of environment_name or environment_uuid
	EnvironmentUUID string    `json:"environment_uuid,omitempty"` // Required: one of environment_name or environment_uuid
	GitHubAppUUID   string    `json:"github_app_uuid"`
	GitRepository   string    `json:"git_repository"`
	GitBranch       string    `json:"git_branch"`
	PortsExposes    string    `json:"ports_exposes"`
	BuildPack       BuildPack `json:"build_pack"`

	// Optional fields
	DestinationUUID         string `json:"destination_uuid,omitempty"`
	Name                    string `json:"name,omitempty"`
	Description             string `json:"description,omitempty"`
	Domains                 string `json:"domains,omitempty"`
	GitCommitSHA            string `json:"git_commit_sha,omitempty"`
	DockerRegistryImageName string `json:"docker_registry_image_name,omitempty"`
	DockerRegistryImageTag  string `json:"docker_registry_image_tag,omitempty"`

	// Static/SPA settings
	IsStatic    *bool  `json:"is_static,omitempty"`
	IsSPA       *bool  `json:"is_spa,omitempty"`
	StaticImage string `json:"static_image,omitempty"`

	// Auto-deploy settings
	IsAutoDeployEnabled  *bool `json:"is_auto_deploy_enabled,omitempty"`
	IsForceHTTPSEnabled  *bool `json:"is_force_https_enabled,omitempty"`
	InstantDeploy        *bool `json:"instant_deploy,omitempty"`

	// Build commands
	InstallCommand string `json:"install_command,omitempty"`
	BuildCommand   string `json:"build_command,omitempty"`
	StartCommand   string `json:"start_command,omitempty"`

	// Directory settings
	BaseDirectory    string `json:"base_directory,omitempty"`
	PublishDirectory string `json:"publish_directory,omitempty"`

	// Port mappings
	PortsMappings string `json:"ports_mappings,omitempty"`

	// Health check settings
	HealthCheckEnabled      *bool  `json:"health_check_enabled,omitempty"`
	HealthCheckPath         string `json:"health_check_path,omitempty"`
	HealthCheckPort         string `json:"health_check_port,omitempty"`
	HealthCheckHost         string `json:"health_check_host,omitempty"`
	HealthCheckMethod       string `json:"health_check_method,omitempty"`
	HealthCheckReturnCode   *int   `json:"health_check_return_code,omitempty"`
	HealthCheckScheme       string `json:"health_check_scheme,omitempty"`
	HealthCheckResponseText string `json:"health_check_response_text,omitempty"`
	HealthCheckInterval     *int   `json:"health_check_interval,omitempty"`
	HealthCheckTimeout      *int   `json:"health_check_timeout,omitempty"`
	HealthCheckRetries      *int   `json:"health_check_retries,omitempty"`
	HealthCheckStartPeriod  *int   `json:"health_check_start_period,omitempty"`

	// Resource limits
	LimitsMemory            string `json:"limits_memory,omitempty"`
	LimitsMemorySwap        string `json:"limits_memory_swap,omitempty"`
	LimitsMemorySwappiness  *int   `json:"limits_memory_swappiness,omitempty"`
	LimitsMemoryReservation string `json:"limits_memory_reservation,omitempty"`
	LimitsCPUs              string `json:"limits_cpus,omitempty"`
	LimitsCPUSet            string `json:"limits_cpuset,omitempty"`
	LimitsCPUShares         *int   `json:"limits_cpu_shares,omitempty"`

	// Docker options
	CustomLabels           string `json:"custom_labels,omitempty"`
	CustomDockerRunOptions string `json:"custom_docker_run_options,omitempty"`

	// Deployment commands
	PostDeploymentCommand          string `json:"post_deployment_command,omitempty"`
	PostDeploymentCommandContainer string `json:"post_deployment_command_container,omitempty"`
	PreDeploymentCommand           string `json:"pre_deployment_command,omitempty"`
	PreDeploymentCommandContainer  string `json:"pre_deployment_command_container,omitempty"`

	// Webhook secrets
	ManualWebhookSecretGitHub    string `json:"manual_webhook_secret_github,omitempty"`
	ManualWebhookSecretGitLab    string `json:"manual_webhook_secret_gitlab,omitempty"`
	ManualWebhookSecretBitbucket string `json:"manual_webhook_secret_bitbucket,omitempty"`
	ManualWebhookSecretGitea     string `json:"manual_webhook_secret_gitea,omitempty"`

	// Redirect settings
	Redirect RedirectType `json:"redirect,omitempty"`

	// Dockerfile settings
	Dockerfile         string `json:"dockerfile,omitempty"`
	DockerfileLocation string `json:"dockerfile_location,omitempty"`

	// Docker Compose settings
	DockerComposeLocation           string                 `json:"docker_compose_location,omitempty"`
	DockerComposeCustomStartCommand string                 `json:"docker_compose_custom_start_command,omitempty"`
	DockerComposeCustomBuildCommand string                 `json:"docker_compose_custom_build_command,omitempty"`
	DockerComposeDomains            []DockerComposeDomain  `json:"docker_compose_domains,omitempty"`

	// Watch paths
	WatchPaths string `json:"watch_paths,omitempty"`

	// Build server
	UseBuildServer *bool `json:"use_build_server,omitempty"`

	// HTTP Basic Auth
	IsHTTPBasicAuthEnabled *bool  `json:"is_http_basic_auth_enabled,omitempty"`
	HTTPBasicAuthUsername  string `json:"http_basic_auth_username,omitempty"`
	HTTPBasicAuthPassword  string `json:"http_basic_auth_password,omitempty"`

	// Network settings
	ConnectToDockerNetwork *bool `json:"connect_to_docker_network,omitempty"`

	// Domain settings
	ForceDomainOverride *bool `json:"force_domain_override,omitempty"`
	AutogenerateDomain  *bool `json:"autogenerate_domain,omitempty"`

	// Label escaping
	IsContainerLabelEscapeEnabled *bool `json:"is_container_label_escape_enabled,omitempty"`
}

// DockerComposeDomain represents a domain mapping for docker compose applications.
type DockerComposeDomain struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// CreateApplicationResponse represents the response from creating an application.
type CreateApplicationResponse struct {
	UUID string `json:"uuid"`
}

// StartApplicationResponse represents the response from starting an application.
type StartApplicationResponse struct {
	Message        string `json:"message"`
	DeploymentUUID string `json:"deployment_uuid"`
}

// ListApplicationsOptions represents the options for listing applications.
type ListApplicationsOptions struct {
	Tag string // Filter applications by tag name
}

// List retrieves all applications.
// See: https://coolify.io/docs/api-reference/api/operations/list-applications
func (s *ApplicationsService) List(ctx context.Context, opts *ListApplicationsOptions) ([]Application, error) {
	var query url.Values
	if opts != nil && opts.Tag != "" {
		query = url.Values{}
		query.Set("tag", opts.Tag)
	}

	var apps []Application
	if err := s.client.do(ctx, "GET", "/api/v1/applications", query, nil, &apps); err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	return apps, nil
}

// Get retrieves an application by its UUID.
// See: https://coolify.io/docs/api-reference/api/operations/get-application-by-uuid
func (s *ApplicationsService) Get(ctx context.Context, uuid string) (*Application, error) {
	if uuid == "" {
		return nil, fmt.Errorf("coolify: uuid is required")
	}

	var app Application
	if err := s.client.do(ctx, "GET", "/api/v1/applications/"+uuid, nil, nil, &app); err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	return &app, nil
}

// CreatePrivateGitHubApp creates a new application from a private GitHub repository using a GitHub App.
// See: https://coolify.io/docs/api-reference/api/operations/create-private-github-app-application
func (s *ApplicationsService) CreatePrivateGitHubApp(ctx context.Context, req *CreatePrivateGitHubAppRequest) (*CreateApplicationResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("coolify: request is required")
	}
	if req.ProjectUUID == "" {
		return nil, fmt.Errorf("coolify: project_uuid is required")
	}
	if req.ServerUUID == "" {
		return nil, fmt.Errorf("coolify: server_uuid is required")
	}
	if req.EnvironmentName == "" && req.EnvironmentUUID == "" {
		return nil, fmt.Errorf("coolify: environment_name or environment_uuid is required")
	}
	if req.GitHubAppUUID == "" {
		return nil, fmt.Errorf("coolify: github_app_uuid is required")
	}
	if req.GitRepository == "" {
		return nil, fmt.Errorf("coolify: git_repository is required")
	}
	if req.GitBranch == "" {
		return nil, fmt.Errorf("coolify: git_branch is required")
	}
	if req.PortsExposes == "" {
		return nil, fmt.Errorf("coolify: ports_exposes is required")
	}
	if req.BuildPack == "" {
		return nil, fmt.Errorf("coolify: build_pack is required")
	}

	var resp CreateApplicationResponse
	if err := s.client.do(ctx, "POST", "/api/v1/applications/private-github-app", nil, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	return &resp, nil
}

// StartOptions represents the options for starting an application.
type StartOptions struct {
	Force         bool // Force rebuild
	InstantDeploy bool // Skip queuing
}

// Start starts (deploys) an application.
// See: https://coolify.io/docs/api-reference/api/operations/start-application-by-uuid
func (s *ApplicationsService) Start(ctx context.Context, uuid string, opts *StartOptions) (*StartApplicationResponse, error) {
	if uuid == "" {
		return nil, fmt.Errorf("coolify: uuid is required")
	}

	var query url.Values
	if opts != nil {
		query = url.Values{}
		if opts.Force {
			query.Set("force", "true")
		}
		if opts.InstantDeploy {
			query.Set("instant_deploy", "true")
		}
	}

	var resp StartApplicationResponse
	if err := s.client.do(ctx, "GET", "/api/v1/applications/"+uuid+"/start", query, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to start application: %w", err)
	}
	return &resp, nil
}

// Stop stops an application.
// See: https://coolify.io/docs/api-reference/api/operations/stop-application-by-uuid
func (s *ApplicationsService) Stop(ctx context.Context, uuid string) error {
	if uuid == "" {
		return fmt.Errorf("coolify: uuid is required")
	}

	if err := s.client.do(ctx, "GET", "/api/v1/applications/"+uuid+"/stop", nil, nil, nil); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	return nil
}

// Restart restarts an application.
// See: https://coolify.io/docs/api-reference/api/operations/restart-application-by-uuid
func (s *ApplicationsService) Restart(ctx context.Context, uuid string) (*StartApplicationResponse, error) {
	if uuid == "" {
		return nil, fmt.Errorf("coolify: uuid is required")
	}

	var resp StartApplicationResponse
	if err := s.client.do(ctx, "GET", "/api/v1/applications/"+uuid+"/restart", nil, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to restart application: %w", err)
	}
	return &resp, nil
}

// Delete deletes an application by its UUID.
// See: https://coolify.io/docs/api-reference/api/operations/delete-application-by-uuid
func (s *ApplicationsService) Delete(ctx context.Context, uuid string) error {
	if uuid == "" {
		return fmt.Errorf("coolify: uuid is required")
	}

	if err := s.client.do(ctx, "DELETE", "/api/v1/applications/"+uuid, nil, nil, nil); err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	return nil
}
