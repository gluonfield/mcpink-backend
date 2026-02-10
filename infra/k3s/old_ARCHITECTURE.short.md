# ARCHITECTURE.short.md (v0) [DRAFT OLD]

- orchestrator: k3s
- ingress: edge nodes + stable IP (provider failover IP) + ingress controller
- dns: ExternalDNS
- tls: cert-manager
- builds: Railpack + BuildKit (dedicated build nodes)
- registry: Harbor (or OCI registry)
- runtime isolation: gVisor (RuntimeClass)
- storage: Longhorn
- autoscaling: (option) Knative Serving OR KEDA + KEDA HTTP Add-on
- progressive delivery: Argo Rollouts
- gitops: Argo CD
- logs: Loki
- metrics: Prometheus
- backups: Velero + Longhorn backups + DB-native backups
- workflows/control: Go + Temporal + Postgres
