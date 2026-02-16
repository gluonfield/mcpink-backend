package gitserver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureBareRepo creates a bare git repository if it doesn't already exist.
// Idempotent: no-op if the repo already exists.
func ensureBareRepo(reposRoot, owner, repo string) (string, error) {
	barePath := filepath.Join(reposRoot, owner, repo+".git")

	if _, err := os.Stat(filepath.Join(barePath, "HEAD")); err == nil {
		return barePath, nil
	}

	if err := os.MkdirAll(filepath.Dir(barePath), 0o755); err != nil {
		return "", fmt.Errorf("mkdir for bare repo: %w", err)
	}

	cmd := exec.Command("git", "init", "--bare", barePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git init --bare: %w\n%s", err, out)
	}

	return barePath, nil
}

// RefSnapshot maps ref names (e.g. "refs/heads/main") to commit SHAs.
type RefSnapshot map[string]string

// snapshotRefs captures current refs in a bare repo via git for-each-ref.
func snapshotRefs(barePath string) (RefSnapshot, error) {
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname) %(objectname)", barePath)
	out, err := cmd.Output()
	if err != nil {
		// Empty repo has no refs — that's fine
		return RefSnapshot{}, nil
	}

	refs := make(RefSnapshot)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			refs[parts[0]] = parts[1]
		}
	}
	return refs, nil
}

// ChangedRef represents a ref that changed between two snapshots.
type ChangedRef struct {
	Name   string // e.g. "refs/heads/main"
	Branch string // e.g. "main"
	OldSHA string // empty if new branch
	NewSHA string
}

// diffRefs compares two ref snapshots and returns changed/new branches.
// Deleted branches (present in before but not after) are ignored — we only
// trigger deploys for branches that have new commits.
func diffRefs(before, after RefSnapshot) []ChangedRef {
	var changes []ChangedRef
	for ref, newSHA := range after {
		if !strings.HasPrefix(ref, "refs/heads/") {
			continue
		}
		oldSHA := before[ref]
		if oldSHA == newSHA {
			continue
		}
		branch := strings.TrimPrefix(ref, "refs/heads/")
		changes = append(changes, ChangedRef{
			Name:   ref,
			Branch: branch,
			OldSHA: oldSHA,
			NewSHA: newSHA,
		})
	}
	return changes
}

// barePath resolves the filesystem path for a bare repo.
func barePath(reposRoot, owner, repo string) string {
	return filepath.Join(reposRoot, owner, repo+".git")
}
