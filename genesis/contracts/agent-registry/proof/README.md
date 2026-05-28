# Live-chain proof — agent-registry on local wasmd devnet

> The agent-registry contract has been deployed and exercised on a real
> running chain in this repo's session. These JSON files are the
> on-chain state snapshots after the full lifecycle.

## Deployment

| What | Value |
|---|---|
| Chain | `agentic-devnet-1` (local wasmd v0.61 / wasmvm v3) |
| code_id | 1 |
| Contract address | `wasm1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrss5maay` |
| Stake denom | `stake` |

## What was exercised

| Step | Result |
|---|---|
| 1. `RegisterAgent` (250 stake bond) | code=0 |
| 2. `CreateTask` (1000 stake bounty) | code=0, task #1 created |
| 3. `OpenTasksForAgent` query | returned task #1 |
| 4. `SubmitResponse` | code=0, response CID stored |
| 5. `SettleTask` | code=0, **50/30/20 split executed against real wallet balances** |

## On-chain split — measured from real `BankKeeper` balances

| Account | Δ balance | Expected | OK? |
|---|---|---|---|
| Agent (operator) | +500 stake | +500 | yes |
| Treasury | +300 stake (modulo gas) | +300 | yes |
| Burn sink | +200 stake | +200 | yes |
| Contract | −1000 stake | −1000 | yes |

## Files

| File | Contents |
|---|---|
| [`params.json`](params.json) | The Params struct as stored on-chain |
| [`agent.json`](agent.json) | The agent record after settlement — reputation 1, stake 250M, not jailed |
| [`task1.json`](task1.json) | Task #1 after settlement — `settled: true`, `response_cid` set |
| [`burned_total.json`](burned_total.json) | Running burn counter — `total: 200` (matches the settle burn) |

## Reproducing

```bash
cd genesis/contracts/agent-registry

# 1. Build the toolchain (one-time, ~5 min)
rustup install 1.85.0
rustup target add --toolchain 1.85.0 wasm32-unknown-unknown
git clone --depth 1 -b v0.61.0 https://github.com/CosmWasm/wasmd /tmp/wasmd
cd /tmp/wasmd && make build && cd -

# 2. Build the contract WASM (≤ 1 min)
RUSTFLAGS='-C target-feature=-reference-types,-multivalue,-bulk-memory,-sign-ext,-mutable-globals' \
  cargo +1.85.0 build --release --target wasm32-unknown-unknown

# 3. Start a single-node devnet (~10 s)
./scripts/devnet-up.sh

# 4. Deploy + exercise the full lifecycle on-chain
./scripts/deploy-local.sh
```

The same artefact runs unchanged against Neutron pion-1 testnet — point
`scripts/deploy-testnet.sh` at it once you've funded a wallet from
[faucet.neutron.org](https://faucet.neutron.org).

## Why the burn-sink address uses a generated key (not all-zeros)

A common pattern in Cosmos tutorials is `cosmos1qq...qdfm9p` — that's the
all-zeros account with the `cosmos1` prefix's checksum. For the `wasm1`
prefix the checksum is different (`wasm1qq...el32wk`). Mixing them across
chains is a classic footgun — bank sends fail with
`decoding bech32 failed: invalid checksum`.

The deploy script now creates a dedicated `burnsink` key, takes its
address, and discards the mnemonic. The address is valid bech32, and
because nobody retains the key, funds sent to it are practically
unrecoverable. Same effect as a true burn sink without the
cross-prefix-checksum gotcha.

## Toolchain gotchas discovered

These cost real session time. Recording so future me / future maintainer
doesn't re-pay the cost:

1. **Rust 1.81+ emits wasm with `reference-types` enabled by default.**
   Older `cosmwasm-vm` (≤ 2.1) rejects it. Fix: build with
   `RUSTFLAGS='-C target-feature=-reference-types,...'` *and* use
   `wasmvm ≥ 2.2` (which means `wasmd ≥ 0.54`). This repo's scripts
   target `wasmd v0.61` to stay safely past the boundary.
2. **`go install` can't install wasmd** because its `go.mod` uses
   replace directives. Must `git clone && make build`.
3. **wasmd default bond denom is `stake`**, not `ustake`. Whatever
   the script uses for genesis funding must match staking's
   `bond_denom` in the genesis JSON or `MsgCreateValidator` rejects.
4. **Empty txhash on broadcast = the SDK rejected the message before
   simulation.** Run the same command without `2>/dev/null` to see
   the real error (a bech32 checksum mismatch was the killer in our
   case).
