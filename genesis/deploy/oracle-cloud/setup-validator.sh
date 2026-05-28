#!/usr/bin/env bash
# setup-validator.sh — provision an Oracle Cloud Always-Free ARM Ampere
# instance (Ubuntu 22.04 ARM64) as an SKYMETRIC validator.
#
# Run as root once per machine:
#   curl -fsSL https://raw.githubusercontent.com/sumitkoul23/ai-agentics-bots/main/genesis/deploy/oracle-cloud/setup-validator.sh | sudo bash
#
# Requires env: MONIKER, CHAIN_ID, SEED_NODE.
set -euxo pipefail

: "${MONIKER:?MONIKER not set}"
: "${CHAIN_ID:?CHAIN_ID not set}"
: "${SEED_NODE:?SEED_NODE not set}"
RELEASE_URL="${RELEASE_URL:-https://github.com/sumitkoul23/agentic-chain/releases/latest/download/skymetricd-linux-arm64.tar.gz}"

apt-get update
apt-get install -y curl jq tar ca-certificates ufw

# Firewall: open only the p2p + rpc ports we need.
ufw allow 22/tcp
ufw allow 26656/tcp
ufw allow 26657/tcp
ufw --force enable

useradd -m -s /bin/bash agentic || true
install -d -o agentic -g agentic /var/lib/skymetricd

# Fetch and install the prebuilt binary.
TMP=$(mktemp -d)
curl -fSL "${RELEASE_URL}" -o "${TMP}/skymetricd.tgz"
tar -xzf "${TMP}/skymetricd.tgz" -C /usr/local/bin skymetricd
chown root:root /usr/local/bin/skymetricd && chmod 0755 /usr/local/bin/skymetricd

sudo -u agentic skymetricd init "${MONIKER}" --chain-id "${CHAIN_ID}" --home /var/lib/skymetricd

# Pull the live genesis.json published by val-1.
GENESIS_URL="${GENESIS_URL:-https://raw.githubusercontent.com/sumitkoul23/agentic-chain/main/networks/${CHAIN_ID}/genesis.json}"
curl -fSL "${GENESIS_URL}" -o /var/lib/skymetricd/config/genesis.json
chown agentic:agentic /var/lib/skymetricd/config/genesis.json

# Wire the seed node.
sed -i "s|^seeds *=.*|seeds = \"${SEED_NODE}\"|" /var/lib/skymetricd/config/config.toml

cat >/etc/systemd/system/skymetricd.service <<UNIT
[Unit]
Description=SKYMETRIC validator (${MONIKER})
After=network-online.target
Wants=network-online.target

[Service]
User=agentic
Group=agentic
ExecStart=/usr/local/bin/skymetricd start --home /var/lib/skymetricd --minimum-gas-prices 0.0001ugen
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
UNIT

systemctl daemon-reload
systemctl enable --now skymetricd

echo
echo "✅ skymetricd is running. Tail logs with:  journalctl -fu skymetricd"
