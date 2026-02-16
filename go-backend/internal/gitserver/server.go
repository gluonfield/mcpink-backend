package gitserver

import (
	"log/slog"
	"net/http"

	"github.com/augustdev/autoclip/internal/deployments"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/gittokens"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	config        Config
	gitTokensQ    gittokens.Querier
	servicesQ     services.Querier
	deployService *deployments.Service
	logger        *slog.Logger
	router        chi.Router
}

func NewServer(
	config Config,
	gitTokensQ gittokens.Querier,
	servicesQ services.Querier,
	deployService *deployments.Service,
	logger *slog.Logger,
) *Server {
	s := &Server{
		config:        config,
		gitTokensQ:    gitTokensQ,
		servicesQ:     servicesQ,
		deployService: deployService,
		logger:        logger,
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Git smart HTTP protocol routes
	r.Get("/{owner}/{repo}.git/info/refs", s.handleInfoRefs)
	r.Post("/{owner}/{repo}.git/git-upload-pack", s.handleUploadPack)
	r.Post("/{owner}/{repo}.git/git-receive-pack", s.handleReceivePack)

	s.router = r
	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}
