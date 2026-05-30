# SKYMETRIC — agent economy on Neutron (CosmWasm), SKY as the settlement coin

> **Current direction:** ship as CosmWasm contracts on Neutron. The
> sovereign-L1 path is preserved in `chain/` for historical reference
> but is no longer the active build. See
> [`docs/10-cosmwasm-pivot.md`](docs/10-cosmwasm-pivot.md) for the
> decision rationale.
>
> SKYMETRIC is an open agent-economy: AI agents register on-chain, bond
> SKY, accept escrowed tasks, and earn (or lose) reputation through
> verifiable work. Built end-to-end with a **$0.00 budget** on open-
> source infrastructure.

## What's here, top to bottom

```
genesis/
├── README.md                     ← you are here
├── STATUS.md                     ← single-page build status
├── docs/                         ← 10 docs: architecture → wallet → cosmwasm-pivot
├── contracts/                    ← ★ CosmWasm contracts (the active build)
│   └── agent-registry/           ← compiles + 5 unit tests pass
├── agents-catalog/               ← specs for agents to operate on SKYMETRIC
│   └── github-experts.md         ← 8 GitHub-specialist agent specs
├── chain/                        ← (deprecated) Cosmos SDK sovereign-L1 attempt
├── deploy/                       ← Docker, Cloudflare, hosting recipes
├── site/                         ← skymetric.dev landing page (static)
├── frontend/dex/                 ← dex.skymetric.dev (Next.js)
├── wallet/                       ← Skymetric Wallet — extension + Keplr fork plan
├── exchange/                     ← DEX product spec + CEX roadmap
└── growth/                       ← Social handles, bios, brand, video scripts
```

## TL;DR

- **Chain ID:** `skymetric-1`
- **Framework:** Cosmos SDK v0.50 + CometBFT PoS
- **Consensus:** Proof-of-Stake, 4 free-tier validators at genesis
- **Native coin:** `SKY` (`usky`, 1 SKY = 10⁶ usky)
- **Supply:** 1,000,000,000 SKY at genesis · 1–7 % tapering inflation · 20 % burn per settled task
- **Utility:** Stake / slash / settle / reputation layer for AI agents
- **Budget:** $0.00

## Documentation index

| Doc | Topic |
|---|---|
| [`docs/01-architecture.md`](docs/01-architecture.md) | Framework choice, consensus, network topology, roadmap, risks |
| [`docs/02-tokenomics.md`](docs/02-tokenomics.md) | Supply, vesting, inflation, burn curve, distribution strategy |
| [`docs/03-devops.md`](docs/03-devops.md) | Free-tier validator quartet, monitoring, runbook |
| [`docs/04-growth-strategy.md`](docs/04-growth-strategy.md) | $0 adoption playbook, dev outreach, KPIs |
| [`docs/05-exchange-strategy.md`](docs/05-exchange-strategy.md) | 4-tier DEX → CEX path, on-chain gate at $100M TVL |
| [`docs/06-financial-instruments.md`](docs/06-financial-instruments.md) | 10 novel instruments ranked by tractability × differentiation |
| [`docs/07-cex-gate-rationale.md`](docs/07-cex-gate-rationale.md) | Why the strict CEX-launch gate is the cheapest LP-acquisition strategy |
| [`docs/08-wallet-strategy.md`](docs/08-wallet-strategy.md) | Why fork Keplr instead of building from scratch |
| [`wallet/HOW_TO_GET_A_WALLET_TODAY.md`](wallet/HOW_TO_GET_A_WALLET_TODAY.md) | **Practical 5-min path to a working Skymetric wallet today.** Same mnemonic works in the v1 branded wallet later. |
| [`docs/09-honest-launch-timeline.md`](docs/09-honest-launch-timeline.md) | **What stands between this scaffold and mainnet — read this before asking "when can we launch."** |

## Code map

### `chain/` — the Cosmos SDK app

```
chain/
├── app/                  ← runtime wiring (config.go, encoding.go, genesis.go, app.go)
├── cmd/skymetricd/         ← the binary entry point + CLI root cmd
├── x/agentic/            ← agent registry · tasks · fraud-proof slashing
├── x/agenticdex/         ← constant-product AMM (the chain's native DEX)
├── x/agenticperps/       ← virtual-AMM perpetuals with funding-rate accrual
├── x/agenticrouter/      ← atomic multi-hop swap aggregator
├── proto/                ← protobuf source (buf-generated in CI)
├── scripts/              ← init-chain.sh · start-node.sh · create-validator.sh
├── config/               ← genesis-overrides.json · app.toml.example
├── go.mod · Makefile
```

### `wallet/` — Skymetric Wallet (3-stage rollout)

```
wallet/
├── extension/            ← MV3 browser-extension scaffold (UX prototype; refuses to sign)
├── keplr-fork/           ← week-by-week fork plan + diff-keeping policy
└── agent-views/          ← React + React-Query components reusable in
                            both the Keplr fork and the DEX frontend
```

### `exchange/` — 4-tier exchange strategy

```
exchange/
├── dex/README.md         ← Tier 1 product spec (DEX frontend at dex.skymetric.dev)
└── cex/                  ← Tier 4 docs (Binance-class licensed CEX, gated at $100M TVL)
    ├── README.md
    ├── roadmap.md
    ├── jurisdictions.md
    ├── architecture.md
    └── compliance-stack.md
```

## Quickstart

### Spin up a local devnet

```bash
cd genesis/chain
make install
./scripts/init-chain.sh   # initialises a single-node devnet
./scripts/start-node.sh   # starts CometBFT + the app
```

RPC at `http://localhost:26657`, REST at `http://localhost:1317`.

### Deploy the landing page

1. Cloudflare Pages → connect to this repo
2. Build command: *(blank)*
3. Output directory: `genesis/site`
4. Save and Deploy

The DEX frontend follows the same pattern with output dir
`genesis/frontend/dex/out` and build command
`cd genesis/frontend/dex && npm i && npm run build`.

### Load the wallet extension (UX-only, won't sign anything)

1. Chrome → `chrome://extensions/` → Developer mode.
2. "Load unpacked" → pick `genesis/wallet/extension/`.
3. Generate the PNG icons first — see
   [`wallet/extension/README.md`](wallet/extension/README.md).

## CI / CD

GitHub Actions workflows in [`../.github/workflows/`](../.github/workflows/):

| Workflow | Trigger | Output |
|---|---|---|
| `genesis-chain.yml` | push / PR touching `chain/` | go vet · build · test · golangci-lint |
| `genesis-release.yml` | tag `genesis-v*` | linux/amd64 + linux/arm64 + darwin/arm64 tarballs, multi-arch docker to GHCR, GitHub Release |
| `genesis-site.yml` | push / PR touching `site/` | static-site sanity checks |

All on free GitHub-hosted runners. No paid runners.

## Status

| Component | State |
|---|---|
| `x/agentic` module | scaffold + full Msg handlers + tests |
| `x/agenticdex` module | scaffold + Msg handlers + swap-math tests |
| `x/agenticperps` module | scaffold + funding/position/liquidation logic |
| `x/agenticrouter` module | scaffold + sync exec (cross-chain v1) |
| `app/app.go` wiring | all four modules registered |
| Landing page | deployable as-is |
| DEX frontend | UI scaffold; wallet signing flow v1 |
| Wallet extension | UX prototype; real keys in v1 (Keplr fork) |
| CI workflows | running on every push |
| Proto generation | **not yet** — handlers are hand-rolled stubs until `buf generate` is wired |
| Audit | **not yet** — required before any mainnet binary release |

## What ships next

The roadmap lives in [`docs/01-architecture.md`](docs/01-architecture.md#7-roadmap)
and is also tracked in PR #7 as the source of truth for the genesis branch.

Priorities for the next batch (ordered by leverage):

1. Wire `buf generate` for proto-gen → flips CI from advisory to hard-fail.
2. Submit SKYMETRIC to `cosmos/chain-registry` so Keplr / Leap auto-detect.
3. Ship financial instrument **#4 streaming payments** (smallest module,
   largest unlock — see `docs/06-financial-instruments.md`).
4. Bootstrap the testnet `skymetric-test-1` per `docs/03-devops.md`.

## License

Apache 2.0 (matches Cosmos SDK + Keplr upstream).
