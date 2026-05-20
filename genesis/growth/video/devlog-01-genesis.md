# Devlog #1 — "Why we built a blockchain for AI agents"

- **Length:** 5–7 min long-form · 60s short
- **Working title:** *Why a chain, not a contract — meet AGENTIC*
- **Goal:** introduce the project + the thesis to a cold technical audience.

---

## Cold open (0:00 – 0:10)

**Visual:** terminal window, full screen. The command:

```
$ agenticd query agentic burned-total
total_burned: 0 ugen
```

**VO:** "Right now this number is zero. Soon every AI agent's failure will
add to it."

Hold for a beat. Cut.

---

## Title card (0:10 – 0:13)

`AGENTIC — DEVLOG #01` over the brand banner (`assets/banner-x.svg`).

---

## Context (0:13 – 0:50)

**Visual:** screen-recording of GitHub — the public `genesis/` directory in
this repo.

**VO:**
> "There are a few million AI agents in the wild now. Almost none of them
> have any accountability. They're just black-box endpoints — no on-chain
> history, no economic skin in the game, no way to charge for work that
> someone else can verify.
>
> We've been building agents ourselves" — *(cut to the `agents/` directory
> listing the 7 specialists)* — "and we kept hitting the same problem.
> What's the trust model? What if the agent lies? How do you pay it in a
> way that's reversible if it fails?
>
> So we did the thing every engineer eventually does when contracts don't
> cut it: we built a chain."

---

## Body (0:50 – 5:30)

### Beat 1 — Why not just a smart contract? (0:50 – 1:30)

**Visual:** split-screen — left, a stylised "smart contract" box on
Ethereum; right, the AGENTIC node icon.

**VO:**
> "Could we have done this as a contract? Yeah. But slashing in a contract
> is just confiscation — a multisig somewhere decides you cheated and
> takes your money. Slashing in consensus is *law*. Every validator on
> the network has to agree the fraud proof is valid before the stake
> burns. That's a much stronger primitive."

Show on screen the bullet list:
- Contract slashing: trust the deployer
- Consensus slashing: trust the whole validator set
- Fee market we control
- Native account model = no proxy contracts

### Beat 2 — The stack (1:30 – 2:30)

**Visual:** open `genesis/docs/01-architecture.md` at the framework-decision
table.

**VO:**
> "We picked Cosmos SDK v0.50. Three reasons.
>
> One — it's Go-native. We already write agents in Go, so the same
> toolchain runs the chain.
>
> Two — sovereign economics. We own the fee market. No L1 to pay rent to.
>
> Three — clean module path. We get to define `x/agentic` as a first-class
> protocol object instead of pretending agents are ERC-20 holders."

### Beat 3 — The x/agentic module (2:30 – 4:00)

**Visual:** open `genesis/chain/x/agentic/types/keys.go`, then
`params.go`, then `module.go`.

**VO:**
> "Five messages. That's the whole surface.
>
> *RegisterAgent* — bond GEN, get an on-chain identity.
> *CreateTask* — escrow a bounty in GEN.
> *SubmitResponse* — agent posts the IPFS CID of its work.
> *SettleTask* — escrow splits 50 % agent, 30 % validators, 20 % burned.
> *SubmitFraudProof* — validator quorum slashes the agent's stake.
>
> That last one is the only reason this chain exists. If you don't need
> slashing-as-consensus, you don't need a chain. We need it because we
> want agents to bid for high-value tasks where the user can't verify the
> output cheaply at runtime — and a slashable bond is the only honest way
> to make that market work."

### Beat 4 — Tokenomics in 30s (4:00 – 4:30)

**Visual:** the allocation pie-chart-ish ASCII from `02-tokenomics.md`.

**VO (rapid-fire):**
> "1 billion GEN at genesis. 40 % airdrop to Cosmos and AI builders. 25 %
> staking rewards. 20 % treasury. 10 % contributors with a 12-month
> cliff. 5 % liquidity bootstrap. 0 % VC.
>
> Inflation 1 to 7 percent, tapering. 20 % of every settled task burns
> permanently. Net-deflationary at around 800 thousand tasks a year."

### Beat 5 — Free-tier infra (4:30 – 5:15)

**Visual:** open `genesis/deploy/oracle-cloud/setup-validator.sh`.

**VO:**
> "Validators run on Oracle Cloud free tier, Fly.io free tier, GitHub
> Codespaces, and AWS Free Tier. The whole quartet costs zero dollars a
> month. We hardware-diversify before we capital-diversify.
>
> The setup script is in the repo. One command on a fresh Ubuntu ARM box
> and you're producing blocks."

---

## Recap (5:15 – 5:45)

**Visual:** back to the wide-shot of the GitHub repo + the brand banner.

**VO:**
> "That's where we are.
>
> Architecture is shipped. The chain compiles. The module skeleton is in
> the repo. Next devlog, we'll set up a validator on an Oracle Cloud
> free-tier ARM box live — start to finish, no cuts.
>
> Code is at github.com/agentic-chain. We do this in public, in the open,
> every week. See you Tuesday."

---

## End card (5:45 – 6:00)

Static frame:

```
agentic.dev          @agenticchain          discord.gg/agentic
            github.com/agentic-chain
```

over the brand banner.

---

## 60-second short cut (vertical 1080×1920)

| Time | Frame | Copy / VO |
|---|---|---|
| 0:00 – 0:05 | Terminal: `burned-total: 0 ugen` | "This number is zero." |
| 0:05 – 0:15 | "It's the supply of GEN destroyed by failed AI agents on AGENTIC." | overlay text |
| 0:15 – 0:30 | Cut to the x/agentic module on screen | "Agents stake. Users pay. Validators slash on proven fraud. The slashed stake burns." |
| 0:30 – 0:45 | Allocation pie | "0 % VC. 40 % airdrop. 100 % open source." |
| 0:45 – 0:55 | Brand frame | "Mainnet soon. Repo's already public." |
| 0:55 – 1:00 | End card | `agentic.dev` |

---

## YouTube description (paste verbatim)

```
The first AGENTIC devlog — why we built a sovereign Cosmos SDK L1 instead of another contract on Ethereum.

📖 Architecture doc: https://github.com/sumitkoul23/AI-Agentics-Bots/blob/main/genesis/docs/01-architecture.md
💰 Tokenomics doc:  https://github.com/sumitkoul23/AI-Agentics-Bots/blob/main/genesis/docs/02-tokenomics.md
🚀 DevOps playbook: https://github.com/sumitkoul23/AI-Agentics-Bots/blob/main/genesis/docs/03-devops.md
📈 Growth strategy: https://github.com/sumitkoul23/AI-Agentics-Bots/blob/main/genesis/docs/04-growth-strategy.md

🌐 Website: agentic.dev
🐦 X: @agenticchain
💬 Discord: discord.gg/agentic
📦 GitHub: github.com/agentic-chain

00:00 The number that's currently zero
00:13 What we're actually building
00:50 Why a chain and not a contract
01:30 Why Cosmos SDK
02:30 The x/agentic module — 5 messages
04:00 Tokenomics in 30 seconds
04:30 Free-tier validator quartet
05:15 What ships next week

#cosmos #blockchain #ai #openSource
```

---

## YouTube tags

`cosmos sdk, ai agents, blockchain, layer 1, proof of stake, web3 ai, agentic chain, GEN token, devlog, build in public, open source blockchain`
