package gitserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/augustdev/autoclip/internal/deployments"
	"github.com/augustdev/autoclip/internal/storage/pg/generated/services"
)

// triggerDeploys finds services matching the given repo+branch with
// git_provider='internal' and starts redeploy workflows for each.
func triggerDeploys(
	ctx context.Context,
	logger *slog.Logger,
	servicesQ services.Querier,
	deployService *deployments.Service,
	repoFullName string,
	changes []ChangedRef,
) {
	for _, change := range changes {
		matchingServices, err := servicesQ.GetServicesByRepoBranchProvider(ctx, services.GetServicesByRepoBranchProviderParams{
			Repo:        repoFullName,
			Branch:      change.Branch,
			GitProvider: "internal",
		})
		if err != nil {
			logger.Error("failed to query services for deploy trigger",
				"repo", repoFullName,
				"branch", change.Branch,
				"error", err)
			continue
		}

		for _, svc := range matchingServices {
			workflowID, err := deployService.RedeployFromInternalGitPush(ctx, svc.ID, change.NewSHA)
			if err != nil {
				logger.Error("failed to start redeploy workflow",
					"serviceID", svc.ID,
					"repo", repoFullName,
					"branch", change.Branch,
					"error", err)
				continue
			}
			logger.Info("triggered redeploy",
				"serviceID", svc.ID,
				"workflowID", workflowID,
				"repo", repoFullName,
				"branch", change.Branch,
				"commitSHA", change.NewSHA)
		}
	}
}

// triggerDeploysForPush is called after a successful git receive-pack.
// It diffs refs before/after and triggers deploys for changed branches.
func (s *Server) triggerDeploysForPush(ctx context.Context, repoFullName string, before, after RefSnapshot) {
	changes := diffRefs(before, after)
	if len(changes) == 0 {
		return
	}

	s.logger.Info("detected ref changes after push",
		"repo", repoFullName,
		"changes", fmt.Sprintf("%d branch(es)", len(changes)))

	triggerDeploys(ctx, s.logger, s.servicesQ, s.deployService, repoFullName, changes)
}
