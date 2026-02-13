package k8sdeployments

import "go.temporal.io/sdk/worker"

func RegisterWorkflowsAndActivities(w worker.Worker, activities *Activities) {
	w.RegisterWorkflow(CreateServiceWorkflow)
	w.RegisterWorkflow(RedeployServiceWorkflow)
	w.RegisterWorkflow(DeleteServiceWorkflow)
	w.RegisterWorkflow(BuildServiceWorkflow)
	w.RegisterWorkflow(AttachCustomDomainWorkflow)
	w.RegisterWorkflow(DetachCustomDomainWorkflow)
	w.RegisterActivity(activities)
}
