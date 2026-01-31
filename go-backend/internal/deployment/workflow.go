package deployment

import "go.temporal.io/sdk/workflow"

type DeployWorkflowInput struct {
}

type DeployWorkflowResults struct {
}

func DeployWorkflow(ctx workflow.Context, input DeployWorkflowInput) (DeployWorkflowResults, error) {
	return DeployWorkflowResults{}, nil
}
