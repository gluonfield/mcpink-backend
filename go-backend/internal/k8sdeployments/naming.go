package k8sdeployments

import (
	"fmt"
	"regexp"
	"strings"
)

var nonAlphanumDash = regexp.MustCompile(`[^a-z0-9-]`)

func sanitizeDNS(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	s = nonAlphanumDash.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	if len(s) > 63 {
		s = s[:63]
		s = strings.TrimRight(s, "-")
	}
	return s
}

func NamespaceName(githubUsername, projectRef string) string {
	return fmt.Sprintf("dp-%s-%s", sanitizeDNS(githubUsername), sanitizeDNS(projectRef))
}

func ServiceName(appName string) string {
	return sanitizeDNS(appName)
}
