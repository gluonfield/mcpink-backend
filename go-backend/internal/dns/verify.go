package dns

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// siblingTXTHost returns the TXT verification hostname placed above the
// delegation cut so it remains queryable even after NS records are set.
// Example: zone "dogs.breacher.org" → "dogs._dp-verify.breacher.org"
func siblingTXTHost(zone string) string {
	zone = NormalizeDomain(zone)
	idx := strings.Index(zone, ".")
	if idx < 0 {
		return "_dp-verify." + zone
	}
	return zone[:idx] + "._dp-verify." + zone[idx+1:]
}

func VerifyTXT(zone, expectedToken string) (bool, error) {
	host := siblingTXTHost(zone)
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

// VerifyNS checks that the zone's NS records resolve to our nameservers.
// This relies on the zone already existing in PowerDNS (created before
// NS verification) so that recursive resolvers can follow the delegation
// chain without getting REFUSED.
func VerifyNS(zone string, expectedNS []string) (bool, error) {
	nsRecords, err := net.LookupNS(NormalizeDomain(zone))
	if err != nil {
		return false, fmt.Errorf("NS lookup failed for %s: %w", zone, err)
	}

	found := make(map[string]bool)
	for _, ns := range nsRecords {
		found[NormalizeDomain(ns.Host)] = true
	}

	for _, expected := range expectedNS {
		if !found[NormalizeDomain(expected)] {
			return false, nil
		}
	}
	return true, nil
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

func ValidateDelegatedZone(zone, platformDomain string) error {
	zone = NormalizeDomain(zone)

	if zone == "" {
		return fmt.Errorf("zone is required")
	}

	if strings.Contains(zone, "*") {
		return fmt.Errorf("wildcard zones are not supported")
	}

	if strings.HasSuffix(zone, "."+NormalizeDomain(platformDomain)) || zone == NormalizeDomain(platformDomain) {
		return fmt.Errorf("cannot delegate a %s subdomain", platformDomain)
	}

	if !strings.Contains(zone, ".") {
		return fmt.Errorf("apex domains cannot be delegated; use a subdomain (e.g. apps.%s)", zone)
	}

	parts := strings.Split(zone, ".")
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid zone format")
		}
	}

	return nil
}

func DelegationInstructions(zone, token string, nameservers []string) string {
	txtHost := siblingTXTHost(zone)
	nsList := ""
	for _, ns := range nameservers {
		nsList += fmt.Sprintf("   %s  NS  %s\n", zone, ns)
	}

	return fmt.Sprintf(
		"Add all records at your DNS provider, then call verify_delegation:\n\n"+
			"   Host: %s\n"+
			"   Type: TXT\n"+
			"   Value: dp-verify=%s\n\n"+
			"%s",
		txtHost, token, nsList,
	)
}
