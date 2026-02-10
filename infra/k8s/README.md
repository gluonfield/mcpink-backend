# k8s Base Manifests

These manifests are the declarative base for the Deploy MCP k3s cluster.

- `dp-system.yml` defines the system namespace.
- `registry.yml`, `registry-gc.yml`, `buildkit.yml` pin infra workloads to ops/build pools.
- `temporal-worker.yml` deploys the Temporal worker on the control pool.
- `gvisor-runtimeclass.yml` constrains customer pods to run nodes.
- `wildcard-cert.yml` and `traefik-tlsstore.yml` set default wildcard TLS.
- `loki-ingress.yml` and `grafana-ingress.yml` expose observability endpoints through Traefik.

Secrets are provisioned by Ansible (`infra/ansible/roles/k8s_addons`) from inventory/vault vars.
