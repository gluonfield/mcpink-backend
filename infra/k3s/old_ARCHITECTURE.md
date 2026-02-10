# ARCHITECTURE.md (v0) [DRAFT OLD]

> v0 is intentionally option-heavy. v1 should delete options, lock choices, and add concrete implementation details.

_Last updated: 2026-02-07_

## Requirements (non-negotiable)

### Platform goals

1. **Bare metal first** (Hetzner + other regions). No managed cloud PaaS dependencies required for core runtime.
2. **Single API transaction**: app + domain + TLS + secrets + (optional) data volume + (optional) database.
3. **Multi-node reliability**: if a run node dies, workloads **reschedule** onto healthy nodes automatically.
4. **No per-node DNS**: customers get stable hostnames; node failures must not require manual DNS edits.
5. **Separation of concerns**: dedicated **build** machines and dedicated **run** machines.
6. **Build from git**: GitHub + private Git (Gitea/Forgejo, GitLab, etc.).
7. **Buildpacks-first**: **Railpack** is the default builder; Dockerfile as escape hatch.
8. **Logs are first-class**: programmatic access to **build logs** and **runtime logs**.
9. **Untrusted code isolation**: run user code with **gVisor** or **Firecrackers**.
10. **Persistent storage primitive**: each app can request durable storage (size, class), attach/mount it, and keep it across deploys/reschedules.
11. **Horizontal scaling**: manual replica scaling + autoscaling where safe.
12. **Progressive delivery**: blue/green (or equivalent) with fast rollback.
13. **Backups**: platform state + customer data have explicit backup/restore stories.
14. **Move fast without foot-guns**: minimal ops burden, strong defaults, and guardrails.

### Non-goals (for v0)

- A full “Kubernetes dashboard for users.” Your external surface is MCP/API; internal ops UI can be separate.
- Solving perfect sandboxing. v0 focuses on strong isolation + network controls; iterate.
- Multi-region active-active routing (v0 can be single-region per app).

---

## High-level shape (keep your 3-plane design)

### Plane A — Control Plane (your product)

- **Deploy MCP API** (Go)
- **Auth**: GitHub OAuth + GitHub App
- **Workflows**: Temporal (idempotent deployments)
- **State**: Postgres (metadata, audit, tenancy, billing)

### Plane B — Factory (build plane)

- **Railpack builds** on dedicated build nodes
- **BuildKit** as the build engine
- **Registry** (durable) for artifacts

### Plane C — Muscle (run plane)

- **k3s cluster** with dedicated run nodes
- **Ingress** at stable edge
- **RuntimeClass** for **gVisor**
- **Persistent volumes** via Longhorn

---

## Stack (v0)

### Orchestrator (winner)

- **k3s**

Why:

- You want a real scheduler (placement, rescheduling, horizontal scaling) without installing “full Kubernetes the hard way.”

Operational notes:

- Separate node pools with labels/taints:
  - `pool=build` (tainted) for builds
  - `pool=run` (tainted) for customer workloads
  - `pool=edge` (tainted) for ingress + scale-to-zero frontends

---

## Categories & choices

### 1) Ingress & traffic entry (stable endpoint)

**Goal:** customers hit one stable endpoint per region; apps can move between nodes without DNS changes.

Options:

1. **Edge nodes + stable IP (recommended on bare metal)**
   - Run ingress controller only on `pool=edge`.
   - Attach a stable IP mechanism (e.g., provider failover IP) to the active edge node(s).
   - Cert-manager terminates TLS at ingress.
   - **Certainty:** High.

2. **MetalLB (LoadBalancer on bare metal)**
   - Works well in true L2/BGP-capable networks, but can be tricky depending on the provider network.
   - **Certainty:** Medium (provider-network dependent).

Notes:

- Keep ingress simple: one controller per cluster; don’t run “a proxy per node” like Coolify.

---

### 2) DNS automation & certificates

**Goal:** `create_app()` can create a hostname and TLS without manual DNS work.

Options:

1. **ExternalDNS + cert-manager (recommended)**
   - ExternalDNS creates/updates DNS records from Kubernetes resources.
   - cert-manager issues/renews ACME certificates.
   - **Certainty:** High.

2. Provider-specific DNS API integration in your Control Plane
   - Only if you want tighter control or don’t trust cluster credentials.
   - **Certainty:** Medium.

---

### 3) Persistent storage (customer “disk” primitive)

**Goal:** apps can request durable volumes and keep them across reschedules.

Winner:

- **Longhorn** (distributed block storage for Kubernetes)

Notes:

- Define a small set of storage classes (e.g., `fast`, `standard`) mapped to Longhorn settings.
- Expose volume ops via your MCP/API:
  - create volume (size/class)
  - attach to app (mountPath)
  - snapshot/backup policy

Certainty: High (this is a common bare-metal pattern), but storage is always the sharpest knife.

---

### 4) Build system (source → image)

Winner:

- **Railpack + BuildKit**

How it integrates:

- Factory runs build Jobs that:
  1. Fetch source (GitHub App installation token or private git creds)
  2. Run `railpack` against a BuildKit daemon (remote or in-cluster)
  3. Push image to registry with immutable tag (commit SHA)
  4. Emit build logs + artifact metadata back to Plane A

Options for “where BuildKit lives”:

1. **Dedicated BuildKit daemon on build nodes**
   - Stable, fast, easy to observe.
   - **Certainty:** High.

2. **Ephemeral BuildKit per build (Job starts buildkitd)**
   - Cleaner isolation; more overhead.
   - **Certainty:** Medium.

---

### 5) Image registry

**Goal:** run nodes pull built images reliably; images are cache/artifacts, but must be available.

Options:

1. **Harbor**
   - Strong enterprise features (RBAC, scanning, replication).
   - **Certainty:** High.

2. **Plain OCI registry (distribution) + object storage backend**
   - Minimal; fewer features.
   - **Certainty:** Medium.

---

### 6) Untrusted workload isolation

**Goal:** run customer code with a smaller attack surface.

Winner (v0):

- **gVisor** via Kubernetes RuntimeClass

Options:

1. **gVisor (runsc)**
   - Good isolation for container workloads; pragmatic.
   - **Certainty:** High.

2. **Firecracker-class isolation (future)**
   - Usually via Kata Containers / microVM approach.
   - **Certainty:** Medium (depends on your appetite for complexity).

Mandatory v0 guardrails:

- NetworkPolicy defaults deny-all for customer namespaces.
- Egress allow-list (block metadata endpoints, SMTP, mining pools, etc.).
- Pod Security standards + drop capabilities.

---

### 7) Autoscaling & “scale to zero”

**Goal:** SSR / admin panels can sleep and wake on HTTP requests.

Options:

1. **Knative Serving**
   - Full request-driven model, scale-to-zero, revisions, traffic splitting.
   - **Certainty:** High (mature), but higher operational complexity.

2. **KEDA + KEDA HTTP Add-on**
   - Lighter than Knative for “just HTTP scale-to-zero,” but the HTTP add-on remains more niche.
   - **Certainty:** Medium.

Pragmatic v0 suggestion:

- Start with **min replicas = 1** for most SSR apps.
- Introduce scale-to-zero only for clearly idle workloads and only after observability is solid.

---

### 8) Deploy strategies (blue/green)

Options:

1. **Argo Rollouts** (recommended)
   - Native blue/green strategy and safe promotion.
   - **Certainty:** High.

2. **Plain Kubernetes + two Deployments + Service switch**
   - DIY; works but you’ll rebuild tooling.
   - **Certainty:** Medium.

---

### 9) GitOps / desired state delivery

Options:

1. **Argo CD** (recommended for your case)
   - Strong UI, multi-cluster story, pairs naturally with Argo Rollouts.
   - **Certainty:** High.

2. **Flux CD**
   - Excellent GitOps toolkit; more “components and composition.”
   - **Certainty:** High.

Platform note:

- Your MCP API is the user interface; GitOps is for **ops/reproducibility**. Choose the one your team can operate calmly at 3AM.

---

### 10) Logs & observability

**Goal:** agent/debugger can fetch logs by app/deployment and not SSH anywhere.

Options:

1. **Loki for logs + Prometheus for metrics**
   - Widely adopted, straightforward.
   - **Certainty:** High.

2. **OpenTelemetry Collector + vendor/hosted backend**
   - Great for traces; potentially more moving parts.
   - **Certainty:** Medium.

Implementation notes:

- Tag every log line with: `tenant`, `project`, `app`, `deployment_id`, `commit_sha`.
- Build logs are separate stream from runtime logs.

---

### 11) Backups & disaster recovery

Options:

1. **Velero** (cluster objects + PV snapshots/backup integrations)
   - **Certainty:** High.

2. **Storage-layer backups (Longhorn snapshots/backups) + DB-native backups**
   - Often necessary anyway; not sufficient alone.
   - **Certainty:** High.

DR drills (v0):

- Practice restore of:
  - Plane A DB + secrets
  - registry (or rebuild policy)
  - one stateful app PV

---

## Control-plane API model (what your product exposes)

### Core objects

- **Tenant** → **Project** → **App** → **Deployment**
- **Volume** (durable disk)
- **Secret** (env + file secrets)
- **Domain** (hostname + TLS)
- **Build** (source → image)

### “One transaction” workflow (ideal)

1. Validate repo access (GitHub App installation token)
2. Allocate:
   - namespace
   - secrets
   - domain + cert
   - volume (optional)
3. Build image (Railpack)
4. Deploy:
   - create Rollout/Deployment
   - wait ready
   - switch traffic (blue/green)
5. Return URL

Idempotency keys:

- workflow id = `tenant/project/app + commit_sha`

---

## Security posture (minimum viable for untrusted code)

- Namespaces per tenant (or per project) with:
  - ResourceQuota + LimitRange
  - default-deny NetworkPolicy
- RuntimeClass = gVisor for customer workloads
- Admission controls:
  - reject privileged pods
  - enforce non-root
  - enforce read-only rootfs (when possible)
- Egress controls:
  - block cloud metadata IPs
  - block SMTP by default
- Audit:
  - every MCP call stored with actor, request id, and resulting k8s object refs

---

## What this v0 intentionally does _not_ decide

- Exact CNI (networking) choice
- Exact registry choice
- Exact scale-to-zero mechanism

Those become v1 decisions after you validate the simplest end-to-end slice:
`git push → build → deploy → logs → rollback → node failure reschedule → volume persists`.
