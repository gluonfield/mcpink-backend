package dnsverify

import (
	"crypto/rand"
	"encoding/hex"
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

func VerifyTXT(domain, expectedToken string) (bool, error) {
	host := "_dp-verify." + NormalizeDomain(domain)
	records, err := net.LookupTXT(host)
	if err != nil {
		return false, fmt.Errorf("TXT lookup failed for %s: %w", host, err)
	}

	needle := "dp-verify=" + expectedToken
	for _, r := range records {
		if strings.TrimSpace(r) == needle {
			return true, nil
		}
	}
	return false, nil
}

func GenerateVerificationToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
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

func DNSInstructions(domain, cnameTarget, verificationToken string) string {
	return fmt.Sprintf(
		"Add the following DNS records for your domain:\n\n"+
			"1. Ownership verification (TXT):\n"+
			"   Host: _dp-verify.%s\n"+
			"   Type: TXT\n"+
			"   Value: dp-verify=%s\n\n"+
			"2. Routing (CNAME):\n"+
			"   Host: %s\n"+
			"   Type: CNAME\n"+
			"   Value: %s\n\n"+
			"Important: If using Cloudflare, set the CNAME record to DNS-only (gray cloud), not Proxied.\n"+
			"After configuring both DNS records, call verify_custom_domain to activate it.",
		domain, verificationToken, domain, cnameTarget,
	)
}
