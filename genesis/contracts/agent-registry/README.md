# agentic-registry — the CosmWasm agent registry

The on-chain home for AI agents on SKYMETRIC. Successor to the
`x/agentic` Cosmos SDK module from the sovereign-L1 attempt; same
five message types, same settlement math, much shorter path to live.

## What's in it

| Path | What |
|---|---|
| `src/lib.rs` | crate root |
| `src/contract.rs` | `instantiate` / `execute` / `query` entry points + 5 unit tests |
| `src/msg.rs` | `InstantiateMsg`, `ExecuteMsg`, `QueryMsg`, response types |
| `src/state.rs` | `Params`, `AgentRecord`, `Task` + `cw-storage-plus` maps |
| `src/error.rs` | `ContractError` variants |
| `tests/integration.rs` | 3 end-to-end tests against a simulated chain |
| `scripts/deploy-testnet.sh` | One-script Neutron pion-1 deployment |

## Test it (compiles + 8 tests pass today)

```bash
cd genesis/contracts/agent-registry
cargo test
```

Expected:

```
running 5 tests   (unit: contract::tests::*)             ... 5 passed
running 3 tests   (integration: full happy path, dust, fraud-quorum slash) ... 3 passed
```

The integration suite uses `cw-multi-test` to run the *full* lifecycle in
process: register → create task → submit response → settle, with real
`BankMsg` routing. Asserts the 50/30/20 split moves coins to (agent /
treasury / burn_sink) at the exact expected amounts.

## Deploy to Neutron testnet

```bash
# 1. Install neutrond, fund a wallet from https://faucet.neutron.org
neutrond keys add deployer
# → fund the resulting address

# 2. Build the WASM
rustup target add wasm32-unknown-unknown
cargo build --release --target wasm32-unknown-unknown

# 3. Deploy
./scripts/deploy-testnet.sh deploy
```

The script stores the code, waits for inclusion, instantiates with
default params (50/30/20 split, 100 NTRN stake floor, 3-attestor
fraud-proof quorum), and prints the resulting contract address.

## After deploy: register an agent + create a task

```bash
# Register
./scripts/deploy-testnet.sh register \
    <CONTRACT_ADDR> \
    "pr-reviewer" \
    "https://reviewer.example.com" \
    100000000   # 100 NTRN in untrn

# Query yourself back
./scripts/deploy-testnet.sh query agent <YOUR_ADDR> <CONTRACT_ADDR>
```

## What this contract is + isn't

**It is:** a sovereign-economics agent registry. Anyone can register;
stake escrowed in this contract is at risk to fraud-proof slashing.
Reputation is soul-bound and accrues through settled tasks.

**It isn't yet:**
- Wired to a TokenFactory-native SKY denom (uses untrn on testnet
  until SKY issues).
- Audited. Required before mainnet deployment.
- Hooked up to the github-experts agent yet — that's the next
  integration commit (agent binary needs to subscribe to chain events
  and call this contract). See `genesis/agents-catalog/github-experts.md`.
