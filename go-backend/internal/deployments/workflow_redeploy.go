package deployments

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func RedeployToCoolifyWorkflow(ctx workflow.Context, input RedeployWorkflowInput) (RedeployWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting RedeployWorkflow",
		"appID", input.AppID,
		"coolifyAppUUID", input.CoolifyAppUUID)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var activities *Activities

	// Step 1: Mark app as building
	err := workflow.ExecuteActivity(ctx, activities.MarkAppBuilding, input.AppID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to mark app as building", "error", err)
		return RedeployWorkflowResult{
			AppID:        input.AppID,
			Status:       string(BuildStatusFailed),
			ErrorMessage: err.Error(),
		}, err
	}

	// Step 2: Call Coolify Deploy API (pulls latest code)
	deployInput := DeployAppInput{
		AppID:          input.AppID,
		CoolifyAppUUID: input.CoolifyAppUUID,
	}

	var deployResult DeployAppResult
	err = workflow.ExecuteActivity(ctx, activities.DeployApp, deployInput).Get(ctx, &deployResult)
	if err != nil {
		logger.Error("Failed to deploy app", "error", err)
		_ = markAppFailed(ctx, activities, input.AppID, err.Error())
		return RedeployWorkflowResult{
			AppID:        input.AppID,
			Status:       string(BuildStatusFailed),
			ErrorMessage: err.Error(),
		}, err
	}

	logger.Info("Deployment triggered", "deploymentUUID", deployResult.DeploymentUUID)

	// Step 3: Wait for app to be running (reuse existing WaitForRunning activity)
	waitCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 4 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    3,
		},
	})

	waitInput := WaitForRunningInput{
		AppID:          input.AppID,
		CoolifyAppUUID: input.CoolifyAppUUID,
		DeploymentUUID: deployResult.DeploymentUUID,
	}

	var waitResult WaitForRunningResult
	err = workflow.ExecuteActivity(waitCtx, activities.WaitForRunning, waitInput).Get(ctx, &waitResult)
	if err != nil {
		logger.Error("App failed to reach running state", "error", err)
		_ = markAppFailed(ctx, activities, input.AppID, err.Error())
		return RedeployWorkflowResult{
			AppID:        input.AppID,
			Status:       string(BuildStatusFailed),
			ErrorMessage: err.Error(),
		}, err
	}

	logger.Info("RedeployWorkflow completed successfully",
		"appID", input.AppID,
		"fqdn", waitResult.FQDN)

	return RedeployWorkflowResult{
		AppID:  input.AppID,
		FQDN:   waitResult.FQDN,
		Status: string(BuildStatusSuccess),
	}, nil
}
