# Hetzner Infrastructure

Infrastructure setup for InkMCP's 3-plane architecture using Hetzner servers.

---

## Architecture Overview

| Plane       | Server   | Role         | What Runs                                               |
| ----------- | -------- | ------------ | ------------------------------------------------------- |
| **Plane B** | Factory  | Build Server | Coolify master (UI + API), Nixpacks, Container Registry |
| **Plane C** | Muscle 1 | Run Server   | Docker + Traefik, User containers                       |
| **Plane C** | Muscle 2 | Run Server   | Docker + Traefik, User containers                       |

```
InkMCP API (Plane A - your backend)
    │
    ▼
Coolify API (Plane B - Factory)
    │
    ├──SSH──▶ Muscle 1 (run apps)
    └──SSH──▶ Muscle 2 (run apps)
```

> **Key:** Coolify is only installed on the Factory server. Run servers are managed via SSH.

---

## Recommended Compute Specs

| Server        | Role             | CPU         | RAM       | Storage    | Notes                                                         |
| ------------- | ---------------- | ----------- | --------- | ---------- | ------------------------------------------------------------- |
| Factory       | Coolify + Builds | 8-16 cores  | 32-64 GB  | 200GB+ SSD | Builds are CPU/RAM intensive but bursty. Registry needs disk. |
| Muscle (each) | Run containers   | 8-16+ cores | 64-128 GB | 100GB+ SSD | RAM is usually the constraint. Scale based on app count.      |

---

## Port Requirements

### Factory Server (Coolify Master)

| Port | Required    | Purpose                                     |
| ---- | ----------- | ------------------------------------------- |
| 22   | ✅ Yes      | SSH access (admin + Coolify internal)       |
| 8000 | ✅ Yes      | Coolify dashboard HTTP                      |
| 6001 | ✅ Yes      | Coolify real-time communications            |
| 6002 | ✅ Yes      | Coolify terminal access (v4.0.0-beta.336+)  |
| 80   | ⚠️ Optional | Only if exposing Coolify dashboard with SSL |
| 443  | ⚠️ Optional | Only if exposing Coolify dashboard with SSL |

### Run Servers (Muscle)

| Port             | Required | Purpose                                   |
| ---------------- | -------- | ----------------------------------------- |
| 22               | ✅ Yes   | SSH (for Coolify to deploy via SSH)       |
| 80               | ✅ Yes   | HTTP traffic + SSL certificate generation |
| 443              | ✅ Yes   | HTTPS traffic to deployed apps            |
| 8000, 6001, 6002 | ❌ No    | Not needed - these are Coolify-specific   |

---

## Firewall Configuration

**Hetzner Auction servers have all ports open by default.** You must configure firewall rules.

### Options

| Method                 | Where         | Notes                                        |
| ---------------------- | ------------- | -------------------------------------------- |
| OS-level firewall      | On the server | `ufw`, `iptables`, `firewalld`               |
| Hetzner Robot Firewall | Robot panel   | Hardware-level, configured via web UI or API |

### Security Recommendations

- Restrict port 22 to your IP or use a bastion host
- Use **vSwitch private network** for Coolify → Run server SSH (avoid exposing 22 publicly on run servers)
- Consider Hetzner Robot Firewall for additional hardware-level protection

## Setup

### Factory

#### Store data in external disk

```sh
umount /mnt/HC_Volume_104528544
mkdir -p /data
mount /dev/sdb /data
```

Persist mount

```sh
nano /etc/fstab
```

Update

```
/dev/disk/by-id/scsi-0HC_Volume_104528544 /data ext4 discard,nofail,defaults 0 0
```

#### Install Coolify

```sh
curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
```

- Choose domain for factory admin panel and add DNS record.

#### Connect Muscle server to Factory

Coolify > Servers > Add > Generate Private Key

```sh
ssh root@<MUSCLE_1_IP>
mkdir -p ~/.ssh
nano ~/.ssh/authorized_keys --- [Paste the public key]
chmod 600 ~/.ssh/authorized_keys
chmod 700 ~/.ssh
```

#### General Wildcard domain for apps across multiple servers

TBD (Cloudflare API, Hetzner API)

#### Wildcard domain per Hetzner server

Add DNS A record `*.s1.ml.ink` pointing to Muscle-1.

Servers -> Muscle 1 -> Wildcard domain -> `s1.ml.ink`
