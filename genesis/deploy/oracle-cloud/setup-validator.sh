#!/usr/bin/env bash
# setup-validator.sh — provision an Oracle Cloud Always-Free ARM Ampere
# instance (Ubuntu 22.04 ARM64) as an AGENTIC validator.
#
# Run as root once per machine:
#   curl -fsSL https://raw.githubusercontent.com/sumitkoul23/ai-agentics-bots/main/genesis/deploy/oracle-cloud/setup-validator.sh | sudo bash
#
# Requires env: MONIKER, CHAIN_ID, SEED_NODE.
set -euxo pipefail

: "${MONIKER:?MONIKER not set}"
: "${CHAIN_ID:?CHAIN_ID not set}"
: "${SEED_NODE:?SEED_NODE not set}"
RELEASE_URL="${RELEASE_URL:-https://github.com/sumitkoul23/agentic-chain/releases/latest/download/agenticd-linux-arm64.tar.gz}"

apt-get update
apt-get install -y curl jq tar ca-certificates ufw

# Firewall: open only the p2p + rpc ports we need.
ufw allow 22/tcp
ufw allow 26656/tcp
ufw allow 26657/tcp
ufw --force enable

useradd -m -s /bin/bash agentic || true
install -d -o agentic -g agentic /var/lib/agenticd

# Fetch and install the prebuilt binary.
TMP=$(mktemp -d)
curl -fSL "${RELEASE_URL}" -o "${TMP}/agenticd.tgz"
tar -xzf "${TMP}/agenticd.tgz" -C /usr/local/bin agenticd
chown root:root /usr/local/bin/agenticd && chmod 0755 /usr/local/bin/agenticd

sudo -u agentic agenticd init "${MONIKER}" --chain-id "${CHAIN_ID}" --home /var/lib/agenticd

# Pull the live genesis.json published by val-1.
GENESIS_URL="${GENESIS_URL:-https://raw.githubusercontent.com/sumitkoul23/agentic-chain/main/networks/${CHAIN_ID}/genesis.json}"
curl -fSL "${GENESIS_URL}" -o /var/lib/agenticd/config/genesis.json
chown agentic:agentic /var/lib/agenticd/config/genesis.json

# Wire the seed node.
sed -i "s|^seeds *=.*|seeds = \"${SEED_NODE}\"|" /var/lib/agenticd/config/config.toml

cat >/etc/systemd/system/agenticd.service <<UNIT
[Unit]
Description=AGENTIC validator (${MONIKER})
After=network-online.target
Wants=network-online.target

[Service]
User=agentic
Group=agentic
ExecStart=/usr/local/bin/agenticd start --home /var/lib/agenticd --minimum-gas-prices 0.0001ugen
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
UNIT

systemctl daemon-reload
systemctl enable --now agenticd

echo
echo "✅ agenticd is running. Tail logs with:  journalctl -fu agenticd"
