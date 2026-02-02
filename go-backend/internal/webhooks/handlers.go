package webhooks

import (
	"log/slog"

	"github.com/augustdev/autoclip/internal/auth"
	"github.com/augustdev/autoclip/internal/deployments"
	"github.com/augustdev/autoclip/internal/githubapp"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/apps"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/githubcreds"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	config        githubapp.Config
	deployService *deployments.Service
	authService   *auth.Service
	appsQ         apps.Querier
	githubCredsQ  githubcreds.Querier
	logger        *slog.Logger
}

func NewHandlers(
	config githubapp.Config,
	deployService *deployments.Service,
	authService *auth.Service,
	appsQ apps.Querier,
	githubCredsQ githubcreds.Querier,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		config:        config,
		deployService: deployService,
		authService:   authService,
		appsQ:         appsQ,
		githubCredsQ:  githubCredsQ,
		logger:        logger,
	}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Post("/webhooks/github", h.HandleGitHubWebhook)
}
