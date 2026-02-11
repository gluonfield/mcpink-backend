# k8s Base Manifests

These manifests are the declarative base for the Deploy MCP k3s cluster.

- `dp-system.yml` defines the system namespace.
- `deployer-server.yml` exposes webhook ingress (`api.ml.ink`) and `/healthz` for deployment control-plane traffic.
- `deployer-worker.yml` runs the `k8s-native` deployment worker queue.
- `registry.yml`, `registry-gc.yml`, `buildkit.yml` pin infra workloads to ops/build pools.
- `gvisor-runtimeclass.yml` constrains customer pods to run nodes.
- `wildcard-cert.yml` and `traefik-tlsstore.yml` set default wildcard TLS.
- `loki-ingress.yml` and `grafana-ingress.yml` expose observability endpoints through Traefik.

Secrets are provisioned by Ansible (`infra/ansible/roles/k8s_addons`) from inventory/vault vars.
Cloudflare LB is the source of truth for public ingress host/origin routing.

## Namespace Model

Each project gets its own namespace: `dp-{tenant}-{project}`. All services within a project share the namespace, which means they can reach each other via K8s DNS (e.g. `http://my-backend:3000` from a frontend pod). Cross-namespace traffic is blocked by network policy.

## Security Architecture

### gVisor: the primary isolation boundary

All customer pods run under gVisor (`runtimeClassName: gvisor`). gVisor interprets every syscall in a userspace kernel — processes inside the sandbox never touch the host kernel directly. This is the same model as GKE Sandbox. Root inside gVisor is a fundamentally different threat than root inside runc.

### What applies to ALL customer pods

| Control | How | Why |
|---|---|---|
| gVisor sandbox | `runtimeClassName: gvisor` | Syscall interception — the real isolation boundary |
| No K8s API access | `automountServiceAccountToken: false` | Pods can't talk to the K8s API |
| No privilege escalation | `allowPrivilegeEscalation: false` | Sets `no_new_privs` — safe for all images including root-based ones |
| Writable root FS | `readOnlyRootFilesystem: false` | Many images write to /tmp, /var, etc. |
| Network isolation | NetworkPolicy (ingress + egress) | Block RFC1918, metadata API; allow only Traefik ingress + DNS |
| PSS baseline | Namespace labels | K8s rejects privileged pods, hostNetwork, hostPID, hostPath, etc. |

One SecurityContext for all build packs — no per-profile split. Capabilities are not dropped because that breaks root-based images (nginx, postgres, redis) that need `CAP_SETUID`/`CAP_SETGID`. Inside gVisor, caps only affect the emulated kernel anyway.

### Pod Security Standards: Baseline enforcement

Customer namespaces carry `pod-security.kubernetes.io/enforce: baseline` labels. This is a K8s-native admission controller that acts as a safety net independent of the Go code in `k8sdeployments/`:
- **Allows**: root, Docker default capabilities, privilege escalation — needed for compatibility.
- **Blocks**: `privileged: true`, hostNetwork, hostPID, hostIPC, hostPath volumes, dangerous sysctls — things that would bypass gVisor or expose the host.

This ensures that even if the Go code has a bug, K8s itself rejects dangerous pod specs before they reach the kubelet.
