#!/usr/bin/env bash
# deploy-testnet.sh — upload + instantiate the agent-registry contract on
# Neutron testnet (pion-1).
#
# Prerequisites on your machine:
#   1. `neutrond` installed: https://docs.neutron.org/neutron/build-and-run/install
#   2. A funded testnet wallet:
#        neutrond keys add deployer
#        # → fund from https://faucet.neutron.org with the resulting address
#   3. The contract built to WASM:
#        cd genesis/contracts/agent-registry
#        rustup target add wasm32-unknown-unknown
#        cargo build --release --target wasm32-unknown-unknown
#        # optimised binary lives at target/wasm32-unknown-unknown/release/agentic_registry.wasm
#
# Usage:
#   ./scripts/deploy-testnet.sh deploy        # store + instantiate
#   ./scripts/deploy-testnet.sh store-only    # just upload the code
#   ./scripts/deploy-testnet.sh instantiate <CODE_ID>
#   ./scripts/deploy-testnet.sh query agent <CONTRACT_ADDR> <OPERATOR_ADDR>
#
# Env overrides:
#   CHAIN_ID         default "pion-1"
#   NODE             default "https://rpc-falcron.pion-1.ntrn.tech:443"
#   WALLET           default "deployer"
#   STAKE_DENOM      default "untrn"  (Neutron's native token, used as stake
#                                      denom until GEN issues as TokenFactory)
#   GAS_PRICES       default "0.05untrn"
#   TREASURY         default = $(neutrond keys show $WALLET -a)
#   BURN_SINK        default "neutron1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqkqyfk7"
#                    (null bech32 — funds are unspendable once sent here)

set -euo pipefail

CHAIN_ID="${CHAIN_ID:-pion-1}"
NODE="${NODE:-https://rpc-falcron.pion-1.ntrn.tech:443}"
WALLET="${WALLET:-deployer}"
STAKE_DENOM="${STAKE_DENOM:-untrn}"
GAS_PRICES="${GAS_PRICES:-0.05untrn}"
TREASURY="${TREASURY:-$(neutrond keys show "$WALLET" -a 2>/dev/null || echo MISSING)}"
BURN_SINK="${BURN_SINK:-neutron1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqkqyfk7}"

WASM_PATH="${WASM_PATH:-target/wasm32-unknown-unknown/release/agentic_registry.wasm}"

COMMON_FLAGS=(
  --chain-id "$CHAIN_ID"
  --node "$NODE"
  --from "$WALLET"
  --gas auto
  --gas-adjustment 1.5
  --gas-prices "$GAS_PRICES"
  --output json
  -y
)

cmd="${1:-deploy}"
shift || true

case "$cmd" in
  deploy)
    [[ "$TREASURY" = "MISSING" ]] && { echo "Wallet '$WALLET' not found. Run: neutrond keys add $WALLET"; exit 1; }
    [[ ! -f "$WASM_PATH" ]] && { echo "WASM not found at $WASM_PATH. Build first:"; \
      echo "  cargo build --release --target wasm32-unknown-unknown"; exit 1; }

    echo "▶ Storing code from $WASM_PATH …"
    TX_HASH=$(neutrond tx wasm store "$WASM_PATH" "${COMMON_FLAGS[@]}" | jq -r .txhash)
    echo "  tx: $TX_HASH"

    # Wait for inclusion and extract the code_id from the events.
    echo "▶ Waiting 12s for block inclusion …"
    sleep 12
    CODE_ID=$(neutrond query tx "$TX_HASH" --node "$NODE" --output json \
      | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
    echo "  code_id: $CODE_ID"

    INIT_MSG=$(cat <<JSON
{
  "stake_denom": "$STAKE_DENOM",
  "burn_sink": "$BURN_SINK",
  "treasury": "$TREASURY",
  "min_agent_stake": "100000000",
  "min_agent_stake_floor": "10000000",
  "split_agent": "0.5",
  "split_treasury": "0.3",
  "split_burn": "0.2",
  "fraud_proof_quorum": 3,
  "reputation_gain_per_task": 1
}
JSON
)
    echo "▶ Instantiating code_id=$CODE_ID …"
    TX_HASH=$(neutrond tx wasm instantiate "$CODE_ID" "$INIT_MSG" \
      --label "agent-registry" \
      --admin "$(neutrond keys show "$WALLET" -a)" \
      "${COMMON_FLAGS[@]}" | jq -r .txhash)
    echo "  tx: $TX_HASH"
    sleep 12
    CONTRACT=$(neutrond query tx "$TX_HASH" --node "$NODE" --output json \
      | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
    echo
    echo "✅ Deployed."
    echo "   code_id:  $CODE_ID"
    echo "   contract: $CONTRACT"
    echo
    echo "Try a query:"
    echo "  $0 query params $CONTRACT"
    ;;

  query)
    sub="${1:?usage: $0 query params|agent <addr>|task <id> <CONTRACT>}"
    shift
    case "$sub" in
      params)
        CONTRACT="${1:?contract address required}"
        neutrond query wasm contract-state smart "$CONTRACT" '{"params":{}}' \
          --node "$NODE" --output json | jq
        ;;
      agent)
        OPERATOR="${1:?operator address required}"
        CONTRACT="${2:?contract address required}"
        neutrond query wasm contract-state smart "$CONTRACT" \
          "{\"agent\":{\"operator\":\"$OPERATOR\"}}" \
          --node "$NODE" --output json | jq
        ;;
      task)
        ID="${1:?task id required}"
        CONTRACT="${2:?contract address required}"
        neutrond query wasm contract-state smart "$CONTRACT" \
          "{\"task\":{\"task_id\":$ID}}" \
          --node "$NODE" --output json | jq
        ;;
      burned)
        CONTRACT="${1:?contract address required}"
        neutrond query wasm contract-state smart "$CONTRACT" '{"burned_total":{}}' \
          --node "$NODE" --output json | jq
        ;;
      *) echo "unknown query subcommand: $sub"; exit 1;;
    esac
    ;;

  register)
    CONTRACT="${1:?contract address required}"
    MONIKER="${2:?moniker required}"
    ENDPOINT="${3:?endpoint URL required}"
    STAKE="${4:?stake amount (in $STAKE_DENOM) required}"

    MSG="{\"register_agent\":{\"moniker\":\"$MONIKER\",\"endpoint\":\"$ENDPOINT\"}}"
    echo "▶ Registering '$MONIKER' with $STAKE$STAKE_DENOM stake …"
    neutrond tx wasm execute "$CONTRACT" "$MSG" \
      --amount "${STAKE}${STAKE_DENOM}" \
      "${COMMON_FLAGS[@]}"
    ;;

  *)
    cat <<USAGE
deploy-testnet.sh — Neutron testnet deployment + interaction helper

  deploy                                   store + instantiate the contract
  query params   <CONTRACT>                show current Params
  query agent    <OPERATOR> <CONTRACT>     query an agent record
  query task     <ID> <CONTRACT>           query a task
  query burned   <CONTRACT>                running total burned
  register       <CONTRACT> <MONIKER> <ENDPOINT> <STAKE>
                                           bond stake + register as an agent

Override defaults via env vars: see top of this file.
USAGE
    ;;
esac
