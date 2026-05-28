# SKYMETRIC CEX — Tier 4 roadmap

> **Status: NOT BUILDING YET.** This folder is the plan we follow *if and
> when* DEX liquidity crosses $100M and the treasury can fund the regulatory
> ramp. Until then, treat this directory as architecture, not implementation.

## Why the gate is non-negotiable

Launching a custodial exchange without compliance is the fastest way to:
1. Get the project's domain seized.
2. Get the maintainers personally criminally exposed (BSA Section 1960 in
   the US carries a 5-year sentence per violation).
3. Lose every user's funds to the inevitable enforcement freeze.

Every "$0 CEX" you can think of in the last cycle is either:
- A DEX-frontend pretending to be a CEX (no custody — fine, but that's
  Tier 3, see `../dex/`)
- A scam that closed within 18 months
- Operating illegally and burning fuse-runway

We will not be those projects.

## Documents in this folder

| File | Purpose |
|---|---|
| [`roadmap.md`](roadmap.md) | Quarter-by-quarter execution plan once the gate opens |
| [`jurisdictions.md`](jurisdictions.md) | Jurisdiction comparison matrix (BVI / Cayman / EU / US / SG / HK / UAE) |
| [`architecture.md`](architecture.md) | Technical stack for a Tier 4 build |
| [`compliance-stack.md`](compliance-stack.md) | KYC / AML / sanctions / transaction monitoring vendor analysis |

## What we DO build now (in this PR)

The on-chain trigger that gates Tier 4. From
`docs/05-exchange-strategy.md`:

```
CEXLaunchPermitted returns true iff aggregate SKYMETRIC DEX liquidity
has been >= $100M (in USDC-equivalent at TWAP) for at least 30
consecutive days, ratified by a successful gov proposal of type
MsgEnableCEXLaunch.
```

This is the strongest commitment device available — the CEX cannot launch
under the SKYMETRIC name until the chain itself has earned it on-chain.
