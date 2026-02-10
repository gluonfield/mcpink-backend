# k3s Ansible Automation

This directory implements `infra/k3s/ARCHITECTURE_PLAN.md` as reproducible Ansible.

It also preserves the original `README.md` platform goals:
- Stable agent contract (`create/redeploy/get/delete`) while swapping infra internals.
- 3-plane isolation (control/build/run) to prevent build contention on runtime nodes.
- Security defaults (gVisor runtime class, egress restrictions, host firewall, no workflow secrets).

## Layout

- `inventory/hosts.yml` - Host inventory and global variables.
- `inventory/group_vars/` - Pool-specific defaults.
- `playbooks/site.yml` - Full cluster bootstrap/reconcile.
- `playbooks/add-run-node.yml` - Add a new run node.
- `playbooks/upgrade-k3s.yml` - Upgrade existing cluster nodes.
- `playbooks/patch-hosts.yml` - Apply OS security updates (apt upgrade + optional reboot).
- `playbooks/cloudflare-lb.yml` - Reconcile Cloudflare Load Balancer origins + hostnames from inventory.
- `roles/` - Node/bootstrap responsibilities.

## Prerequisites

1. `ansible-core` installed on your local machine.
2. SSH access as root to all target nodes.
3. Hetzner private networking configured (cloud network + vSwitch).
4. DNS records in place for:
   - `*.ml.ink`
   - `grafana.ml.ink`
   - `loki.ml.ink`

## First Run

```bash
cd infra/ansible
ansible-playbook playbooks/site.yml
```

If you keep secrets outside inventory, pass them at runtime:

```bash
ansible-playbook playbooks/site.yml \
  -e cloudflare_api_token="$CLOUDFLARE_API_TOKEN" \
  -e loki_basic_auth_users="deploy:$LOKI_BCRYPT"
```

Apply Cloudflare Load Balancer config (wildcard + Grafana + Loki) from inventory run nodes:

```bash
ansible-playbook playbooks/cloudflare-lb.yml
```

## Security Patching

Run vulnerability and package updates separately from cluster reconciliation:

```bash
ansible-playbook playbooks/patch-hosts.yml --limit all
```

Useful options:

```bash
# No reboot
ansible-playbook playbooks/patch-hosts.yml -e security_patch_reboot_if_required=false

# Patch one node at a time (default) or increase serial batch size
ansible-playbook playbooks/patch-hosts.yml -e serial=2
```

## Post-bootstrap secrets (Ansible-managed)

`playbooks/site.yml` creates/updates required runtime secrets directly when vars are set. Manual `infra/k8s/*.example.yml` secret applies are no longer used.

Managed by `roles/k8s_addons`:

- `github-app` (`dp-system`): `github_app_id` + `github_app_private_key`
- `temporal-creds` (`dp-system`): `temporal_cloud_api_key` as `cloud-api-key`
- `temporal-worker-config` (`dp-system`): `temporal_address` + `temporal_namespace`
- `cloudflare-api-token` (`cert-manager`): `cloudflare_api_token`
- `loki-auth-secret` (`dp-system`): `loki_basic_auth_users`

After setting values in inventory/vault (or via `-e`), re-run:

```bash
ansible-playbook playbooks/site.yml
```
