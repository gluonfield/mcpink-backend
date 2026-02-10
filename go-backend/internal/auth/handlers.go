package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/augustdev/autoclip/internal/github_oauth"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	service     *Service
	githubOAuth *github_oauth.OAuthService
	config      Config
	logger      *slog.Logger
}

func NewHandlers(
	service *Service,
	githubOAuth *github_oauth.OAuthService,
	config Config,
	logger *slog.Logger,
) *Handlers {
	return &Handlers{
		service:     service,
		githubOAuth: githubOAuth,
		config:      config,
		logger:      logger,
	}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/auth/github/callback", h.HandleGitHubCallback)
	r.Get("/auth/githubapp/callback", h.HandleGitHubAppCallback)
}

// HandleGitHubConnect initiates GitHub OAuth for linking a GitHub account.
// Requires that getUserID extracts the authenticated user from context.
// POST /auth/github/connect
func (h *Handlers) HandleGitHubConnect(getUserID func(r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getUserID(r)
		if userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		state, err := h.service.CreateOAuthState(userID)
		if err != nil {
			h.logger.Error("failed to create oauth state", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		var body struct {
			Scope string `json:"scope"`
		}
		if r.Body != nil {
			json.NewDecoder(r.Body).Decode(&body)
		}

		var scopes []string
		if body.Scope != "" {
			scopes = []string{body.Scope}
		}

		authURL := h.githubOAuth.GetAuthURLWithScopes(state, scopes)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"url": authURL})
	}
}

// HandleGitHubCallback handles the GitHub OAuth callback for the connect flow.
func (h *Handlers) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	redirectURL := h.config.FrontendURL + "/auth/callback"

	if errorParam == "access_denied" {
		h.logger.Info("user denied oauth access")
		http.Redirect(w, r, redirectURL+"?error=access_denied", http.StatusTemporaryRedirect)
		return
	}

	if state == "" {
		h.logger.Error("no state in callback")
		http.Redirect(w, r, redirectURL+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	userID, err := h.service.ValidateOAuthState(state)
	if err != nil {
		h.logger.Error("invalid state", "error", err)
		http.Redirect(w, r, redirectURL+"?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	if code == "" {
		h.logger.Error("no code in callback")
		http.Redirect(w, r, redirectURL+"?error=no_code", http.StatusTemporaryRedirect)
		return
	}

	if err := h.service.LinkGitHub(r.Context(), userID, code); err != nil {
		h.logger.Error("failed to link github", "error", err)
		http.Redirect(w, r, redirectURL+"?error=link_failed", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// HandleGitHubAppCallback redirects to frontend after GitHub App installation.
func (h *Handlers) HandleGitHubAppCallback(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.config.FrontendURL+"/githubapp/callback", http.StatusTemporaryRedirect)
}
