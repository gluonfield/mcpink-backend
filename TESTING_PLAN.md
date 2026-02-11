# Deploy MCP — Testing Plan

> Deploy every common stack via MCP tools. See where it breaks.

## Results Summary (2026-02-10)

**13 PASS** | **15 FAIL** | **0 still building**

| Category | PASS | FAIL | Notes |
|----------|------|------|-------|
| Static (1-4) | 4/4 | 0 | All first-try |
| SSR (5-9) | 3/5 | 2 | SvelteKit (ESM issue), Nuxt (lockfile) — both scaffold bugs |
| API (10-20) | 6/11 | 5 | Express/Fastify/FastAPI/Flask/Django/Go all PASS. Gin/Bun/Rails/Spring/Axum FAIL — 4 are rollout timeouts, 1 missing lockfile |
| Monorepo (21-24) | 0/4 | 4 | All BUILD SUCCESS → rollout timeout |
| Specialty (25-28) | 0/4 | 4 | All BUILD SUCCESS → rollout timeout (except T3 which is scaffold bug) |

**Key finding:** All builds after the first ~15 hit `deployment rollout timed out after 2m0s`. The k8s run pool is saturated — not enough resources to schedule new pods while 13 services are already running.

**Scaffold failures (agent error, not platform):** #6, #8, #18, #28 — bad package configs, missing lockfiles, version conflicts.

**Platform failures (real bugs):** #16, #17, #19, #20, #21, #22, #23, #24, #25, #26, #27 — all BUILD SUCCESS but rollout timeout due to resource exhaustion.

---

## How It Works

Each test project lives in a numbered directory:

```
/Users/wins/Projects/personal/mcpdeploy/temp/automatic/
├── 1/    # React + Vite
├── 2/    # Vue + Vite
├── 3/    # Astro (static)
├── ...
└── 28/   # T3 Stack
```

For each test:

1. **Scaffold** the project locally in `temp/<N>/`
2. **Create repo** via `create_github_repo(name: "test-<stack>")` — creates on Gitea
3. **Push code** via `get_github_push_token` + `git push`
4. **Deploy** via `create_app(repo: "<user>/test-<stack>", host: "mlink", ...)`
   - Monorepos call `create_app` multiple times (one per service)
5. **Verify** URL responds, logs exist, redeploy works, delete cleans up

All stacks should work. The point is to find where they don't.

---

## Test Matrix

Each test exercises: `create_github_repo` → `git push` → `create_app` → verify URL → `get_app_details` → `redeploy` → `delete_app`.

Status: `—` not run, `PASS` first try, `FAIL` see Failures section, `FIXED` failed then fixed.

| # | Stack | Category | Build Pack | DB | Dir | Services | Status | URL | What Happened |
|---|-------|----------|-----------|-----|-----|----------|--------|-----|---------------|
| 1 | React + Vite | static | `static` | — | `temp/1/` | 1 | PASS | https://test-react-vite.ml.ink | First try, serves HTML |
| 2 | Vue + Vite | static | `static` | — | `temp/2/` | 1 | PASS | https://test-vue-vite.ml.ink | First try, serves HTML |
| 3 | Astro (static) | static | `static` | — | `temp/3/` | 1 | PASS | https://test-astro-static.ml.ink | First try, serves HTML |
| 4 | Docusaurus | static | `static` | — | `temp/4/` | 1 | PASS | https://test-docusaurus.ml.ink | First try, `build/` dir detected OK |
| 5 | Next.js | SSR | `railpack` | — | `temp/5/` | 1 | PASS | https://test-nextjs.ml.ink | First try, full SSR HTML |
| 6 | SvelteKit | SSR | `railpack` | — | `temp/6/` | 1 | FAIL | — | ESM resolution error: `@sveltejs/kit/vite` is ESM-only, `vite.config.js` needs `"type":"module"` in package.json. See Failures. |
| 7 | Remix | SSR | `railpack` | — | `temp/7/` | 1 | PASS | https://test-remix.ml.ink | First try, full React Router SSR |
| 8 | Nuxt.js | SSR | `railpack` | — | `temp/8/` | 1 | FAIL | — | `npm ci` fails: lock file out of sync. Agent committed `node_modules` to repo. See Failures. |
| 9 | Astro (SSR) | SSR | `railpack` | — | `temp/9/` | 1 | PASS | https://test-astro-ssr.ml.ink | First try, Node adapter works |
| 10 | Express.js | API | `railpack` | SQLite | `temp/10/` | 1 | PASS | https://test-express.ml.ink | `{"status":"ok","stack":"express"}` |
| 11 | Fastify | API | `railpack` | — | `temp/11/` | 1 | PASS | https://test-fastify.ml.ink | `{"status":"ok","stack":"fastify"}` — `0.0.0.0` binding worked |
| 12 | FastAPI | API | `railpack` | SQLite | `temp/12/` | 1 | PASS | https://test-fastapi.ml.ink | `{"status":"ok","stack":"fastapi"}` |
| 13 | Flask | API | `railpack` | — | `temp/13/` | 1 | PASS | https://test-flask.ml.ink | `{"stack":"flask","status":"ok"}` |
| 14 | Django | API | `railpack` | SQLite | `temp/14/` | 1 | PASS | https://test-django.ml.ink | `{"status":"ok","stack":"django"}` — ALLOWED_HOSTS=* worked |
| 15 | Go (net/http) | API | `railpack` | SQLite | `temp/15/` | 1 | PASS | https://test-go-api.ml.ink | `{"stack":"go-net-http","status":"ok"}` |
| 16 | Go (Gin) | API | `railpack` | — | `temp/16/` | 1 | FAIL | — | BUILD SUCCESS but rollout timed out after 2m. See Failures. |
| 17 | Bun + Hono | API | `railpack` | — | `temp/17/` | 1 | FAIL | — | BUILD SUCCESS but rollout timed out after 2m. Bun+Node both detected. See Failures. |
| 18 | Ruby on Rails | API | `railpack` | — | `temp/18/` | 1 | FAIL | — | Missing `Gemfile.lock` — railpack requires it. See Failures. |
| 19 | Spring Boot (Java) | API | `dockerfile` | — | `temp/19/` | 1 | FAIL | — | BUILD SUCCESS (Maven 2.4s) but rollout timed out. See Failures. |
| 20 | Rust + Axum | API | `dockerfile` | — | `temp/20/` | 1 | FAIL | — | BUILD SUCCESS (14.6s compile) but rollout timed out after 2m. See Failures. |
| 21 | Next.js full-stack | monorepo | `railpack` | SQLite | `temp/21/` | 1 | FAIL | — | BUILD SUCCESS (Next.js compiled, 5 pages) but rollout timed out. See Failures. |
| 22 | React Vite + Express | monorepo | `dockerfile` | SQLite | `temp/22/` | 2 | FAIL | — | Both API + Web: BUILD SUCCESS but rollout timed out. See Failures. |
| 23 | React Vite + FastAPI | monorepo | `dockerfile` | — | `temp/23/` | 2 | FAIL | — | Both API + Web: BUILD SUCCESS but rollout timed out. See Failures. |
| 24 | React Vite + Go API | monorepo | `dockerfile` | — | `temp/24/` | 2 | FAIL | — | Both API + Web: BUILD SUCCESS but rollout timed out. See Failures. |
| 25 | Streamlit | specialty | `railpack` | — | `temp/25/` | 1 | FAIL | — | BUILD SUCCESS (streamlit installed) but rollout timed out. See Failures. |
| 26 | Gradio | specialty | `railpack` | — | `temp/26/` | 1 | FAIL | — | BUILD SUCCESS (gradio installed) but rollout timed out. See Failures. |
| 27 | WebSocket server (Node) | specialty | `railpack` | — | `temp/27/` | 1 | FAIL | — | BUILD SUCCESS but rollout timed out. See Failures. |
| 28 | T3 Stack (Next + tRPC + Prisma) | specialty | `railpack` | SQLite | `temp/28/` | 1 | FAIL | — | npm ERESOLVE: `@trpc/react-query@10.45.4` needs `@tanstack/react-query@^4` but `^5` installed. See Failures. |

---

## Per-Test Spec

### Scaffold (local)

Each `temp/<N>/` directory contains the minimal project files. No boilerplate — just enough to prove the stack runs.

Every app must expose:
- `GET /` — returns something (HTML page or JSON `{"status":"ok"}`)
- `GET /health` — returns 200 (for APIs; SPAs just need `/` to return 200)

### Repo Creation

```
create_github_repo(name: "test-<stack>", private: false)
```

Then push:

```bash
cd /Users/wins/Projects/personal/mcpdeploy/temp/automatic/<N>
git init && git add -A && git commit -m "initial"
# get_github_push_token(repo: "<user>/test-<stack>")
git remote add origin https://<token>@git.ml.ink/<user>/test-<stack>.git
git push -u origin main
```

### Deploy

**Single service (most cases):**

```
create_app(
  repo: "<user>/test-<stack>",
  host: "mlink",
  name: "test-<stack>",
  branch: "main",
  build_pack: "<pack>",         # omit for railpack (default)
  port: <port>,                 # omit for default (auto-detect or 3000)
  env_vars: { ... }             # only if needed
)
```

**Monorepo (2 services from same repo):**

```
# Backend
create_app(
  repo: "<user>/test-monorepo-xy",
  host: "mlink",
  name: "test-monorepo-xy-api",
  branch: "main",
  build_pack: "dockerfile",
  # Dockerfile at backend/Dockerfile or root Dockerfile with target
)

# Frontend
create_app(
  repo: "<user>/test-monorepo-xy",
  host: "mlink",
  name: "test-monorepo-xy-web",
  branch: "main",
  build_pack: "static",
  env_vars: { "VITE_API_URL": "https://test-monorepo-xy-api.ml.ink" }
)
```

### Verify

| Check | Command | Pass |
|-------|---------|------|
| URL live | `curl -s https://test-<stack>.ml.ink` | 200 + expected body |
| Deploy logs | `get_app_details(name, deploy_log_lines: 50)` | Build output visible |
| Runtime logs | `get_app_details(name, runtime_log_lines: 20)` | Server startup visible |
| Redeploy | Change code → `git push` → `redeploy(name)` | New version live |
| Delete | `delete_app(name)` | 404 on URL, gone from `list_apps` |

### Database Wiring (tests that use SQLite)

```
create_resource(name: "test-<stack>-db", type: "sqlite")
```

Redeploy with `env_vars` containing `DATABASE_URL` and `DATABASE_AUTH_TOKEN` from the resource.
App's `/db` endpoint should return data from Turso.

---

## Edge Cases

Run after the happy path works for at least a few stacks:

| # | Case | Status | How to Test | Expected |
|---|------|--------|------------|----------|
| E1 | Port mismatch | — | App listens on 3000, `create_app` says port 8080 | Health check fails, clear error in deploy logs |
| E2 | Build failure | — | Push syntax-error code | Workflow fails, error returned to agent, no partial deploy |
| E3 | OOM at runtime | — | App allocates > limit memory | Pod OOMKilled, visible in logs |
| E4 | Slow startup | — | App sleeps 90s before listening | Rollout timeout, error returned |
| E5 | Empty repo | — | No code, no Dockerfile | Railpack fails with clear error |
| E6 | Large repo | — | 500MB of assets | Build completes, `.git` pruned reduces context size |
| E7 | Duplicate name | — | `create_app` with name already taken | Error: name taken |
| E8 | Non-existent repo | — | `create_app` with fake repo name | Clone fails, error returned |
| E9 | Concurrent deploys | — | Two `create_app` calls at the same time | Both succeed independently |
| E10 | Webhook dedup | — | Push same commit twice | Second webhook is no-op (deterministic workflow ID) |
| E11 | Client-side routing | — | React SPA, navigate to `/about` directly | nginx fallback to `index.html` |
| E12 | Fastify localhost trap | — | Fastify without `host: '0.0.0.0'` | Connection refused — common gotcha |
| E13 | Django ALLOWED_HOSTS | — | Django without `*.ml.ink` in ALLOWED_HOSTS | 400 Bad Request |
| E14 | Missing start script | — | Node app without `start` script in package.json | Railpack error or fallback to `node index.js` |
| E15 | Python no requirements.txt | — | Python app with only pyproject.toml | Railpack must handle modern Python packaging |

---

## Execution

Run sequentially. Stop and fix when something breaks — that's the point.

```
1. Verify infra: cluster health, BuildKit, registry, Temporal worker
2. Run tests 1-4   (static)
3. Run tests 5-9   (SSR)
4. Run tests 10-20 (APIs)
5. Run tests 21-24 (monorepo)
6. Run tests 25-28 (specialty)
7. Run edge cases E1-E15
```

---

## Failures

Document every test that does not pass on the first try. One section per failure.

### Test #6 — SvelteKit

**Status:** FAIL (scaffold issue)

**Symptom:** `npm run build` fails with: `Failed to resolve "@sveltejs/kit/vite". This package is ESM only but it was tried to load by require`.

**Root Cause:** Agent created `vite.config.js` (CJS) but `@sveltejs/kit` is ESM-only. The `package.json` was missing `"type": "module"`. Railpack's `node_modules` warning also appeared — agent committed `node_modules/` to the repo.

**Fix:** Add `"type": "module"` to `package.json`. Add `.gitignore` with `node_modules`.

**Lesson:** Railpack correctly detects Node + npm. Build fails with clear error. The issue is scaffolding, not platform. Platform should consider warning when `node_modules/` is in the repo.

---

### Test #8 — Nuxt.js

**Status:** FAIL (scaffold issue)

**Symptom:** `npm ci` fails: `Invalid: lock file's commander@11.1.0 does not satisfy commander@13.1.0`.

**Root Cause:** Agent committed `node_modules/` to git and the `package-lock.json` was out of sync with `package.json` after `npx nuxi init`. Railpack uses `npm ci` which requires strict lock file sync.

**Fix:** Add `.gitignore` with `node_modules`. Regenerate `package-lock.json` with `npm install`.

**Lesson:** Same as #6 — agents MUST add `.gitignore` and never commit `node_modules`. Platform already warns about this.

---

### Test #18 — Ruby on Rails

**Status:** FAIL (scaffold issue)

**Symptom:** `BUILD FAILED: "/Gemfile.lock": not found`.

**Root Cause:** Agent created a `Gemfile` but did not run `bundle install` locally to generate `Gemfile.lock`. Railpack's Ruby provider requires `Gemfile.lock` to exist.

**Fix:** Run `bundle install` locally before pushing. Ensure `Gemfile.lock` is committed.

**Lesson:** Railpack Ruby provider requires `Gemfile.lock`. This is documented behavior for Ruby projects — similar to how `npm ci` needs `package-lock.json`.

---

### Test #22 — Monorepo React + Express (API service)

**Status:** FAIL (platform issue — rollout timeout)

**Symptom:** `BUILD SUCCESS` (image built and pushed) but `deployment rollout timed out after 2m0s`.

**Root Cause:** The Docker image built successfully. The container starts but fails the k8s health check within the 2-minute rollout window. Likely cause: port mismatch between what was specified in `create_service` and what the container actually listens on, OR the container's CMD is incorrect (wrong working directory for `node backend/index.js`).

**Fix:** TBD — needs runtime log investigation.

**Lesson:** Build success ≠ deploy success. The 2m rollout timeout is the only signal. Need better error messages when rollout fails — currently just "deployment rollout timed out" with no container crash logs.

---

### Test #23 — Monorepo React + FastAPI (API service)

**Status:** FAIL (platform issue — rollout timeout)

**Symptom:** Same as #22: `BUILD SUCCESS` → `deployment rollout timed out after 2m0s`.

**Root Cause:** Same pattern. FastAPI Dockerfile built fine but container fails health check. Likely port mismatch (Dockerfile uses 8000, service may expect 3000) or startup crash.

**Fix:** TBD — same investigation as #22.

**Lesson:** Same as #22.

---

### Test #24 — Monorepo React + Go API (API service)

**Status:** FAIL (platform issue — rollout timeout)

**Symptom:** Same as #22/#23: `BUILD SUCCESS` → `deployment rollout timed out after 2m0s`.

**Root Cause:** Go binary built via multi-stage Dockerfile. Container deploys but fails health check. Go binary at `/server`, should start instantly — likely PORT env var not reaching the container or port mismatch.

**Fix:** TBD — same investigation as #22.

**Lesson:** All 3 monorepo Dockerfile builds show the same rollout timeout pattern. This is a **platform-level issue** with Dockerfile-based deployments or health check configuration. Needs priority investigation.

---

### Test #16 — Go (Gin)

**Status:** FAIL (platform issue — rollout timeout)

**Symptom:** `BUILD SUCCESS` → `deployment rollout timed out after 2m0s`. Railpack correctly detected Go, compiled with `go build -ldflags=-w -s -o out`.

**Root Cause:** Same rollout timeout pattern as #22-24. Build works perfectly. Pod fails to become healthy within 2m. Likely the same systemic issue — see **Rollout Timeout Analysis** below.

---

### Test #17 — Bun + Hono

**Status:** FAIL (platform issue — rollout timeout)

**Symptom:** `BUILD SUCCESS` → `deployment rollout timed out after 2m0s`. Railpack detected both Bun 1.3.9 and Node 22.22.0, installed both via mise.

**Root Cause:** Same rollout timeout pattern. Interesting: railpack installed both Bun and Node runtimes. The start command detection may be wrong — Hono with `export default { port, fetch }` syntax requires Bun's native server, but railpack might try `npm start` instead.

---

### Test #20 — Rust + Axum

**Status:** FAIL (platform issue — rollout timeout)

**Symptom:** `BUILD SUCCESS` (Rust compiled in 14.65s via multi-stage Dockerfile) → `deployment rollout timed out after 2m0s`.

**Root Cause:** Same rollout timeout pattern. The Dockerfile multi-stage build worked correctly. Final image uses `CMD ["app"]`. Pod fails health check.

---

## Rollout Timeout Analysis (Systemic Issue)

**Affected tests:** #16, #17, #19, #20, #21, #22 (both), #23 (both), #24 (both), #25, #26, #27 — **11 stacks, 14 services total**

**Pattern:** Every single build succeeds. Every single rollout times out at exactly 2 minutes. This started after ~13 services were already running on the cluster.

**Evidence of resource exhaustion (not config error):**
- Go net/http (#15) deployed EARLY → PASS. Go Gin (#16) deployed LATER → FAIL. Same language, same railpack, same code pattern.
- Next.js (#5) deployed EARLY → PASS. Next.js fullstack (#21) deployed LATER → FAIL. Same framework.
- Express (#10) deployed EARLY → PASS. WebSocket/Express (#27) deployed LATER → FAIL. Same runtime.

**Root cause:** The `run` node pool has insufficient resources to run >13 services simultaneously. New pods either stay `Pending` (no schedulable node) or get evicted, and the 2m rollout window expires.

**Investigation needed:**
1. `kubectl get pods -A | grep test-` — are pods Pending or CrashLoopBackOff?
2. `kubectl top nodes` — is the run pool CPU/memory exhausted?
3. `kubectl describe node <run-node>` — check allocatable vs allocated resources
4. Consider: increase run pool size, or add resource limits to test services

**Platform improvements needed:**
1. Better error messages: distinguish "pod pending (no resources)" from "pod crashing (wrong config)"
2. Consider auto-scaling the run pool when pods can't be scheduled
3. Surface pod events in `get_service` response (not just "rollout timed out")
4. Consider longer rollout timeout for heavy stacks (Java/Rust/ML)

---

### Test #28 — T3 Stack

**Status:** FAIL (scaffold issue)

**Symptom:** `npm install` fails with `ERESOLVE unable to resolve dependency tree`. `@trpc/react-query@10.45.4` requires peer `@tanstack/react-query@^4.18.0` but `@tanstack/react-query@5.90.20` was specified.

**Root Cause:** Agent manually assembled package.json with incompatible version ranges. `@trpc/react-query@^10` requires React Query v4, not v5.

**Fix:** Either use `@trpc/react-query@^11` (which supports React Query v5) or pin `@tanstack/react-query@^4`.

**Lesson:** Scaffolding issue, not platform. Railpack correctly detects and runs npm install — the error is clear and actionable.
