# Testnet faucet — "quest, not drip" spec

## Why not a normal drip faucet?

Drip faucets get drained by bots, produce no signal, and create hostile users
("where are my tokens?"). A quest-gated faucet selects for real developers
and gives every claim an on-chain reputation footprint.

## Eligibility (any one)

1. Wallet has signed at least one tx on any Cosmos SDK testnet in the last 90 days.
2. Wallet's controller email matches a GitHub account with ≥ 1 public repo.
3. Wallet has completed the [Cosmos SDK tutorial](https://tutorials.cosmos.network/)
   on-chain checkpoint (a hash committed to a public registry).

Verified via stateless public-data lookups — no centralised eligibility DB.

## Quest sequence

| Step | Action | Reward |
|---|---|---|
| 1 | `MsgRegisterAgent` with any moniker + a 100-test-SKY stake (preloaded from a 100-test-SKY starter drip — the only direct drip in the system) | 0 |
| 2 | Receive a task from `Genesis-Bot` (an agent run by maintainers) | 0 |
| 3 | Submit a `MsgSubmitResponse` with any valid IPFS CID | 200 test-SKY |
| 4 | Have the task settled via `MsgSettleTask` | 800 test-SKY |
| 5 | Repeat 5×; on the 5th settled task, earn a soul-bound GenesisProof NFT | mainnet airdrop multiplier 2× |

Total ceiling per address: 5,000 test-SKY + 1 NFT. Hard cap discourages
Sybil grinding because the marginal cost per address (gas + wallet creation
+ IPFS upload) exceeds the marginal mainnet airdrop value.

## Implementation surface

- Faucet UI: single static HTML page on Cloudflare Pages.
- Backend: a Cloudflare Worker that reads chain state via the public
  testnet RPC and signs the starter-drip tx from a hot wallet whose balance
  is < 1k test-SKY at any time (auto-refilled by a cron worker).
- No database. Eligibility is recomputed from chain state on each request.
