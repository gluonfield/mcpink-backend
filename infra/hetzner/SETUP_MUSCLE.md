# Muscle Server Setup (User Workload)

Muscle servers run untrusted user code (MCP servers). They require full hardening including gVisor sandboxing and egress rules.

## Server Info Template

| Field          | Value                           |
| -------------- | ------------------------------- |
| **Name**       | hetzner-muscle-N                |
| **Type**       | Hetzner Dedicated / Cloud       |
| **Public IP**  | (fill in)                       |
| **Private IP** | (assign from vSwitch: 10.0.1.X) |
| **Role**       | Run user containers             |

## Prerequisites

- [ ] Server provisioned
- [ ] SSH access working
- [ ] Root access

---

## Phase 1: Basic Setup

### 1.1 Update System

```bash
apt update && apt upgrade -y
apt install -y curl wget git jq
```

### 1.2 Install Docker

```bash
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker
```

---

## Phase 2: Hardening (Mandatory)

Muscle servers run untrusted user code - hardening is mandatory.

### 2.1 Full Hardening Script

```bash
# Clone/copy the hardening scripts to the server
# Then run:
cd /path/to/infra/hetzner/hardening
./setup-muscle.sh
```

Or run each step manually:

### 2.2 Install gVisor (Container Sandbox)

```bash
# See hardening/install-gvisor.sh
curl -fsSL https://gvisor.dev/archive.key | gpg --dearmor -o /usr/share/keyrings/gvisor-archive-keyring.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/gvisor-archive-keyring.gpg] https://storage.googleapis.com/gvisor/releases release main" > /etc/apt/sources.list.d/gvisor.list
apt update && apt install -y runsc
```

### 2.3 Configure Docker with gVisor

```bash
# See hardening/daemon.json
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "runtimes": {
    "runsc": {
      "path": "/usr/bin/runsc",
      "runtimeArgs": [
        "--platform=systrap",
        "--network=sandbox"
      ]
    }
  }
}
EOF
systemctl restart docker
```

### 2.4 SSH Hardening

```bash
# See hardening/harden-ssh.sh
./harden-ssh.sh
```

### 2.5 Host Firewall

```bash
# See hardening/setup-host-firewall.sh
./setup-host-firewall.sh
```

### 2.6 Egress Rules (Limit Outbound)

```bash
# See hardening/setup-egress-rules.sh
./setup-egress-rules.sh
```

### 2.7 Verify Hardening

```bash
# See hardening/verify-hardening.sh
./verify-hardening.sh
```

**Reference:** See `hardening/README.md` for detailed documentation.

---

## Phase 3: Network (vSwitch)

### 3.1 For Cloud VPS

In Hetzner Cloud Console:

- Go to server → Networking → Attach to network
- Select network #11898981 (ink-vpc)
- Assign IP from range 10.0.1.X

### 3.2 For Dedicated Server

Configure netplan:

```yaml
# /etc/netplan/50-vswitch.yaml
network:
  version: 2
  vlans:
    vlan4000:
      id: 4000
      link: enp0s31f6 # Check actual interface: ip link show
      addresses:
        - 10.0.1.X/24
      routes:
        - to: 10.0.1.0/24
          via: 10.0.1.1
```

```bash
netplan apply
```

### 3.3 Verify Connectivity

```bash
ping 10.0.1.2  # Factory
ping 10.0.1.4  # Muscle-Ops-1
ping 10.0.1.5  # Builder-1
```

---

## Phase 4: Registry Access

### 4.1 Docker Login to Registry

```bash
docker login registry.tops.subj.org
# Credentials in Coolify registry service settings
```

### 4.2 Test Pull

```bash
docker pull registry.tops.subj.org/test:v1 || echo "No test image yet - OK"
```

---

## Phase 5: Monitoring

### 5.1 Install Monitoring Agent

```bash
# See monitoring/install-alloy.sh
cd /path/to/infra/hetzner/monitoring
./install-alloy.sh
```

### 5.2 Deploy cAdvisor (Container Metrics)

```bash
# See monitoring/cadvisor-compose.yaml
docker compose -f cadvisor-compose.yaml up -d
```

**Reference:** See `monitoring/README.md` for Grafana/Prometheus setup.

---

## Phase 6: Coolify Integration

### 6.1 Add Coolify SSH Key

```bash
mkdir -p ~/.ssh && chmod 700 ~/.ssh
echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAVyYsJThwW7tXS6rKjAk0IePULrF6glVn2tjpwhLm/U' >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

### 6.2 Add Server to Coolify

```
Coolify UI → Servers → Add Server
  - Name: Muscle-N
  - IP: <public-ip> (or private IP if Factory can reach via vSwitch)
  - Port: 22
  - User: root
  - Private Key: (select or generate)
```

### 6.3 Validate Server

Click "Validate Server" - Coolify will:

- Check SSH connectivity
- Verify Docker is running
- Install Coolify agent

### 6.4 Verify NOT a Build Server

Ensure "Use it as a build server?" is **unchecked** - Muscle servers run apps, not builds.

---

## Phase 7: Verification

### 7.1 Verify gVisor Working

```bash
docker run --runtime=runsc --rm alpine dmesg 2>&1 | grep -i gvisor && echo "gVisor OK"
```

### 7.2 Verify Hardening

```bash
./hardening/verify-hardening.sh
```

### 7.3 Test Deployment

1. Deploy a test app to this Muscle server
2. Verify it runs with `--runtime=runsc`
3. Check container is sandboxed

---

## Checklist Summary

### Basic Setup

- [ ] Server provisioned
- [ ] System updated
- [ ] Docker installed

### Hardening (Mandatory)

- [ ] gVisor installed
- [ ] Docker configured with runsc runtime
- [ ] SSH hardened (no password auth)
- [ ] Host firewall enabled
- [ ] Egress rules configured
- [ ] Hardening verified

### Network

- [ ] Attached to vSwitch
- [ ] Private IP assigned (10.0.1.X)
- [ ] Can ping other servers

### Registry

- [ ] Docker login to registry
- [ ] Test pull working

### Monitoring

- [ ] Alloy agent installed
- [ ] cAdvisor running
- [ ] Metrics visible in Grafana

### Coolify

- [ ] SSH key added
- [ ] Added to Coolify
- [ ] Validated successfully
- [ ] Test deployment working

---

## Maintenance

### Update gVisor

```bash
apt update && apt install -y runsc
systemctl restart docker
```

### Check Container Runtime

```bash
docker inspect <container_id> | jq '.[0].HostConfig.Runtime'
# Should output: "runsc"
```

### Miner Detection (Run Periodically)

```bash
./hardening/detect-miners.sh
```

### Update Registry Login (When Muscle-Ops-1 Ready)

```bash
docker logout registry.tops.subj.org

# For internal vSwitch registry (no auth)
# Add to /etc/docker/daemon.json:
# "insecure-registries": ["10.0.1.4:5000"]
systemctl restart docker
```

---

## Troubleshooting

### gVisor Issues

See `hardening/GVISOR_ISSUE.md` for known issues and workarounds.

### Container Won't Start with runsc

```bash
# Check gVisor logs
journalctl -u docker | grep runsc

# Test gVisor directly
runsc --version
docker run --runtime=runsc hello-world
```

### Network Connectivity Issues

```bash
# Check vSwitch interface
ip addr show | grep 10.0.1

# Check routes
ip route | grep 10.0.1
```
