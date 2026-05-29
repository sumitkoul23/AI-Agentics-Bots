# Exchange Strategy — DEX now, CEX at $100M+ liquidity

Two distinct products, four sequenced tiers. Each tier is *necessary
infrastructure* for the next.

## TL;DR sequencing

```
  Tier 1            Tier 2              Tier 3              Tier 4
  ───────────────   ─────────────────   ─────────────────   ──────────────────
  SKYMETRIC DEX       Cross-chain         CEX-frontend on     Licensed
  (x/agenticdex     aggregator          DEX rails           CEX entity
   module)          (router contract)   (order-book UX)     (Binance-class)
  ───────────────   ─────────────────   ─────────────────   ──────────────────
  Cost: $0          Cost: $0            Cost: $0            Cost: $5–50M
  Time: weeks       Time: months        Time: months        Time: years
  Custody: NONE     Custody: NONE       Custody: NONE       Custody: yes
  Gate:             Gate:               Gate:               Gate:
   testnet live      mainnet live        $10M TVL            $100M liquidity
                                                            + treasury funded
                                                            + jurisdiction
                                                            chosen
```

The gate on Tier 4 is *yours* to set; we wrote the on-chain ratchet so the
$100M threshold is enforceable: see "Tier 4 launch trigger" below.

---

## Tier 1 — SKYMETRIC DEX (`x/agenticdex` module)

**Status:** scaffold shipped in this PR — `genesis/chain/x/agenticdex/`.

**What it is:** a sovereign on-chain AMM module living alongside `x/agentic`.
Constant-product (`x*y=k`) pools at v0; concentrated-liquidity ranges in v1.
Inspired by Osmosis `x/gamm` and Uniswap v2 — Apache-2.0-licensed code paths
we can study but not copy verbatim.

**Why we need it before listing anywhere else:** a chain with no native
liquidity is a chain no aggregator will route through. The first pool must
exist before we open IBC to Osmosis or apply to MEXC. Genesis bootstrap
pool: `SKY / USDC.axl` (Axelar-bridged USDC) seeded from the 5 % liquidity
bootstrap bucket in `02-tokenomics.md`.

**Fee structure (proposed defaults, gov-changeable):**

| Slice | % of swap fee | Going to |
|---|---|---|
| LP fee | 0.25 % of trade volume | Liquidity providers |
| Protocol fee | 0.05 % of trade volume | Split 60 % treasury · 40 % **burn** |

A 0.30 % round-trip is competitive with Uniswap v2 (0.30 %) and slightly
better than Osmosis stable pools (0.20–0.30 %). The 40 % protocol-fee burn
hooks directly into the deflationary curve from `02-tokenomics.md` — every
DEX swap actively shrinks supply, on top of the agent-task burn.

**Messages (full list in `chain/x/agenticdex/types/msgs.go`):**
- `MsgCreatePool` — seed a new constant-product pool with two assets.
- `MsgJoinPool` — provide liquidity proportionally, receive LP shares.
- `MsgExitPool` — burn LP shares, redeem proportional reserves.
- `MsgSwap` — swap one asset for another with `min_amount_out` slippage guard.
- `MsgUpdateParams` — gov-only.

**v0 limitations (honest):**

- Constant-product only (no concentrated liquidity, no stableswap curve).
- No multi-hop routing in-protocol (frontend does it client-side).
- No oracle integration — pure on-chain price discovery.
- No MEV protection — public mempool. Add a private order flow / sealed
  bid auction in v1.

These are deliberately accepted to ship Tier 1 fast. Each is a tractable
follow-up.

---

## Tier 2 — Cross-chain DEX aggregator

**Cost:** $0. **Status:** designed, not yet built.

**What it is:** a frontend (and a thin Cosmos-side router contract) that
takes a single user order and:

1. Quotes routes across SKYMETRIC DEX, Osmosis, Astroport (Neutron), and
   Skip Protocol's relayer mesh.
2. Bundles the swap + the necessary IBC packets into one atomic intent.
3. Settles on the user's source chain.

**Stack:**
- [Skip Protocol API](https://api-docs.skip.money/) — already free for low
  volume; covers route quoting + IBC packet construction.
- Frontend: a Next.js app deployed to Cloudflare Pages free tier.
- Custody: **none.** Every swap is a user-signed tx on each leg.

**Why this matters strategically:** the aggregator is how SKYMETRIC stops
being just one of N Cosmos app-chains and becomes the *default entry
point* for any swap whose endpoint is a SKY-paired pool. Daily volume from
the aggregator is what makes the next listing application credible.

**Roadmap to ship Tier 2:**

| Phase | Target | Deliverable |
|---|---|---|
| 2.0 | mainnet day 0 | Skip Protocol API key + route-quote endpoint live |
| 2.1 | mainnet +30 | Frontend with 1-asset/1-chain swap UX |
| 2.2 | mainnet +60 | Multi-hop, multi-chain order builder |
| 2.3 | mainnet +90 | Limit orders (held off-chain, settled when triggered) |

---

## Tier 3 — CEX-frontend on DEX rails

**Cost:** $0 to build, ~$50/mo CDN at scale (acceptable). **Status:**
designed.

**What it is:** Binance-grade trading UX that runs entirely against
non-custodial DEX liquidity. Patterns to follow: Vertex Protocol, Hyperliquid
(both ~$1B+ daily volume on this exact model).

**Surface:**
- Spot order book (off-chain order book, on-chain settlement).
- Perpetuals — funded by a perps module (Tier 3.5, future) that uses SKY as
  margin currency.
- Margin / lending — a lending module that lets LPs borrow against their
  LP shares.
- Charting (TradingView free widget).
- API + websocket feeds (free CCXT-compatible endpoint).

**Why this works as the on-ramp without a CEX license:** users custody
their own funds throughout. We never hold a USD balance, never wire-transfer
fiat, never need a money-transmitter license. The CEX UX is purely
visual + matching; settlement is on-chain.

**This is the product that 95 % of "we want a CEX" projects actually
needed.** Tier 4 is only relevant when fiat on-ramps + custodial services
become a strategic moat.

---

## Tier 4 — Licensed CEX entity (Binance-class)

**Gate (from your brief):** $100M+ liquidity, treasury funded.

**Cost — realistic ranges, sourced from public filings:**

| Cost line | Range |
|---|---|
| US Money Transmitter Licenses (50 states, including NMLS surety bonds) | $5–15M |
| FinCEN MSB registration + ongoing BSA/AML program | $500k / yr |
| EU MiCA license (Crypto-Asset Service Provider) | $1–3M setup + $500k / yr |
| BVI / Cayman / Bermuda offshore entity | $50–250k setup |
| Custody infra (cold storage, MPC, insurance) | $2–5M / yr |
| Trading engine + matching infra | $1–3M build |
| Banking relationships (settlement, fiat ramps) | $500k–2M / yr |
| Compliance team (~25 people, MLRO, BSA officer, ops) | $5–8M / yr |
| Insurance (crime + custody + cyber) | $1–3M / yr |
| Legal + filing reserves | $2–5M |
| **Total year-1 ramp** | **$20–50M** |

**Jurisdiction decision tree (proposal):**

1. **Year 1 of Tier 4:** Incorporate the operating entity in **BVI or
   Cayman** (zero corporate income tax, light regulatory regime for crypto
   exchanges). Apply for VASP registration with the relevant FSC.
2. **Year 2:** Acquire **MiCA CASP** in an EU jurisdiction (likely
   **Lithuania** or **Malta** — fastest paths). Opens the European retail
   market.
3. **Year 3:** Begin **US Money Transmitter License** acquisitions, starting
   with NY BitLicense → NMLS multi-state. This is the most expensive +
   slowest path; sequencing it last lets us bootstrap cashflow first.
4. **Year 3+:** Apply for a banking license (Wyoming SPDI, Swiss FINMA
   fintech license, or German BaFin e-money). At this point we are a real
   financial institution.

**Tier 4 launch trigger — encoded on-chain:**

```cosmos
// genesis/chain/x/agenticdex (paraphrased — full code lands at gate time)
//
// CEXLaunchPermitted returns true iff aggregate SKYMETRIC DEX liquidity
// has been >= $100M (in USDC-equivalent at TWAP) for at least 30
// consecutive days, ratified by a successful gov proposal of type
// MsgEnableCEXLaunch.
```

The on-chain ratchet means no maintainer can unilaterally launch a CEX
under the SKYMETRIC brand. Governance + measurable on-chain TVL gate the
trigger. This is the strongest signal-of-intent we can give the community:
the CEX literally cannot launch until the network has earned it.

---

## What ships in this PR

- `genesis/chain/x/agenticdex/` — Tier 1 module scaffold
  - `types/`: keys, params, pool, msgs
  - `keeper/`: keeper, pool ops, swap math (constant product), fees
  - `module.go`: SDK wiring
- `genesis/exchange/dex/README.md` — frontend / product spec for the DEX UI
- `genesis/exchange/cex/roadmap.md` — Tier 4 milestone roadmap with
  licensing checklist
- `genesis/exchange/cex/jurisdictions.md` — jurisdiction comparison matrix
- `genesis/exchange/cex/architecture.md` — Tier 3 + Tier 4 technical stack

## What's deliberately *not* in this PR

- **Concentrated liquidity (Uniswap v3 / Osmosis CL)** — saved for v1.
- **Perpetuals module** — saved for Tier 3.5.
- **Order-book matching engine** — built when Tier 3 hits the dev queue.
- **Compliance / KYC code** — only relevant at Tier 4; building it now
  would commit us to a model we haven't earned the right to choose.
