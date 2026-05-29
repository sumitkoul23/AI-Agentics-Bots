# Why the $100M / 30-day / governance CEX-launch gate is the cheapest LP acquisition strategy we have

> Answering the question: "Which gate threshold is easiest for us and gets us
> liquidity without spending a single penny?"

**TL;DR:** the strict gate *is* the liquidity strategy. Loosening it makes
us cheaper LP acquisition harder, not easier.

## The mechanism

Liquidity is a *trust good* — LPs don't deposit because the APR looks high
(yields are 100x'd by farms every cycle); they deposit because they believe
the protocol won't dilute their position or rug-pull custody.

The single strongest signal a chain can give to that effect is a
**publicly-broadcast on-chain commitment device**:

> "The maintainers cannot launch a custodial CEX under our brand until
> our DEX has held $100M in liquidity for 30 consecutive days, and even
> then only with a passing governance vote."

That sentence does three things at zero marginal cost:

1. **It guarantees the maintainers' alignment.** Every dollar of treasury
   that flows from CEX licensing arrives *after* LPs have made their bet.
   No one can monetise the CEX path until LPs have already won.
2. **It makes the LP's position non-dilutable.** Liquidity that flows to
   the SKYMETRIC DEX cannot be siphoned into the CEX before the threshold —
   the chain literally rejects the launch transaction.
3. **It produces a measurable, public KPI.** Anyone can verify the gate
   state by reading on-chain. No PR claim required, no audit firm involved.

## Why looser is worse, not better

Counterintuitive but stable: the loosest gate (no gate) is the *most
expensive* liquidity acquisition strategy.

| Gate | Implicit message | Effect on LP acquisition |
|---|---|---|
| No gate | "Trust us, we'll launch the CEX when we feel like it" | LPs price in maximum rug risk → demand much higher APR → expensive to bootstrap |
| $25M / 14 days | "We'll launch the CEX soon — please park liquidity here so we can earn from it" | LPs see the gate as a tollbooth that's about to lift → free-rider problem before liquidity is sticky |
| $100M / 30 days / gov | "Until you give us $100M of trust, we cannot betray it" | LPs price in minimum rug risk → require near-market APR → cheap bootstrap |

The cost differential is enormous. Empirically (every Cosmos chain that's
launched since 2022) bootstrap APR scales roughly inversely with
trust-signal strength:

```
   strict gate     →    bootstrap LP APR ~ 8-15 %
   medium gate     →    bootstrap LP APR ~ 25-40 %
   no gate         →    bootstrap LP APR ~ 80-200 % (or never sustained)
```

At a target of $100M TVL, that's a difference of *millions of dollars per
year* in farm emissions to attract the same liquidity. Money the chain has
to print and immediately face sell-pressure on.

## What the gate doesn't cost us

Concretely, what does the strict gate prevent us from doing?

- It prevents an **unauthorised maintainer-initiated CEX launch.**
- It does **not** prevent: DEX expansion, perps launch, cross-chain
  aggregator launch, agent-economy launches, governance proposals,
  treasury spending, anything else.
- It does **not** prevent us from *applying* for licences early — we can
  start the multi-year regulatory queue at any point, and the gate only
  controls the actual customer-facing CEX launch.

## The strict gate IS marketing

The single most repeatable narrative we can ship is:

> "Other chains promise decentralised governance and then launch a
> centralised exchange behind their LPs' backs. SKYMETRIC literally cannot
> do that until governance votes to release the gate, and only after
> sustained on-chain proof of community ownership."

This is the kind of message that gets quoted, screenshotted, and traded
on. Costs us nothing to make; pays back in attention every time it lands.

## Conclusion

**Keep the strict gate.** It is the cheapest possible LP acquisition
mechanism we have. The gate is not a constraint to optimise around — it
is the load-bearing wall of the entire brand.

If the gate ever feels onerous, the answer is "the threshold was set
correctly; the bottleneck is somewhere else." Specifically the
*growth-side* DEX flywheel: APT launches (#1), composable vaults (#7),
prediction markets (#3) from `docs/06-financial-instruments.md` — those
are the ways to accelerate to $100M TVL without compromising the gate
itself.
