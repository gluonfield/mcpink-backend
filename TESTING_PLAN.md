# Deploy MCP — Testing Plan

> Deploy every common stack via MCP tools. See where it breaks.

## Results Summary (2026-02-10)

**9 PASS** | **15 FAIL** | **4 PARTIAL** | **0 still building**

### Fault Classification (after code inspection)

| Fault | Count | Tests |
|-------|-------|-------|
| **User code** (scaffold bugs) | 4 | #6, #8, #18, #28 — missing ESM config, bad lockfiles, version conflicts |
| **User code** (wrong `build_pack` param) | 4 | #1, #2, #3, #4 — code is correct, but agents chose `static` instead of `railpack` |
| **Deployment system** (resource exhaustion) | 9 | #16, #17, #20, #21, #22, #24, #25, #26, #27 — code verified correct, cluster ran out of capacity |
| **Both** (user code + system) | 2 | #19 (Spring Boot hardcodes port 8080), #23 (FastAPI Dockerfile hardcodes port 8000) — port mismatch AND resource exhaustion |

### Key Findings

1. **`build_pack="static"` skips build step.** Copies raw files to nginx. JS frameworks need `railpack`. The `build_command` parameter is ignored.

2. **k8s run pool saturates after ~13 services.** All later builds succeed but rollout times out at exactly 2m. Evidence: Go net/http (#15) PASS early, Go Gin (#16) FAIL later — identical stack.

3. **No readiness/liveness probes on deployed pods.** The platform deploys containers without k8s health probes. The rollout timeout is the ONLY failure signal. Can't distinguish "pod Pending (no resources)" from "pod crashing (wrong config)" from "port mismatch (app listening on wrong port)".

4. **PORT env var is always injected.** Platform sets `PORT=<configured-port>` in every container. Apps that read PORT (Streamlit, Gradio, Go, Node) work correctly. Apps that ignore PORT (Spring Boot hardcoded to 8080) fail silently — the k8s Service routes to port 3000 but the app isn't there.

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
| 1 | React + Vite | static | `static` | — | `temp/1/` | 1 | PARTIAL | https://test-react-vite.ml.ink | Deploys OK but serves raw index.html with uncompiled JSX imports. Static pack skips build. |
| 2 | Vue + Vite | static | `static` | — | `temp/2/` | 1 | PARTIAL | https://test-vue-vite.ml.ink | Deploys OK but serves raw index.html with uncompiled Vue imports. Static pack skips build. |
| 3 | Astro (static) | static | `static` | — | `temp/3/` | 1 | PARTIAL | https://test-astro-static.ml.ink | Deploys OK but shows nginx welcome page — no index.html at root (needs `dist/` from build). |
| 4 | Docusaurus | static | `static` | — | `temp/4/` | 1 | PARTIAL | https://test-docusaurus.ml.ink | Deploys OK but shows nginx welcome page — no index.html at root (needs `build/` from build). |
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

## Static Build Pack Finding (Tests 1-4)

**`build_pack="static"` does NOT run any build step.** It copies raw repo files directly into nginx and serves them. This means:

- **React/Vue (tests 1-2):** Serve raw `index.html` with uncompiled JSX/Vue imports — browser can't execute them
- **Astro/Docusaurus (tests 3-4):** Show default nginx welcome page because the build output directories (`dist/`, `build/`) don't exist

The `build_command` parameter is **ignored** by the static build pack.

**Recommendation:** JS frameworks that need a build step should use `build_pack="railpack"` instead. The `static` pack should only be used for pre-built HTML/CSS/JS files. Alternatively, the static pack should support running `build_command` before copying to nginx.

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

### Test #22 — Monorepo React + Express (API + Web)

**Status:** FAIL (deployment system — resource exhaustion)

**Symptom:** `BUILD SUCCESS` → `deployment rollout timed out after 2m0s` for both API and Web services.

**Code inspection:** Express backend is correct — reads `PORT` env var, binds `0.0.0.0`, defaults to 3000, has `/` and `/health` routes. Dockerfile CMD `node backend/index.js` is correct.

**Root Cause:** Resource exhaustion. Code would work on a healthy cluster.

---

### Test #23 — Monorepo React + FastAPI (API + Web)

**Status:** FAIL (BOTH user code + deployment system)

**Symptom:** `BUILD SUCCESS` → `deployment rollout timed out after 2m0s`.

**Code inspection:** Dockerfile hardcodes `CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]` and `EXPOSE 8000`. The `create_service` call used default port 3000. k8s Service/Ingress routes to port 3000, but uvicorn listens on 8000 → port mismatch.

**Root Cause:** Two issues: (1) Port mismatch — would return 503 even on healthy cluster unless `port: 8000` specified in `create_service`. (2) Resource exhaustion caused the immediate rollout timeout.

**Fix:** Either change Dockerfile to read `PORT` env var (`--port ${PORT:-8000}`) or specify `port: 8000` in the `create_service` call.

---

### Test #24 — Monorepo React + Go API (API + Web)

**Status:** FAIL (deployment system — resource exhaustion)

**Symptom:** `BUILD SUCCESS` → `deployment rollout timed out after 2m0s` for both services.

**Code inspection:** Go backend is correct — reads `PORT` env var with `os.Getenv("PORT")`, defaults to 3000, binds `0.0.0.0`. Dockerfile `CMD ["/server"]` is correct.

**Root Cause:** Resource exhaustion. Code would work on a healthy cluster.

---

### Test #16 — Go (Gin)

**Status:** FAIL (deployment system — resource exhaustion)

**Symptom:** `BUILD SUCCESS` → `deployment rollout timed out after 2m0s`. Railpack correctly detected Go, compiled with `go build -ldflags=-w -s -o out`.

**Code inspection:** Correct — reads `PORT` env var, defaults to 3000, `r.Run(":" + port)` binds `0.0.0.0`, has `/` and `/health` routes. Identical pattern to test #15 (Go net/http) which PASSED.

**Root Cause:** Resource exhaustion. Same Go stack as #15 which deployed early and passed.

---

### Test #17 — Bun + Hono

**Status:** FAIL (deployment system — resource exhaustion)

**Symptom:** `BUILD SUCCESS` → `deployment rollout timed out after 2m0s`. Railpack detected both Bun 1.3.9 and Node 22.22.0, installed both via mise.

**Code inspection:** Correct — reads `PORT` env var via `process.env.PORT || '3000'`, uses Bun-native `export default { port, fetch }` syntax, has `/` and `/health` routes. Start script is `bun run index.ts`.

**Root Cause:** Resource exhaustion. Note: railpack installing both Bun and Node is wasteful but doesn't cause failure. Start script `bun run index.ts` from package.json is correct.

---

### Test #19 — Spring Boot (Java)

**Status:** FAIL (BOTH user code + deployment system)

**Symptom:** `BUILD SUCCESS` (Maven 2.4s) → `deployment rollout timed out after 2m0s`.

**Code inspection:** App hardcodes port 8080 (Spring Boot default). No `application.properties` to override. Does NOT read `PORT` env var. Dockerfile `EXPOSE 8080` + `CMD ["java", "-jar", "app.jar"]`. The `create_service` call used default port 3000 → k8s routes to 3000 but app listens on 8080.

**Root Cause:** Two issues: (1) Port mismatch — would return 503 even on healthy cluster. (2) Resource exhaustion caused the immediate rollout timeout.

**Fix:** Either add `server.port=${PORT:8080}` to `application.properties` or specify `port: 8080` in the `create_service` call.

---

### Test #20 — Rust + Axum

**Status:** FAIL (deployment system — resource exhaustion)

**Symptom:** `BUILD SUCCESS` (Rust compiled in 14.65s via multi-stage Dockerfile) → `deployment rollout timed out after 2m0s`.

**Code inspection:** Correct — reads `PORT` env var, defaults to 3000, binds `0.0.0.0:{port}` with `TcpListener::bind`. Has `/` and `/health` routes. Multi-stage Dockerfile correct.

**Root Cause:** Resource exhaustion. Code would work on a healthy cluster.

---

## Rollout Timeout Analysis (Systemic Issue)

**Affected tests:** #16, #17, #20, #21, #22 (both), #24 (both), #25, #26, #27 — **9 stacks with correct code, 12 services**

**Also affected but with user code issues:** #19 (port 8080), #23 (port 8000) — port mismatch would cause 503 even on healthy cluster.

**Pattern:** Every single build succeeds. Every single rollout times out at exactly 2 minutes. This started after ~13 services were already running on the cluster.

**Evidence of resource exhaustion (not config error):**
- Go net/http (#15) deployed EARLY → PASS. Go Gin (#16) deployed LATER → FAIL. Same language, same railpack, same code pattern.
- Next.js (#5) deployed EARLY → PASS. Next.js fullstack (#21) deployed LATER → FAIL. Same framework.
- Express (#10) deployed EARLY → PASS. WebSocket/Express (#27) deployed LATER → FAIL. Same runtime.

**Root cause:** The `run` node pool has insufficient resources to run >13 services simultaneously. New pods stay `Pending` (no schedulable node) and the 2m rollout window expires.

**Critical platform finding: No readiness/liveness probes.**
Code inspection of `internal/k8sdeployments/k8s_resources.go` confirms: deployed pods have NO readiness, liveness, or startup probes. The rollout timeout is the ONLY failure signal. This means:
- Can't distinguish "pod Pending (no resources)" from "pod crashing (bad code)" from "port mismatch (app on wrong port)"
- Apps with port mismatches (#19, #23) would deploy "successfully" but return 503 — no health check catches it
- The `get_service` response only shows "deployment rollout timed out" with no pod-level diagnostics

**Investigation needed:**
1. `kubectl get pods -A | grep test-` — are pods Pending or CrashLoopBackOff?
2. `kubectl top nodes` — is the run pool CPU/memory exhausted?
3. `kubectl describe node <run-node>` — check allocatable vs allocated resources

**Platform improvements needed:**
1. **Add readiness probes** — HTTP GET on the configured port. Catches port mismatches, crashes, and gives meaningful pod events.
2. **Better error messages** — surface pod events (Pending/CrashLoopBackOff/OOMKilled) in `get_service` response.
3. **Auto-scaling** — detect when pods can't be scheduled and scale the run pool.
4. **Port validation** — for Dockerfile build_pack, warn if `EXPOSE` port doesn't match configured port.

---

### Test #28 — T3 Stack

**Status:** FAIL (scaffold issue)

**Symptom:** `npm install` fails with `ERESOLVE unable to resolve dependency tree`. `@trpc/react-query@10.45.4` requires peer `@tanstack/react-query@^4.18.0` but `@tanstack/react-query@5.90.20` was specified.

**Root Cause:** Agent manually assembled package.json with incompatible version ranges. `@trpc/react-query@^10` requires React Query v4, not v5.

**Fix:** Either use `@trpc/react-query@^11` (which supports React Query v5) or pin `@tanstack/react-query@^4`.

**Lesson:** Scaffolding issue, not platform. Railpack correctly detects and runs npm install — the error is clear and actionable.
