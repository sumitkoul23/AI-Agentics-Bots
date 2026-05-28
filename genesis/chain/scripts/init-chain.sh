#!/usr/bin/env bash
# init-chain.sh — bring up a fresh single-node devnet of the Skymetric chain.
#
# Run from `genesis/chain/`:
#   make install
#   ./scripts/init-chain.sh
#
# Idempotency: rm -rf $AGENTICD_HOME first if you want a clean slate.
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-skymetric-devnet-1}"
MONIKER="${MONIKER:-genesis-node}"
KEYRING="${KEYRING:-test}"        # `test` is fine for devnet; do NOT use on mainnet
AGENTICD_HOME="${AGENTICD_HOME:-$HOME/.skymetricd}"
DENOM="usky"

VAL_KEY="validator"
FAUCET_KEY="faucet"

GENESIS_BALANCE="500000000000000${DENOM}" # 500M SKY — validator + dev faucet share
SELF_DELEGATION="1000000000${DENOM}"      # 1 000 SKY self-stake at genesis

echo "▶ Wiping any prior state at ${AGENTICD_HOME}"
rm -rf "${AGENTICD_HOME}"

echo "▶ skymetricd init"
skymetricd init "${MONIKER}" --chain-id "${CHAIN_ID}" --home "${AGENTICD_HOME}"

echo "▶ Creating validator + faucet keys (keyring: ${KEYRING})"
skymetricd keys add "${VAL_KEY}"    --keyring-backend "${KEYRING}" --home "${AGENTICD_HOME}"
skymetricd keys add "${FAUCET_KEY}" --keyring-backend "${KEYRING}" --home "${AGENTICD_HOME}"

VAL_ADDR=$(skymetricd keys show "${VAL_KEY}"    -a --keyring-backend "${KEYRING}" --home "${AGENTICD_HOME}")
FAUCET_ADDR=$(skymetricd keys show "${FAUCET_KEY}" -a --keyring-backend "${KEYRING}" --home "${AGENTICD_HOME}")

echo "  validator: ${VAL_ADDR}"
echo "  faucet:    ${FAUCET_ADDR}"

echo "▶ Pre-funding genesis accounts"
skymetricd genesis add-genesis-account "${VAL_ADDR}"    "${GENESIS_BALANCE}" --home "${AGENTICD_HOME}"
skymetricd genesis add-genesis-account "${FAUCET_ADDR}" "${GENESIS_BALANCE}" --home "${AGENTICD_HOME}"

echo "▶ Generating gentx"
skymetricd genesis gentx "${VAL_KEY}" "${SELF_DELEGATION}" \
    --chain-id "${CHAIN_ID}" \
    --keyring-backend "${KEYRING}" \
    --home "${AGENTICD_HOME}"

echo "▶ Collecting gentxs into final genesis.json"
skymetricd genesis collect-gentxs --home "${AGENTICD_HOME}"

echo "▶ Validating genesis"
skymetricd genesis validate-genesis --home "${AGENTICD_HOME}"

echo
echo "✅ Devnet initialised at ${AGENTICD_HOME}"
echo "   Next: ./scripts/start-node.sh"
