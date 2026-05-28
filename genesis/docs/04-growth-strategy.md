# Agent 5 — Growth Hacker: $0 Adoption Playbook

## North-star: first 1,000 active addresses on testnet, then a 10× lift at mainnet.

Everything below is **organic, dev-led, content-led, and zero-budget.** Paid
ads are explicitly excluded — they're the laziest way to spend a treasury and
the easiest to detect as Sybil farming.

---

## 1. Pre-launch (week –4 → week 0): "build in public"

| Channel | Cadence | Owner | Asset |
|---|---|---|---|
| X / Twitter thread series | 3× / week | this repo's @sumitkoul23 | "Building a chain for AI agents with $0" — devlog format, 6 threads tracing each Genesis agent's work, screenshots of the code in this very PR |
| GitHub repo README + issue board | continuous | maintainers | Tag every roadmap item as `good-first-issue` to invite drive-by PRs |
| Cosmos Discord (#dev-general, #cosmwasm) | weekly office hours | maintainers | Public Q&A on Cosmos SDK v0.50 gotchas — earn reputation in the ecosystem we're plugging into |
| HackerNews "Show HN" | once, at testnet launch | maintainers | "Show HN: A blockchain where AI agents stake on the quality of their own work" |

Output of this phase: the testnet faucet sees its first 100 wallets *before*
the airdrop announcement, all organically interested.

## 2. Testnet (`skymetric-test-1`): the developer flywheel

### 2.1 Faucet quest

The testnet faucet doesn't dispense free tokens — it dispenses *quests*. To
earn 1,000 test-SKY a developer must:

1. Register an agent with `MsgRegisterAgent`.
2. Receive a task from the on-chain "Genesis Bot" (a public agent run by the
   maintainers using one of the specialists already in `agents/`).
3. Submit a response CID.
4. Get the response settled.

That's the entire onboarding funnel — by the time a dev qualifies for the
faucet, they have already touched four message types and understand the
chain. They're a real user.

### 2.2 "Adopt-an-agent" program

Each of the 7 specialists in `agents/` becomes a publicly-bonded agent on
testnet. Devs can fork any of them, register their own variant, and
immediately compete for the same `Task` queue. Forks that out-earn the
original get featured in our weekly devlog — public reputation is the
incentive.

### 2.3 Hackathons we don't pay for

Apply (free) to be a track sponsor at:
- [ETHGlobal](https://ethglobal.com/) — Cosmos × AI cross-track, prizes paid
  in SKY from the bootstrap bucket, not USD.
- [Cosmoverse](https://cosmoverse.org/) hackathon — natural fit, the Cosmos
  Foundation actively wants new app-chains to showcase.
- [Encode Club](https://www.encode.club/) AI hack — 100% online, free
  participation, large global audience.

Cost: maintainer time judging.

### 2.4 Content seeding

- Open-source the entire `genesis/` directory (already done as of this PR).
- Cross-post the architecture doc to the [Cosmos Forum](https://forum.cosmos.network/)
  — historically generates 3–5k organic eyeballs per technical post.
- Submit `docs/01-architecture.md` to the [`awesome-cosmos`](https://github.com/cosmos/awesome) list.

## 3. Mainnet launch: the airdrop is the marketing

When the airdrop snapshot Merkle root publishes, **the eligibility checker is
the landing page** — every wallet that pastes its address sees a personalised
"you are eligible for X SKY" message and a "share your eligibility" button.
The viral coefficient on these eligibility checkers is empirically > 1.0
across every chain that's tried it (Optimism, Arbitrum, Celestia, etc.).

Total cost: a single static HTML file on Cloudflare Pages.

## 4. Beyond launch: compounding loops

| Loop | Mechanic |
|---|---|
| **Reputation NFT envy** | Top-100 reputation agents are auto-tweeted by an on-chain bot every Monday. Public leaderboard → competition for inclusion. |
| **Burn ticker** | Every fee-burn block emits an event; a Cloudflare Worker tweets the running total when it crosses each million-SKY milestone. Free "number go down" content. |
| **Cross-chain agent migration** | When IBC opens to Osmosis/Neutron, existing agents on those chains can register on SKYMETRIC and inherit their reputation via IBC packets. Pulls users from neighbouring ecosystems. |
| **Quadratic-funded grants** | Quarterly QF rounds (5% treasury) turn every grantee into a marketer of the chain they got funded on. |

## 5. KPIs (and the dashboard)

A free Grafana Cloud dashboard tracks the only five numbers that matter:

1. Daily active addresses
2. Tasks settled / day
3. usky burned / day
4. Registered agents (cumulative)
5. Validator count + bonded ratio

Targets:

| Milestone | Time | DAA | Agents | Bonded |
|---|---|---|---|---|
| Testnet | week +4 | 1,000 | 50 | 1M test-SKY |
| Mainnet day 1 | month +2 | 5,000 | 200 | 100M SKY |
| Mainnet month 6 | month +8 | 25,000 | 1,500 | 500M SKY |
| Mainnet year 1 | month +14 | 100,000 | 10,000 | 670M SKY (goal-bonded) |
