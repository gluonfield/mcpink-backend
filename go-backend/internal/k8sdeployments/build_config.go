package k8sdeployments

import "encoding/json"

type BuildConfig struct {
	RootDirectory    string `json:"root_directory,omitempty"`
	DockerfilePath   string `json:"dockerfile_path,omitempty"`
	PublishDirectory string `json:"publish_directory,omitempty"`
}

func parseBuildConfig(raw []byte) BuildConfig {
	var bc BuildConfig
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &bc)
	}
	return bc
}
