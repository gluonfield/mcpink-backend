package k8sdeployments

import (
	"os"
	"strconv"
	"strings"
)

func effectiveAppPort(buildPack, appPort, publishDir string) string {
	port := strings.TrimSpace(appPort)
	if port == "" {
		port = "3000"
	}

	switch buildPack {
	case "static":
		port = "8080"
	case "railpack", "nixpacks":
		if strings.TrimSpace(publishDir) != "" {
			port = "8080"
		}
	}

	return port
}

// extractPortFromDockerfile parses the last EXPOSE directive from a Dockerfile.
// In multi-stage builds the final stage's EXPOSE is the relevant one.
func extractPortFromDockerfile(dockerfilePath string) string {
	data, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return ""
	}
	var lastPort string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trimmed), "EXPOSE") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				portStr := strings.Split(parts[1], "/")[0] // strip /tcp, /udp
				if _, err := strconv.Atoi(portStr); err == nil {
					lastPort = portStr
				}
			}
		}
	}
	return lastPort
}
