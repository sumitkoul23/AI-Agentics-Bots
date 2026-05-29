#!/usr/bin/env bash
# deploy-local.sh — upload + instantiate + exercise the agent-registry
# contract on the local wasmd devnet started by ./scripts/devnet-up.sh.
#
# Walks the FULL lifecycle so you see, end-to-end, that the contract works
# on a real running chain:
#   1. store the WASM (uploads agentic_registry.wasm)
#   2. instantiate with default params
#   3. operator RegisterAgent  → check stake escrowed
#   4. requester CreateTask    → check bounty escrowed + task_id assigned
#   5. operator SubmitResponse → check response_cid set
#   6. requester SettleTask    → check 50/30/20 payouts via balance queries
#   7. query BurnedTotal + Reputation
set -euo pipefail

WASMD="${WASMD:-/tmp/wasmd/build/wasmd}"
HOME_DIR="${HOME_DIR:-/tmp/wasmd-data}"
NODE="${NODE:-tcp://localhost:26657}"
CHAIN_ID="${CHAIN_ID:-skymetric-devnet-1}"
DENOM="${DENOM:-stake}"
WASM_PATH="${WASM_PATH:-target/wasm32-unknown-unknown/release/agentic_registry.wasm}"

TX_FLAGS=(
  --chain-id "$CHAIN_ID"
  --node "$NODE"
  --home "$HOME_DIR"
  --keyring-backend test
  --gas auto --gas-adjustment 1.5
  --gas-prices "0.0001${DENOM}"
  --output json
  -y
)

# Fixed addresses derived from the devnet-up keys.
VAL=$("$WASMD"  keys show validator -a --keyring-backend test --home "$HOME_DIR")
OPR=$("$WASMD"  keys show operator  -a --keyring-backend test --home "$HOME_DIR")
REQ=$("$WASMD"  keys show requester -a --keyring-backend test --home "$HOME_DIR")
# Burn sink — create a key whose mnemonic we discard, so its address is
# valid bech32 for the chain prefix but practically unspendable. This is
# the cosmos-ecosystem-standard approach (vs the "all-zeros" address,
# which has a different checksum per prefix and is a footgun).
if ! "$WASMD" keys show burnsink --keyring-backend test --home "$HOME_DIR" > /dev/null 2>&1; then
  "$WASMD" keys add burnsink --keyring-backend test --home "$HOME_DIR" >/dev/null 2>&1
fi
BURN_SINK=$("$WASMD" keys show burnsink -a --keyring-backend test --home "$HOME_DIR")
TREASURY="$VAL"   # validator address doubles as treasury for the test

bal() { "$WASMD" query bank balances "$1" --node "$NODE" --output json 2>/dev/null | jq -r ".balances[] | select(.denom==\"$DENOM\") | .amount // \"0\"" | head -1; }

# Wait for tx inclusion. wasmd takes ~5s per block.
wait_tx() {
  local txhash="$1"
  for _ in $(seq 1 12); do
    sleep 1
    local code
    code=$("$WASMD" query tx "$txhash" --node "$NODE" --output json 2>/dev/null | jq -r '.code // empty')
    if [[ -n "$code" ]]; then
      [[ "$code" == "0" ]] || { echo "  tx $txhash failed with code $code"; "$WASMD" query tx "$txhash" --node "$NODE" --output json | jq -r '.raw_log'; exit 1; }
      return 0
    fi
  done
  echo "tx $txhash not included within 12s"
  exit 1
}

[[ -f "$WASM_PATH" ]] || { echo "build first: cargo build --release --target wasm32-unknown-unknown"; exit 1; }

echo "═══ 1. STORE WASM ═══"
echo "▶ Uploading $WASM_PATH"
TX=$("$WASMD" tx wasm store "$WASM_PATH" --from validator "${TX_FLAGS[@]}" | jq -r .txhash)
echo "  tx: $TX"
wait_tx "$TX"
CODE_ID=$("$WASMD" query tx "$TX" --node "$NODE" --output json | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
echo "  ✅ code_id: $CODE_ID"
echo

echo "═══ 2. INSTANTIATE ═══"
INIT_MSG=$(jq -nc \
  --arg denom "$DENOM" \
  --arg burn "$BURN_SINK" \
  --arg treasury "$TREASURY" \
  '{
    stake_denom: $denom,
    burn_sink: $burn,
    treasury: $treasury,
    min_agent_stake: "100000000",
    min_agent_stake_floor: "10000000",
    split_agent: "0.5",
    split_treasury: "0.3",
    split_burn: "0.2",
    fraud_proof_quorum: 3,
    reputation_gain_per_task: 1
  }')
echo "▶ init msg: $INIT_MSG"
TX=$("$WASMD" tx wasm instantiate "$CODE_ID" "$INIT_MSG" \
    --label agent-registry --admin "$VAL" --from validator "${TX_FLAGS[@]}" | jq -r .txhash)
wait_tx "$TX"
CONTRACT=$("$WASMD" query tx "$TX" --node "$NODE" --output json | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
echo "  ✅ contract: $CONTRACT"
echo

echo "═══ 3. OPERATOR REGISTERS ═══"
echo "  operator pre-bal: $(bal "$OPR") $DENOM"
MSG='{"register_agent":{"moniker":"pr-reviewer","endpoint":"https://reviewer.example.com"}}'
TX=$("$WASMD" tx wasm execute "$CONTRACT" "$MSG" --amount "250000000${DENOM}" --from operator "${TX_FLAGS[@]}" | jq -r .txhash)
wait_tx "$TX"
echo "  operator post-bal: $(bal "$OPR") $DENOM"
echo "  contract bal:      $(bal "$CONTRACT") $DENOM"
echo

echo "═══ 4. REQUESTER CREATES TASK ═══"
echo "  requester pre-bal: $(bal "$REQ") $DENOM"
MSG=$(jq -nc --arg agent "$OPR" '{create_task:{agent:$agent,spec:"Review PR #7 at github.com/sumitkoul23/ai-agentics-bots"}}')
TX=$("$WASMD" tx wasm execute "$CONTRACT" "$MSG" --amount "1000${DENOM}" --from requester "${TX_FLAGS[@]}" | jq -r .txhash)
wait_tx "$TX"
echo "  requester post-bal: $(bal "$REQ") $DENOM"
echo "  contract bal:       $(bal "$CONTRACT") $DENOM"
echo

echo "═══ 5. QUERY OPEN TASKS (the watch-loop path) ═══"
QUERY=$(jq -nc --arg agent "$OPR" '{open_tasks_for_agent:{agent:$agent}}')
"$WASMD" query wasm contract-state smart "$CONTRACT" "$QUERY" --node "$NODE" --output json | jq '.data.tasks'
echo

echo "═══ 6. OPERATOR SUBMITS RESPONSE ═══"
MSG='{"submit_response":{"task_id":1,"response_cid":"QmExampleReviewCidLOCAL12345"}}'
TX=$("$WASMD" tx wasm execute "$CONTRACT" "$MSG" --from operator "${TX_FLAGS[@]}" | jq -r .txhash)
wait_tx "$TX"
echo "  ✅ response submitted"
echo

echo "═══ 7. REQUESTER SETTLES TASK ═══"
OP_PRE=$(bal "$OPR"); TR_PRE=$(bal "$TREASURY"); BN_PRE=$(bal "$BURN_SINK"); CN_PRE=$(bal "$CONTRACT")
echo "  operator pre:  $OP_PRE"
echo "  treasury pre:  $TR_PRE"
echo "  burnsink pre:  $BN_PRE"
echo "  contract pre:  $CN_PRE"
MSG='{"settle_task":{"task_id":1}}'
TX=$("$WASMD" tx wasm execute "$CONTRACT" "$MSG" --from requester "${TX_FLAGS[@]}" | jq -r .txhash)
wait_tx "$TX"
OP_POST=$(bal "$OPR"); TR_POST=$(bal "$TREASURY"); BN_POST=$(bal "$BURN_SINK"); CN_POST=$(bal "$CONTRACT")
echo "  operator post: $OP_POST  (Δ$((OP_POST - OP_PRE)) — expect +500)"
echo "  treasury post: $TR_POST  (Δ$((TR_POST - TR_PRE)) — expect +300)"
echo "  burnsink post: $BN_POST  (Δ$((BN_POST - BN_PRE)) — expect +200)"
echo "  contract post: $CN_POST  (Δ$((CN_POST - CN_PRE)) — expect -1000)"
echo

echo "═══ 8. QUERY FINAL STATE ═══"
echo "Agent:"
"$WASMD" query wasm contract-state smart "$CONTRACT" \
  "$(jq -nc --arg op "$OPR" '{agent:{operator:$op}}')" --node "$NODE" --output json | jq .data.agent
echo
echo "BurnedTotal:"
"$WASMD" query wasm contract-state smart "$CONTRACT" '{"burned_total":{}}' --node "$NODE" --output json | jq .data
echo
echo "✅ Full lifecycle deployed and exercised on local devnet."
echo
echo "   chain_id: $CHAIN_ID"
echo "   contract: $CONTRACT"
echo "   code_id:  $CODE_ID"
