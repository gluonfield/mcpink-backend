package internalgit

import "time"

type Config struct {
	BaseURL               string
	AdminToken            string
	UserPrefix            string
	WebhookSecret         string
	SSHURL                string
	SSHPort               int
	CoolifyPrivateKeyUUID string
	DeployPublicKey       string
}

const (
	DefaultTimeout       = 30 * time.Second
	DefaultTokenDuration = 1 * time.Hour

	// DeployBotUsername is the Gitea user that Coolify authenticates as via SSH.
	// This user must be created manually and have the deploy SSH key added to its account.
	// Repos are made accessible to Coolify by adding this user as a read-only collaborator.
	DeployBotUsername = "coolify-deploy"
)
