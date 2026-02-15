package k8sdeployments

import "go.temporal.io/sdk/worker"

func RegisterWorkflowsAndActivities(w worker.Worker, activities *Activities) {
	w.RegisterWorkflow(CreateServiceWorkflow)
	w.RegisterWorkflow(RedeployServiceWorkflow)
	w.RegisterWorkflow(DeleteServiceWorkflow)
	w.RegisterWorkflow(BuildServiceWorkflow)

	w.RegisterActivity(activities.CloneRepo)
	w.RegisterActivity(activities.ResolveImageRef)
	w.RegisterActivity(activities.ResolveBuildContext)
	w.RegisterActivity(activities.ImageExists)
	w.RegisterActivity(activities.RailpackBuild)
	w.RegisterActivity(activities.RailpackStaticBuild)
	w.RegisterActivity(activities.DockerfileBuild)
	w.RegisterActivity(activities.StaticBuild)
	w.RegisterActivity(activities.CleanupSource)
	w.RegisterActivity(activities.Deploy)
	w.RegisterActivity(activities.WaitForRollout)
	w.RegisterActivity(activities.DeleteService)
	w.RegisterActivity(activities.UpdateDeploymentBuilding)
	w.RegisterActivity(activities.UpdateDeploymentDeploying)
	w.RegisterActivity(activities.MarkDeploymentActive)
	w.RegisterActivity(activities.MarkDeploymentFailed)
	w.RegisterActivity(activities.UpdateDeploymentBuildProgress)
	w.RegisterActivity(activities.SoftDeleteService)
}
