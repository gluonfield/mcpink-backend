package k8sdeployments

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// recordHeartbeat sends a heartbeat to Temporal if ctx is an activity context.
// Safe to call from tests with a plain context.
func recordHeartbeat(ctx context.Context, details ...any) {
	if activity.IsActivity(ctx) {
		activity.RecordHeartbeat(ctx, details...)
	}
}
