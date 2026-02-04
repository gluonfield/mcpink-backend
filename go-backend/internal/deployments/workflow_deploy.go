package deployments

import (
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func DeployToCoolifyWorkflow(ctx workflow.Context, input DeployWorkflowInput) (DeployWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DeployWorkflow",
		"repo", input.Repo,
		"branch", input.Branch)

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

	appID := input.AppID

	// Determine git provider
	if input.GitProvider == "" {
		return DeployWorkflowResult{}, fmt.Errorf("git provider is required")
	}

	// Step 1: Create app record in database
	envVarsJSON, _ := json.Marshal(input.EnvVars)
	workflowInfo := workflow.GetInfo(ctx)
	createRecordInput := CreateAppRecordInput{
		AppID:         appID,
		UserID:        input.UserID,
		ProjectID:     input.ProjectID,
		WorkflowID:    workflowInfo.WorkflowExecution.ID,
		WorkflowRunID: workflowInfo.WorkflowExecution.RunID,
		Repo:          input.Repo,
		Branch:        input.Branch,
		Name:          input.Name,
		BuildPack:     input.BuildPack,
		Port:          input.Port,
		EnvVars:       envVarsJSON,
		GitProvider:   input.GitProvider,
	}

	err := workflow.ExecuteActivity(ctx, activities.CreateAppRecord, createRecordInput).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create app record", "error", err)
		return DeployWorkflowResult{
			AppID:        appID,
			Status:       string(BuildStatusFailed),
			ErrorMessage: err.Error(),
		}, err
	}

	logger.Info("Created app record", "appID", appID)

	// Step 2: Create app in Coolify (branch based on git provider)
	var createAppResult CoolifyAppResult

	if input.GitProvider == "gitea" {
		// Internal git (Gitea) flow
		internalGitInput := InternalGitAppInput{
			AppID:          appID,
			PrivateKeyUUID: input.PrivateKeyUUID,
			SSHCloneURL:    input.SSHCloneURL,
			Branch:         input.Branch,
			Name:           input.Name,
			BuildPack:      input.BuildPack,
			Port:           input.Port,
			Memory:         input.Memory,
			CPU:            input.CPU,
			InstallCommand: input.InstallCommand,
			BuildCommand:   input.BuildCommand,
			StartCommand:   input.StartCommand,
		}

		err = workflow.ExecuteActivity(ctx, activities.CreateAppFromInternalGit, internalGitInput).Get(ctx, &createAppResult)
		if err != nil {
			logger.Error("Failed to create app from internal git", "error", err)
			_ = markAppFailed(ctx, activities, appID, err.Error())
			return DeployWorkflowResult{
				AppID:        appID,
				Status:       string(BuildStatusFailed),
				ErrorMessage: err.Error(),
			}, err
		}
	} else {
		// GitHub flow (default)
		createAppInput := CoolifyAppInput{
			AppID:          appID,
			GitHubAppUUID:  input.GitHubAppUUID,
			Repo:           input.Repo,
			Branch:         input.Branch,
			Name:           input.Name,
			BuildPack:      input.BuildPack,
			Port:           input.Port,
			Memory:         input.Memory,
			CPU:            input.CPU,
			InstallCommand: input.InstallCommand,
			BuildCommand:   input.BuildCommand,
			StartCommand:   input.StartCommand,
		}

		err = workflow.ExecuteActivity(ctx, activities.CreateAppFromPrivateGithub, createAppInput).Get(ctx, &createAppResult)
		if err != nil {
			logger.Error("Failed to create app from GitHub", "error", err)
			_ = markAppFailed(ctx, activities, appID, err.Error())
			return DeployWorkflowResult{
				AppID:        appID,
				Status:       string(BuildStatusFailed),
				ErrorMessage: err.Error(),
			}, err
		}
	}

	// Step 2.5: Create DNS record (non-blocking - continues even if it fails)
	dnsInput := CreateDNSRecordInput{
		AppID:      appID,
		AppName:    input.Name,
		ServerUUID: createAppResult.ServerID,
	}
	var dnsResult CreateDNSRecordResult
	dnsErr := workflow.ExecuteActivity(ctx, activities.CreateDNSRecord, dnsInput).Get(ctx, &dnsResult)
	if dnsErr != nil {
		logger.Warn("Failed to create DNS record, will use Coolify-generated domain",
			"appID", appID,
			"error", dnsErr)
	}

	// Step 2.6: Update Coolify with our FQDN (if DNS was created)
	if dnsResult.FQDN != "" {
		updateDomainInput := UpdateCoolifyDomainInput{
			CoolifyAppUUID: createAppResult.CoolifyAppUUID,
			Domain:         dnsResult.FQDN,
		}
		domainErr := workflow.ExecuteActivity(ctx, activities.UpdateCoolifyDomain, updateDomainInput).Get(ctx, nil)
		if domainErr != nil {
			logger.Warn("Failed to update Coolify domain, will use auto-generated",
				"appID", appID,
				"error", domainErr)
		}
	}

	// Step 3: Bulk update environment variables (if any)
	if len(input.EnvVars) > 0 {
		bulkUpdateInput := BulkUpdateEnvsInput{
			CoolifyAppUUID: createAppResult.CoolifyAppUUID,
			EnvVars:        input.EnvVars,
		}

		err = workflow.ExecuteActivity(ctx, activities.BulkUpdateEnvs, bulkUpdateInput).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to set environment variables", "error", err)
			_ = markAppFailed(ctx, activities, appID, err.Error())
			return DeployWorkflowResult{
				AppID:        appID,
				AppUUID:      createAppResult.CoolifyAppUUID,
				Status:       string(BuildStatusFailed),
				ErrorMessage: err.Error(),
			}, err
		}
	}

	// Step 4: Start the app (triggers deployment)
	startAppInput := StartAppInput{
		CoolifyAppUUID: createAppResult.CoolifyAppUUID,
	}

	var startAppResult StartAppResult
	err = workflow.ExecuteActivity(ctx, activities.StartApp, startAppInput).Get(ctx, &startAppResult)
	if err != nil {
		logger.Error("Failed to start app", "error", err)
		_ = markAppFailed(ctx, activities, appID, err.Error())
		return DeployWorkflowResult{
			AppID:        appID,
			AppUUID:      createAppResult.CoolifyAppUUID,
			Status:       string(BuildStatusFailed),
			ErrorMessage: err.Error(),
		}, err
	}

	logger.Info("Deployment triggered", "deploymentUUID", startAppResult.DeploymentUUID)

	// Step 5: Wait for app to be running (polls for 3min, retries 3x = 9min total)
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
		AppID:          appID,
		CoolifyAppUUID: createAppResult.CoolifyAppUUID,
		DeploymentUUID: startAppResult.DeploymentUUID,
	}

	var waitResult WaitForRunningResult
	err = workflow.ExecuteActivity(waitCtx, activities.WaitForRunning, waitInput).Get(ctx, &waitResult)
	if err != nil {
		logger.Error("App failed to reach running state", "error", err)
		_ = markAppFailed(ctx, activities, appID, err.Error())
		return DeployWorkflowResult{
			AppID:        appID,
			AppUUID:      createAppResult.CoolifyAppUUID,
			Status:       string(BuildStatusFailed),
			ErrorMessage: err.Error(),
		}, err
	}

	logger.Info("DeployWorkflow completed successfully",
		"appID", appID,
		"appUUID", createAppResult.CoolifyAppUUID,
		"fqdn", waitResult.FQDN)

	return DeployWorkflowResult{
		AppID:   appID,
		AppUUID: createAppResult.CoolifyAppUUID,
		FQDN:    waitResult.FQDN,
		Status:  string(BuildStatusSuccess),
	}, nil
}

func markAppFailed(ctx workflow.Context, activities *Activities, appID, errorMsg string) error {
	failedInput := UpdateAppFailedInput{
		AppID:        appID,
		ErrorMessage: errorMsg,
	}
	return workflow.ExecuteActivity(ctx, activities.UpdateAppFailed, failedInput).Get(ctx, nil)
}
