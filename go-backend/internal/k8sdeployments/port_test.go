package k8sdeployments

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEffectiveAppPort(t *testing.T) {
	tests := []struct {
		name       string
		buildPack  string
		appPort    string
		publishDir string
		want       string
	}{
		{
			name:      "default railpack port",
			buildPack: "railpack",
			want:      "3000",
		},
		{
			name:      "respects explicit port",
			buildPack: "railpack",
			appPort:   "4200",
			want:      "4200",
		},
		{
			name:       "railpack publish directory forces 8080",
			buildPack:  "railpack",
			appPort:    "3000",
			publishDir: "dist",
			want:       "8080",
		},
		{
			name:      "static buildpack forces 8080",
			buildPack: "static",
			appPort:   "80",
			want:      "8080",
		},
		{
			name:       "nixpacks with publish directory forces 8080",
			buildPack:  "nixpacks",
			appPort:    "3000",
			publishDir: "dist",
			want:       "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveAppPort(tt.buildPack, tt.appPort, tt.publishDir)
			if got != tt.want {
				t.Fatalf("effectiveAppPort(%q, %q, %q) = %q, want %q", tt.buildPack, tt.appPort, tt.publishDir, got, tt.want)
			}
		})
	}
}

func TestExtractPortFromDockerfile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple EXPOSE",
			content: "FROM python:3.11\nEXPOSE 5000\nCMD [\"python\", \"app.py\"]",
			want:    "5000",
		},
		{
			name:    "multi-stage picks last EXPOSE",
			content: "FROM node:18 AS build\nEXPOSE 3000\nRUN npm build\nFROM nginx:alpine\nEXPOSE 8080\nCOPY --from=build /app/dist /usr/share/nginx/html",
			want:    "8080",
		},
		{
			name:    "EXPOSE with /tcp protocol",
			content: "FROM python:3.11\nEXPOSE 8080/tcp\nCMD [\"python\", \"app.py\"]",
			want:    "8080",
		},
		{
			name:    "no EXPOSE",
			content: "FROM python:3.11\nCMD [\"python\", \"app.py\"]",
			want:    "",
		},
		{
			name:    "invalid EXPOSE non-numeric",
			content: "FROM python:3.11\nEXPOSE abc\nCMD [\"python\", \"app.py\"]",
			want:    "",
		},
		{
			name:    "EXPOSE with no port value",
			content: "FROM python:3.11\nEXPOSE\nCMD [\"python\", \"app.py\"]",
			want:    "",
		},
		{
			name:    "lowercase expose ignored",
			content: "FROM python:3.11\nexpose 5000\nCMD [\"python\", \"app.py\"]",
			want:    "5000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "Dockerfile")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			got := extractPortFromDockerfile(path)
			if got != tt.want {
				t.Fatalf("extractPortFromDockerfile() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("file does not exist", func(t *testing.T) {
		got := extractPortFromDockerfile("/nonexistent/Dockerfile")
		if got != "" {
			t.Fatalf("extractPortFromDockerfile(nonexistent) = %q, want empty", got)
		}
	})
}
