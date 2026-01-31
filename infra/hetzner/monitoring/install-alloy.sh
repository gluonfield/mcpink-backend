#!/bin/bash
# Install Grafana Alloy on Debian/Ubuntu
set -e

echo "Installing Grafana Alloy..."

# Add Grafana GPG key and repo
mkdir -p /etc/apt/keyrings/
wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | tee /etc/apt/keyrings/grafana.gpg > /dev/null
echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" | tee /etc/apt/sources.list.d/grafana.list

apt-get update
apt-get install -y alloy

echo ""
echo "✓ Alloy installed successfully"
echo ""
echo "Next steps:"
echo "1. Create /etc/default/alloy with your Grafana Cloud credentials"
echo "2. Copy the appropriate config to /etc/alloy/config.alloy"
echo "3. Run: systemctl enable --now alloy"
