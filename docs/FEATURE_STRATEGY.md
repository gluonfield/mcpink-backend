# Feature Strategy

> Agents deploy 100x more apps than humans. Design every feature for the agent first, human second.

---

## Design Constraints

Agents are:
- **Dashboard-blind** — tool responses are their entire universe
- **Stateless across sessions** — tomorrow's agent doesn't know what today's agent did
- **Retry-happy** — timeouts happen; agents retry aggressively
- **High-throughput** — many deploys per session, dozens of services per user

The contract must optimize for: **URL + debuggable state**, **safe retries**, and **backpressure + cleanup**.

---

## Core Principles

### 1. One Transaction, Full Stack
An agent should go from "I built a Next.js app with a database" to "it's live" in one or two tool calls.

### 2. Errors Are Documentation
Every error must answer: what happened, what to do next, relevant state now, log tail for context. Agents can't Google things.

### 3. Explicit State Over Silent Idempotency
Create operations fail with full context (existing state + suggestions). Deletes are idempotent (desired end state already achieved). Two separate concerns:
- **Create semantics**: 409 conflict + full existing state + next actions
- **Retry safety**: `idempotency_key` dedupes identical replays without turning create into upsert

### 4. Self-Describing Responses
Every tool response includes enough context for the agent to take the next action without calling another tool.

### 5. Small Tool Surface, High Power
Aim for ~15 tools. Every tool description consumes tokens on every agent request. Keep descriptions lean and scannable.

### 6. Lifecycle Governance
Agents create garbage. TTL, quotas, and GC are not nice-to-haves — they're required.

---

## What's Already Shipped

These features are implemented and live. Do not re-plan them.

| Feature | Tools / Implementation |
|---|---|
| **Core deploy flow** (railpack, dockerfile, static, dockercompose) | `create_service`, `redeploy_service`, `delete_service` |
| **Service inspection with logs** (deploy + runtime logs via Loki) | `get_service` with `deploy_log_lines`, `runtime_log_lines`, `include_env` |
| **Service listing** | `list_services` |
| **SQLite databases via Turso** | `create_resource`, `get_resource`, `list_resources`, `delete_resource` |
| **Git: internal (Gitea) + GitHub** | `create_repo`, `get_git_token` with `host=ml.ink` / `host=github.com` |
| **Auto-deploy on push** (GitHub App + Gitea webhooks) | Temporal workflows with deterministic IDs from commit SHA |
| **Custom domains** (add, verify DNS, remove) | `add_custom_domain`, `verify_custom_domain`, `remove_custom_domain` |
| **Auth** (Firebase + GitHub OAuth + MCP OAuth with PKCE) | API keys, MCP OAuth flow, GitHub App installation |
| **User identity** | `whoami` |
| **Build packs** (railpack auto-detect, dockerfile, static, railpack+publish_directory for SPAs) | Build pack selection with validation |
| **Resource limits** (memory/vCPU enums stored in DB) | Configurable via `create_service`, applied to pod spec |
| **Monorepo support** | `root_directory`, `dockerfile_path` params |
| **Per-namespace ResourceQuota** | Reconciled by infra (40 CPU, 40Gi memory, 200 pods) |
| **Pod security** (drop ALL caps, no privilege escalation, no SA token) | K8s resources in `k8s_resources.go` |
| **Network isolation** (default-deny ingress, block private ranges on egress) | NetworkPolicy in k8s resources |
| **Deployment tracking** | Separate `deployments` table, deployment state independent of service |
| **Env vars** (set on create, viewable via get_service) | Encrypted at rest, injected at runtime |

---

## Tier 1 — Ship Next

High-impact features that make agents significantly more effective.

### 1.1 Resource Auto-Wiring in `create_service`

**Status: Not started**

Allow `create_service` to provision resources inline and auto-inject connection env vars:

```
create_service({
  "name": "my-saas",
  "repo": "my-saas",
  "resources": [{"type": "sqlite", "name": "main-db"}],
  "env_vars": [{"key": "NODE_ENV", "value": "production"}]
})
```

Server provisions the SQLite database and injects `DATABASE_URL` + `DATABASE_AUTH_TOKEN` automatically. Response includes what got wired.

### 1.2 `update_service`

**Status: Not started**

Explicit mutation tool for modifying live services without full recreate:
- Env vars (add/update/remove)
- Memory/CPU scaling
- Command overrides (build, start)
- Build pack changes

### 1.3 Conflict Responses with Full State

**Status: Partial** — creates currently return generic errors on duplicates.

When `create_service("my-app")` hits a duplicate, return 409 with:
- Existing service state (URL, status, config, resources)
- Suggestions: use `update_service`, `redeploy_service`, or `delete_service` + recreate

### 1.4 Structured Health Status

**Status: Not started** — `get_service` returns deployment/runtime status but no health analysis.

Add to `get_service` response:
```json
{
  "health": {
    "status": "unhealthy",
    "reason": "crash_loop",
    "restart_count": 3,
    "last_error_line": "Error: Cannot find module 'express'",
    "suggestion": "Check that 'express' is in package.json dependencies."
  }
}
```

### 1.5 Retry Safety via `idempotency_key`

**Status: Not started**

All mutating calls accept optional `idempotency_key`:
- Same key + same request hash → return cached result
- Same key + different hash → `idempotency_key_reuse_conflict` error
- Keys expire after 24h

Prevents duplicate builds/workflows from network retries without making creates idempotent.

### 1.6 Backpressure + Quotas

**Status: Partial** — per-namespace ResourceQuota exists in k8s; no API-level rate limits or service caps.

Add:
- Max services per tenant/project
- Concurrent builds per tenant
- Rate limits on mutating calls
- "deploy in progress" dedup (partially handled via Temporal deterministic workflow IDs)

### 1.7 Shared BuildKit Cache

**Status: Not started**

Currently each service has its own BuildKit registry cache (`registry/cache/{namespace}/{name}:buildcache`). Language runtime layers (Go, Node, Python installs) are re-downloaded per service.

Add a shared cache import alongside per-service cache so common layers (language runtimes) are downloaded once across all customers:

```go
solveOpt.CacheImports = []client.CacheOptionsEntry{
    {Type: "registry", Attrs: map[string]string{"ref": perServiceCacheRef}},
    {Type: "registry", Attrs: map[string]string{"ref": sharedCacheRef}},
}
```

Optionally maintain a warm-up job that pre-builds common runtimes (Go, Node LTS, Python) and exports to the shared cache ref. This reduces first-build times for new customers and provides resilience if the build node is reprovisioned.

BuildKit's local content-addressable store already deduplicates identical layers on the same daemon, but the shared registry cache makes this survive daemon restarts and node reprovisioning.

---

## Tier 2 — Ship Soon

Features that build competitive advantage and make the platform sticky.

### 2.1 Service Linking

Allow env var templates referencing other services:
```
"env_vars": [{"key": "API_URL", "value": "{{service:api:internal_url}}"}]
```

Resolves to internal cluster URL. Updates automatically if the referenced service moves.

### 2.2 Preview Environments

Branch/PR deploys with auto-cleanup:
- `https://my-app-pr-42.ml.ink`
- TTL (e.g., 72h) with auto-delete
- "Promote preview → production" flow

### 2.3 Exec One-Off Commands

Run migrations, seed data, or debug in service containers:
```
exec_command(service="my-app", command="npx prisma migrate deploy")
→ Returns stdout/stderr
```

With timeouts, output truncation, and audit logging.

### 2.4 Rollback

Revert to previous working deployment:
```
rollback_service(name="my-app")
rollback_service(name="my-app", deployment_id="dep_abc")
```

Model: Service → points to Release (immutable image + config). Rollbacks are pointer swaps to previous releases.

### 2.5 Metrics via Tool

Add `include_metrics` to `get_service`:
```json
{
  "metrics": {
    "cpu_usage_percent": 85,
    "memory_usage_mb": 450,
    "memory_limit_mb": 512
  },
  "warnings": ["Memory at 88% of limit. Consider increasing to 1024Mi."]
}
```

Data comes from Prometheus (already deployed on ops-1).

### 2.6 Deployment History

```
get_service_history(name="my-app")
```

List past deployments with commit SHA, status, duration. Enables informed rollback decisions.

---

## Tier 3 — Ship Later

Features that build a moat.

### 3.1 Stacks / Templates

One-call architecture templates encoding best practices:
```
create_stack(template="nextjs-sqlite", name="my-saas", repo="my-saas")
→ Creates service + SQLite + correct env wiring + healthcheck
```

### 3.2 Scheduled Jobs (Cron)

First-class cron tool for SaaS apps:
```
create_cron(name="daily-cleanup", service="my-app", schedule="0 2 * * *", command="node scripts/cleanup.js")
```

### 3.3 Multi-Region

Deploy stateless services to multiple regions with latency-based routing. Volumes are single-region only (local block storage). Turso databases handle multi-region natively.

Enforced constraints:
- Multi-region + volume → reject with clear error
- Multi-replica + volume → reject with clear error
- Multi-region + stateless or database → allowed

### 3.4 Cost Estimation + Budgets

Pre-flight cost estimate before deploying. Tenant-level budgets/spend caps to prevent runaway agent spending.

### 3.5 Event Webhooks

Let users receive webhooks for deploy.success, deploy.failed, service.unhealthy. Enables agent-built monitoring (e.g., deploy notifications to Slack).

---

## Tool Design Guidelines

### Lean Descriptions
```
// BAD: 200 tokens
"Deploy a service from a git repository. This tool will clone the repository,
 detect the build pack, build the application using BuildKit..."

// GOOD: 40 tokens
"Deploy a service from a git repo. Auto-detects language, builds, deploys,
 and returns the public URL."
```

### Flat Parameters
Agents work better with flat params than nested objects. Use `"memory": "512Mi"` not `"service.config.resources.memory": "512Mi"`.

### Status Enums Not Free Text
```
// BAD
"status": "The service is currently experiencing issues"

// GOOD
"status": "crash_loop",
"health_message": "Container exited 3 times in 5 minutes."
```

Standard status enums: `queued`, `building`, `deploying`, `active`, `failed`, `cancelled`, `crash_loop`, `stopped`, `deleting`.

### Error Response Pattern

Every error includes: code, message, suggestion, relevant state.

```json
{
  "error": "service_not_found",
  "message": "No service named 'my-ap' found.",
  "suggestion": "Did you mean 'my-app'? Use list_services() to see all services.",
  "available_services": ["my-app", "my-api"]
}
```

Build failures inline the last 30 lines of build output directly in the error response so the agent doesn't need a follow-up call.

---

## Tool Surface

### Shipped (15 tools)

| Tool | Purpose |
|---|---|
| `whoami` | Auth check + account status |
| `create_service` | Deploy from git repo |
| `get_service` | Status + logs + env |
| `list_services` | All services overview |
| `redeploy_service` | Trigger redeploy |
| `delete_service` | Remove service |
| `create_resource` | Provision database |
| `get_resource` | Connection details |
| `list_resources` | All databases |
| `delete_resource` | Remove database |
| `create_repo` | Create git repo |
| `get_git_token` | Temp push token |
| `add_custom_domain` | Attach custom domain |
| `verify_custom_domain` | Verify DNS + activate |
| `remove_custom_domain` | Detach custom domain |

### Next (Tier 1-2)

| Tool | Purpose |
|---|---|
| `update_service` | Modify env/scaling/config without recreate |
| `exec_command` | Run one-off command in container |
| `rollback_service` | Revert to previous deployment |
| `get_service_history` | Past deployments with status |

### Future (Tier 3)

| Tool | Purpose |
|---|---|
| `create_cron` | Scheduled tasks |
| `estimate_cost` | Cost estimation |
| `create_webhook` | Event notifications |
| `create_stack` | Template-based deploy |
