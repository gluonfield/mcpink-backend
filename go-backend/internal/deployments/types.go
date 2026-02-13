package deployments

type EnvVar struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime bool   `json:"is_build_time,omitempty"`
}
