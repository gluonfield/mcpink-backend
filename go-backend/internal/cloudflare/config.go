package cloudflare

type Config struct {
	APIToken    string
	ZoneID      string
	BaseDomain  string
	CNAMETarget string
}
