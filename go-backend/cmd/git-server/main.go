package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/augustdev/autoclip/internal/bootstrap"
	"github.com/augustdev/autoclip/internal/deployments"
	"github.com/augustdev/autoclip/internal/gitserver"
	"github.com/augustdev/autoclip/internal/storage/pg"
	"go.uber.org/fx"
)

type config struct {
	fx.Out

	Db        pg.DbConfig
	Temporal  bootstrap.TemporalClientConfig
	GitServer gitserver.Config
}

func main() {
	fx.New(
		fx.StopTimeout(15*time.Second),
		fx.Provide(
			bootstrap.NewLogger,
			bootstrap.LoadConfig[config],
			pg.NewDatabase,
			pg.NewServiceQueries,
			pg.NewDeploymentQueries,
			pg.NewProjectQueries,
			pg.NewUserQueries,
			pg.NewGitHubCredsQueries,
			pg.NewGitTokenQueries,
			pg.NewClusterMap,
			bootstrap.CreateTemporalClient,
			deployments.NewService,
			gitserver.NewServer,
		),
		fx.Invoke(
			startGitServer,
		),
	).Run()
}

func startGitServer(lc fx.Lifecycle, server *gitserver.Server, config gitserver.Config, logger *slog.Logger) {
	port := config.Port
	if port == "" {
		port = "3000"
	}

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: server.Handler(),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting git server", "port", port)
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Git server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down git server...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return httpServer.Shutdown(shutdownCtx)
		},
	})
}
