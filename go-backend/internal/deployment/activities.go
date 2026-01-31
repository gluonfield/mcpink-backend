package deployment

import (
	"context"
	"log/slog"

	"go.temporal.io/sdk/worker"
)

type Activities struct {
	logger *slog.Logger
}

func NewActivities(logger *slog.Logger) *Activities {
	return &Activities{
		logger: logger,
	}
}

func (a *Activities) UpdatePreviewStatus(ctx context.Context) error {
	return nil
}

func RegisterWorkflowsAndActivities(w worker.Worker, activities *Activities) {
	w.RegisterWorkflow(DeployWorkflow)
	w.RegisterActivity(activities.UpdatePreviewStatus)
}
