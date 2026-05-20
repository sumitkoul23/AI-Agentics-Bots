# Jurisdiction comparison matrix

Public information as of 2025–2026. Always re-verify with local counsel
before filing — these regimes change frequently.

| Jurisdiction | Licence | Setup time | Setup cost | Annual cost | US users | EU users | Public reputation |
|---|---|---|---|---|---|---|---|
| **BVI** | FSC Approved Crypto Service | 4–6 mo | $50–150k | $80–150k/yr | ❌ | ❌ (need MiCA) | Neutral |
| **Cayman Islands** | CIMA VASP | 6–9 mo | $100–250k | $120–200k/yr | ❌ | ❌ | Neutral / mature |
| **Switzerland** | FINMA VASP + AMLA | 9–12 mo | $500k–1M | $300–800k/yr | ⚠️ via separate US entity | ✅ | Strong |
| **Liechtenstein** | FMA TVTG | 6–9 mo | $300–500k | $250–500k/yr | ⚠️ | ✅ via EEA passport | Strong |
| **Lithuania** | MiCA CASP | 6–9 mo | $200–400k | $200–400k/yr | ⚠️ | ✅ (entire EEA) | Mid |
| **Malta** | MFSA VFA → MiCA CASP | 9–12 mo | $400–700k | $250–500k/yr | ⚠️ | ✅ | Mid (post-2022 reforms) |
| **Singapore** | MAS PSA DPT Service | 12–24 mo | $300–600k | $400–800k/yr | ⚠️ | ❌ | Very strong |
| **Hong Kong** | SFC VATP | 12–18 mo | $400–700k | $400–700k/yr | ❌ | ❌ | Strong (post-2023) |
| **UAE** | VARA (Dubai) | 9–12 mo | $500–800k | $400–700k/yr | ⚠️ | ⚠️ | Strong |
| **US (state-by-state)** | MTL × 50 + NY BitLicense + FinCEN MSB | 24–36 mo | $5–15M | $5–8M/yr | ✅ | — | Highest barrier |
| **UK** | FCA cryptoasset registration | 12–24 mo | $300–600k | $300–600k/yr | ⚠️ | ⚠️ | Mid (high refusal rate) |

## Decision tree (default)

```
                       Year 1                Year 2              Year 3
  ─────────────────    ─────────────────    ─────────────────    ─────────────────
  Cayman + BVI    →    Lithuania (MiCA) →   Singapore (PSA)  →   US (MTL + NYDFS)
  Speed: fast          Reach: EEA           Reach: APAC          Reach: USD market
  Users: rest-of-      Adds: ~450M          Adds: ~700M          Adds: ~350M
  world ex US/EU       potential users      potential users      potential users
```

## Edge cases

- **If treasury > $50M at Tier 4 gate:** skip Cayman, start directly in
  Switzerland (FINMA). Slow but produces the strongest brand and the
  longest moat against competitors.
- **If treasury < $20M at Tier 4 gate:** Cayman first, defer MiCA and
  US entirely. Service only KYC'd non-US/EU users via VPN-blocked
  geos. Reassess in 18 months.
- **If we're cashflow-positive from DEX fees alone:** consider skipping
  Tier 4 entirely. Cashflow-positive DEXs (Uniswap Labs) have higher
  enterprise value per dollar of revenue than custodial CEXs (Coinbase,
  Kraken) in every public-comparable. The cost of a licence is also the
  cost of accepting a regulatory ceiling.

## What we don't do

- **Forum-shopping the cheapest jurisdiction.** The market reads through
  this in 30 seconds; the brand cost outweighs the licence cost.
- **Operating in stealth from a non-jurisdiction.** Every survivor of the
  2022 collapse had a defensible domicile; every casualty did not.
- **Using a Seychelles / Marshall Islands / Saint Vincent IBC as the
  customer-facing entity.** Acceptable as a parent or special-purpose
  vehicle, never as the user-facing brand.
