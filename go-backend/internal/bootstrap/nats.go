package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
)

func NewNatsClient(lc fx.Lifecycle, config NATSConfig, logger *slog.Logger) (*nats.Conn, error) {
	conn, err := nats.Connect(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			conn.Close()
			return nil
		},
	})

	logger.Info("Connected to NATS", "url", config.URL)
	return conn, nil
}
