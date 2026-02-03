package internalgit

import "time"

type Config struct {
	Enabled               bool
	BaseURL               string
	AdminToken            string
	UserPrefix            string
	WebhookSecret         string
	SSHURL                string
	CoolifyPrivateKeyUUID string
}

func (c Config) IsEnabled() bool {
	return c.Enabled && c.BaseURL != "" && c.AdminToken != ""
}

const (
	DefaultTimeout       = 30 * time.Second
	DefaultTokenDuration = 1 * time.Hour
)
