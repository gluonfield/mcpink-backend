package graph

import (
	"log/slog"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/augustdev/autoclip/internal/auth"
	"github.com/augustdev/autoclip/internal/coolify"
	"github.com/augustdev/autoclip/internal/githubapp"
	"github.com/augustdev/autoclip/internal/storage/pg"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/apps"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/projects"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/resources"
)

type Resolver struct {
	Db               *pg.DB
	Logger           *slog.Logger
	AuthService      *auth.Service
	GitHubAppService *githubapp.Service
	CoolifyClient    *coolify.Client
	AppQueries       apps.Querier
	ProjectQueries   projects.Querier
	ResourceQueries  resources.Querier
	FirebaseAuth     *firebaseauth.Client
}
