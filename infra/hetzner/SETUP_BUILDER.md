# Builder Server Setup

Builder servers handle Docker image builds and push to the registry. They don't run user workloads, so they need less hardening than Muscle servers.

## Server Info

| Field          | Value                                 |
| -------------- | ------------------------------------- |
| **Name**       | hetzner-builder-1                     |
| **Type**       | Hetzner Cloud VPS                     |
| **Public IP**  | 46.225.92.127                         |
| **Private IP** | 10.0.1.5                              |
| **Role**       | Build server for Coolify              |

## Prerequisites

- [x] Server provisioned on Hetzner Cloud
- [x] SSH access working
- [x] Root or sudo access

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

### 1.3 Configure Docker Daemon

```bash
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
systemctl restart docker
```

---

## Phase 2: Network (vSwitch)

### 2.1 Attach to Cloud Network

In Hetzner Cloud Console:

- Go to server → Networking → Attach to network
- Select network #11898981 (ink-vpc)
- Assign IP: 10.0.1.5

### 2.2 Verify Connectivity

```bash
# Should be able to ping other servers on vSwitch
ping 10.0.1.2  # Factory
ping 10.0.1.3  # Muscle-1
ping 10.0.1.4  # Muscle-Ops-1
```

---

## Phase 3: Registry Access

### 3.1 Docker Login to Registry

```bash
docker login registry.tops.subj.org
# Credentials in Coolify registry service settings
```

### 3.2 Test Push

```bash
docker pull alpine
docker tag alpine registry.tops.subj.org/builder-test:v1
docker push registry.tops.subj.org/builder-test:v1
```

---

## Phase 4: Basic Security

Builder doesn't run user code, so minimal hardening needed.

### 4.1 SSH Hardening

```bash
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/PermitRootLogin yes/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config
systemctl restart sshd
```

### 4.2 Basic Firewall

```bash
apt install -y ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw --force enable
```

**Note:** No gVisor needed on builder - it only builds images, doesn't run user containers.

---

## Phase 5: Coolify Integration

### 5.1 Add Coolify SSH Key

```bash
mkdir -p ~/.ssh && chmod 700 ~/.ssh
echo 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAVyYsJThwW7tXS6rKjAk0IePULrF6glVn2tjpwhLm/U' >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

### 5.2 Add Server to Coolify

```
Coolify UI → Servers → Add Server
  - Name: Builder-1
  - IP: 46.225.92.127 (or 10.0.1.5 via vSwitch)
  - Port: 22
  - User: root
  - Private Key: (select Coolify key)
```

### 5.3 Enable as Build Server

```
Coolify UI → Servers → Builder-1 → Settings
  - "Use it as a build server?" → Enable ✓
```

### 5.4 Validate Server

Click "Validate Server" in Coolify to ensure connectivity and Docker are working.

---

## Phase 6: Verification

### 6.1 Check Build Server Status

In Coolify UI, Builder-1 should show as a build server.

### 6.2 Test Full Build Flow

1. Create a test app with `use_build_server = true`
2. Deploy it
3. Check logs - should show:
   - Build happening on Builder-1
   - Push to registry
   - Pull on Muscle-1

---

## Checklist Summary

- [x] Server provisioned
- [x] System updated
- [x] Docker installed (using 200GB volume)
- [ ] Attached to vSwitch (10.0.1.5)
- [x] Docker login to registry
- [x] SSH hardened
- [x] Basic firewall enabled
- [ ] Coolify SSH key added
- [ ] Added to Coolify
- [ ] Marked as build server
- [ ] Validated in Coolify
- [ ] Test build completed

---

## Maintenance

### Cleanup Build Cache

Coolify handles cleanup, but manual cleanup if needed:

```bash
docker system prune -af --volumes
```

### Update Docker Login (When Muscle-Ops-1 Ready)

```bash
# Remove old login
docker logout registry.tops.subj.org

# Configure internal registry
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "insecure-registries": ["10.0.1.4:5000"]
}
EOF
systemctl restart docker
```

---

## Troubleshooting

### Build Fails

```bash
# Check Docker logs
journalctl -u docker -f

# Check disk space
df -h

# Clean up if needed
docker system prune -af
```

### Can't Push to Registry

```bash
# Test registry connectivity
curl -v http://10.0.1.4:5000/v2/

# Check Docker login
cat ~/.docker/config.json
```

### vSwitch Not Working

```bash
# Check if IP assigned
ip addr show

# In Hetzner Cloud Console, verify network attachment
```
