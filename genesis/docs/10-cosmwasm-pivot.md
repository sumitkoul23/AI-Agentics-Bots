# Pivot: CosmWasm on Neutron, not sovereign Cosmos SDK L1

**Decision date:** session of 2026-05-21
**Decision-maker:** project maintainer
**Status:** active

## What changed

Previously: the Skymetric chain was being scaffolded as a sovereign Cosmos
SDK L1 (own validators, own consensus, own fee market). After the
honest-launch-timeline assessment in
[`09-honest-launch-timeline.md`](09-honest-launch-timeline.md) and an
hour of real compile-fixing in
[the chain code](../chain/), the sovereign-L1 path is no longer the
correct bet for this team's resources.

New direction: **SKYMETRIC ships as a suite of CosmWasm smart contracts on
Neutron.** Same name, same brand, same token economics — different
substrate.

## Why this is the right call (not a retreat)

| Dimension | Sovereign L1 | CosmWasm on Neutron |
|---|---|---|
| Time to live product | 9–24+ months | **3–4 months** |
| Audit cost | $80–200k (whole chain) | **$30–50k** (per contract) |
| Validator recruitment | 12–20 external operators needed | **0** — inherit Neutron's set |
| Consensus halt risk | Real — bug = chain dead | **None** — Neutron's consensus is independent |
| IBC connectivity | Manual channel setup | **Inherited** |
| Day-1 liquidity | Zero — bootstrap from scratch | **Tens of millions** in Neutron's existing pools |
| Token issuance | Native denom | CW20 or Token Factory denom (both equivalent for our use) |
| Brand | "Our own chain" | "Currently a CosmWasm protocol — migrating to sovereign L1 when traction warrants" |

The only column where sovereign L1 wins is the abstract brand claim.
That claim doesn't pay for the audit, doesn't recruit validators, and
doesn't ship the product. Time-to-market beats narrative — by a lot.

## What we keep

Almost everything. The pivot changes *where* contracts run, not *what
they do*:

- **SKY coin** — issues as a CW20 on Neutron, or via Token Factory.
  Same supply (1B), same vesting schedule, same burn mechanics.
- **`x/agentic` semantics** → `contracts/agent-registry/` —
  RegisterAgent, CreateTask, SubmitResponse, SettleTask, SubmitFraudProof.
- **`x/agenticdex`** → `contracts/agentic-dex/` (v0.5) — but realistically
  we instead **partner with Astroport** (the existing CosmWasm DEX on
  Neutron) rather than rebuild. Day-one liquidity.
- **`x/agenticperps`** → `contracts/agentic-perps/` (v1) — or partner with
  Levana (perps on Sei / Osmosis) the same way.
- **`x/agenticrouter`** → unnecessary. Skip Protocol already covers
  cross-chain routing for Neutron.
- **All docs** — architecture, tokenomics, growth, wallet, exchange —
  still apply with minor terminology updates (s/chain/protocol/ in some
  places).
- **Wallet, DEX frontend, landing page** — unchanged. They never cared
  whether the backend was a sovereign L1 or a contract suite.

## What we drop

- `genesis/chain/` — the entire Go module tree. Kept in git history for
  reference; new code goes in `genesis/contracts/`.
- Validator recruitment, genesis-key ceremony, free-tier validator
  quartet from `docs/03-devops.md` — Neutron does this for us.
- The on-chain CEX-launch gate from `docs/07-cex-gate-rationale.md`
  needs to be re-encoded as a contract-side check instead of an L1-level
  rule. Same threshold, same rationale.

## The new sequencing

| Phase | Deliverable | ETA |
|---|---|---|
| 0 — now | This doc + first contract (`agent-registry`) compiling | session +0 |
| 1 — week +2 | `agent-registry` deployed to Neutron testnet | 2 weeks |
| 2 — week +6 | `agentic-dex` contract OR Astroport integration spec | 6 weeks |
| 3 — week +10 | First SKY airdrop snapshot + claim contract | 10 weeks |
| 4 — week +14 | Mainnet deployment to Neutron — **live product** | 14 weeks |

That's 3.5 months to mainnet. Sovereign L1 from the prior path was
12+ months. The trade is honest.

## The migration story (year +1 if traction warrants)

If usage grows past Neutron's natural per-chain ceiling (~$500M TVL on a
single subprotocol), we run a sovereign-L1 migration event. This is the
exact path Berachain and Celestia followed and it became a marketing
moment in both cases:

1. Snapshot SKYMETRIC contract state on Neutron.
2. Launch sovereign chain seeded with the snapshot.
3. Open a one-way migration bridge (Neutron → Skymetric L1).
4. Sunset the Neutron contracts.

This is a year-out problem. We do not over-commit to it now. **The right
default assumption is that we run on Neutron permanently** and treat
the sovereign-L1 narrative as optional.

## Open questions deliberately left for later

- **Astroport partnership vs own DEX contract?** Probably partner first,
  build later. The DEX is not the SKYMETRIC moat — the agent registry +
  reputation layer is.
- **Token Factory vs CW20?** Token Factory is cleaner (native denom
  semantics) but newer. CW20 is battle-tested. Pick at SKY-issuance
  time, not now.
- **Multi-chain expansion?** Neutron has full IBC; SKYMETRIC contracts on
  Neutron are reachable from every Cosmos chain. We don't need to
  re-deploy on Osmosis / Stargaze / Sei. If product-market fit demands
  it, do it then.

## What this doc replaces

The "sovereign L1 vs CosmWasm" question raised in
`09-honest-launch-timeline.md` §"The single decision that defines the
timeline." That question is now answered. The remaining ambiguity in
`09` (timelines for the sovereign-L1 path) is preserved as historical
context but does not bind future work.
