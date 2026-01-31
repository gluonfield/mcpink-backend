package githubapp

type Config struct {
	AppID      int64  `mapstructure:"appid"`
	PrivateKey string `mapstructure:"privatekey"`
}
