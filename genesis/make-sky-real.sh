#!/usr/bin/env bash
# make-sky-real.sh — Run this on YOUR machine to make SKY exist on a real chain.
#
# What it does (15 minutes, $0 cost):
#   1. Install neutrond (Neutron CLI)
#   2. Create an admin wallet — YOU save the mnemonic
#   3. YOU fund it from faucet.neutron.org
#   4. Mint 1,000,000,000 SKY via Neutron TokenFactory
#   5. Build the agent-registry WASM (Rust required)
#   6. Deploy the contract with SKY as stake denom
#   7. Print the live contract URL + write DEPLOYMENT.md
#
# Requirements: curl, jq, git, Rust   (https://rustup.rs)
#
# Usage:
#   chmod +x make-sky-real.sh && ./make-sky-real.sh

set -euo pipefail

GRN='\033[0;32m'; CYN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${CYN}▶ $*${NC}"; }
ok()    { echo -e "${GRN}✅ $*${NC}"; }
die()   { echo -e "${RED}❌ $*${NC}"; exit 1; }
pause() {
  echo -e "\n${CYN}━━━━ $* ━━━━${NC}"
  read -rp "   Press Enter when ready → "
}

CHAIN_ID="pion-1"
NODE="https://rpc-falcron.pion-1.ntrn.tech:443"
GAS_PRICES="0.05untrn"
WALLET="skymetric-admin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ── 1. Install neutrond ───────────────────────────────────────────────────────
info "Step 1/7 — neutrond CLI"

if ! command -v neutrond &>/dev/null; then
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)
  [[ "$ARCH" == "x86_64" ]]   && ARCH="amd64"
  [[ "$ARCH" =~ arm|aarch64 ]] && ARCH="arm64"

  # Pinned to the pion-1 testnet release. If you want mainnet swap the tag.
  VERSION="v11.0.1-testnet"
  URL="https://github.com/neutron-org/neutron/releases/download/${VERSION}/neutrond-${OS}-${ARCH}"
  DEST="$HOME/.local/bin/neutrond"
  mkdir -p "$HOME/.local/bin"
  info "  Downloading neutrond ${VERSION} …"
  curl -fL "$URL" -o "$DEST" || die "Download failed. Install manually: https://docs.neutron.org/neutron/build-and-run/install"
  chmod +x "$DEST"
  export PATH="$HOME/.local/bin:$PATH"
fi

command -v neutrond &>/dev/null || export PATH="$HOME/.local/bin:$PATH"
ok "neutrond: $(neutrond version 2>&1 | head -1)"

# ── 2. Create admin wallet ────────────────────────────────────────────────────
info "Step 2/7 — Create admin wallet"

if ! neutrond keys show "$WALLET" &>/dev/null; then
  echo ""
  echo "  Creating wallet '$WALLET'."
  echo "  ⚠️  WRITE DOWN THE MNEMONIC NOW. It will not be shown again."
  echo ""
  neutrond keys add "$WALLET"
fi

ADMIN_ADDR=$(neutrond keys show "$WALLET" -a)
ok "Wallet: $ADMIN_ADDR"

# ── 3. Fund from faucet ───────────────────────────────────────────────────────
pause "Step 3/7 — Fund from testnet faucet

  Open this in your browser → https://faucet.neutron.org

  Paste:  $ADMIN_ADDR

  Request NTRN, wait ~30 seconds, then press Enter."

BALANCE=$(neutrond query bank balances "$ADMIN_ADDR" --node "$NODE" --output json 2>/dev/null \
  | jq -r '.balances[] | select(.denom=="untrn") | .amount' || echo "0")
[[ -z "$BALANCE" ]] && BALANCE="0"
[[ "$BALANCE" -lt 1000000 ]] && die "Balance ${BALANCE}untrn — need ≥1 NTRN. Fund first: https://faucet.neutron.org"
ok "Balance: ${BALANCE}untrn ($(( BALANCE / 1000000 )) NTRN)"

# ── 4. Create SKY denom via TokenFactory ─────────────────────────────────────
info "Step 4/7 — Mint 1B SKY via TokenFactory"

SKY_DENOM="factory/${ADMIN_ADDR}/usky"

EXISTING=$(neutrond query tokenfactory denom-authority-metadata "$SKY_DENOM" \
  --node "$NODE" --output json 2>/dev/null | jq -r '.authority_metadata.admin' || echo "")

if [[ "$EXISTING" != "$ADMIN_ADDR" ]]; then
  info "  Creating denom …"
  neutrond tx tokenfactory create-denom usky \
    --chain-id "$CHAIN_ID" --node "$NODE" \
    --from "$WALLET" --gas auto --gas-adjustment 1.5 \
    --gas-prices "$GAS_PRICES" -y --output json | jq -r '"  create-denom tx: " + .txhash'
  sleep 8
fi
ok "Denom: $SKY_DENOM"

info "  Minting 1,000,000,000 SKY (1_000_000_000_000_000 usky) …"
MINT_TX=$(neutrond tx tokenfactory mint "1000000000000000${SKY_DENOM}" \
  --chain-id "$CHAIN_ID" --node "$NODE" \
  --from "$WALLET" --gas auto --gas-adjustment 1.5 \
  --gas-prices "$GAS_PRICES" -y --output json | jq -r .txhash)
echo "  mint tx: $MINT_TX"
sleep 8

SKY_BAL=$(neutrond query bank balances "$ADMIN_ADDR" --node "$NODE" --output json \
  | jq -r --arg d "$SKY_DENOM" '.balances[] | select(.denom==$d) | .amount' || echo "0")
ok "SKY balance: ${SKY_BAL:-0} usky"

# ── 5. Build the WASM ────────────────────────────────────────────────────────
info "Step 5/7 — Build CosmWasm contract"

CONTRACT_DIR="$SCRIPT_DIR/contracts/agent-registry"
WASM="$CONTRACT_DIR/target/wasm32-unknown-unknown/release/agentic_registry.wasm"

if [[ ! -f "$WASM" ]]; then
  command -v rustup &>/dev/null || die "Rust not found. Install at https://rustup.rs"
  rustup target add wasm32-unknown-unknown
  cd "$CONTRACT_DIR"
  info "  Building WASM (~60 s first time) …"
  RUSTFLAGS='-C target-feature=-reference-types,-multivalue,-bulk-memory,-sign-ext,-mutable-globals' \
    cargo build --release --target wasm32-unknown-unknown
  cd -
fi
ok "WASM: $(du -sh "$WASM" | cut -f1) at $WASM"

# ── 6. Burn sink — valid bech32 for neutron1 prefix ─────────────────────────
#
# Why not a hardcoded "all-zeros" address?
# The bech32 checksum is computed from the prefix ("neutron1"), so the
# all-zeros account valid for "cosmos1" has a *different* checksum here.
# Pasting the cosmos1 null address causes "invalid checksum" on Neutron.
# We generate a real key, take its address, and throw the mnemonic away —
# funds sent to it are unspendable because nobody has the private key.
#
if ! neutrond keys show burnsink &>/dev/null; then
  info "  Generating burn-sink key (mnemonic discarded after this) …"
  neutrond keys add burnsink --output json >/dev/null 2>&1
fi
BURN_SINK=$(neutrond keys show burnsink -a)
info "  Burn sink: $BURN_SINK"

# ── 7. Deploy contract ────────────────────────────────────────────────────────
info "Step 6/7 — Upload WASM"

STORE_TX=$(neutrond tx wasm store "$WASM" \
  --chain-id "$CHAIN_ID" --node "$NODE" \
  --from "$WALLET" --gas auto --gas-adjustment 2 \
  --gas-prices "$GAS_PRICES" -y --output json | jq -r .txhash)
echo "  store tx: $STORE_TX"
sleep 15

CODE_ID=$(neutrond query tx "$STORE_TX" --node "$NODE" --output json \
  | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
[[ -z "$CODE_ID" ]] && die "Could not extract code_id from tx $STORE_TX. Run: neutrond query tx $STORE_TX --node $NODE --output json | jq"
ok "code_id: $CODE_ID"

info "Step 7/7 — Instantiate"

# admin is optional (defaults to info.sender) but we pass it explicitly
# so the contract's admin field is unambiguous.
INIT_MSG=$(cat <<JSON
{
  "stake_denom": "$SKY_DENOM",
  "burn_sink": "$BURN_SINK",
  "treasury": "$ADMIN_ADDR",
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

INST_TX=$(neutrond tx wasm instantiate "$CODE_ID" "$INIT_MSG" \
  --label "skymetric-agent-registry" \
  --admin  "$ADMIN_ADDR" \
  --chain-id "$CHAIN_ID" --node "$NODE" \
  --from "$WALLET" --gas auto --gas-adjustment 2 \
  --gas-prices "$GAS_PRICES" -y --output json | jq -r .txhash)
echo "  instantiate tx: $INST_TX"
sleep 15

CONTRACT=$(neutrond query tx "$INST_TX" --node "$NODE" --output json \
  | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
[[ -z "$CONTRACT" ]] && die "Could not read contract address from tx $INST_TX"

# ── Write DEPLOYMENT.md ───────────────────────────────────────────────────────
cat >"$SCRIPT_DIR/DEPLOYMENT.md" <<MD
# Skymetric — Live Deployment

| | |
|---|---|
| Chain | Neutron testnet \`pion-1\` |
| SKY denom | \`$SKY_DENOM\` |
| Code ID | $CODE_ID |
| Contract | \`$CONTRACT\` |
| Admin | \`$ADMIN_ADDR\` |
| Burn sink | \`$BURN_SINK\` |
| Deployed | $(date -u +"%Y-%m-%d %H:%M UTC") |

## Explorer

https://neutron.celat.one/pion-1/contracts/$CONTRACT

## Admin dashboard

Open \`genesis/site/admin/index.html\` in a browser and enter:
- Contract address: \`$CONTRACT\`
- Chain ID: \`pion-1\`
- RPC: \`https://rpc-falcron.pion-1.ntrn.tech:443\`

Connect Keplr (import your \`$WALLET\` mnemonic or use the same address).

## Register as an agent

\`\`\`bash
neutrond tx wasm execute "$CONTRACT" \\
  '{"register_agent":{"moniker":"my-agent","endpoint":"https://my-agent.example.com"}}' \\
  --amount "100000000${SKY_DENOM}" \\
  --from $WALLET \\
  --chain-id pion-1 --node https://rpc-falcron.pion-1.ntrn.tech:443 \\
  --gas auto --gas-adjustment 1.5 --gas-prices 0.05untrn -y
\`\`\`

## Whitelist a denom (e.g. peaq PHN once you have the real IBC hash)

\`\`\`bash
neutrond tx wasm execute "$CONTRACT" \\
  '{"add_stake_denom":{"denom":"ibc/<REAL-HASH>"}}' \\
  --from $WALLET --chain-id pion-1 --node https://rpc-falcron.pion-1.ntrn.tech:443 \\
  --gas auto --gas-adjustment 1.5 --gas-prices 0.05untrn -y
\`\`\`
MD

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════"
echo ""
ok "SKY IS LIVE ON NEUTRON TESTNET"
echo ""
echo "  Denom    : $SKY_DENOM"
echo "  Contract : $CONTRACT"
echo "  Explorer : https://neutron.celat.one/pion-1/contracts/$CONTRACT"
echo ""
echo "  You hold ${SKY_BAL:-0} usky in $ADMIN_ADDR"
echo "  Deployment saved → $SCRIPT_DIR/DEPLOYMENT.md"
echo ""
echo "══════════════════════════════════════════════════"
