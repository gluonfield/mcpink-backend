package github_oauth

type Config struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	Scopes       []string
}
