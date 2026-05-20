# Genesis System — AGENTIC Chain (`GEN`)

> An autonomous, multi-agent build of a brand-new L1 blockchain whose native coin
> (`GEN`) settles work between AI agents. Built end-to-end with a **$0.00 budget**
> using open-source tooling and free-tier infrastructure.

The Genesis System is composed of five cooperating sub-agents, each owning a
slice of the lifecycle. This directory is the output of their first build pass.

| Agent | Role | Deliverables |
|---|---|---|
| 1 | Chief Architect | [`docs/01-architecture.md`](docs/01-architecture.md) |
| 2 | Core Developer | [`chain/`](chain/) |
| 3 | Tokenomics Engineer | [`docs/02-tokenomics.md`](docs/02-tokenomics.md) |
| 4 | DevOps Engineer | [`docs/03-devops.md`](docs/03-devops.md), [`deploy/`](deploy/) |
| 5 | Growth Hacker | [`docs/04-growth-strategy.md`](docs/04-growth-strategy.md), [`growth/`](growth/) |

## TL;DR

- **Chain ID:** `agentic-1`
- **Framework:** Cosmos SDK v0.50.x + CometBFT
- **Consensus:** Proof-of-Stake (4 genesis validators, free-tier hosted)
- **Native coin:** `GEN` (base denom `ugen`, 1 GEN = 10⁶ ugen)
- **Supply:** 1,000,000,000 GEN at genesis; tapering inflation; tx-fee burn
- **Utility:** Stake / slash / settle / reputation layer for AI agents
- **Budget:** $0.00

## Quickstart (local dev node)

```bash
cd genesis/chain
make install            # builds the `agenticd` binary
./scripts/init-chain.sh # initialises a single-node devnet
./scripts/start-node.sh # starts CometBFT + the app
```

Once running, hit the REST endpoint at `http://localhost:1317` and the RPC at
`tcp://localhost:26657`.

## Status

This is the **initial scaffold** produced by the Genesis System's first build
pass. It compiles to a single-node devnet; testnet hardening, governance
modules, and the agent-staking module are tracked in
[`docs/01-architecture.md`](docs/01-architecture.md#roadmap).
