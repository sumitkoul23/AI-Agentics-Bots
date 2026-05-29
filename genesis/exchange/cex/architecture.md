# Tier 4 CEX — technical architecture

This is the stack we build *when* the on-chain gate opens. Designed to be
operationally indistinguishable from a top-10 venue, while custody and
settlement stay rooted in SKYMETRIC.

## High-level topology

```
            ┌──────────────────────────────────────────────────────────┐
            │              cex.skymetric.dev (Next.js, edge)              │
            └─────┬────────────────────────┬──────────────────┬────────┘
                  │ REST                   │ WS               │ FIX
                  ▼                        ▼                  ▼
   ┌──────────────────┐    ┌────────────────────┐   ┌─────────────────┐
   │   API Gateway    │    │ Market-Data Gateway │   │  FIX 4.4 Adapter│
   │  (Kong / Envoy)  │    │   (zero-mq fanout)  │   │  (institutional)│
   └────────┬─────────┘    └─────────┬──────────┘   └─────────┬───────┘
            │                        │                         │
            ▼                        ▼                         ▼
   ┌──────────────────┐    ┌────────────────────┐   ┌─────────────────┐
   │   Order Service  │◄──►│  Match Engine (Rust)│  │ Risk + Margin   │
   │  (Postgres-backed)│   │   tonic gRPC bus    │◄─┤ Engine          │
   └────────┬─────────┘    └─────────┬──────────┘   └─────────┬───────┘
            │                        │                         │
            └────────────────┬───────┴─────────────────────────┘
                             ▼
              ┌─────────────────────────────────────┐
              │     Settlement Service              │
              │  (signs txs to x/agenticdex pools)  │
              └──────────────────┬──────────────────┘
                                 ▼
                 ┌──────────────────────────────────┐
                 │     Custody (MPC, Fireblocks)    │
                 │  Hot 5 %  /  Warm 15 %  /  Cold 80 %│
                 └──────────────────────────────────┘
```

Key invariant: every CEX trade settles on-chain within N seconds (we
target ≤ 10 s p99). Off-chain matching is purely a UX optimisation; nothing
of value lives only off-chain.

## Why this won't collapse like FTX did

The single architectural decision that prevents FTX-class catastrophe:

> **The CEX operating entity has no direct access to the cold-storage
> keys.** Cold storage is governed by a separate multi-jurisdictional
> trustee with quorum signing tied to on-chain attestations.

This is the same pattern Kraken uses (cold storage with Bitgo as
quorum-co-signer in a separate jurisdiction). It rules out the entire
class of "borrowed from custody to plug a sister-trading-firm hole"
failures.

## Components

### Frontend — `cex.skymetric.dev`

- Next.js 14 with App Router + RSC
- TradingView Advanced Charts (free for non-broker use)
- @cosmos-kit/* for in-app wallet handoff (users can withdraw to their
  own wallet without leaving the site)
- React-Window virtualised orderbook (depth-of-market for 1000+ levels)
- Service worker for offline tolerance

### API Gateway

- Kong Open Source (Apache 2.0) — handles routing, auth, rate-limit
- Envoy as the data-plane proxy
- mTLS between every internal service

### Match Engine

- Rust, single-threaded per-symbol with a sharded supervisor for parallel
  symbols. Latency target ≤ 30 µs intra-symbol match.
- Reference: open-source [LMAX-style disruptor pattern](https://lmax-exchange.github.io/disruptor/)
- Persists every accepted order to a write-ahead log on NVMe before ACK
- Replay-from-log boot in < 60 s

### Risk + Margin Engine

- Real-time risk recompute every 100 ms
- Tiered margin: 1×, 3×, 5× spot; 10× perps initially
- Auto-deleveraging at the contract level (no socialised losses across
  unrelated users — every position is collateral-explicit)

### Settlement Service

- Watches the match-engine event bus; signs an SKYMETRIC tx to
  `x/agenticdex` for each fill batch (every 100 ms)
- Treats on-chain settlement as the source of truth — if the match
  engine claims a fill that doesn't settle, balances roll back

### Custody (the part everyone screws up)

| Tier | Allocation | Tech | Access |
|---|---|---|---|
| Hot | ≤ 5 % | Lit Protocol / Fireblocks MPC | Settlement service only |
| Warm | ~15 % | Multi-sig HSM (3-of-5, hot keys) | Treasury ops, 24h timelock |
| Cold | ≥ 80 % | Trustee-held shamir shards across 3+ jurisdictions | 72h timelock + on-chain attestation |

### Proof-of-Reserves

- On-chain Merkle attestation every block
- Quarterly Big-Four audit attesting the public Merkle root matches
  custody-side balances
- Users can verify their balance is in the audited root via a Merkle
  inclusion proof rendered in their account page (Kraken / Bitget /
  Bybit all do this now; it's table stakes)

## Why we still need this if we have the DEX

Honest answer: for casual users, **we mostly don't.** Tier 3 (CEX-frontend
on DEX rails) covers ≥ 95 % of the population.

The 5 % we build Tier 4 for:
- **Fiat on-ramp users.** A DEX cannot accept USD wires legally without
  becoming a money transmitter — so we'd be doing this anyway.
- **Institutional desks.** Need FIX, prime brokerage, swap-line credit,
  insurance-backed custody. Real institutions categorically will not
  trade on a non-custodial DEX.
- **Geographies where DEX UX is illegal.** Some jurisdictions
  (notably China + parts of the Middle East) ban DEX participation
  outright but permit licensed local CEXs.

If those three audiences don't matter to us at the gate, we save $30M+
by skipping Tier 4 entirely. That decision should be re-made by
governance once the gate opens, not pre-committed now.
