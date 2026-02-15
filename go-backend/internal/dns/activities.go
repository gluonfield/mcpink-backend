package dns

import (
	"context"
	"log/slog"

	"github.com/augustdev/autoclip/internal/powerdns"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/delegatedzones"
	"go.temporal.io/sdk/activity"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Activities struct {
	logger          *slog.Logger
	k8s             kubernetes.Interface
	dynClient       dynamic.Interface
	delegatedZonesQ delegatedzones.Querier
	pdns            *powerdns.Client
}

func NewActivities(
	logger *slog.Logger,
	k8s kubernetes.Interface,
	dynClient dynamic.Interface,
	delegatedZonesQ delegatedzones.Querier,
	pdns *powerdns.Client,
) *Activities {
	return &Activities{
		logger:          logger,
		k8s:             k8s,
		dynClient:       dynClient,
		delegatedZonesQ: delegatedZonesQ,
		pdns:            pdns,
	}
}

func recordHeartbeat(ctx context.Context, details ...any) {
	if activity.IsActivity(ctx) {
		activity.RecordHeartbeat(ctx, details...)
	}
}
