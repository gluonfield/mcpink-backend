package dnsverify

import (
	"fmt"
	"net"
	"strings"
)

func VerifyCNAME(domain, expectedTarget string) (bool, error) {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return false, fmt.Errorf("CNAME lookup failed for %s: %w", domain, err)
	}

	normalized := NormalizeDomain(cname)
	expected := NormalizeDomain(expectedTarget)

	return normalized == expected, nil
}

func NormalizeDomain(d string) string {
	d = strings.TrimSpace(d)
	d = strings.ToLower(d)
	d = strings.TrimSuffix(d, ".")
	return d
}

func ValidateCustomDomain(domain, platformDomain string) error {
	domain = NormalizeDomain(domain)

	if domain == "" {
		return fmt.Errorf("domain is required")
	}

	if strings.Contains(domain, "*") {
		return fmt.Errorf("wildcard domains are not supported")
	}

	if strings.HasSuffix(domain, "."+NormalizeDomain(platformDomain)) || domain == NormalizeDomain(platformDomain) {
		return fmt.Errorf("cannot use a %s subdomain as a custom domain", platformDomain)
	}

	if !strings.Contains(domain, ".") {
		return fmt.Errorf("apex domains are not supported; use a subdomain (e.g. app.%s)", domain)
	}

	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid domain format")
		}
	}

	return nil
}

func DNSInstructions(domain, cnameTarget string) string {
	return fmt.Sprintf(
		"Add a CNAME record for your domain:\n\n"+
			"  Host: %s\n"+
			"  Type: CNAME\n"+
			"  Value: %s\n\n"+
			"Important: If using Cloudflare, set the record to DNS-only (gray cloud), not Proxied.\n"+
			"After configuring DNS, call verify_custom_domain to activate it.",
		domain, cnameTarget,
	)
}
