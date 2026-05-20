#!/usr/bin/env bash
# create-validator.sh — submit a CreateValidator tx from an already-funded key
# to join an existing AGENTIC network (testnet or mainnet).
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-agentic-test-1}"
KEYRING="${KEYRING:-os}"
MONIKER="${MONIKER:-$(hostname)}"
AMOUNT="${AMOUNT:-1000000000ugen}"      # 1 000 GEN self-bond
COMMISSION_RATE="${COMMISSION_RATE:-0.05}"
COMMISSION_MAX_RATE="${COMMISSION_MAX_RATE:-0.20}"
COMMISSION_MAX_CHANGE="${COMMISSION_MAX_CHANGE:-0.01}"
KEY_NAME="${KEY_NAME:-validator}"
NODE="${NODE:-tcp://localhost:26657}"

PUBKEY_JSON=$(agenticd tendermint show-validator)

cat > /tmp/validator.json <<JSON
{
  "pubkey": ${PUBKEY_JSON},
  "amount": "${AMOUNT}",
  "moniker": "${MONIKER}",
  "commission-rate": "${COMMISSION_RATE}",
  "commission-max-rate": "${COMMISSION_MAX_RATE}",
  "commission-max-change-rate": "${COMMISSION_MAX_CHANGE}",
  "min-self-delegation": "1"
}
JSON

agenticd tx staking create-validator /tmp/validator.json \
    --from "${KEY_NAME}" \
    --chain-id "${CHAIN_ID}" \
    --keyring-backend "${KEYRING}" \
    --node "${NODE}" \
    --gas auto --gas-adjustment 1.4 --gas-prices 0.0005ugen -y
