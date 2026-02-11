package k8sdeployments

import (
	"strings"
	"testing"

	"github.com/augustdev/autoclip/internal/storage/pg/generated/apps"
)

func TestBuildImageTag(t *testing.T) {
	commit := "0123456789abcdef"
	dist := "dist"

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
				BuildPack:        "railpack",
				PublishDirectory: &dist,
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
	dist := "dist"
	public := "public"

	railpackDist := apps.App{BuildPack: "railpack", PublishDirectory: &dist}
	railpackPublic := apps.App{BuildPack: "railpack", PublishDirectory: &public}
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
