package k8sdeployments

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/augustdev/autoclip/internal/storage/pg/generated/apps"
)

func mustMarshalBuildConfig(bc BuildConfig) []byte {
	b, _ := json.Marshal(bc)
	return b
}

func TestBuildImageTag(t *testing.T) {
	commit := "0123456789abcdef"

	tests := []struct {
		name string
		app  apps.App
		want string
	}{
		{
			name: "railpack without publish directory keeps legacy commit tag",
			app: apps.App{
				BuildPack: "railpack",
			},
			want: commit,
		},
		{
			name: "railpack with publish directory includes config hash",
			app: apps.App{
				BuildPack:   "railpack",
				BuildConfig: mustMarshalBuildConfig(BuildConfig{PublishDirectory: "dist"}),
			},
			want: commit + "-",
		},
		{
			name: "dockerfile includes config hash",
			app: apps.App{
				BuildPack: "dockerfile",
			},
			want: commit + "-",
		},
		{
			name: "railpack with root_directory includes config hash",
			app: apps.App{
				BuildPack:   "railpack",
				BuildConfig: mustMarshalBuildConfig(BuildConfig{RootDirectory: "frontend"}),
			},
			want: commit + "-",
		},
		{
			name: "dockerfile with dockerfile_path includes config hash",
			app: apps.App{
				BuildPack:   "dockerfile",
				BuildConfig: mustMarshalBuildConfig(BuildConfig{DockerfilePath: "worker.Dockerfile"}),
			},
			want: commit + "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildImageTag(commit, tt.app)
			if tt.want == commit {
				if got != commit {
					t.Fatalf("buildImageTag() = %q, want %q", got, commit)
				}
				return
			}

			if !strings.HasPrefix(got, tt.want) {
				t.Fatalf("buildImageTag() = %q, expected prefix %q", got, tt.want)
			}

			suffix := strings.TrimPrefix(got, tt.want)
			if len(suffix) != 8 {
				t.Fatalf("buildImageTag() hash length = %d, want 8", len(suffix))
			}
		})
	}
}

func TestBuildImageTag_ConfigDrivesTag(t *testing.T) {
	commit := "0123456789abcdef"

	railpackDist := apps.App{
		BuildPack:   "railpack",
		BuildConfig: mustMarshalBuildConfig(BuildConfig{PublishDirectory: "dist"}),
	}
	railpackPublic := apps.App{
		BuildPack:   "railpack",
		BuildConfig: mustMarshalBuildConfig(BuildConfig{PublishDirectory: "public"}),
	}
	dockerfile := apps.App{BuildPack: "dockerfile"}

	distTag := buildImageTag(commit, railpackDist)
	if distTag != buildImageTag(commit, railpackDist) {
		t.Fatalf("buildImageTag() is not deterministic for identical config")
	}

	if distTag == buildImageTag(commit, railpackPublic) {
		t.Fatalf("buildImageTag() should differ when publish_directory changes")
	}

	if distTag == buildImageTag(commit, dockerfile) {
		t.Fatalf("buildImageTag() should differ when build_pack changes")
	}
}

func TestBuildImageTag_RootDirectoryDrivesTag(t *testing.T) {
	commit := "0123456789abcdef"

	frontend := apps.App{
		BuildPack:   "railpack",
		BuildConfig: mustMarshalBuildConfig(BuildConfig{RootDirectory: "frontend"}),
	}
	backend := apps.App{
		BuildPack:   "railpack",
		BuildConfig: mustMarshalBuildConfig(BuildConfig{RootDirectory: "backend"}),
	}
	plain := apps.App{BuildPack: "railpack"}

	frontendTag := buildImageTag(commit, frontend)
	backendTag := buildImageTag(commit, backend)
	plainTag := buildImageTag(commit, plain)

	if frontendTag == backendTag {
		t.Fatalf("different root_directory should produce different tags")
	}
	if frontendTag == plainTag {
		t.Fatalf("root_directory vs no root_directory should produce different tags")
	}
}

func TestBuildImageTag_DockerfilePathDrivesTag(t *testing.T) {
	commit := "0123456789abcdef"

	server := apps.App{
		BuildPack:   "dockerfile",
		BuildConfig: mustMarshalBuildConfig(BuildConfig{DockerfilePath: "server.Dockerfile"}),
	}
	worker := apps.App{
		BuildPack:   "dockerfile",
		BuildConfig: mustMarshalBuildConfig(BuildConfig{DockerfilePath: "worker.Dockerfile"}),
	}
	defaultDF := apps.App{BuildPack: "dockerfile"}

	if buildImageTag(commit, server) == buildImageTag(commit, worker) {
		t.Fatalf("different dockerfile_path should produce different tags")
	}
	if buildImageTag(commit, server) == buildImageTag(commit, defaultDF) {
		t.Fatalf("dockerfile_path vs default should produce different tags")
	}
}
