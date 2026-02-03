package internalgit

// CreateRepoResult is returned after creating a repo
type CreateRepoResult struct {
	Repo      string `json:"repo"`
	GitRemote string `json:"git_remote"`
	ExpiresAt string `json:"expires_at"`
	Message   string `json:"message"`
}

// GetPushTokenResult is returned when getting a push token
type GetPushTokenResult struct {
	GitRemote string `json:"git_remote"`
	ExpiresAt string `json:"expires_at"`
}

// WebhookPayload represents the incoming webhook payload from Gitea
type WebhookPayload struct {
	Ref        string            `json:"ref"`
	Before     string            `json:"before"`
	After      string            `json:"after"`
	CompareURL string            `json:"compare_url"`
	Repository WebhookRepository `json:"repository"`
	Pusher     WebhookUser       `json:"pusher"`
	Sender     WebhookUser       `json:"sender"`
}

// WebhookRepository represents repository info in webhook payload
type WebhookRepository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	SSHURL   string `json:"ssh_url"`
	CloneURL string `json:"clone_url"`
	Private  bool   `json:"private"`
}

// WebhookUser represents user info in webhook payload
type WebhookUser struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}
