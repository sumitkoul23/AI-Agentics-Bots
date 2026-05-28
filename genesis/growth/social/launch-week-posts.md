# Launch week — 7 days of posts, ready to paste

> Replace `@skymetric` with the actual handle you secured. Each row is
> one post; the same beat is delivered in platform-native form. Schedule
> with [Buffer free tier](https://buffer.com/) (3 channels, 10 scheduled
> posts — fits a launch week).

---

## Day 0 · Monday — "We exist"

### X (280 ch)
```
introducing SKYMETRIC — a sovereign L1 for AI-agent work.

agents stake $SKY to bid on tasks.
users pay $SKY, get verifiable outputs.
proven fraud burns the agent's stake.

cosmos sdk · proof of stake · 100% open · built with $0.

code: github.com/agentic-chain
```

### Telegram channel
```
Hi everyone — welcome to SKYMETRIC.

We're building a sovereign Cosmos SDK Layer 1 where AI agents stake on the quality of their own work. The native coin is $SKY.

Why a new chain instead of a smart contract? Because the primitive — slashable agent stake — only makes sense as a first-class on-chain object, not a contract on a chain that doesn't know what an agent is.

This channel is announcements-only. For chat, join @skymetricchat.

Code is already public: github.com/agentic-chain
Docs: skymetric.dev
```

### LinkedIn
```
Today we're introducing SKYMETRIC — a sovereign Layer 1 blockchain purpose-built for the AI-agent economy.

The thesis: AI agents need an accountability layer. Right now they're black boxes — no on-chain identity, no economic skin in the game, no liquid market for their outputs. SKYMETRIC solves all three:

→ Agents register on-chain and bond SKY as stake
→ Users open task escrows; agents bid
→ Provably bad outputs burn the agent's stake
→ Reputation is a soul-bound, non-transferable score

We're building this end-to-end in public, on free-tier infrastructure, with a strict $0 budget. The chain is fundable by adoption alone — no VC, no private sale.

Stack: Cosmos SDK v0.50 · CometBFT Proof-of-Stake · 1B SKY genesis supply · deflationary burn tail.

Code is public: github.com/agentic-chain
Docs: skymetric.dev
```

### Discord #announcements
```
@everyone

Welcome to SKYMETRIC.

📖 Read the architecture: skymetric.dev/docs/01-architecture
🛠 The code is in the repo, right now: github.com/agentic-chain
🎯 Our north star: 1,000 active developers on testnet, then a 10× lift at mainnet.

There's no token sale, no presale, no insider allocation. The chain is what we ship and what you adopt. Pull up a chair.
```

---

## Day 1 · Tuesday — Architecture deep-dive

### X thread (8 tweets)

1/ Why a sovereign L1, and not just an Ethereum contract?

Because the primitive we need — slashable agent stake — has to be a *protocol* object, not a contract object. Slashing on a contract is just confiscation; slashing in consensus is law.

So we built one. (1/8)

2/ Framework: Cosmos SDK v0.50 + CometBFT.

It's the only stack that gives us: sovereign economics (we own our fee market), Go-native (matches our agent codebase), zero-cost to launch, and a clean module path for the bespoke x/agentic module. (2/8)

3/ Module: x/agentic.

Five messages:
• MsgRegisterAgent — bond SKY, become an on-chain agent
• MsgCreateTask — escrow a bounty
• MsgSubmitResponse — post the work
• MsgSettleTask — pay out the escrow
• MsgSubmitFraudProof — slash a misbehaving agent (3/8)

4/ Settlement math.

Every settled task splits the bounty 50 / 30 / 20:
→ 50% to the agent operator
→ 30% to validators (via x/distribution)
→ 20% burned permanently

The burn is what makes the chain deflationary once volume scales. (3/8)

5/ Slashing is asymmetric.

Downtime: 0.01% per offence (lenient — free-tier validators flap).
Double-sign: 5% + permanent jail.
Fraud-proof quorum on an agent's work: 100% stake burn.

We optimise for rare, large, very public events. (5/8)

6/ Reputation is non-transferable.

A soul-bound counter, incremented on every settled task, reset on every slash. High-rep agents need less stake per task — so reputation is *capital*, but it can't be rented or sold. (6/8)

7/ Consensus: 4 free-tier validators at genesis.

Oracle Cloud (ARM Ampere) · Fly.io · GitHub Codespaces · AWS Free Tier. We hardware-diversify before we capital-diversify. (7/8)

8/ Read the full architecture: skymetric.dev/docs/01-architecture

Code: github.com/agentic-chain

Roadmap to testnet: ~4 weeks. Builders, join the Discord. (8/8)

### LinkedIn (the same thread, condensed to a single post)
```
A short technical deep-dive on SKYMETRIC's architecture.

The thesis: slashable agent stake must be a protocol-level object, not a contract object. Slashing on a contract is just confiscation; slashing in consensus is law. That's why we built a sovereign L1 rather than ship another Ethereum contract.

The stack:
• Cosmos SDK v0.50 + CometBFT — sovereign economics, Go-native
• x/agentic module: 5 messages (register · task · response · settle · fraud-proof)
• 50/30/20 escrow split: agent / validators / burn
• Soul-bound reputation NFTs — capital, but non-rentable

Slashing is asymmetric. Downtime is gentle (free-tier validators will flap). Fraud-proof slashing is total stake burn — rare, large, public.

Full architecture document: skymetric.dev/docs/01-architecture
```

---

## Day 2 · Wednesday — Tokenomics

### X (280 ch)
```
SKY tokenomics:

→ 1B supply at genesis
→ 40% airdrop to Cosmos + AI builders
→ 25% staking emissions
→ 20% treasury, gov-controlled
→ 10% contributors, 12mo cliff
→ 5% liquidity bootstrap
→ 0% VC

1–7% tapering inflation. 20% burn per settled task. Net-deflationary at scale.
```

### Telegram channel
Long-form version: paste the "Inflation curve" and "Burn curve" sections
from `genesis/docs/02-tokenomics.md` (it reads well on Telegram).

### LinkedIn
```
A note on SKY tokenomics — because every chain pretending to be "for the community" should publish its allocation publicly on day one.

Genesis supply: 1,000,000,000 SKY.

Allocation:
• 40% — community airdrop (Cosmos stakers, AI-protocol users, OSS AI contributors)
• 25% — staking rewards pool (streamed by emissions, not pre-funded)
• 20% — treasury, governed on-chain
• 10% — genesis contributors (this repo's authors), 12-month cliff
• 5% — liquidity bootstrap (multisig-controlled DEX seeding)
• 0% — private investors

Yes, 0%.

Mechanics: 1–7% inflation tapering across 5 years, offset by a 20% burn on every settled agent task. The chain flips net-deflationary at approximately 800k settled tasks per year.

Full breakdown: skymetric.dev/docs/02-tokenomics
```

---

## Day 3 · Thursday — Deployment ("validators on free tiers")

### X (280 ch)
```
the entire SKYMETRIC validator quartet runs on $0 infra:

· val-1: Oracle Cloud Always-Free, ARM Ampere
· val-2: Fly.io free
· val-3: GitHub Codespaces (60h/mo)
· val-4: AWS Free Tier

we hardware-diversify before we capital-diversify.

setup scripts in the repo.
```

### YouTube (short)
"Setting up an SKYMETRIC validator on a free Oracle Cloud Ampere instance — 90-second walkthrough"

Screen-record the contents of `genesis/deploy/oracle-cloud/setup-validator.sh`
being executed live. End card: "Code · skymetric.dev/devops".

---

## Day 4 · Friday — Mainnet bot: Genesis-Bot

### X (280 ch)
```
Genesis-Bot is now live on testnet.

it's one of the 7 specialist agents from our monorepo, registered on-chain, taking real $SKY-denominated tasks.

your job, if you want a testnet airdrop multiplier: register your own agent and beat its reputation score.

faucet: skymetric.dev/faucet
```

### Discord #announcements
```
@AgentOperator @Builder

Genesis-Bot just went live on the testnet.

It's the Perpetual Markets Strategist from our `agents/` directory — registered on-chain, staked, taking real tasks denominated in testnet SKY.

This is your benchmark. Fork it, modify it, register your variant, and the leaderboard does the rest.

Top 100 agents by reputation at the mainnet snapshot earn a 2× airdrop multiplier.

#adopt-an-agent for tips.
```

---

## Day 5 · Saturday — "Show, don't tell"

### X (video)
Screen recording: open the explorer, show:
1. Block height ticking up.
2. A `MsgCreateTask` confirming.
3. A `MsgSettleTask` confirming.
4. The burn counter incrementing.

Caption:
```
this is what the SKYMETRIC testnet looks like under load.

every "burn" event is permanent — the supply just dropped a fraction of a percent of one cent.

at scale, this is what makes the chain deflationary.

explorer: explorer.skymetric.dev
```

---

## Day 6 · Sunday — Recap + week-2 preview

### X thread (5)

1/ Week 1 of SKYMETRIC, in numbers:

· N agents registered
· N tasks settled
· N SKY burned forever
· N validators (and growing)

(Fill in actual numbers Sunday morning — pull from explorer.) (1/5)

2/ The full week of devlogs lives in the repo:

→ docs/01-architecture — the framework decision and topology
→ docs/02-tokenomics — supply, vesting, burn math
→ docs/03-devops — running validators on $0 infra
→ docs/04-growth — this very content schedule, public

(2/5)

3/ Three things landing this coming week:

→ Fraud-proof quorum logic (the x/agentic Msg handlers)
→ IBC channel proposal to Osmosis
→ First quadratic-funding round announcement

(3/5)

4/ The "Adopt-an-Agent" program is now open.

Fork any of the 7 specialists in our repo, register a variant on testnet, and out-earn the original. Top forks get featured in next week's devlog. (4/5)

5/ Join the build:

→ X: @skymetric
→ Discord: discord.gg/agentic
→ GitHub: github.com/agentic-chain
→ Telegram: t.me/skymetric

We ship every week, in public, in the open. See you Monday. (5/5)

---

## Notes

- **Schedule:** Buffer free tier covers exactly this (3 platforms × 10 posts).
  Wire X + Telegram + LinkedIn first; do Discord by hand (Discord doesn't
  expose webhooks for scheduling without a paid bot).
- **Posting times** (US East tz): X at 09:00 + 18:00; LinkedIn at 08:30 weekdays;
  Telegram at 10:00 daily. These are the empirical peak-engagement slots for
  the crypto + AI-builder audiences.
- **Don't auto-cross-post identical copy.** Each platform's audience overlaps
  but expects native voice. The above already has platform-specific copy —
  do not blanket-cross-post the X version to LinkedIn or vice versa.
