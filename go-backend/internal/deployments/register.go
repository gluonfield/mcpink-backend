package deployments

import (
	"go.temporal.io/sdk/worker"
)

func RegisterWorkflowsAndActivities(w worker.Worker, activities *Activities) {
	w.RegisterWorkflow(DeployToCoolifyWorkflow)
	w.RegisterWorkflow(RedeployToCoolifyWorkflow)
	w.RegisterWorkflow(DeleteAppWorkflow)
	w.RegisterActivity(activities.CreateAppRecord)
	w.RegisterActivity(activities.CreateAppFromPrivateGithub)
	w.RegisterActivity(activities.CreateAppFromInternalGit)
	w.RegisterActivity(activities.BulkUpdateEnvs)
	w.RegisterActivity(activities.StartApp)
	w.RegisterActivity(activities.WaitForRunning)
	w.RegisterActivity(activities.UpdateAppFailed)
	w.RegisterActivity(activities.DeployApp)
	w.RegisterActivity(activities.MarkAppBuilding)
	w.RegisterActivity(activities.CreateDNSRecord)
	w.RegisterActivity(activities.UpdateCoolifyDomain)
	w.RegisterActivity(activities.DeleteDNSRecord)
	w.RegisterActivity(activities.DeleteAppFromCoolify)
	w.RegisterActivity(activities.SoftDeleteApp)
}
