# Agent 3 — Tokenomics Engineer: SKY Token Economy

## 1. Supply

| Metric | Value |
|---|---|
| Total supply at genesis | **1,000,000,000 SKY** |
| Base denom | `usky` (1 SKY = 10⁶ usky) |
| Supply cap | None at protocol level; effective cap emerges from the burn ↔ mint equilibrium below |
| Decimals on listings | 6 |

## 2. Genesis allocation

```
┌───────────────────────────────────────────────────────────────────┐
│ Community airdrop                                            40 % │██████████████████████████
│ Validators / staking rewards pool (streamed by x/mint)       25 % │████████████████
│ Treasury (gov-controlled)                                    20 % │█████████████
│ Genesis contributors (this branch's authors + early devs)    10 % │██████
│ Liquidity bootstrap                                           5 % │███
└───────────────────────────────────────────────────────────────────┘
                                                  Total = 100 %
```

### Vesting rules

| Bucket | Cliff | Vesting | Mechanism |
|---|---|---|---|
| Airdrop | None | Linear over 6 months **from claim time** | `x/auth` periodic-vesting accounts created at claim |
| Validators pool | None | Streamed block-by-block by `x/mint` | n/a — emitted, not pre-funded |
| Treasury | 6 mo | Linear over 36 mo | Continuous vesting account owned by `x/gov` module address |
| Contributors | 12 mo | Linear over 24 mo | Continuous vesting accounts per address |
| Liquidity bootstrap | None | Unlocked at mainnet genesis | Multisig of 3 maintainers; spent only on DEX seeding |

## 3. Inflation curve

Standard Cosmos `x/mint` parameters (already encoded in
`chain/config/genesis-overrides.json`):

| Param | Value | Rationale |
|---|---|---|
| `inflation_min` | 1 % | Floor — security budget once chain is mature |
| `inflation_max` | 7 % | Ceiling — match early Cosmos Hub for staker appeal |
| `inflation_rate_change` | 1 % / yr | Slow glide toward equilibrium |
| `goal_bonded` | 67 % | Push toward Cosmos-typical bonded ratio |
| `blocks_per_year` | 10,512,000 | ≈ 3-second blocks |

Year-1 emission, at goal-bonded 67 %, is ≈ 70 M SKY. Real-yield staker APR is
inflation / bonded ratio ≈ **10.4 %** at genesis, sliding to ≈ **1.5 %** by
year 5 as inflation tapers and bonded ratio normalises.

## 4. Burn curve (the deflationary tail)

Two distinct burn streams cancel emissions over time:

1. **Tx-fee burn** — `BurnFromEscrow(20 % of every settled agent task)`.
   Implemented in `chain/x/agentic/keeper/burn.go`.
2. **Slashing burn** — full agent-stake burn on `MsgSubmitFraudProof` quorum.
   Asymmetric: rarely triggered, but each event is large and very public.

The chain flips **net-deflationary** when:

```
annual_tx_burn  +  annual_slashing_burn   >   annual_emission
```

Modelled break-even (see `growth/economics-model.md` once the growth agent
seeds it) lands at ~ 800k settled tasks / year at an average bounty of 50 SKY
— well within the addressable market of the seven specialists already in this
repo plus a public testnet flywheel.

## 5. Zero-cost distribution strategy

### a. Cosmos + AI-builder airdrop (40 % bucket)

Eligibility checked against publicly-queryable on-chain state — **no KYC, no
form, no Discord** — so distribution is a pure SQL job over open data:

| Cohort | Source | Cap per address |
|---|---|---|
| Cosmos stakers (any Cosmos-SDK chain, ≥ 30 days bonded at snapshot) | Numia / public LCD endpoints | 250 SKY |
| Ethereum AI-protocol users (Bittensor, Akash, Ritual, Hyperbolic) | Public Ethereum logs | 250 SKY |
| Active GitHub contributors to OSS AI repos (≥ 5 merged PRs in 2025) | GitHub API | 500 SKY |
| Holders of any Cosmos NFT collection ≥ 6 months | NFT registry snapshot | 100 SKY |

Snapshot is a Cloudflare Worker (free tier) that publishes a Merkle root.
Claims happen on-chain via a `MsgClaimAirdrop` against the published root —
**zero off-chain infra cost**.

### b. Testnet faucet → mainnet warp (5 % bucket)

Every testnet `MsgCreateTask` settlement above a reputation threshold mints a
soul-bound "GenesisProof" NFT on mainnet for the agent operator. Holders
receive a fixed pro-rata mainnet SKY drop. Mechanically encourages real usage
during testnet rather than empty wallet-spinning.

### c. Public goods quadratic match

5 % of treasury annually allocated via a quarterly
[Gitcoin-style quadratic-funding round](https://wtfisqf.com/) — runs entirely
on chain via `x/gov` + a thin off-chain pool ([Allo Protocol](https://allo.gitcoin.co/)
self-hosted on free tier). Bootstraps developer flywheel without paying any
single grantee from treasury directly.

## 6. Listing path (still $0)

- **Day 0:** Open IBC channels to Osmosis + Neutron via standard gov proposal
  (free). Bootstrap a SKY/OSMO 50/50 LP from the 5 % liquidity bucket.
- **Day 30:** Apply to MEXC / Gate.io free listing programs (they comp the
  fee for chains with > 10k active addresses, which the airdrop guarantees).
- **Day 90:** CoinGecko + CoinMarketCap listings — both have free self-serve
  paths once a DEX pool exists with > $25k 24h volume.
- **Year 1+:** Binance Innovation Zone — once governance is genuinely
  decentralised and TVL > $5M, both prerequisites achievable with no further
  capital expenditure.

## 7. Anti-Sybil notes

- Airdrop cohorts are intersections of **costly-to-fake** on-chain histories
  (≥ 30 days bonded, ≥ 5 merged PRs).
- Per-address cap < $20 worth at any plausible FDV → not worth the Sybil
  overhead at zero-cost scale.
- Reputation NFTs are non-transferable (enforced in `x/agentic`), so resale
  markets cannot rent identity.
