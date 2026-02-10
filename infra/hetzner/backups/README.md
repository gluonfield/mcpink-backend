# Backup Scripts

This directory previously contained Coolify backup/restore scripts. Coolify has been replaced by K3s/Kubernetes for container orchestration.

## Current Backup Strategy

- **Gitea**: Hosted on muscle-ops-1 with RAID1 + `/backups` partition for built-in redundancy
- **Database (PostgreSQL)**: Backed up via Coolify S3 backup (managed by the hosting platform)
- **Customer apps**: Stateless by design — no backup needed unless they mount volumes
