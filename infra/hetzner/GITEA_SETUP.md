# Internal Git (Gitea) Setup

This file is for future me.

## How it works (high level)

- A user calls `create_repo` MCP tool → we create a repo in Gitea under their namespace.
- We ensure the repo is readable by Coolify:
  - add `GITEA_DEPLOYPUBLICKEY` as a repo deploy key (read-only)
  - add `coolify-deploy` as a read-only collaborator (redundant on purpose)
- User pushes code via the returned HTTPS remote.
- User calls `create_app` → we create a Coolify application using **Private Repository (with Deploy Key)**.
- Gitea push webhook → `/webhooks/internal-git` → triggers redeploy.

## MCP usage (repo formats)

The `create_app` tool only accepts:

- repo name: `exp20`
- owner/repo: `gluonfield/exp20`

Examples:

```json
{ "repo": "exp20", "branch": "main", "name": "exp20" }
```

```json
{
  "repo": "gluonfield/exp20",
  "host": "mlink",
  "branch": "main",
  "name": "exp20"
}
```

```json
{ "repo": "exp20", "host": "github", "branch": "main", "name": "exp20" }
```

It intentionally rejects URLs and embedded credentials (too easy to mis-route, and it leaks secrets).

## Caveats / gotchas

- **Coolify validation:** Coolify rejects `ssh://...` in `git_repository`. Use `GITEA_SSHURL=git@host` and `GITEA_SSHPORT=...`.
- **Custom SSH port:** Coolify only supports custom SSH port through the weird format `git@host:PORT/owner/repo.git` (not `ssh://git@host:PORT/...`).
- **Port 22 != Gitea SSH:** in our infra Gitea SSH is exposed on `2222`, port `22` is the host’s SSH daemon.

## Minimal checklist when something breaks

- 422 “git repository must start with …”: check we’re not passing `ssh://...` anywhere and that `GITEA_SSHURL` is `git@...`.
- “Permission denied (publickey)”: check `GITEA_SSHPORT` and that deploy key/collaborator exists on the repo.
- “Bad Gateway”: check exposed port vs buildpack (`static` should be port 80).
