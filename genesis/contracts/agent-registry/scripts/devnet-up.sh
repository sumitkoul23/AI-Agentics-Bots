#!/usr/bin/env bash
# devnet-up.sh — start a single-node wasmd devnet locally, ready for the
# agent-registry contract to be uploaded against. ~30 seconds end-to-end.
#
# This is the script the user can run on any laptop to reproduce
# everything we do in this session, without paying for or depending on
# Neutron testnet.
set -euo pipefail

WASMD="${WASMD:-/tmp/wasmd/build/wasmd}"
CHAIN_ID="${CHAIN_ID:-agentic-devnet-1}"
HOME_DIR="${HOME_DIR:-/tmp/wasmd-data}"
DENOM="${DENOM:-stake}"
RPC_PORT="${RPC_PORT:-26657}"

if [[ ! -x "$WASMD" ]]; then
  echo "wasmd binary not at $WASMD. Build first:" >&2
  echo "  git clone --depth 1 -b v0.53.0 https://github.com/CosmWasm/wasmd /tmp/wasmd && cd /tmp/wasmd && make build" >&2
  exit 1
fi

echo "▶ Wiping any prior state at $HOME_DIR"
rm -rf "$HOME_DIR"

echo "▶ wasmd init"
"$WASMD" init devnet --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null 2>&1

# Two test keys: validator (genesis self-stake + admin) and operator (agent).
echo "▶ Creating test keys (keyring: test)"
"$WASMD" keys add validator --keyring-backend test --home "$HOME_DIR" 2>&1 | tail -n 1 >/dev/null
"$WASMD" keys add operator  --keyring-backend test --home "$HOME_DIR" 2>&1 | tail -n 1 >/dev/null
"$WASMD" keys add requester --keyring-backend test --home "$HOME_DIR" 2>&1 | tail -n 1 >/dev/null

VAL=$("$WASMD"  keys show validator -a --keyring-backend test --home "$HOME_DIR")
OPR=$("$WASMD"  keys show operator  -a --keyring-backend test --home "$HOME_DIR")
REQ=$("$WASMD"  keys show requester -a --keyring-backend test --home "$HOME_DIR")

echo "  validator: $VAL"
echo "  operator:  $OPR"
echo "  requester: $REQ"

echo "▶ Pre-funding accounts"
"$WASMD" genesis add-genesis-account "$VAL" "1000000000000${DENOM}" --home "$HOME_DIR" >/dev/null 2>&1
"$WASMD" genesis add-genesis-account "$OPR" "1000000000000${DENOM}" --home "$HOME_DIR" >/dev/null 2>&1
"$WASMD" genesis add-genesis-account "$REQ" "1000000000000${DENOM}" --home "$HOME_DIR" >/dev/null 2>&1

echo "▶ Creating gentx"
"$WASMD" genesis gentx validator "100000000${DENOM}" \
  --chain-id "$CHAIN_ID" --keyring-backend test --home "$HOME_DIR" >/dev/null 2>&1

echo "▶ Collecting gentxs"
"$WASMD" genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1

# Set tiny gas prices in app.toml so test txs cost ~0.
sed -i 's|^minimum-gas-prices.*|minimum-gas-prices = "0.0001'"${DENOM}"'"|' "$HOME_DIR/config/app.toml"

# Bind RPC to a known port.
sed -i "s|tcp://127.0.0.1:26657|tcp://0.0.0.0:${RPC_PORT}|" "$HOME_DIR/config/config.toml"

echo "▶ Starting node in background — logs at /tmp/wasmd-data/devnet.log"
nohup "$WASMD" start --home "$HOME_DIR" --minimum-gas-prices "0.0001${DENOM}" \
  > /tmp/wasmd-data/devnet.log 2>&1 &
echo $! > /tmp/wasmd-data/wasmd.pid
sleep 6

# Verify it's producing blocks.
HEIGHT=$("$WASMD" status --node "tcp://localhost:${RPC_PORT}" 2>/dev/null | grep -oE '"height":"[0-9]+"' | head -1 | grep -oE '[0-9]+')
if [[ -z "${HEIGHT:-}" ]]; then
  echo "❌ node didn't start. Tail of log:" >&2
  tail -20 /tmp/wasmd-data/devnet.log >&2
  exit 1
fi
echo
echo "✅ devnet up at tcp://localhost:${RPC_PORT}"
echo "   height: $HEIGHT"
echo "   pid:    $(cat /tmp/wasmd-data/wasmd.pid)"
echo "   logs:   tail -f /tmp/wasmd-data/devnet.log"
echo
echo "Next: ./scripts/deploy-local.sh"
