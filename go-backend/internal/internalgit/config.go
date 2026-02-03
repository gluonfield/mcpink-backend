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
	DeployBotUsername     string
}

const (
	DefaultTimeout       = 30 * time.Second
	DefaultTokenDuration = 1 * time.Hour
)
