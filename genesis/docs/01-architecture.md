# Agent 1 — Chief Architect: System Design

## 1. Framework choice: Cosmos SDK v0.50 + CometBFT

| Option considered | Verdict | Why |
|---|---|---|
| **Cosmos SDK v0.50 + CometBFT** | ✅ Chosen | Sovereign L1, Go-native (matches this monorepo), no L1 settlement fees, free to run, app-chain ergonomics ideal for an AI-agent specific module |
| Polkadot Substrate | ❌ | Requires parachain auction or solochain isolation; Rust toolchain adds friction for an agent-first repo |
| OP-Stack / Arbitrum Orbit L2 | ❌ | Inherits L1 gas economics; sequencer requires custodial infra and Ethereum capital eventually |
| Avalanche Subnet | ❌ | Subnet validators must also validate the P-chain → non-trivial AVAX bond |

**Decision:** Cosmos SDK is the only stack that satisfies all of: zero-cost to
launch, sovereign economics, Go-native (so the agents in `agents/` can speak to
it via the same toolchain), and a clean module path for the bespoke
agent-staking logic Agent 2 will implement.

## 2. Consensus: CometBFT Proof-of-Stake

- **Validator set:** 4 at genesis, capped at 100 active by month 12.
- **Block time target:** 3 s.
- **Bond denom:** `ugen`.
- **Min self-delegation:** 1 GEN (low to allow CI / Codespaces validators).
- **Unbonding period:** 14 days (half of Cosmos Hub's 21 — friendlier for an
  agent-velocity chain, still long enough to make long-range attacks
  uneconomical).
- **Slashing:**
  - Downtime: 0.01 % per offence (lenient — free-tier nodes will flap).
  - Double-sign: 5 % + permanent jail.

No mining hardware. No GPUs. A `t2.micro`-class instance can run a validator
indefinitely.

## 3. Network topology (zero-cost)

```
                      ┌────────────────────┐
                      │   seed.agentic.dev │  (Oracle Cloud Always-Free, ARM)
                      └─────────┬──────────┘
                                │ p2p :26656
        ┌───────────────────────┼─────────────────────────┐
        │                       │                         │
┌───────▼────────┐    ┌─────────▼────────┐     ┌──────────▼────────┐
│ validator-1    │    │ validator-2      │     │ validator-3       │
│ Oracle Cloud   │    │ Fly.io free      │     │ GitHub Codespaces │
│ (always-free)  │    │ (3 × 256MB VM)   │     │ (60 hr/mo free)   │
└────────────────┘    └──────────────────┘     └───────────────────┘
                                │
                      ┌─────────▼──────────┐
                      │ explorer (Ping.pub)│  (Cloudflare Pages, free)
                      └────────────────────┘
```

Validator-4 (genesis-key holder) lives on a maintainer's laptop until a fourth
free tier comes online.

## 4. Module composition

Standard SDK modules (`auth`, `bank`, `staking`, `slashing`, `distribution`,
`gov`, `mint`, `feegrant`, `authz`, `ibc`) **plus** the bespoke module Agent 2
implements:

```
x/agentic/  → Agent registry, task escrow, reputation NFT, slash on fraud proof
```

A deliberately small surface area for v0. We do **not** ship our own VM in v0;
EVM (`evmos`) or CosmWasm bake in as a later proposal once the chain has
governance participants.

## 5. Genesis allocation (summary — full table in `02-tokenomics.md`)

| Bucket | % | Cliff | Vesting |
|---|---|---|---|
| Community airdrop (Cosmos + Ethereum AI builders) | 40 % | — | linear over 6 months from claim |
| Validators / staking rewards pool | 25 % | — | streamed by `x/mint` |
| Treasury (governance-controlled) | 20 % | 6 mo | linear over 36 months |
| Genesis contributors | 10 % | 12 mo | linear over 24 months |
| Liquidity bootstrapping (DEX seeding via testnet faucet flips) | 5 % | — | unlocked at mainnet |

No private sale. No VC allocation. The chain has literally never seen a dollar
of equity capital, which is the entire point.

## 6. Roadmap

| Phase | Target | Deliverable |
|---|---|---|
| 0 — *now* | week 0 | This scaffold compiles to a single-node devnet |
| 1 — Devnet  | week 2 | `x/agentic` module skeleton, 4-validator devnet on free tiers |
| 2 — Public testnet `agentic-test-1` | month 2 | Faucet, explorer, first agents from this repo staking GEN |
| 3 — Mainnet `agentic-1` | month 6 | Audited via open-source toolchain (Slither for EVM contracts isn't applicable; we'll use [`Cosmos-SDK`'s built-in invariants](https://docs.cosmos.network/v0.50/build/building-modules/invariants) and community review) |
| 4 — IBC | month 9 | Open IBC channels to Osmosis, Neutron |
| 5 — Agent VM | year 1+ | CosmWasm or custom executor for on-chain agent logic |

## 7. Risk register (and the $0 mitigations)

| Risk | Mitigation |
|---|---|
| Free-tier node eviction | 4-of-N validators across distinct providers; auto-restart via `systemd` units in `deploy/` |
| Sybil airdrop farming | Eligibility scoped to verifiable on-chain history (Cosmos staker / Ethereum AI-protocol user) |
| Key compromise on shared dev infra | Genesis keys generated **offline**, only the operator-account keys live on free-tier hosts |
| Lack of liquidity at mainnet | Liquidity bootstrap pool seeded from the testnet-faucet flip strategy in `04-growth-strategy.md` |
| No audit budget | Public bug-bounty in GEN; module-level invariants enforced at every block; long unbonding to make exploits unprofitable |
