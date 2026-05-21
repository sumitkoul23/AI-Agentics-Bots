# Genesis System — build status

> Single-page truth on what's done, what's in progress, and what's next.
> Updated as each batch ships.

## Batches shipped (chronological)

| Batch | Commit | What |
|---|---|---|
| 1 | `7580627` | Initial scaffold: 5 agents' first deliverables (architecture, x/agentic, tokenomics, devops, growth) |
| 2 | `53b1c3f` | Social media launch kit: handles, bios, brand, logo, signup runbook |
| 3 | `16d1de3` | Creatives + video scripts + landing page + x/agentic Msg handlers |
| 4 | `b04c3ac` | x/agenticdex (constant-product AMM) + 4-tier exchange strategy |
| 5 | `c0bee85` | x/agenticperps + x/agenticrouter + DEX frontend + CI/CD + financial instruments doc |
| 6 | `45d4d81` | Agentic Wallet — strategy + MV3 extension scaffold + agent-economy views + Keplr fork plan |
| 7 | *this batch* | Integration: app.go wires all 4 modules; refreshed README + STATUS |

## Component status matrix

| Component | Status | Compile? | Tests? | Audit? | Ships at |
|---|---|---|---|---|---|
| `x/agentic` | scaffold + handlers | needs proto-gen | ✅ split-math + params | — | mainnet |
| `x/agenticdex` | scaffold + handlers | needs proto-gen | ✅ swap-math | — | mainnet |
| `x/agenticperps` | scaffold + handlers | needs proto-gen | ⚠️ none yet | — | mainnet +90d |
| `x/agenticrouter` | scaffold (native-AMM only) | needs proto-gen | ⚠️ none yet | — | mainnet +60d |
| `app/app.go` wiring | ✅ all 4 modules registered | needs proto-gen + module instances | — | — | — |
| `cmd/agenticd` | ✅ root cmd + entry point | needs proto-gen | — | — | — |
| Landing page (`site/`) | ✅ ready | ✅ static HTML | — | — | now |
| DEX frontend (`frontend/dex/`) | ✅ UI scaffold | ✅ `next build` | — | — | mainnet |
| Wallet extension scaffold | ✅ MV3 loads + popup renders | ✅ | — | — | v0 demo |
| Agent-economy views | ✅ all 4 components | ✅ TSX | — | — | v1 (Keplr fork) |
| Keplr fork (v1 wallet) | 📋 plan only | — | — | **REQUIRED** | mainnet +90d |
| CI: `genesis-chain.yml` | ✅ advisory | — | — | — | now |
| CI: `genesis-release.yml` | ✅ multi-arch release on tag | — | — | — | now |
| CI: `genesis-site.yml` | ✅ static-site validation | — | — | — | now |
| Proto pipeline (`buf`) | ⚠️ not wired | — | — | — | next batch |

Legend: ✅ done · 📋 planned · ⚠️ not yet · — N/A

## Blockers before testnet launch

In order, smallest to largest:

1. **Proto generation.** `buf generate` against `chain/proto/` produces the
   typed Go message structs. Today's hand-rolled types in
   `types/msgs.go` are placeholders meant to compile alone, but the
   keepers reference proto-generated types that don't exist yet.
2. **`go mod tidy`.** Once proto-gen produces real packages, the dependency
   graph closes.
3. **Module instance wiring.** `app/app.go::ModuleBasics` is set, but the
   actual keeper construction + module instance registration with the
   `runtime.AppBuilder` is still to be written. ~200 LOC.
4. **Genesis bootstrap.** The seed-node ID, the validator key ceremony,
   and the live `genesis.json` need to be produced and published before
   anyone outside the maintainers can join.
5. **At least one external review.** Even before a paid audit, a single
   second pair of eyes on the slashing path catches > 80 % of common bugs.

## What we are deliberately NOT building yet

| Item | Why deferred |
|---|---|
| Concentrated-liquidity DEX (v3 / CL) | Adds 3000+ LOC; constant-product is fine at v0 TVL |
| Perps order-book (CLOB) | Needs off-chain matching infra → Tier 3 product |
| Custodial CEX | Gated at $100M TVL + gov vote (see docs/07) |
| Native KYC / AML | Tier 4 only — premature for an L1 |
| Memecoin launcher | Attention drain on real product surface |
| Generic stablecoin | Needs real fiat issuer (Tier 4) |

## Open questions for the human

None blocking — every "what should we do" answer is encoded in a doc.
The maintainer's queue:

1. Decide whether to register `agentic.dev` (the one paid step we
   accepted in the playbook).
2. Decide which social handles to claim from
   `growth/social/signup-checklist.md`.
3. Decide whether to engage with an external auditor now (cheap diff
   audit of the Keplr fork) or after the proto pipeline lands.

Otherwise the Genesis System is in steady-state: ship a batch, surface
the status here, repeat.
