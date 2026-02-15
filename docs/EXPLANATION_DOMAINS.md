# Custom Domains - Complete Implementation Reference

This document explains exactly how custom domains work in MCPDeploy, from user-facing API to infrastructure. It covers the database schema, Go business logic, DNS verification, Temporal workflows, K8s resource creation, TLS certificate provisioning, and traffic routing through Hetzner + Traefik.

---

## Infrastructure Overview

### Cluster Topology (Single k3s Cluster)

| Node | Public IP | Private IP | Pool | Specs | Role |
|------|-----------|------------|------|-------|------|
| k3s-1 | 46.225.100.234 | 10.0.0.4 | ctrl | 8GB | Control plane |
| build-1 | 46.225.92.127 | 10.0.0.3 | build | 16GB | BuildKit (taint: `pool=build:NoSchedule`) |
| ops-1 | 116.202.163.209 | 10.0.1.4 | ops | Dedicated Xeon | Registry/Gitea/Monitoring (taint: `pool=ops:NoSchedule`) |
| run-1 | 157.90.130.187 | 10.0.1.3 | run | 256GB EPYC | Customer workloads + Traefik |

### Hetzner Load Balancer

- **Public IP**: `46.225.35.234`
- **Type**: lb11 in fsn1
- **Services**: TCP passthrough only (no TLS termination)
  - Port 80 -> run-1:80
  - Port 443 -> run-1:443
- **Why TCP passthrough**: cert-manager needs raw HTTP-01 challenges on port 80. A TLS-terminating LB would break certificate provisioning.

### Traffic Flow

```
Browser → DNS lookup (user's registrar)
       → CNAME: app.customer.com → my-service.cname.ml.ink
       → A record: *.cname.ml.ink → 46.225.35.234 (Hetzner LB)
       → Hetzner LB: TCP passthrough port 443
       → run-1 node (10.0.1.3)
       → Traefik DaemonSet (hostNetwork, port 443)
       → TLS termination (using cert from K8s Secret)
       → Route by Host header → Ingress match
       → K8s Service → Customer Pod (gVisor sandbox)
```

### Firewall Rules on Run Nodes

Only the Hetzner LB and Cloudflare IPs can reach ports 80/443 on run nodes:

```yaml
# infra/eu-central-1/inventory/group_vars/run.yml
traefik_public_allowed_sources:
  - 46.225.35.234/32    # Hetzner LB (custom domain traffic)
  # Cloudflare IPv4 ranges (*.ml.ink traffic via Cloudflare LB)
  - 173.245.48.0/20
  - 103.21.244.0/22
  - 103.22.200.0/22
  - 103.31.4.0/22
  - 141.101.64.0/18
  - 108.162.192.0/18
  - 190.93.240.0/20
  - 188.114.96.0/20
  - 197.234.240.0/22
  - 198.41.128.0/17
  - 162.158.0.0/15
  - 104.16.0.0/13
  - 104.24.0.0/14
  - 172.64.0.0/13
  - 131.0.72.0/22
```

---

## Traefik Configuration

Traefik runs as a **DaemonSet** on all `pool=run` nodes with `hostNetwork: true`, binding directly to ports 80 and 443.

```yaml
# infra/eu-central-1/k8s/values/traefik-values.yml
deployment:
  kind: DaemonSet

nodeSelector:
  pool: run

hostNetwork: true

ingressClass:
  enabled: true
  isDefaultClass: true
  name: traefik

providers:
  kubernetesIngress:
    enabled: true
  kubernetesCRD:
    enabled: true

ports:
  web:
    port: 80
    hostPort: 80
  websecure:
    port: 443
    hostPort: 443

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

---

## TLS Certificate Strategy

Two distinct certificate strategies exist:

### 1. Platform Domains (`*.ml.ink`) - Wildcard via DNS-01

A single wildcard certificate covers all `*.ml.ink` subdomains:

```yaml
# infra/eu-central-1/k8s/system/wildcard-cert.yml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: wildcard-apps
  namespace: dp-system
spec:
  secretName: wildcard-apps-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - "*.ml.ink"
```

This wildcard cert is set as Traefik's default via TLSStore:

```yaml
# infra/eu-central-1/k8s/system/traefik-tlsstore.yml
apiVersion: traefik.io/v1alpha1
kind: TLSStore
metadata:
  name: default
  namespace: dp-system
spec:
  defaultCertificate:
    secretName: wildcard-apps-tls
```

Standard service Ingresses do **NOT** have a TLS section - they rely on this default wildcard cert.

### 2. Custom Domains (`app.customer.com`) - Per-Domain via HTTP-01

Each custom domain gets its own certificate via Let's Encrypt HTTP-01 challenge. The cert-manager ClusterIssuer handles both:

```yaml
# infra/eu-central-1/k8s/system/cert-manager-issuer.yml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: ops@ml.ink
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
      # DNS-01 for *.ml.ink wildcard (via Cloudflare API)
      - dns01:
          cloudflare:
            email: ops@ml.ink
            apiTokenSecretRef:
              name: cloudflare-api-token
              key: api-token
        selector:
          dnsZones:
            - ml.ink
      # HTTP-01 for custom domains (via Traefik temporary ingress)
      - http01:
          ingress:
            ingressClassName: traefik
            ingressTemplate:
              metadata:
                annotations:
                  traefik.ingress.kubernetes.io/router.entrypoints: web
```

The `traefik.ingress.kubernetes.io/router.entrypoints: web` annotation ensures the ACME challenge ingress only listens on port 80 (no TLS redirect).

---

## K8s Templates (Design Spec)

These YAML templates define the expected K8s resources. The Go code in `k8sdeployments/` must produce equivalent resources.

### Standard Service (platform subdomain)

```yaml
# infra/eu-central-1/k8s/templates/customer-service-template.yml

# Secret for env vars
apiVersion: v1
kind: Secret
metadata:
  name: "{{ service }}-env"
  namespace: "dp-{{ tenant }}-{{ project }}"
type: Opaque
stringData:
  PORT: "{{ port }}"
---
# Deployment with gVisor
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "{{ service }}"
  namespace: "dp-{{ tenant }}-{{ project }}"
  labels:
    app: "{{ service }}"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: "{{ service }}"
  template:
    metadata:
      labels:
        app: "{{ service }}"
    spec:
      runtimeClassName: gvisor          # REQUIRED: security sandbox
      automountServiceAccountToken: false
      containers:
        - name: app
          image: "registry.internal:5000/dp-{{ tenant }}-{{ project }}/{{ service }}:{{ sha }}"
          ports:
            - containerPort: !!int "{{ port }}"
          envFrom:
            - secretRef:
                name: "{{ service }}-env"
          resources:
            requests:
              cpu: "{{ vcpus }}"
              memory: "{{ memory }}"
            limits:
              cpu: "{{ vcpus }}"
              memory: "{{ memory }}"
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: false
---
# ClusterIP Service
apiVersion: v1
kind: Service
metadata:
  name: "{{ service }}"
  namespace: "dp-{{ tenant }}-{{ project }}"
spec:
  selector:
    app: "{{ service }}"
  ports:
    - port: !!int "{{ port }}"
      targetPort: !!int "{{ port }}"
---
# Ingress for *.ml.ink (NO TLS section - uses wildcard default)
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: "{{ service }}"
  namespace: "dp-{{ tenant }}-{{ project }}"
spec:
  ingressClassName: traefik
  rules:
    - host: "{{ name }}.ml.ink"
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: "{{ service }}"
                port:
                  number: !!int "{{ port }}"
```

### Custom Domain Ingress (separate resource)

```yaml
# infra/eu-central-1/k8s/templates/customer-custom-domain-template.yml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: "{{ service }}-cd"
  namespace: "dp-{{ tenant }}-{{ project }}"
spec:
  ingressClassName: traefik
  tls:                              # HAS TLS section (per-domain cert)
    - hosts:
        - "{{ custom_domain }}"
      secretName: "{{ service }}-cd-tls"
  rules:
    - host: "{{ custom_domain }}"
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: "{{ service }}"
                port:
                  number: !!int "{{ port }}"
```

**Key differences from standard Ingress:**
- Name: `{{ service }}-cd` (suffix `-cd` for "custom domain")
- Has `tls` section with per-domain secret
- Host is the custom domain, not `*.ml.ink`
- Same backend Service (routes to same pods)

---

## Database Schema

### custom_domains Table

```sql
-- go-backend/internal/storage/pg/migrations/0001_initial.sql (lines 201-216)
CREATE TABLE custom_domains (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    domain TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending_dns',
    expected_record_target TEXT NOT NULL,
    expires_at TIMESTAMPTZ DEFAULT (NOW() + INTERVAL '7 days'),
    verified_at TIMESTAMPTZ,
    last_checked_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(service_id),   -- One custom domain per service
    UNIQUE(domain)        -- One service per domain (prevents squatting)
);
```

```sql
-- go-backend/internal/storage/pg/migrations/0002_custom_domain_verification.sql
ALTER TABLE custom_domains ADD COLUMN verification_token TEXT NOT NULL DEFAULT '';
```

### Status Progression

```
pending_dns → provisioning → active
                           → failed (retryable)
```

- `pending_dns`: User needs to configure DNS records (7-day expiry)
- `provisioning`: DNS verified, Temporal workflow creating K8s resources
- `active`: TLS cert ready, Ingress created, domain is live
- `failed`: Certificate or DNS verification failed (user can retry)

### Constraints

- `UNIQUE(service_id)`: One custom domain per service. Must remove before adding another.
- `UNIQUE(domain)`: Domain globally unique. Prevents squatting across services.
- `ON DELETE CASCADE`: Removing a service auto-deletes its custom domain record.
- `expires_at`: 7-day window. Unclaimed `pending_dns` records auto-expire (reclaimable).

### SQL Queries

```sql
-- go-backend/internal/storage/pg/queries/customdomains/customdomains.sql

-- name: CreateCustomDomain :one
INSERT INTO custom_domains (service_id, domain, expected_record_target, verification_token)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetByServiceID :one
SELECT * FROM custom_domains WHERE service_id = $1;

-- name: GetByDomain :one
SELECT * FROM custom_domains WHERE lower(domain) = lower($1);

-- name: UpdateStatus :one
UPDATE custom_domains SET status = $2, last_checked_at = NOW(), updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: UpdateVerified :one
UPDATE custom_domains SET status = 'active', verified_at = NOW(), updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: UpdateError :one
UPDATE custom_domains SET last_error = $2, last_checked_at = NOW(), updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: Delete :exec
DELETE FROM custom_domains WHERE id = $1;

-- name: DeleteByServiceID :exec
DELETE FROM custom_domains WHERE service_id = $1;

-- name: ExpireStale :exec
UPDATE custom_domains SET status = 'failed', last_error = 'expired'
WHERE status = 'pending_dns' AND expires_at < NOW();
```

### Clusters Table (Referenced by Custom Domain Flow)

```sql
-- go-backend/internal/storage/pg/migrations/0003_clusters.sql
CREATE TABLE clusters (
    region TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    task_queue TEXT NOT NULL UNIQUE,
    apps_domain TEXT NOT NULL,
    cname_target TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Pre-seeded:
INSERT INTO clusters (region, name, task_queue, apps_domain, cname_target)
VALUES ('eu-central-1', 'Europe Central 1', 'deployer-eu-central-1', 'ml.ink', 'cname.ml.ink');
```

The `cname_target` field is used to compute per-service CNAME targets: `<service-name>.<cname_target>` (e.g., `my-app.cname.ml.ink`).

---

## DNS Verification Module

```go
// go-backend/internal/dnsverify/verify.go

// VerifyCNAME checks that the user's domain has a CNAME pointing to the expected target.
func VerifyCNAME(domain, expectedTarget string) (bool, error) {
    cname, err := net.LookupCNAME(domain)
    if err != nil {
        return false, fmt.Errorf("CNAME lookup failed for %s: %w", domain, err)
    }
    normalized := NormalizeDomain(cname)
    expected := NormalizeDomain(expectedTarget)
    return normalized == expected, nil
}

// VerifyTXT checks for a TXT record at _dp-verify.<domain> containing dp-verify=<token>.
func VerifyTXT(domain, expectedToken string) (bool, error) {
    host := "_dp-verify." + NormalizeDomain(domain)
    records, err := net.LookupTXT(host)
    if err != nil {
        return false, fmt.Errorf("TXT lookup failed for %s: %w", host, err)
    }
    needle := "dp-verify=" + expectedToken
    for _, r := range records {
        if strings.TrimSpace(r) == needle {
            return true, nil
        }
    }
    return false, nil
}

// GenerateVerificationToken creates a random 32-char hex token.
func GenerateVerificationToken() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}

// NormalizeDomain lowercases and strips trailing dots.
func NormalizeDomain(d string) string {
    d = strings.TrimSpace(d)
    d = strings.ToLower(d)
    d = strings.TrimSuffix(d, ".")
    return d
}

// ValidateCustomDomain rejects wildcards, apex domains, and platform subdomains.
func ValidateCustomDomain(domain, platformDomain string) error {
    domain = NormalizeDomain(domain)
    if domain == "" {
        return fmt.Errorf("domain is required")
    }
    if strings.Contains(domain, "*") {
        return fmt.Errorf("wildcard domains are not supported")
    }
    if strings.HasSuffix(domain, "."+NormalizeDomain(platformDomain)) || domain == NormalizeDomain(platformDomain) {
        return fmt.Errorf("cannot use a %s subdomain as a custom domain", platformDomain)
    }
    if !strings.Contains(domain, ".") {
        return fmt.Errorf("apex domains are not supported; use a subdomain (e.g. app.%s)", domain)
    }
    parts := strings.Split(domain, ".")
    for _, part := range parts {
        if part == "" {
            return fmt.Errorf("invalid domain format")
        }
    }
    return nil
}

// DNSInstructions returns user-facing instructions for configuring DNS.
func DNSInstructions(domain, cnameTarget, verificationToken string) string {
    return fmt.Sprintf(
        "Add the following DNS records for your domain:\n\n"+
            "1. Ownership verification (TXT):\n"+
            "   Host: _dp-verify.%s\n"+
            "   Type: TXT\n"+
            "   Value: dp-verify=%s\n\n"+
            "2. Routing (CNAME):\n"+
            "   Host: %s\n"+
            "   Type: CNAME\n"+
            "   Value: %s\n\n"+
            "Important: If using Cloudflare, set the CNAME record to DNS-only (gray cloud), not Proxied.\n"+
            "After configuring both DNS records, call verify_custom_domain to activate it.",
        domain, verificationToken, domain, cnameTarget,
    )
}
```

### DNS Records Required from User

For a service `my-app` with domain `app.customer.com`:

| Type | Host | Value | Purpose |
|------|------|-------|---------|
| TXT | `_dp-verify.app.customer.com` | `dp-verify=<32-char-hex-token>` | Ownership proof |
| CNAME | `app.customer.com` | `my-app.cname.ml.ink` | Traffic routing |

The CNAME target `my-app.cname.ml.ink` resolves (via `*.cname.ml.ink` wildcard A record at Cloudflare) to `46.225.35.234` (Hetzner LB).

---

## Business Logic Layer

### AddCustomDomain

```go
// go-backend/internal/deployments/service.go (lines 574-638)

func (s *Service) AddCustomDomain(ctx context.Context, params AddCustomDomainParams) (*AddCustomDomainResult, error) {
    // 1. Normalize domain (lowercase, strip trailing dot)
    domain := dnsverify.NormalizeDomain(params.Domain)

    // 2. Fetch service (verifies user owns it)
    svc, err := s.GetServiceByName(ctx, GetServiceByNameParams{
        Name: params.Name, Project: params.Project, UserID: params.UserID,
    })

    // 3. Get cluster config for the service's region
    cluster, ok := s.clusters[svc.Region]

    // 4. Validate: no wildcards, no apex, no platform subdomains
    if err := dnsverify.ValidateCustomDomain(domain, cluster.AppsDomain); err != nil {
        return nil, err
    }

    // 5. Anti-squat: reclaim expired/failed domains
    existing, err := s.customDomainsQ.GetByDomain(ctx, domain)
    if err == nil {
        canReclaim := existing.Status == "failed" ||
            (existing.Status == "pending_dns" && existing.ExpiresAt.Valid &&
             existing.ExpiresAt.Time.Before(time.Now()))
        if canReclaim {
            _ = s.customDomainsQ.Delete(ctx, existing.ID)
        } else {
            return nil, fmt.Errorf("domain %s is already attached to a service", existing.Domain)
        }
    }

    // 6. Check service doesn't already have a custom domain
    _, err = s.customDomainsQ.GetByServiceID(ctx, svc.ID)
    if err == nil {
        return nil, fmt.Errorf("service %s already has a custom domain; remove it first", params.Name)
    }

    // 7. Generate verification token + compute per-service CNAME target
    verificationToken := dnsverify.GenerateVerificationToken()
    perServiceTarget := *svc.Name + "." + cluster.CnameTarget
    // e.g., "my-app" + "." + "cname.ml.ink" = "my-app.cname.ml.ink"

    // 8. Create DB record (status: pending_dns, expires in 7 days)
    cd, err := s.customDomainsQ.CreateCustomDomain(ctx, customdomains.CreateCustomDomainParams{
        ServiceID:            svc.ID,
        Domain:               domain,
        ExpectedRecordTarget: perServiceTarget,
        VerificationToken:    verificationToken,
    })

    // 9. Generate DNS instructions for user
    instructions := dnsverify.DNSInstructions(domain, perServiceTarget, verificationToken)

    return &AddCustomDomainResult{
        ServiceID: svc.ID, Domain: domain,
        Status: cd.Status, Instructions: instructions,
    }, nil
}
```

**No K8s resources are created yet.** The user must configure DNS and call `verify_custom_domain`.

### VerifyCustomDomain

```go
// go-backend/internal/deployments/service.go (lines 653-769)

func (s *Service) VerifyCustomDomain(ctx context.Context, params VerifyCustomDomainParams) (*VerifyCustomDomainResult, error) {
    svc, err := s.GetServiceByName(ctx, ...)
    cd, err := s.customDomainsQ.GetByServiceID(ctx, svc.ID)

    // Short-circuit if already done
    if cd.Status == "active" {
        return &VerifyCustomDomainResult{Status: "active", Message: "Custom domain is already active"}, nil
    }
    if cd.Status == "provisioning" {
        return &VerifyCustomDomainResult{Status: "provisioning", Message: "please wait"}, nil
    }

    // Step 1: Verify TXT ownership record
    txtOK, txtErr := dnsverify.VerifyTXT(cd.Domain, cd.VerificationToken)
    if txtErr != nil || !txtOK {
        // Update DB with error, return failure message
        return &VerifyCustomDomainResult{Status: cd.Status, Message: errMsg}, nil
    }

    // Step 2: Verify CNAME routing record
    cnameOK, cnameErr := dnsverify.VerifyCNAME(cd.Domain, cd.ExpectedRecordTarget)
    if cnameErr != nil || !cnameOK {
        // Update DB with error, return failure message
        return &VerifyCustomDomainResult{Status: cd.Status, Message: errMsg}, nil
    }

    // Step 3: Both passed -> transition to provisioning
    s.customDomainsQ.UpdateStatus(ctx, customdomains.UpdateStatusParams{
        ID: cd.ID, Status: "provisioning",
    })

    // Step 4: Compute K8s namespace and service name
    cluster, _ := s.clusters[svc.Region]
    user, _ := s.usersQ.GetUserByID(ctx, svc.UserID)
    proj, _ := s.projectsQ.GetProjectByID(ctx, svc.ProjectID)
    namespace := k8sdeployments.NamespaceName(user.ID, proj.Ref)   // "dp-<userID>-<projectRef>"
    serviceName := k8sdeployments.ServiceName(*svc.Name)           // K8s-safe name

    // Step 5: Kick off Temporal workflow (async)
    workflowID := fmt.Sprintf("attach-cd-%s", cd.ID)
    s.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
        ID:        workflowID,
        TaskQueue: cluster.TaskQueue,   // "deployer-eu-central-1" → runs on deployer-worker in cluster
    }, k8sdeployments.AttachCustomDomainWorkflow, k8sdeployments.AttachCustomDomainWorkflowInput{
        CustomDomainID: cd.ID,
        ServiceID:      svc.ID,
        Namespace:      namespace,
        ServiceName:    serviceName,
        CustomDomain:   cd.Domain,
    })

    return &VerifyCustomDomainResult{
        Status: "provisioning", Message: "DNS verified! Provisioning TLS certificate...",
    }, nil
}
```

### RemoveCustomDomain

```go
// go-backend/internal/deployments/service.go (lines 782-839)

func (s *Service) RemoveCustomDomain(ctx context.Context, params RemoveCustomDomainParams) (*RemoveCustomDomainResult, error) {
    svc, err := s.GetServiceByName(ctx, ...)
    cd, err := s.customDomainsQ.GetByServiceID(ctx, svc.ID)

    cluster, _ := s.clusters[svc.Region]
    user, _ := s.usersQ.GetUserByID(ctx, svc.UserID)
    proj, _ := s.projectsQ.GetProjectByID(ctx, svc.ProjectID)
    namespace := k8sdeployments.NamespaceName(user.ID, proj.Ref)
    serviceName := k8sdeployments.ServiceName(*svc.Name)

    // Start cleanup workflow BEFORE deleting DB record
    workflowID := fmt.Sprintf("detach-cd-%s", cd.ID)
    s.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
        ID:        workflowID,
        TaskQueue: cluster.TaskQueue,
    }, k8sdeployments.DetachCustomDomainWorkflow, k8sdeployments.DetachCustomDomainWorkflowInput{
        CustomDomainID: cd.ID,
        ServiceID:      svc.ID,
        Namespace:      namespace,
        ServiceName:    serviceName,
    })

    // Then delete DB record
    s.customDomainsQ.Delete(ctx, cd.ID)

    return &RemoveCustomDomainResult{
        ServiceID: svc.ID,
        Message:   fmt.Sprintf("Custom domain %s removed", cd.Domain),
    }, nil
}
```

---

## Temporal Workflows

### AttachCustomDomainWorkflow (Certificate-First Pattern)

```go
// go-backend/internal/k8sdeployments/workflows.go (lines 310-389)

func AttachCustomDomainWorkflow(ctx workflow.Context, input AttachCustomDomainWorkflowInput) (AttachCustomDomainWorkflowResult, error) {
    // Short timeout for quick K8s operations
    shortCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval: time.Second, BackoffCoefficient: 2.0,
            MaximumInterval: 30 * time.Second, MaximumAttempts: 3,
        },
    })

    // Long timeout for certificate provisioning (Let's Encrypt can be slow)
    waitCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        HeartbeatTimeout:    30 * time.Second,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval: 5 * time.Second, BackoffCoefficient: 1.0,
            MaximumAttempts: 1,  // No retry - cert failures are usually permanent
        },
    })

    var activities *Activities

    // Helper: mark as failed in DB and return error
    markFailed := func(errMsg string) (AttachCustomDomainWorkflowResult, error) {
        _ = workflow.ExecuteActivity(shortCtx, activities.UpdateCustomDomainDBStatus,
            UpdateCustomDomainStatusInput{CustomDomainID: input.CustomDomainID, Status: "failed"}).Get(ctx, nil)
        return AttachCustomDomainWorkflowResult{Status: "failed", ErrorMessage: errMsg},
            fmt.Errorf("attach custom domain failed: %s", errMsg)
    }

    // === PHASE 1: Create Certificate CR (NO INGRESS YET) ===
    // CRITICAL: Ingress must NOT exist during cert provisioning.
    // If Ingress with TLS section exists, Traefik auto-redirects HTTP->HTTPS (308).
    // This breaks cert-manager's HTTP-01 ACME challenge because Let's Encrypt
    // follows HTTP on port 80 and does NOT follow 308 redirects.
    certName := input.ServiceName + "-cd"
    workflow.ExecuteActivity(shortCtx, activities.ApplyCustomDomainCertificate,
        ApplyCustomDomainCertificateInput{
            Namespace: input.Namespace, ServiceName: input.ServiceName, Domain: input.CustomDomain,
        }).Get(ctx, nil)

    // === PHASE 2: Wait for Certificate to be Ready ===
    // Polls cert.status.conditions[type=Ready] every 5 seconds.
    // Terminal failures (InvalidDomain, CAA, RateLimited, Denied) abort immediately.
    workflow.ExecuteActivity(waitCtx, activities.WaitForCertificateReady,
        WaitForCertificateReadyInput{
            Namespace: input.Namespace, CertificateName: certName,
        }).Get(ctx, nil)

    // === PHASE 3: Create Ingress WITH TLS ===
    // Safe now because cert already exists. Traefik 308 redirect is fine.
    workflow.ExecuteActivity(shortCtx, activities.ApplyCustomDomainIngress,
        ApplyCustomDomainIngressInput{
            Namespace: input.Namespace, ServiceName: input.ServiceName, Domain: input.CustomDomain,
        }).Get(ctx, nil)

    // === PHASE 4: Mark as Active in DB ===
    workflow.ExecuteActivity(shortCtx, activities.UpdateCustomDomainDBStatus,
        UpdateCustomDomainStatusInput{CustomDomainID: input.CustomDomainID, Status: "active"}).Get(ctx, nil)

    return AttachCustomDomainWorkflowResult{Status: "active"}, nil
}
```

**Why Certificate-First?** This is the most critical design decision:
1. Traefik v3 automatically redirects HTTP -> HTTPS (308) for any Ingress with a `tls` section
2. Let's Encrypt HTTP-01 challenge sends `GET http://domain/.well-known/acme-challenge/<token>` on port 80
3. If Traefik redirects this to HTTPS, Let's Encrypt **does not follow the redirect** -> challenge fails
4. Solution: Create the Certificate CR first (cert-manager creates a temporary challenge Ingress on port 80 with `entrypoints: web` annotation). Once cert is ready, create the real Ingress with TLS.

### DetachCustomDomainWorkflow

```go
// go-backend/internal/k8sdeployments/workflows.go (lines 391-420)

func DetachCustomDomainWorkflow(ctx workflow.Context, input DetachCustomDomainWorkflowInput) (DetachCustomDomainWorkflowResult, error) {
    actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
        StartToCloseTimeout: 2 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            InitialInterval: time.Second, BackoffCoefficient: 2.0,
            MaximumInterval: 30 * time.Second, MaximumAttempts: 3,
        },
    })

    var activities *Activities

    // Single activity: Delete Ingress, Certificate CR, and TLS Secret
    workflow.ExecuteActivity(actCtx, activities.DeleteCustomDomainIngress,
        DeleteCustomDomainIngressInput{
            Namespace: input.Namespace, ServiceName: input.ServiceName,
        }).Get(ctx, nil)

    return DetachCustomDomainWorkflowResult{Status: "deleted"}, nil
}
```

---

## K8s Activities (Resource CRUD)

### ApplyCustomDomainCertificate

Creates a cert-manager `Certificate` custom resource:

```go
// go-backend/internal/k8sdeployments/activity_custom_domain.go (lines 23-47)

func (a *Activities) ApplyCustomDomainCertificate(ctx context.Context, input ApplyCustomDomainCertificateInput) error {
    cert := buildCustomDomainCertificate(input.Namespace, input.ServiceName, input.Domain)
    data, _ := json.Marshal(cert)

    // Server-side apply (idempotent, declarative)
    _, err = a.dynClient.Resource(certGVR).Namespace(input.Namespace).Patch(
        ctx, cert.GetName(), types.ApplyPatchType, data,
        metav1.PatchOptions{FieldManager: "temporal-worker"},
    )
    return err
}
```

The Certificate resource it builds:

```go
// go-backend/internal/k8sdeployments/k8s_resources.go (lines 282-304)

func buildCustomDomainCertificate(namespace, serviceName, domain string) *unstructured.Unstructured {
    certName := serviceName + "-cd"        // e.g., "my-app-cd"
    secretName := certName + "-tls"        // e.g., "my-app-cd-tls"

    return &unstructured.Unstructured{
        Object: map[string]any{
            "apiVersion": "cert-manager.io/v1",
            "kind":       "Certificate",
            "metadata": map[string]any{
                "name":      certName,
                "namespace": namespace,
            },
            "spec": map[string]any{
                "secretName": secretName,
                "issuerRef": map[string]any{
                    "name": "letsencrypt-prod",
                    "kind": "ClusterIssuer",
                },
                "dnsNames": []any{domain},
            },
        },
    }
}
```

Equivalent YAML:
```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-app-cd
  namespace: dp-<userID>-<projectRef>
spec:
  secretName: my-app-cd-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - app.customer.com
```

### WaitForCertificateReady

Polls the Certificate CR status every 5 seconds:

```go
// go-backend/internal/k8sdeployments/activity_custom_domain.go (lines 49-94)

func (a *Activities) WaitForCertificateReady(ctx context.Context, input WaitForCertificateReadyInput) error {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()   // Timeout after 10 minutes (set by workflow)
        case <-ticker.C:
            recordHeartbeat(ctx, "polling certificate status")

            cert, _ := a.dynClient.Resource(certGVR).Namespace(input.Namespace).Get(ctx, input.CertificateName, metav1.GetOptions{})
            conditions, found, _ := unstructuredConditions(cert.Object)
            if !found { continue }

            for _, c := range conditions {
                if c.Type == "Ready" && c.Status == "True" {
                    return nil   // SUCCESS
                }
                if c.Type == "Ready" && c.Status == "False" {
                    if isTerminalCertFailure(c.Reason) {
                        // InvalidDomain, CAA, RateLimited, Denied -> non-retryable
                        return temporal.NewNonRetryableApplicationError(...)
                    }
                }
            }
        }
    }
}

func isTerminalCertFailure(reason string) bool {
    switch reason {
    case "InvalidDomain", "CAA", "RateLimited", "Denied":
        return true
    }
    return false
}
```

### ApplyCustomDomainIngress

Creates the Ingress with TLS section (only after cert is ready):

```go
// go-backend/internal/k8sdeployments/activity_custom_domain.go (lines 96-126)

func (a *Activities) ApplyCustomDomainIngress(ctx context.Context, input ApplyCustomDomainIngressInput) error {
    // Get the existing K8s Service to extract the port
    svc, _ := a.k8s.CoreV1().Services(input.Namespace).Get(ctx, input.ServiceName, metav1.GetOptions{})
    port := svc.Spec.Ports[0].Port

    ingress := buildCustomDomainIngress(input.Namespace, input.ServiceName, input.Domain, port)
    data, _ := json.Marshal(ingress)

    ingressName := input.ServiceName + "-cd"
    _, err = a.k8s.NetworkingV1().Ingresses(input.Namespace).Patch(ctx, ingressName,
        types.ApplyPatchType, data,
        metav1.PatchOptions{FieldManager: "temporal-worker"})
    return err
}
```

The Ingress resource it builds:

```go
// go-backend/internal/k8sdeployments/k8s_resources.go (lines 306-350)

func buildCustomDomainIngress(namespace, serviceName, customDomain string, port int32) *networkingv1.Ingress {
    pathType := networkingv1.PathTypePrefix
    ingressClassName := "traefik"
    ingressName := serviceName + "-cd"      // e.g., "my-app-cd"

    return &networkingv1.Ingress{
        TypeMeta: metav1.TypeMeta{Kind: "Ingress", APIVersion: "networking.k8s.io/v1"},
        ObjectMeta: metav1.ObjectMeta{
            Name:      ingressName,
            Namespace: namespace,
        },
        Spec: networkingv1.IngressSpec{
            IngressClassName: &ingressClassName,
            TLS: []networkingv1.IngressTLS{
                {
                    Hosts:      []string{customDomain},
                    SecretName: ingressName + "-tls",   // "my-app-cd-tls" (populated by cert-manager)
                },
            },
            Rules: []networkingv1.IngressRule{
                {
                    Host: customDomain,
                    IngressRuleValue: networkingv1.IngressRuleValue{
                        HTTP: &networkingv1.HTTPIngressRuleValue{
                            Paths: []networkingv1.HTTPIngressPath{
                                {
                                    Path:     "/",
                                    PathType: &pathType,
                                    Backend: networkingv1.IngressBackend{
                                        Service: &networkingv1.IngressServiceBackend{
                                            Name: serviceName,
                                            Port: networkingv1.ServiceBackendPort{
                                                Number: port,
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

### DeleteCustomDomainIngress

Cleans up all 3 resources when removing a custom domain:

```go
// go-backend/internal/k8sdeployments/activity_custom_domain.go (lines 128-153)

func (a *Activities) DeleteCustomDomainIngress(ctx context.Context, input DeleteCustomDomainIngressInput) error {
    ingressName := input.ServiceName + "-cd"
    tlsSecretName := ingressName + "-tls"
    certName := ingressName

    // 1. Delete Ingress
    err := a.k8s.NetworkingV1().Ingresses(input.Namespace).Delete(ctx, ingressName, metav1.DeleteOptions{})
    // (ignores NotFound)

    // 2. Delete Certificate CR
    err = a.dynClient.Resource(certGVR).Namespace(input.Namespace).Delete(ctx, certName, metav1.DeleteOptions{})
    // (ignores NotFound)

    // 3. Delete TLS Secret
    err = a.k8s.CoreV1().Secrets(input.Namespace).Delete(ctx, tlsSecretName, metav1.DeleteOptions{})
    // (ignores NotFound)

    return nil
}
```

---

## Naming Conventions Summary

For a service named `my-app` in namespace `dp-<userID>-<projectRef>`:

| Resource | Name | Notes |
|----------|------|-------|
| Standard Ingress | `my-app` | Host: `my-app.ml.ink`, no TLS section |
| Custom Domain Ingress | `my-app-cd` | Host: `app.customer.com`, has TLS section |
| Certificate CR | `my-app-cd` | cert-manager.io/v1, references `letsencrypt-prod` |
| TLS Secret | `my-app-cd-tls` | Populated by cert-manager with the issued cert |
| K8s Service | `my-app` | Shared by both Ingresses |
| Deployment | `my-app` | `runtimeClassName: gvisor` |
| Env Secret | `my-app-env` | Contains PORT and user-defined env vars |

---

## Complete Lifecycle Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ STEP 1: add_custom_domain(name="my-app", domain="app.co.com")  │
│                                                                  │
│  → Validate domain (no wildcard, no apex, no *.ml.ink)          │
│  → Anti-squat check (reclaim expired/failed)                    │
│  → Generate verification token (32-char hex)                    │
│  → Compute CNAME target: my-app.cname.ml.ink                   │
│  → INSERT INTO custom_domains (status: pending_dns)             │
│  → Return DNS instructions to user                              │
│                                                                  │
│  K8s resources created: NONE                                    │
│  DB status: pending_dns                                         │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2: User configures DNS at their registrar                  │
│                                                                  │
│  TXT:   _dp-verify.app.co.com → dp-verify=<token>              │
│  CNAME: app.co.com → my-app.cname.ml.ink                       │
│                                                                  │
│  (Wait 5-30 min for DNS propagation)                            │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ STEP 3: verify_custom_domain(name="my-app")                    │
│                                                                  │
│  → net.LookupTXT("_dp-verify.app.co.com")                      │
│    Check for "dp-verify=<token>" → PASS                        │
│  → net.LookupCNAME("app.co.com")                               │
│    Check matches "my-app.cname.ml.ink" → PASS                  │
│  → UPDATE status = 'provisioning'                               │
│  → Start Temporal workflow: AttachCustomDomainWorkflow          │
│    TaskQueue: deployer-eu-central-1 (runs on k3s cluster)      │
│                                                                  │
│  DB status: provisioning                                        │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ STEP 4: Temporal Workflow (deployer-worker in k3s cluster)      │
│                                                                  │
│  Phase 1: Create Certificate CR                                 │
│    → cert-manager picks it up                                   │
│    → Creates temporary Ingress for HTTP-01 challenge            │
│    → Let's Encrypt validates on port 80                         │
│    → TLS cert written to Secret: my-app-cd-tls                 │
│                                                                  │
│  Phase 2: Wait for cert Ready (poll every 5s, max 10min)       │
│    → Check cert.status.conditions[type=Ready,status=True]       │
│                                                                  │
│  Phase 3: Create Ingress with TLS                               │
│    → Host: app.co.com                                           │
│    → TLS secret: my-app-cd-tls                                 │
│    → Backend: Service/my-app on port from K8s Service          │
│    → Traefik picks up Ingress, starts routing                  │
│                                                                  │
│  Phase 4: UPDATE status = 'active', verified_at = NOW()        │
│                                                                  │
│  DB status: active                                              │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ LIVE: Custom domain is active                                   │
│                                                                  │
│  Browser → app.co.com                                           │
│    → CNAME → my-app.cname.ml.ink → 46.225.35.234 (Hetzner LB)│
│    → TCP passthrough → run-1:443                               │
│    → Traefik: TLS termination with my-app-cd-tls cert         │
│    → Route by Host header → Ingress my-app-cd                 │
│    → K8s Service my-app → Pod (gVisor sandbox)                 │
│                                                                  │
│  cert-manager auto-renews cert 30 days before expiry            │
└─────────────────────────────────────────────────────────────────┘
```

### Removal Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ remove_custom_domain(name="my-app")                             │
│                                                                  │
│  → Start Temporal workflow: DetachCustomDomainWorkflow           │
│    (BEFORE deleting DB record → prevents orphaned K8s objects)  │
│  → Workflow deletes: Ingress, Certificate CR, TLS Secret        │
│  → DELETE FROM custom_domains WHERE id = ...                    │
│                                                                  │
│  Domain freed for other services to claim                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## MCP Tool Definitions

Three MCP tools exposed to AI agents:

```go
// go-backend/internal/mcpserver/types.go

type AddCustomDomainInput struct {
    Name    string  // Service name
    Domain  string  // Custom domain (e.g., "app.customer.com")
    Project string  // Project name (default: "default")
}

type AddCustomDomainOutput struct {
    ServiceID    string  // Service ID
    Domain       string  // Domain name
    Status       string  // "pending_dns"
    Instructions string  // DNS setup instructions
}

type VerifyCustomDomainInput struct {
    Name    string  // Service name
    Project string  // Project name
}

type VerifyCustomDomainOutput struct {
    ServiceID string  // Service ID
    Domain    string  // Domain name
    Status    string  // "active", "provisioning", "failed", "pending_dns"
    Message   string  // Status message or error
}

type RemoveCustomDomainInput struct {
    Name    string  // Service name
    Project string  // Project name
}

type RemoveCustomDomainOutput struct {
    ServiceID string  // Service ID
    Message   string  // Confirmation message
}
```

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **TCP passthrough LB** (no TLS termination at Hetzner) | cert-manager HTTP-01 needs raw port 80 requests; TLS-terminating LB breaks ACME challenges |
| **Certificate-first, Ingress-second** | Traefik v3 auto-redirects HTTP->HTTPS when TLS section exists; this breaks HTTP-01 challenge |
| **Per-service CNAME target** (`my-app.cname.ml.ink`) | Enables unique routing per service; wildcard `*.cname.ml.ink` DNS resolves all to Hetzner LB |
| **TXT ownership verification** before provisioning | Prevents domain hijacking; only domain owner can set `_dp-verify` TXT records |
| **7-day expiry on pending_dns** | Anti-squat: unclaimed domains auto-expire and become reclaimable |
| **Temporal for async provisioning** | Cert issuance can take minutes; workflow handles retry, heartbeat, failure state |
| **Server-side apply** (`types.ApplyPatchType`) | Idempotent, declarative K8s operations; safe to retry without conflicts |
| **Separate Ingress per custom domain** (`-cd` suffix) | Standard `*.ml.ink` Ingress has no TLS section (uses wildcard default); custom domain needs its own TLS section |
| **gVisor runtime** on all customer pods | Security sandbox; syscall interception prevents container escapes |
| **Firewall restricts port 80/443** to LB + Cloudflare IPs | Prevents direct access to Traefik, only trusted sources reach application layer |

---

## File Reference

| File | Purpose |
|------|---------|
| `go-backend/internal/dnsverify/verify.go` | DNS verification (TXT, CNAME), validation, token generation |
| `go-backend/internal/deployments/service.go` (lines 574-847) | Business logic: AddCustomDomain, VerifyCustomDomain, RemoveCustomDomain |
| `go-backend/internal/k8sdeployments/workflows.go` (lines 310-420) | Temporal workflows: AttachCustomDomainWorkflow, DetachCustomDomainWorkflow |
| `go-backend/internal/k8sdeployments/activity_custom_domain.go` | K8s activities: cert, ingress, cleanup, DB status updates |
| `go-backend/internal/k8sdeployments/k8s_resources.go` (lines 282-350) | K8s resource builders: buildCustomDomainCertificate, buildCustomDomainIngress |
| `go-backend/internal/mcpserver/types.go` | MCP input/output types |
| `go-backend/internal/mcpserver/tools.go` (lines 662-758) | MCP tool handlers |
| `go-backend/internal/storage/pg/migrations/0001_initial.sql` (lines 201-216) | DB schema: custom_domains table |
| `go-backend/internal/storage/pg/migrations/0002_custom_domain_verification.sql` | Added verification_token column |
| `go-backend/internal/storage/pg/queries/customdomains/customdomains.sql` | SQL queries |
| `infra/eu-central-1/k8s/templates/customer-service-template.yml` | Standard service template (no TLS section) |
| `infra/eu-central-1/k8s/templates/customer-custom-domain-template.yml` | Custom domain Ingress template (with TLS) |
| `infra/eu-central-1/k8s/system/cert-manager-issuer.yml` | ClusterIssuer: DNS-01 for *.ml.ink, HTTP-01 for custom domains |
| `infra/eu-central-1/k8s/system/wildcard-cert.yml` | Wildcard cert for *.ml.ink |
| `infra/eu-central-1/k8s/system/traefik-tlsstore.yml` | Default TLS cert (wildcard) for Traefik |
| `infra/eu-central-1/k8s/values/traefik-values.yml` | Traefik Helm values: DaemonSet, hostNetwork, ports |
| `infra/eu-central-1/inventory/group_vars/run.yml` | Firewall: allowed sources for ports 80/443 |
