package mcpserver

import (
	"testing"

	"github.com/augustdev/autoclip/internal/storage/pg/generated/users"
)

func TestNormalizeCreateAppRepo(t *testing.T) {
	u := &users.User{GithubUsername: "gluonfield"}

	cases := []struct {
		name     string
		input    CreateAppInput
		wantHost string
		wantRepo string
		wantErr  bool
	}{
		{
			name:     "defaults to ml.ink",
			input:    CreateAppInput{Repo: "exp20"},
			wantHost: "ml.ink",
			wantRepo: "ml.ink/gluonfield/exp20",
		},
		{
			name:     "github.com host expands repo name",
			input:    CreateAppInput{Repo: "exp20", Host: "github.com"},
			wantHost: "github.com",
			wantRepo: "gluonfield/exp20",
		},
		{
			name:    "rejects owner/repo format",
			input:   CreateAppInput{Repo: "gluonfield/exp20", Host: "ml.ink"},
			wantErr: true,
		},
		{
			name:    "rejects url",
			input:   CreateAppInput{Repo: "https://git.ml.ink/gluonfield/exp20.git"},
			wantErr: true,
		},
		{
			name:    "rejects prefixed repo",
			input:   CreateAppInput{Repo: "ml.ink/gluonfield/exp20", Host: "ml.ink"},
			wantErr: true,
		},
		{
			name:    "rejects embedded creds",
			input:   CreateAppInput{Repo: "gluonfield:token@git.ml.ink/gluonfield/exp20"},
			wantErr: true,
		},
		{
			name:    "rejects paths with slashes",
			input:   CreateAppInput{Repo: "a/b/c"},
			wantErr: true,
		},
		{
			name:    "invalid host",
			input:   CreateAppInput{Repo: "exp20", Host: "gitlab"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotHost, gotRepo, err := normalizeCreateAppRepo(u, tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (host=%q repo=%q)", gotHost, gotRepo)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotHost != tc.wantHost {
				t.Fatalf("host: got %q, want %q", gotHost, tc.wantHost)
			}
			if gotRepo != tc.wantRepo {
				t.Fatalf("repo: got %q, want %q", gotRepo, tc.wantRepo)
			}
		})
	}
}
