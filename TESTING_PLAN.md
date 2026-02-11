# Deploy MCP — Testing Plan

Projects live in `/Users/wins/Projects/personal/mcpdeploy/temp/automatic/<N>/`

## Test Matrix

| # | Dir | Stack | Build Pack | Status | URL | Issue |
|---|-----|-------|-----------|--------|-----|-------|
| 1 | `1/` | React + Vite | `railpack` | ✅ | https://test-react-vite-v2.ml.ink | Fixed: publish_directory=dist |
| 2 | `2/` | Vue + Vite | `railpack` | ✅ | https://test-vue-vite.ml.ink | Fixed: publish_directory=dist |
| 3 | `3/` | Astro (static) | `railpack` | ✅ | https://test-astro-static.ml.ink | Fixed: publish_directory=dist |
| 4 | `4/` | Docusaurus | `railpack` | ✅ | https://test-docusaurus.ml.ink | Fixed: publish_directory=build |
| 5 | `5/` | Next.js | `railpack` | ✅ | https://test-nextjs.ml.ink | |
| 6 | `6/` | SvelteKit | `railpack` | ❌ | https://test-sveltekit.ml.ink (503) | Proxy routing — container running but 503 from Cloudflare/Traefik |
| 7 | `7/` | Remix | `railpack` | ✅ | https://test-remix.ml.ink | |
| 8 | `8/` | Nuxt.js | `railpack` | ✅ | https://test-nuxtjs.ml.ink | Fixed: was resource exhaustion |
| 9 | `9/` | Astro (SSR) | `railpack` | ✅ | https://test-astro-ssr.ml.ink | |
| 10 | `10/` | Express.js | `railpack` | ✅ | https://test-express.ml.ink | |
| 11 | `11/` | Fastify | `railpack` | ✅ | https://test-fastify.ml.ink | |
| 12 | `12/` | FastAPI | `railpack` | ✅ | https://test-fastapi.ml.ink | |
| 13 | `13/` | Flask | `railpack` | ✅ | https://test-flask.ml.ink | |
| 14 | `14/` | Django | `railpack` | ✅ | https://test-django.ml.ink | |
| 15 | `15/` | Go (net/http) | `railpack` | ✅ | https://test-go-api.ml.ink | |
| 16 | `16/` | Go (Gin) | `railpack` | ✅ | https://test-go-gin.ml.ink | Fixed: was resource exhaustion |
| 17 | `17/` | Bun + Hono | `railpack` | ✅ | https://test-bun-hono.ml.ink | Fixed: was resource exhaustion |
| 18 | `18/` | Ruby on Rails | `railpack` | ❌ | https://test-rails.ml.ink (503) | start_command passed but 503 — proxy routing issue |
| 19 | `19/` | Spring Boot | `dockerfile` | ✅ | https://test-spring-boot.ml.ink | Fixed: was resource exhaustion |
| 20 | `20/` | Rust + Axum | `dockerfile` | ❌ | — | Build OK but WaitForRollout timeout — pod never ready |
| 21 | `21/` | Next.js full-stack | `railpack` | ✅ | https://test-nextjs-fullstack.ml.ink | Fixed: was resource exhaustion |
| 22 | `22/` | React + Express (mono) | `dockerfile` | ✅ | https://test-mono-re.ml.ink | Single service: backend serves API + React frontend |
| 23 | `23/` | React + FastAPI (mono) | `dockerfile` | ✅ | https://test-mono-rf.ml.ink | Single service: backend serves API + React frontend |
| 24 | `24/` | React + Go (mono) | `dockerfile` | ✅ | https://test-mono-rg.ml.ink | Single service: backend serves API + React frontend |
| 25 | `25/` | Streamlit | `railpack` | ✅ | https://test-streamlit.ml.ink | Fixed: was resource exhaustion |
| 26 | `26/` | Gradio | `railpack` | ✅ | https://test-gradio.ml.ink | Fixed: was resource exhaustion |
| 27 | `27/` | WebSocket (Node) | `railpack` | ✅ | https://test-websocket.ml.ink | Fixed: was resource exhaustion |
| 28 | `28/` | T3 Stack | `railpack` | ✅ | https://test-t3-stack.ml.ink | |
| 29 | `29/` | Flask (Dockerfile) | `dockerfile` | ✅ | https://test-flask-dockerfile.ml.ink | Validates dockerfile build pack with Python |
| 30 | `30/` | Plain HTML + assets | `static` | ✅ | https://test-plain-html.ml.ink | No build step — raw HTML/CSS/JS served via nginx |

**27 ✅** | **3 ❌** (all builds succeed, failures are runtime/routing)

---

## Static Sites (require `publish_directory`)

These frameworks build to static files and need `publish_directory` set so the platform uses the railpack static build flow (build app, then serve output via nginx on port 8080):

| Stack | `publish_directory` |
|-------|-------------------|
| React + Vite | `dist` |
| Vue + Vite | `dist` |
| Astro (static) | `dist` |
| Docusaurus | `build` |

---

## Full-Stack / Monorepo Apps

Full-stack tests (#22-24) use a single `dockerfile` service with a multi-stage Docker build:
1. **Stage 1**: Build React frontend with Vite (`npm run build` → `dist/`)
2. **Stage 2**: Build/install backend
3. **Runtime**: Backend serves both the API (`/api/items`) and the built React static files (`/`)

The React frontend fetches `/api/items` on load and renders the response — demonstrating real end-to-end connectivity (not two independent hello-world services).

---

## Remaining Failures

### #6 — SvelteKit (503 despite "running")

Build succeeds, `runtime_status: running`, but Cloudflare/Traefik returns 503. Traffic never reaches the pod. Not a port issue (adapter-node listens on PORT/3000). Needs Ingress/Traefik investigation.

### #18 — Ruby on Rails (503 despite "running")

Build succeeds with `start_command` fix, container running, but 503 from proxy. Same symptom as SvelteKit — likely same root cause.

### #20 — Rust + Axum (WaitForRollout timeout)

Build succeeds but pod never becomes ready within the 3-minute timeout. May need longer timeout or investigation into why the container doesn't start.
