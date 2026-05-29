# SKYMETRIC DEX — product spec

> The user-facing frontend for `x/agenticdex`. Lives at `dex.skymetric.dev`.

## Stack ($0)

- **Framework:** Next.js 14 (App Router) — already free-tier-friendly on
  Cloudflare Pages
- **Wallet:** [Cosmos Kit](https://cosmoskit.com/) — supports Keplr, Leap,
  Cosmostation, Ledger, plus 15+ others out of the box
- **Routing:** in-protocol single-hop at v0; client-side multi-hop pathfinder
  at v1
- **Charts:** TradingView Lightweight Charts (free, MIT)
- **Indexer:** [Subquery](https://subquery.network/) or [Numia](https://numia.xyz/)
  free tiers for historical price + volume

## Views

| Route | Purpose |
|---|---|
| `/swap` | The default landing — single-line swap with auto-routing |
| `/pools` | Pool list with TVL / APR / volume, sortable |
| `/pools/[id]` | Pool detail — add / remove liquidity, your position, price chart |
| `/portfolio` | Your LP positions + agent reward history (cross-references `x/agentic`) |
| `/governance` | Active proposals (DEX param changes route through here) |

## Differentiation vs Osmosis / Astroport

The DEX itself is intentionally undifferentiated — constant-product AMMs are
a commodity, and trying to be "the best AMM" is a losing strategy at $0.

The product wedge is the **integration with `x/agentic`**:

1. **Agent quotes.** A swap UI option to ask any registered agent for a
   pre-trade analysis ("Is this a fair price?"). The agent earns SKY per
   quote, scaled by reputation.
2. **Strategy LP positions.** Agents can manage LP positions on a user's
   behalf — escrowed via `x/agentic` task semantics with slashable bonds.
3. **Reputation-weighted execution.** Aggregator routes prefer pools whose
   creators have high-reputation agent footprint. Surfaces the chain's
   reputation primitive directly inside the trade flow.

None of this is needed to ship v0 — but every one is a follow-up that uses
infrastructure we *already shipped* in this PR.

## Roadmap

| Phase | Target | Deliverable |
|---|---|---|
| 0 — now | this PR | Backend module (`x/agenticdex`) scaffold |
| 1 — testnet +14d | week +6 | `/swap`, `/pools` views; Cosmos Kit wallet flow |
| 2 — mainnet day 0 | month +2 | `/portfolio`, real-time prices, TradingView chart |
| 3 — mainnet +30d | month +3 | Agent-quote integration (uses `x/agentic` MsgCreateTask) |
| 4 — mainnet +60d | month +4 | Multi-hop client routing across pools |
| 5 — mainnet +90d | month +5 | Cross-chain aggregator via Skip Protocol API |
| 6 — TVL $10M | conditional | Concentrated-liquidity pools (v1 module proposal) |
| 7 — TVL $100M | conditional | CEX-frontend (see `../cex/`) goes live |

## What's *not* on this DEX

- **No CLOB / limit orders.** A constant-product AMM only. Limit orders
  arrive at Tier 3 (CEX-frontend), not here.
- **No leverage.** Same — Tier 3.
- **No options / structured products.** Not in scope.
- **No custodial wallet feature.** The DEX never holds user funds outside
  of signed transactions.
