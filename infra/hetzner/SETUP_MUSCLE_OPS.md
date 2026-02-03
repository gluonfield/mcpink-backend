# Muscle Ops Server Setup

Ops servers run trusted infrastructure software (Registry, Gitea, Grafana, Prometheus). They do NOT require gVisor or egress rules since they only run software we control.

## Architecture

```
                    Deployment Flow
┌────────────┐    ┌─────────────┐    ┌────────────┐    ┌────────────┐
│  GitHub    │───▶│  Builder-1  │───▶│ Muscle-Ops │───▶│  Muscle-1  │
│            │    │  (build)    │    │ (registry) │    │   (run)    │
└────────────┘    │  10.0.0.3   │    │  10.0.1.4  │    │  10.0.1.3  │
                  └─────────────┘    └────────────┘    └────────────┘
                         │                 │                 │
                         └─────────────────┴─────────────────┘
                              vSwitch + Cloud Network
```

## Server Info

| Field          | Value                                              |
| -------------- | -------------------------------------------------- |
| **Name**       | hetzner-muscle-ops-1                               |
| **Type**       | Hetzner Dedicated                                  |
| **Public IP**  | 116.202.163.209                                    |
| **Private IP** | 10.0.1.4                                           |
| **Role**       | Docker Registry, Gitea, Grafana, Prometheus        |

## Prerequisites

- [x] Server provisioned
- [x] SSH access working
- [x] Root access

---

## Phase 1: Verify Hardware & RAID

Ops servers typically use dedicated hardware with RAID. Verify before proceeding.

### 1.1 Check RAID Status

```bash
# Check software RAID
cat /proc/mdstat

# Check hardware RAID (if applicable)
lspci | grep -i raid

# List block devices
lsblk

# Check disk health
smartctl -a /dev/sda  # or nvme0n1
```

**Expected for muscle-ops-1:**
- 2×960GB NVMe in RAID1 → `/` and `/data`
- 2×2TB HDD in RAID1 → `/backups`

### 1.2 Verify Mount Points

```bash
df -h
```

Expected output:
```
/dev/md0    ~900G   /
/dev/md1    ~1.8T   /backups
```

### 1.3 Setup HDD RAID for Backups

The 2×2TB HDDs (sda, sdb) need to be configured as RAID1 for `/backups`:

```bash
# Create RAID1 mirror
mdadm --create /dev/md3 --level=1 --raid-devices=2 /dev/sda /dev/sdb

# Format
mkfs.ext4 /dev/md3

# Mount
mkdir -p /backups
mount /dev/md3 /backups

# Add to fstab for persistence
echo '/dev/md3 /backups ext4 defaults 0 0' >> /etc/fstab

# Save RAID config
mdadm --detail --scan >> /etc/mdadm/mdadm.conf
update-initramfs -u
```

### 1.4 Create Data Directory

```bash
mkdir -p /data
```

---

## Phase 2: Basic Setup

### 2.1 Update System

```bash
apt update && apt upgrade -y
apt install -y curl wget git jq
```

### 2.2 Install Docker (or let Coolify do it)

If setting up manually:
```bash
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker
```

Or skip and let Coolify install Docker when you validate the server.

---

## Phase 3: Basic Security

No gVisor or egress rules needed - just basic hardening.

### 3.1 SSH Hardening

```bash
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/PermitRootLogin yes/PermitRootLogin prohibit-password/' /etc/ssh/sshd_config
systemctl restart sshd
```

### 3.2 Basic Firewall

```bash
apt install -y ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow from 10.0.1.0/24  # Allow all vSwitch traffic
ufw --force enable
```

---

## Phase 4: Network (vSwitch)

### 4.1 Configure vSwitch (Dedicated Server)

```bash
# Find network interface
ip link show
```

Create netplan config:

```yaml
# /etc/netplan/50-vswitch.yaml
network:
  version: 2
  vlans:
    vlan4000:
      id: 4000
      link: eno1  # Check with: ip link show
      mtu: 1400   # Required for vSwitch
      addresses:
        - 10.0.1.4/24
      routes:
        - to: 10.0.0.0/16  # Entire cloud network
          via: 10.0.1.1
```

```bash
netplan apply
```

### 4.2 Verify Connectivity

```bash
ping 10.0.1.3  # Muscle-1 (vSwitch)
ping 10.0.0.3  # Builder-1 (Cloud Network)
```

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
  - Name: Muscle-Ops-1
  - IP: 116.202.163.209 (or 10.0.1.4 via vSwitch)
  - Port: 22
  - User: root
  - Private Key: (select Coolify key)
```

### 5.3 Validate Server

Click "Validate Server" - Coolify will:
- Check SSH connectivity
- Install Docker if needed
- Install Coolify agent

### 5.4 Verify NOT a Build Server

Ensure "Use it as a build server?" is **unchecked**.

---

## Phase 6: Deploy Services

Deploy these via Coolify:

| Service         | Image             | Port  | Purpose                |
| --------------- | ----------------- | ----- | ---------------------- |
| Docker Registry | `registry:2`      | 5000  | Internal image storage |
| Gitea           | `gitea/gitea`     | 3000  | Git hosting            |
| Grafana         | `grafana/grafana` | 3001  | Dashboards             |
| Prometheus      | `prom/prometheus` | 9090  | Metrics                |

### 6.1 Docker Registry ✅

Deployed via Coolify:
- **URL:** `registry.tops.subj.org`
- **Credentials:** See Coolify service settings
- **Storage:** Docker volume (~437MB used)
- **Image:** `registry:3`

### 6.2 Gitea ✅

Deployed via Coolify:
- **URL:** `git.ml.ink`
- **Ports:** 3000 (web), 22 (SSH)
- **Storage:** Docker volume

### 6.3 Grafana & Prometheus

Deploy via Coolify for monitoring stack.

---

## Phase 7: Configure Other Servers ✅

Registry uses HTTPS with auth via public domain.

### 7.1 Current Setup (Public Domain)

```bash
docker login registry.tops.subj.org
# Credentials in Coolify registry service settings
```

**Configured servers:**
- [x] muscle-1 (157.90.130.187)
- [x] builder-1 (46.225.92.127)
- [x] factory (46.225.65.56)

---

## TODO: Switch to Internal vSwitch Registry

Currently traffic goes via public internet. To use vSwitch (faster, free bandwidth):

### Step 1: Expose Registry on Port 5000

In Coolify, add port mapping `5000:5000` to the registry service.

### Step 2: Configure All Servers

On each server (muscle-1, builder-1, factory), add to `/etc/docker/daemon.json`:

```json
{
  "insecure-registries": ["10.0.1.4:5000"]
}
```

Then restart Docker:
```bash
systemctl restart docker
```

### Step 3: Update Backend Environment

Change `COOLIFY_REGISTRYURL` from `registry.tops.subj.org` to `10.0.1.4:5000`

### Step 4: Test Internal Access

```bash
# From any server on vSwitch
docker pull alpine
docker tag alpine 10.0.1.4:5000/test:v1
docker push 10.0.1.4:5000/test:v1
```

### Benefits

| Route | Latency | Bandwidth Cost |
|-------|---------|----------------|
| Public (`registry.tops.subj.org`) | ~1ms | Metered |
| vSwitch (`10.0.1.4:5000`) | ~0.1ms | Free |

---

## Checklist Summary

### Hardware

- [x] NVMe RAID1 verified (md0, md1, md2)
- [x] NVMe RAID resync complete
- [x] HDD RAID1 setup for `/backups` (md3, 1.8TB)
- [x] Disk space adequate (800GB NVMe + 1.8TB HDD)

### Basic Setup

- [x] System updated
- [x] Docker installed (via Coolify)

### Security

- [x] SSH hardened (no password auth)
- [x] Firewall enabled (ufw)

### Network

- [x] vSwitch configured (10.0.1.4, MTU 1400)
- [x] Can ping other servers (muscle-1, builder)

### Coolify

- [x] SSH key added
- [x] Added to Coolify
- [x] Validated successfully

### Services

- [x] Docker Registry deployed (`registry.tops.subj.org`)
- [x] Gitea deployed (`git.ml.ink`)
- [ ] Grafana deployed
- [ ] Prometheus deployed
- [x] Other servers configured to use registry

---

## Maintenance

### Check RAID Health

```bash
cat /proc/mdstat
mdadm --detail /dev/md0
mdadm --detail /dev/md2
mdadm --detail /dev/md3  # HDD backup mirror
```

### Registry Garbage Collection

Registry GC removes unreferenced blobs. Run weekly during low-traffic:

```bash
# /etc/cron.weekly/registry-gc
#!/bin/bash
CONTAINER=$(docker ps -qf "name=registry")
docker exec $CONTAINER registry garbage-collect /etc/docker/registry/config.yml
```

### Backup to /backups

```bash
# Sync critical data to HDD backup drive
rsync -av /var/lib/docker/volumes/*registry*/ /backups/registry/
rsync -av /data/gitea/ /backups/gitea/
rsync -av /data/grafana/ /backups/grafana/
```

### Offsite Backup (restic to S3)

```bash
# Install restic
apt install restic

# Init repo (once)
export RESTIC_REPOSITORY="s3:https://fsn1.your-objectstorage.com/backups"
restic init

# Daily backup cron
restic backup /backups/gitea /backups/grafana
restic forget --keep-daily 7 --keep-weekly 4 --prune
```

### Monitor Disk Usage

```bash
df -h / /backups
```

---

## Troubleshooting

### RAID Degraded

```bash
# Check status
cat /proc/mdstat

# If degraded, check which disk failed
mdadm --detail /dev/md0

# Replace failed disk (contact Hetzner support for hardware)
```

### Registry Not Accessible

```bash
# Check if running
docker ps | grep registry

# Check logs
docker logs <registry_container>

# Test locally
curl http://localhost:5000/v2/
```

### vSwitch Not Working

```bash
# Check interface
ip addr show | grep 10.0.1

# Check netplan config
cat /etc/netplan/50-vswitch.yaml

# Reapply
netplan apply
```
