package dns

import "go.temporal.io/sdk/worker"

func RegisterWorkflowsAndActivities(w worker.Worker, activities *Activities) {
	w.RegisterWorkflow(ActivateZoneWorkflow)
	w.RegisterWorkflow(DeactivateZoneWorkflow)
	w.RegisterWorkflow(AttachSubdomainWorkflow)
	w.RegisterWorkflow(DetachSubdomainWorkflow)

	w.RegisterActivity(activities.CreateZone)
	w.RegisterActivity(activities.DeleteZone)
	w.RegisterActivity(activities.UpsertRecord)
	w.RegisterActivity(activities.DeleteRecord)
	w.RegisterActivity(activities.ApplyWildcardCert)
	w.RegisterActivity(activities.WaitForCertReady)
	w.RegisterActivity(activities.UpdateZoneStatus)
	w.RegisterActivity(activities.DeleteCertificate)
	w.RegisterActivity(activities.ApplySubdomainIngress)
	w.RegisterActivity(activities.DeleteIngress)
}
