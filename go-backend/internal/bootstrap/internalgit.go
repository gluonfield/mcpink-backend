package bootstrap

import (
	"log/slog"

	"github.com/augustdev/autoclip/internal/internalgit"
	"github.com/augustdev/autoclip/internal/storage/pg"
)

func NewInternalGitService(config internalgit.Config, db *pg.DB, logger *slog.Logger) *internalgit.Service {
	if !config.IsEnabled() {
		logger.Info("Internal git service not configured, skipping")
		return nil
	}

	client, err := internalgit.NewClient(config)
	if err != nil {
		logger.Error("failed to create internal git client", "error", err)
		return nil
	}

	// Webhook URL should point to our server's public endpoint
	// For now, we construct it from the auth frontend URL or use a placeholder
	// This will be called by Gitea when pushes occur
	webhookURL := "https://api.ml.ink/webhooks/internal-git"

	svc := internalgit.NewService(internalgit.ServiceConfig{
		Client:     client,
		DB:         db.Pool,
		WebhookURL: webhookURL,
	})

	logger.Info("Internal git service initialized", "baseURL", config.BaseURL)
	return svc
}
