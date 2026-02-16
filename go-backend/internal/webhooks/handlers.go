package webhooks

import (
	"log/slog"
	"net/http"

	"github.com/augustdev/autoclip/internal/deployments"
	"github.com/augustdev/autoclip/internal/githubapp"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/githubcreds"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/services"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	config        githubapp.Config
	deployService *deployments.Service
	servicesQ     services.Querier
	githubCredsQ  githubcreds.Querier
	logger        *slog.Logger
}

func NewHandlers(
	config githubapp.Config,
	deployService *deployments.Service,
	servicesQ services.Querier,
	githubCredsQ githubcreds.Querier,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		config:        config,
		deployService: deployService,
		servicesQ:     servicesQ,
		githubCredsQ:  githubCredsQ,
		logger:        logger,
	}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.Post("/webhooks/github", h.HandleGitHubWebhook)
}
