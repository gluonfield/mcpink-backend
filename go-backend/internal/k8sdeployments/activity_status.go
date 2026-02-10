package k8sdeployments

import (
	"context"
	"fmt"

	"github.com/augustdev/autoclip/internal/storage/pg/generated/apps"
)

func (a *Activities) UpdateBuildStatus(ctx context.Context, input UpdateBuildStatusInput) error {
	_, err := a.appsQ.UpdateBuildStatus(ctx, apps.UpdateBuildStatusParams{
		ID:          input.ServiceID,
		BuildStatus: input.BuildStatus,
	})
	if err != nil {
		return fmt.Errorf("update build status to %s: %w", input.BuildStatus, err)
	}
	a.logger.Info("Updated build status", "serviceID", input.ServiceID, "status", input.BuildStatus)
	return nil
}

func (a *Activities) MarkAppRunning(ctx context.Context, input MarkAppRunningInput) error {
	_, err := a.appsQ.UpdateAppRunning(ctx, apps.UpdateAppRunningParams{
		ID:         input.ServiceID,
		Fqdn:       &input.URL,
		CommitHash: &input.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("mark app running: %w", err)
	}
	a.logger.Info("Marked app running", "serviceID", input.ServiceID, "url", input.URL)
	return nil
}

func (a *Activities) MarkAppFailed(ctx context.Context, input MarkAppFailedInput) error {
	_, err := a.appsQ.UpdateAppFailed(ctx, apps.UpdateAppFailedParams{
		ID:           input.ServiceID,
		ErrorMessage: &input.ErrorMessage,
	})
	if err != nil {
		return fmt.Errorf("mark app failed: %w", err)
	}
	a.logger.Info("Marked app failed", "serviceID", input.ServiceID, "error", input.ErrorMessage)
	return nil
}

func (a *Activities) SoftDeleteApp(ctx context.Context, serviceID string) error {
	_, err := a.appsQ.SoftDeleteApp(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("soft delete app: %w", err)
	}
	a.logger.Info("Soft-deleted app", "serviceID", serviceID)
	return nil
}
