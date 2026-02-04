package deployments

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func DeleteAppWorkflow(ctx workflow.Context, input DeleteAppWorkflowInput) (DeleteAppWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DeleteAppWorkflow", "appID", input.AppID)

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

	// Step 1: Delete DNS record
	err := workflow.ExecuteActivity(ctx, activities.DeleteDNSRecord, input.AppID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to delete DNS record", "error", err)
		return DeleteAppWorkflowResult{
			AppID:        input.AppID,
			Status:       "failed",
			ErrorMessage: err.Error(),
		}, err
	}

	// Step 2: Delete from Coolify (if deployed)
	if input.CoolifyAppUUID != "" {
		deleteInput := DeleteAppFromCoolifyInput{
			AppID:          input.AppID,
			CoolifyAppUUID: input.CoolifyAppUUID,
		}
		err = workflow.ExecuteActivity(ctx, activities.DeleteAppFromCoolify, deleteInput).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to delete app from Coolify", "error", err)
			return DeleteAppWorkflowResult{
				AppID:        input.AppID,
				Status:       "failed",
				ErrorMessage: err.Error(),
			}, err
		}
	}

	// Step 3: Soft delete from database
	err = workflow.ExecuteActivity(ctx, activities.SoftDeleteApp, input.AppID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to soft delete app", "error", err)
		return DeleteAppWorkflowResult{
			AppID:        input.AppID,
			Status:       "failed",
			ErrorMessage: err.Error(),
		}, err
	}

	logger.Info("DeleteAppWorkflow completed successfully", "appID", input.AppID)

	return DeleteAppWorkflowResult{
		AppID:  input.AppID,
		Status: "deleted",
	}, nil
}
