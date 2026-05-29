# Tier 4 CEX roadmap — quarter by quarter

Activated only after `CEXLaunchPermitted == true` (on-chain).

## Q1 — Foundation

**Cost:** $500k–1M (50–80 % of cash from treasury, balance from a closed
strategic round with crypto-native funds — *not* generalist VCs).

- [ ] Incorporate operating entity in BVI (or Cayman, see `jurisdictions.md`)
- [ ] Establish parent foundation in Cayman as governance counterparty
- [ ] Engage compliance counsel (Anderson Kill / Reed Smith / Carey Olsen)
- [ ] Apply for VASP registration in chosen jurisdiction
- [ ] Engage Big-Four audit firm (Big Four required for licensing reciprocity)
- [ ] Hire founding compliance team (5 FTEs minimum: BSA officer,
      MLRO, sanctions lead, ops, paralegal)
- [ ] Sign banking MOU with at least 2 of: Bank Frick (LI), Sygnum (CH),
      DBS (SG), Standard Chartered (HK)

## Q2 — Build

**Cost:** $2–3M cumulative.

- [ ] Trading engine v1 — fork or license:
      Match Engine ([Cosmos-X](https://github.com/cosmosforce/cosmosx))
      / [DexFinex Pro](https://github.com/Bitfinexcom/grenache) for the
      websocket gateway / [zero-mq](https://zeromq.org/) for the order bus
- [ ] Custody — MPC architecture via Fireblocks ($1k+/month at our scale)
      or self-host [Lit Protocol](https://litprotocol.com/) (free OSS)
- [ ] KYC / AML stack — see `compliance-stack.md` vendor matrix
- [ ] Customer-support tooling (Zendesk + a 24/7 outsourced first-tier
      team in PH / KE — $80k/mo at expected ticket volume)
- [ ] Insurance — Marsh / Aon broker engagement for crime + custody + cyber
- [ ] Production infra — AWS/GCP enterprise contracts (this is the
      moment we accept paid infra; $5–20k/mo for HA in two regions)

## Q3 — Soft launch

**Cost:** $5–8M cumulative.

- [ ] Soft-launch in BVI / Cayman with invite-only access (1k users)
- [ ] Apply for **MiCA CASP** in Lithuania (fastest EU path) or Malta
- [ ] Apply for **MAS PSA** in Singapore (12–18 month timeline)
- [ ] Apply for **FinCEN MSB** registration in the US
- [ ] Begin **NY BitLicense** application (most expensive single
      jurisdiction; budget $1–2M legal alone)
- [ ] Public bug bounty program (Immunefi — fees apply at payout time only)

## Q4 — Public launch

**Cost:** $10–20M cumulative.

- [ ] Public launch in EU (post-MiCA approval)
- [ ] Liquidity provisioning programs targeting market-makers (Wintermute,
      GSR, Amber, B2C2)
- [ ] First fiat on-ramp (EUR via SEPA through banking partner)
- [ ] Native SKY/USDC orderbook with > $5M daily volume target
- [ ] Listing program for vetted Cosmos-ecosystem tokens
- [ ] Insurance fund seeded with 1 % of treasury (returned via protocol-fee
      diversion until fund reaches 5 % of platform AUM)

## Year 2

- [ ] US Money Transmitter Licenses (start with the easier states:
      Wyoming → Washington → New Hampshire → Texas → Florida)
- [ ] NY BitLicense final approval
- [ ] UK FCA registration as cryptoasset business
- [ ] Hong Kong VATP licence
- [ ] UAE VARA licence
- [ ] Derivatives infrastructure (Tier 3.5 perps module ported on-chain
      first, then mirrored on Tier 4)

## Year 3+

- [ ] **Banking license** — choose one:
      Wyoming SPDI (cheapest path to US banking)
      / Swiss FINMA fintech license
      / German BaFin e-money institution
      / Hong Kong virtual bank
- [ ] Cross-listing reciprocity with major DEXs (Uniswap, Curve, Aerodrome)
- [ ] Institutional custody offering (split out as separate entity if
      the regulatory split is meaningful)

## Hard rules we follow throughout

1. **Never deviate from KYC.** No "tier 1 unverified" sneaks for early
   adopters. Once you set a precedent, regulators audit it forever.
2. **No US-resident users until US licensing is complete.** Geo-block at
   the L4 firewall, not just the frontend.
3. **Custody is non-negotiable.** Cold storage > 95 % at all times.
   On-chain proof-of-reserves quarterly, audited.
4. **No proprietary trading against users.** Operating entity does not
   take market positions on the venue, ever. This single rule precludes
   95 % of the catastrophes of the last cycle.
5. **No exchange token besides $SKY.** And $SKY's economics are entirely
   on-chain, governed by the chain — the CEX cannot mint or favour it.
