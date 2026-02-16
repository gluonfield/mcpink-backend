package internalgit

import "time"

type Config struct {
	PublicGitURL string // e.g. https://git.ml.ink
}

const (
	DefaultTimeout       = 30 * time.Second
	DefaultTokenDuration = 1 * time.Hour
)
