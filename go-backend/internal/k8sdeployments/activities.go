package k8sdeployments

import (
	"log/slog"

	"github.com/augustdev/autoclip/internal/githubapp"
	"github.com/augustdev/autoclip/internal/internalgit"
	deploymentsdb "github.com/augustdev/autoclip/internal/storage/pg/generated/deployments"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/projects"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/services"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/users"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Activities struct {
	logger         *slog.Logger
	k8s            kubernetes.Interface
	dynClient      dynamic.Interface
	githubApp      *githubapp.Service
	internalGitSvc *internalgit.Service
	servicesQ      services.Querier
	deploymentsQ   deploymentsdb.Querier
	projectsQ      projects.Querier
	usersQ         users.Querier
	config         Config
}

func NewActivities(
	logger *slog.Logger,
	k8s kubernetes.Interface,
	dynClient dynamic.Interface,
	githubApp *githubapp.Service,
	internalGitSvc *internalgit.Service,
	servicesQ services.Querier,
	deploymentsQ deploymentsdb.Querier,
	projectsQ projects.Querier,
	usersQ users.Querier,
	config Config,
) *Activities {
	return &Activities{
		logger:         logger,
		k8s:            k8s,
		dynClient:      dynClient,
		githubApp:      githubApp,
		internalGitSvc: internalGitSvc,
		servicesQ:      servicesQ,
		deploymentsQ:   deploymentsQ,
		projectsQ:      projectsQ,
		usersQ:         usersQ,
		config:         config,
	}
}
