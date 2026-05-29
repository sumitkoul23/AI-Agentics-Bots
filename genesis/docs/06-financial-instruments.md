# Novel financial instruments — the SKYMETRIC product menu

> Every chain ships an AMM. Spot-DEX-and-perps is now table stakes. The
> point of having a *bespoke* agent-economy chain is to mint instruments
> nobody else can — instruments whose underlying is the on-chain
> reputation, behaviour, and revenue of AI agents themselves.

Ten ideas, ranked by tractability × differentiation. Each entry includes
the on-chain primitive needed, the user it serves, and a one-line napkin
on volume.

---

## 1. Agent Performance Tokens (APT)  — *tractable now*

**What:** Tokenised claim on a specific agent's *future revenue*. Each
agent's operator can mint `APT-<agent-id>` shares against a fraction of
their forward task earnings, capped at gov-set max.

**On-chain primitive:** new module `x/aperp` (different from `x/agenticperps`
— this is an **agent-perpetual**, not asset-perpetual). Holders receive
streaming SKY proportional to their share whenever the agent settles a task.

**Useful when:** an agent operator wants to bootstrap working-capital
without selling stake. Buyers want exposure to top-N agents without running
infra.

**Volume hypothesis:** if 5 % of registered agents mint APT and avg float
is $10k, $5M TVL at 1k agents → meaningful niche product.

**Status:** new module, ~3 weeks of work.

---

## 2. Reputation-Collateralised Loans  — *tractable now*

**What:** an agent operator deposits their soul-bound Reputation NFT as
*risk-pricing* collateral (not seizable) and borrows SKY against expected
future earnings. Default = NFT burn + jail flag → no future earnings →
borrower out of business.

**On-chain primitive:** `x/agenticlend` module + a borrow-curve where APR
is a function of (reputation_decile, current_revenue, slash_history).

**Useful when:** agents need SKY to bid on high-stake tasks but don't want
to dilute APT supply.

**Volume hypothesis:** lower per-loan tickets (~$500–5k) but very high
turnover. Standard DeFi-money-market shape.

**Status:** ~4 weeks. Depends on Compound-v2-style rate model + reputation
oracle (we already have the latter in `x/agentic.Agents`).

---

## 3. Prediction markets on agent performance  — *tractable now*

**What:** a `x/agenticmarkets` module that creates binary markets like:
- "Will agent `priya-hub-z7w4v9` be in the top-100 reputation cohort at
  block N?"
- "Will SKY per settled task average > X over the next 7 days?"
- "Will validator V be jailed before block N?"

Liquidity bootstrapped by the same constant-product math as `x/agenticdex`
(YES / NO tokens trade against USDC pools).

**Useful when:** the chain's primitive (reputation) becomes itself the
betting market. Strong reflexive flywheel.

**Volume hypothesis:** Polymarket-style numbers — 100s of markets × thin
margins × high-turnover.

**Status:** ~3 weeks. Cleanest pure on-chain module of the lot.

---

## 4. Streaming payments to agents  — *near-tractable*

**What:** Sablier-style continuous money streams from a user to an agent.
"Pay this agent 0.01 SKY per minute it watches my Twitter mentions." Lets
agents accept long-running retainer work without one-off task creation.

**On-chain primitive:** extend `x/agentic` with a `Stream` type (start /
end / rate / locked balance). Settle on every block tick or on-demand.

**Useful when:** monitoring / surveillance / always-on agent work. Today's
`MsgCreateTask` is a one-shot; streams unlock the recurring-revenue
business model.

**Volume hypothesis:** lower TPS impact (1 tx open + 1 close per stream),
but unlocks 10×+ the total addressable agent-work market.

**Status:** ~2 weeks.

---

## 5. Slashable insurance pools  — *tractable now*

**What:** LPs deposit USDC into a `x/agenticinsurance` pool that backs
agent task outcomes. Premiums flow to LPs per settled task. When a
fraud-proof slash happens but the agent's stake doesn't cover the full
task bounty, the pool absorbs the shortfall.

**On-chain primitive:** insurance keeper holds USDC, gets credited from
`x/agentic.SettleTask` (small slice of the validator-30 % cut), pays out
on `slashAgentAndCloseTask` shortfalls.

**Useful when:** high-value task requesters need a "guarantee" beyond the
agent's stake. Insurance underwriters earn yield from premium flow.

**Volume hypothesis:** scales with task TVL; 0.5 % of settled volume → LP
APR is competitive with Aave at similar size.

**Status:** ~3 weeks. Mirrors Nexus Mutual's model.

---

## 6. Yield-bearing reputation NFTs  — *experimental*

**What:** convert each agent's soul-bound reputation NFT into a fractional
yield-bearing wrapper. Wrap mints share tokens proportional to reputation;
unwrap burns the shares. The wrapper accrues 5 % of the agent's settled
revenue.

**On-chain primitive:** ERC-4626-style vault but for reputation. The
underlying "asset" is an abstract scoring number; the yield is real SKY.

**Useful when:** agent operators want to liquidate part of their reputation
"value" without actually transferring the soul-bound NFT.

**Volume hypothesis:** narrow but high-margin. Probably only the top 10 %
of agents have enough reputation to be wrappable.

**Status:** ~4 weeks. Novel — no real precedent.

---

## 7. Composable agent strategies (vaults)  — *tractable*

**What:** Yearn / Beefy-style vaults that route user capital through
*multiple* agents in sequence. e.g. "deposit USDC, agent-A picks a
trade, agent-B sizes it, agent-C executes via `x/agenticperps`."

**On-chain primitive:** thin vault contract / module (`x/agenticvault`)
that ranks agents by reputation × specialty tag and rotates capital.
Each step is a settled task — vault pays the agents from the deposit.

**Useful when:** the average user can't pick which agent to use. Vaults
become the *user-facing* product, agents become the *backend* labour
market. This is the inversion that wins crypto-AI long-term.

**Volume hypothesis:** the killer app of the chain. If just one vault
goes viral, it carries the chain.

**Status:** ~4 weeks. The most strategic ship.

---

## 8. Agent IPO launchpad  — *near-tractable*

**What:** a new agent registering on-chain can hold a *bootstrap auction*
for the first batch of their APT shares (see #1). Standard Liquidity
Bootstrap Pool (LBP) curve — initial high price decays over 72 hours,
buyers stake SKY to discover fair value.

**On-chain primitive:** a `x/agenticlaunchpad` wrapper around #1's
APT-mint flow. Liquidity automatically seeds the post-IPO APT/USDC pool
on `x/agenticdex`.

**Useful when:** new high-quality agents want capital + audience. Removes
the "cold start" reputation problem.

**Volume hypothesis:** Echo/Hyperliquid-style ticker churn — 50–200 IPOs
per year at meaningful float.

**Status:** ~5 weeks. Depends on #1 shipping first.

---

## 9. Synthetic agent index baskets  — *tractable*

**What:** a single token whose value tracks the basket of the top-N
reputation agents' APTs. Rebalances quarterly. "Buy 1 unit of GENI-25" =
own a slice of the top 25 agents.

**On-chain primitive:** rebalancing vault that holds APT-shares of N
agents. Mint / redeem at NAV. Standard ETF math.

**Useful when:** retail wants exposure to the agent economy without
picking individual agents. The "QQQ for AI agents."

**Volume hypothesis:** if it lists on any external CEX, this is the
single token that does $50M+ daily volume. Indexes are the trade-everyone-
agrees-to-make in every crypto cycle.

**Status:** ~4 weeks. Depends on #1.

---

## 10. Options on SKY  — *non-tractable v0*

**What:** standard European calls/puts on SKY with USDC collateral.
American-style on the longer dated.

**On-chain primitive:** `x/agenticoptions` — Lyra-style AMM, or simpler
Premia-style fixed strike grid. Significant complexity; this is the only
module of the ten that justifies a separate audit.

**Useful when:** validators want to hedge directional risk on their
delegations; large APT-holders want downside protection.

**Volume hypothesis:** the slowest to mature but eventually structural —
options volume = 5-10 % of spot at a mature chain.

**Status:** v2 product. Don't ship until we have $50M+ in mainnet
liquidity and a real audit budget.

---

## Recommended sequencing

```
  Month +2 (mainnet day 0)                Month +6                       Month +12
  ─────────────────────────────           ─────────────────              ───────────────────────
  Ship in v0:                             Ship in v0.5:                  Ship in v1:
    #4  Streaming payments                 #1  APT (agent-perp)            #6  Yield-bearing rep
    #5  Insurance pool                     #3  Prediction markets          #8  Agent IPO launchpad
                                           #2  Reputation loans            #9  Synthetic baskets
                                           #7  Composable vaults

  Year 2+:
    #10 Options
```

The first wave (4 + 5) ships *now* because both are sub-300-LOC additions
to keepers we already have, and both deepen the chain's existing primitives
without requiring brand-new infrastructure.

The composable-vaults wave (#7) is the strategic one — it inverts the
relationship between user and agent, and that inversion is the actual
product moat. Ship it second once the underlying instruments exist.

---

## What we do NOT build

- **Memecoin launchers.** Every chain attracts a "fair launch" memecoin
  launcher. We deliberately don't. Each memecoin launched is one less unit
  of attention on the agent primitives we actually care about.
- **NFT marketplace.** Cosmos NFT marketplaces (Stargaze) already exist;
  building one is rent-seeking on the same audience. We open IBC to
  Stargaze instead.
- **Generic stablecoin issuance.** Stablecoins need a USD peg, which
  needs a real fiat issuer (see `exchange/cex/`). Issuing our own
  uncollateralised stable is the single fastest way to fail.
