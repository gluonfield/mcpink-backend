package deployments

type EnvVar struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime bool   `json:"is_build_time,omitempty"`
}

type BuildStatus string

const (
	BuildStatusQueued   BuildStatus = "queued"
	BuildStatusBuilding BuildStatus = "building"
	BuildStatusSuccess  BuildStatus = "success"
	BuildStatusFailed   BuildStatus = "failed"
)

type RuntimeStatus string

const (
	RuntimeStatusRunning RuntimeStatus = "running"
	RuntimeStatusStopped RuntimeStatus = "stopped"
	RuntimeStatusExited  RuntimeStatus = "exited"
)
