package internalgit

import "time"

// GiteaUser represents a Gitea user
type GiteaUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	IsAdmin   bool   `json:"is_admin"`
}

// CreateUserRequest is the request to create a new Gitea user
type CreateUserRequest struct {
	Username           string `json:"username"`
	Email              string `json:"email"`
	Password           string `json:"password"`
	MustChangePassword bool   `json:"must_change_password"`
	Visibility         string `json:"visibility,omitempty"`
}

// GiteaRepo represents a Gitea repository
type GiteaRepo struct {
	ID            int64     `json:"id"`
	Owner         GiteaUser `json:"owner"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	Private       bool      `json:"private"`
	Empty         bool      `json:"empty"`
	DefaultBranch string    `json:"default_branch"`
	SSHURL        string    `json:"ssh_url"`
	CloneURL      string    `json:"clone_url"`
	HTMLURL       string    `json:"html_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateRepoRequest is the request to create a new repository
type CreateRepoRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Private       bool   `json:"private"`
	AutoInit      bool   `json:"auto_init"`
	DefaultBranch string `json:"default_branch,omitempty"`
	Readme        string `json:"readme,omitempty"`
}

// GiteaAccessToken represents a Gitea access token
type GiteaAccessToken struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	SHA1           string   `json:"sha1"`
	TokenLastEight string   `json:"token_last_eight"`
	Scopes         []string `json:"scopes,omitempty"`
}

// CreateAccessTokenRequest is the request to create a new access token
type CreateAccessTokenRequest struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes,omitempty"`
}

// GiteaWebhook represents a Gitea webhook
type GiteaWebhook struct {
	ID          int64           `json:"id"`
	Type        string          `json:"type"`
	URL         string          `json:"url,omitempty"`
	Config      WebhookConfig   `json:"config"`
	Events      []string        `json:"events"`
	Active      bool            `json:"active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// WebhookConfig contains the webhook configuration
type WebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Secret      string `json:"secret,omitempty"`
}

// CreateWebhookRequest is the request to create a new webhook
type CreateWebhookRequest struct {
	Type                string        `json:"type"`
	Config              WebhookConfig `json:"config"`
	Events              []string      `json:"events"`
	BranchFilter        string        `json:"branch_filter,omitempty"`
	Active              bool          `json:"active"`
	AuthorizationHeader string        `json:"authorization_header,omitempty"`
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

// CreateRepoResult is returned after creating a repo
type CreateRepoResult struct {
	Repo      string `json:"repo"`
	GitRemote string `json:"git_remote"`
	Message   string `json:"message"`
}

// GetPushTokenResult is returned when getting a push token
type GetPushTokenResult struct {
	GitRemote string `json:"git_remote"`
	ExpiresAt string `json:"expires_at"`
}
