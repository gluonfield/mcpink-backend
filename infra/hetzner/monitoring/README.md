# Grafana Cloud Monitoring Setup

## Prerequisites

1. Grafana Cloud account (free tier works)
2. SSH access to servers

## Getting Grafana Cloud Credentials

1. Log into https://grafana.com
2. Go to **Connections** → **Data sources** → Click your Prometheus datasource
3. Copy:
   - Remote write URL
   - Username (numeric ID)
4. Go to **Security** → **Access Policies** → Create policy with `metrics:write` scope
5. Create a token under that policy

## Deployment

### 1. Set hostname (all servers)

Set a descriptive hostname - this appears in dashboards:

```bash
hostnamectl set-hostname factory    # or muscle-1, muscle-2, etc.
```

### 2. Install Alloy (all servers)

```bash
bash install-alloy.sh
```

### 3. Set credentials (all servers)

```bash
cat > /etc/default/alloy << 'EOF'
GRAFANA_CLOUD_PROMETHEUS_URL="https://prometheus-prod-XX-prod-XX.grafana.net/api/prom/push"
GRAFANA_CLOUD_USERNAME="YOUR_ID"
GRAFANA_CLOUD_API_KEY="glc_YOUR_TOKEN"
CONFIG_FILE="/etc/alloy/config.alloy"
EOF
```

### 4. Deploy config

**Factory servers:**

```bash
cp configs/factory.alloy /etc/alloy/config.alloy
cp docker-daemon.json /etc/docker/daemon.json
systemctl restart docker
systemctl enable --now alloy
```

**Muscle servers:**

```bash
# Edit instance label in muscle.alloy first (muscle-1, muscle-2, etc.)
cp configs/muscle.alloy /etc/alloy/config.alloy
cp docker-daemon.json /etc/docker/daemon.json
docker compose -f cadvisor-compose.yaml up -d
systemctl restart docker
systemctl enable --now alloy
```

## Verification

```bash
# Check Alloy status
systemctl status alloy
journalctl -u alloy -f

# Test metrics endpoints
curl -s localhost:9323/metrics | head  # Docker
curl -s localhost:8081/metrics | head  # cAdvisor (Muscle only)
```

In Grafana Cloud → Explore:
```
up{instance="factory"}
up{instance=~"muscle-.*"}
node_cpu_seconds_total
container_memory_usage_bytes
```

## Dashboards

Import in Grafana Cloud (**Dashboards** → **New** → **Import**):

| ID | Name | Use for |
|----|------|---------|
| 1860 | Node Exporter Full | CPU, memory, disk, network (best for servers) |
| 11600 | Docker Container Monitoring | Container metrics |

Select your Prometheus datasource (`grafanacloud-*-prom`) when importing.

## Adding New Muscle Servers

1. Set hostname: `hostnamectl set-hostname muscle-N`
2. Run `install-alloy.sh`
3. Copy `muscle.alloy`, change `instance` label to `muscle-N`
4. Set credentials in `/etc/default/alloy`
5. Deploy cAdvisor, restart services
6. Dashboard auto-includes new server

## Verification Checklist

- [ ] `/etc/default/alloy` has credentials
- [ ] Alloy running: `systemctl status alloy`
- [ ] Docker metrics: `curl -s localhost:9323/metrics | head`
- [ ] cAdvisor (Muscle): `curl -s localhost:8081/metrics | head`
- [ ] Grafana Cloud shows data for `up{instance="..."}`
