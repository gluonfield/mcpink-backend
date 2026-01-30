package bootstrap

import (
	"github.com/go-chi/chi/v5"

	"github.com/augustdev/autoclip/internal/auth"
)

func NewAuthRouter(authHandlers *auth.Handlers) chi.Router {
	router := chi.NewRouter()
	authHandlers.RegisterRoutes(router)
	return router
}
