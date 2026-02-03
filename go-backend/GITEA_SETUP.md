# Gitea Integration Setup

## Implementation Status: COMPLETE ✓

All code has been implemented. The following configuration is needed to activate the integration.

---

## Required Configuration

| Variable | Description | Status |
|----------|-------------|--------|
| `GITEA_ADMINTOKEN` | Gitea API token with admin privileges | ✅ DONE |
| `GITEA_WEBHOOKSECRET` | Secret for verifying webhook signatures | ⏳ PENDING |
| `GITEA_COOLIFYPRIVATEKEYUUID` | Coolify's UUID for the SSH deploy key | ⏳ PENDING |

---

## Setup Steps

### Step 1: Create Gitea Admin Token

1. Log into `git.ml.ink` as admin
2. Go to **Settings** → **Applications**
3. Under "Manage Access Tokens", create a new token:
   - **Name:** `mcp-deploy-admin`
   - **Scopes:** Select all (or at minimum: `admin`, `write:user`, `write:repository`, `write:org`)
4. Copy the token

**Provide me:** The token value (I'll add it to config)

---

### Step 2: Generate Webhook Secret

Run this command and save the output:
```bash
openssl rand -hex 32
```

**Provide me:** The generated secret

---

### Step 3: SSH Deploy Key for Coolify

This allows Coolify to pull code from Gitea via SSH.

#### 3a. Generate SSH Keypair
```bash
ssh-keygen -t ed25519 -C "coolify-gitea-deploy" -f gitea-deploy-key -N ""
```

This creates:
- `gitea-deploy-key` (private key)
- `gitea-deploy-key.pub` (public key)

#### 3b. Add Public Key to Gitea

**Option A: As a Machine User (Recommended)**
1. Create a new Gitea user (e.g., `deploy-bot`)
2. Add the public key to that user's SSH keys
3. This user will be added as collaborator to repos automatically

**Option B: As Global Deploy Key**
1. Admin → Settings → Deploy Keys (if available)
2. Add the public key

#### 3c. Upload Private Key to Coolify
1. Go to Coolify → **Keys** → **Add New**
2. Paste the contents of `gitea-deploy-key` (private key)
3. Save and copy the **UUID** shown

**Provide me:** The Coolify private key UUID

---

## What I Need From You

1. ✅ **Gitea Admin Token:** `48d83c13d5c33...` (saved to .env)

2. ⏳ **Webhook Secret:** Generate with `openssl rand -hex 32` and provide

3. ⏳ **Coolify Private Key UUID:**
   - Generate SSH key: `ssh-keygen -t ed25519 -f gitea-deploy-key -N ""`
   - Add `gitea-deploy-key.pub` to Gitea (as deploy key or machine user)
   - Upload `gitea-deploy-key` (private) to Coolify → Keys
   - Provide the UUID

4. ✅ **Gitea Base URL:** `https://git.ml.ink`

5. ✅ **User prefix:** `u` (creates users like `u-abc123`)

---

## Current Config in .env

```bash
GITEA_ENABLED=true
GITEA_BASEURL=https://git.ml.ink
GITEA_ADMINTOKEN=48d83c13d5c33...  # ✅ DONE
GITEA_USERPREFIX=u
GITEA_WEBHOOKSECRET=               # ⏳ NEEDS: Step 2
GITEA_SSHURL=git@git.ml.ink
GITEA_COOLIFYPRIVATEKEYUUID=       # ⏳ NEEDS: Step 3
```

---

## After Configuration

Once configured, the following MCP tools will be available:

| Tool | Description |
|------|-------------|
| `create_repo(name, source="private")` | Creates repo at `ml.ink/u-{user_id}/{name}` |
| `get_push_token(repo="ml.ink/...")` | Gets fresh push credentials |
| `create_app(repo="ml.ink/...")` | Deploys from internal git |

Webhook endpoint: `POST /webhooks/internal-git` (auto-redeploy on push)

---

## Files Modified

```
go-backend/
├── application.yaml                          # Added gitea config section
├── sqlc.yaml                                 # Added internalrepos
├── internal/
│   ├── storage/pg/
│   │   ├── migrations/
│   │   │   ├── 0021_internal_repos.sql      # NEW: internal_repos table + gitea_username
│   │   │   └── 0022_apps_git_provider.sql   # NEW: git_provider column
│   │   └── queries/
│   │       ├── internalrepos/internalrepos.sql  # NEW
│   │       ├── users/users.sql              # Added gitea queries
│   │       └── apps/apps.sql                # Added git_provider
│   ├── internalgit/                         # NEW: entire package
│   │   ├── config.go
│   │   ├── types.go
│   │   ├── client.go
│   │   ├── repos.go
│   │   └── service.go
│   ├── coolify/
│   │   └── applications.go                  # Added CreatePrivateDeployKey
│   ├── deployments/
│   │   ├── types.go                         # Added GitProvider fields
│   │   ├── activities.go                    # Added CreateAppFromInternalGit
│   │   ├── workflow.go                      # Branch for gitea vs github
│   │   └── service.go                       # Added RedeployFromInternalGitPush
│   ├── webhooks/
│   │   ├── handlers.go                      # Added gitea config, new route
│   │   └── internalgit.go                   # NEW: webhook handler
│   ├── mcpserver/
│   │   ├── server.go                        # Added internalGitSvc
│   │   ├── types.go                         # Added CreateRepo/GetPushToken types
│   │   ├── tools.go                         # Updated create_app for gitea
│   │   └── tools_repo.go                    # NEW: unified repo tools
│   └── bootstrap/
│       ├── config.go                        # Added Gitea config
│       └── internalgit.go                   # NEW: service provider
└── cmd/server/main.go                       # Added providers
```

---

## Testing Checklist

After configuration:

- [ ] `create_repo(name="test-app")` returns `ml.ink/u-xxx/test-app` + git remote
- [ ] Can push code to the git remote
- [ ] `create_app(repo="ml.ink/u-xxx/test-app", branch="main", name="test")` deploys
- [ ] Push to repo triggers auto-redeploy via webhook
- [ ] GitHub flow still works (regression test)
