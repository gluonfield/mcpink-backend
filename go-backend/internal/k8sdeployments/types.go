package k8sdeployments

const TaskQueue = "k8s-native"

// --- Parent Workflow types ---

type CreateServiceWorkflowInput struct {
	ServiceID      string
	Repo           string
	Branch         string
	GitProvider    string
	InstallationID int64
	CommitSHA      string
	AppsDomain     string
}

type CreateServiceWorkflowResult struct {
	ServiceID    string
	Status       string
	URL          string
	CommitSHA    string
	ErrorMessage string
}

type RedeployServiceWorkflowInput struct {
	ServiceID      string
	Repo           string
	Branch         string
	GitProvider    string
	InstallationID int64
	CommitSHA      string
	AppsDomain     string
}

type RedeployServiceWorkflowResult struct {
	ServiceID    string
	Status       string
	URL          string
	CommitSHA    string
	ErrorMessage string
}

type DeleteServiceWorkflowInput struct {
	ServiceID string
	Namespace string
	Name      string
}

type DeleteServiceWorkflowResult struct {
	ServiceID    string
	Status       string
	ErrorMessage string
}

// --- BuildServiceWorkflow (child) types ---

type BuildServiceWorkflowInput struct {
	ServiceID      string
	Repo           string
	Branch         string
	GitProvider    string
	InstallationID int64
	CommitSHA      string
}

type BuildServiceWorkflowResult struct {
	ImageRef  string
	CommitSHA string
}

// --- Activity types ---

type CloneRepoInput struct {
	ServiceID      string
	Repo           string
	Branch         string
	GitProvider    string
	InstallationID int64
	CommitSHA      string
}

type CloneRepoResult struct {
	SourcePath string
	CommitSHA  string
}

type ResolveBuildContextInput struct {
	ServiceID  string
	SourcePath string
	CommitSHA  string
}

type ResolveBuildContextResult struct {
	BuildPack string
	ImageRef  string
	Namespace string
	Name      string
	Port      string
	EnvVars   map[string]string
}

type BuildImageInput struct {
	SourcePath string
	ImageRef   string
	BuildPack  string
	Name       string
	Namespace  string
	EnvVars    map[string]string
}

type BuildImageResult struct {
	ImageRef string
}

type DeployInput struct {
	ServiceID  string
	ImageRef   string
	CommitSHA  string
	AppsDomain string
}

type DeployResult struct {
	Namespace      string
	DeploymentName string
	URL            string
}

type WaitForRolloutInput struct {
	Namespace      string
	DeploymentName string
}

type WaitForRolloutResult struct {
	Status string
}

type DeleteServiceInput struct {
	ServiceID string
	Namespace string
	Name      string
}

type DeleteServiceResult struct {
	Status string
}
