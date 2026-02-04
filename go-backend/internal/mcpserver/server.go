package mcpserver

import (
	"log/slog"
	"net/http"

	"github.com/augustdev/autoclip/internal/auth"
	"github.com/augustdev/autoclip/internal/coolify"
	"github.com/augustdev/autoclip/internal/deployments"
	"github.com/augustdev/autoclip/internal/githubapp"
	"github.com/augustdev/autoclip/internal/internalgit"
	"github.com/augustdev/autoclip/internal/logs"
	"github.com/augustdev/autoclip/internal/resources"
	"github.com/invopop/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var reflector = jsonschema.Reflector{
	DoNotReference: true,
}

func schemaFor[T any]() any {
	return reflector.Reflect(new(T))
}

type Server struct {
	mcpServer        *mcp.Server
	authService      *auth.Service
	coolifyClient    *coolify.Client
	deployService    *deployments.Service
	resourcesService *resources.Service
	githubAppService *githubapp.Service
	internalGitSvc   *internalgit.Service
	logProvider      logs.Provider
	logger           *slog.Logger
}

func NewServer(authService *auth.Service, coolifyClient *coolify.Client, deployService *deployments.Service, resourcesService *resources.Service, githubAppService *githubapp.Service, internalGitSvc *internalgit.Service, logProvider logs.Provider, logger *slog.Logger) *Server {
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "Ink MCP",
			Title:   "Ink MCP - deploy your apps and servers on the cloud",
			Version: "0.1.0",
		},
		&mcp.ServerOptions{
			Instructions: "This server has capabilities to deploy most single port apps NextJS, React, flask etc as well as many other servers and returns the URL of the deployed application. It can provision sqlite databases too. 99% of apps should be deployable with this MCP.",
		},
	)

	s := &Server{
		mcpServer:        mcpServer,
		authService:      authService,
		coolifyClient:    coolifyClient,
		deployService:    deployService,
		resourcesService: resourcesService,
		githubAppService: githubAppService,
		internalGitSvc:   internalGitSvc,
		logProvider:      logProvider,
		logger:           logger,
	}

	s.registerTools()

	logger.Info("MCP server initialized")
	return s
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "whoami",
		Description: "Get information about the authenticated user",
		InputSchema: schemaFor[WhoamiInput](),
	}, s.handleWhoami)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_app",
		Description: "Create and deploy an application. Use host='ml.ink' (default) for private repos or host='github.com' for GitHub.",
		InputSchema: schemaFor[CreateAppInput](),
	}, s.handleCreateApp)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "redeploy",
		Description: "Redeploy an existing application to pull latest code",
		InputSchema: schemaFor[RedeployInput](),
	}, s.handleRedeploy)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_apps",
		Description: "List all deployed applications",
		InputSchema: schemaFor[ListAppsInput](),
	}, s.handleListApps)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_resource",
		Description: "Create a new resource (e.g., sqlite database). Returns connection URL and auth token.",
		InputSchema: schemaFor[CreateResourceInput](),
	}, s.handleCreateResource)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_resources",
		Description: "List all resources (databases, etc.) for the authenticated user",
		InputSchema: schemaFor[ListResourcesInput](),
	}, s.handleListResources)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_resource",
		Description: "Get detailed information about a resource including connection URL and auth token",
		InputSchema: schemaFor[GetResourceDetailsInput](),
	}, s.handleGetResourceDetails)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_resource",
		Description: "Delete a resource (e.g., sqlite database). This permanently removes the resource.",
		InputSchema: schemaFor[DeleteResourceInput](),
	}, s.handleDeleteResource)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_app_details",
		Description: "Get detailed information about a deployed application, optionally including environment variables and logs",
		InputSchema: schemaFor[GetAppDetailsInput](),
	}, s.handleGetAppDetails)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_app",
		Description: "Delete an application. This permanently removes the deployment.",
		InputSchema: schemaFor[DeleteAppInput](),
	}, s.handleDeleteApp)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_repo",
		Description: "Create a git repository. Use host='ml.ink' (default) for instant private repos, or host='github.com' for GitHub.",
		InputSchema: schemaFor[CreateRepoInput](),
	}, s.handleCreateRepo)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_git_token",
		Description: "Get a temporary git token to push code. Example: name='myapp', host='ml.ink' (default) or host='github.com'.",
		InputSchema: schemaFor[GetGitTokenInput](),
	}, s.handleGetGitToken)
}

func (s *Server) Handler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, &mcp.StreamableHTTPOptions{
		Stateless: true,
	})
}
