# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

Each agent is an independent Go module. All commands run from within the agent's own directory.

```bash
# Build
cd agents/<agent-dir>
go build .

# Run (Linux/Mac)
PRIVATE_KEY=<wallet-key> ./agent-binary <metadata-file.json>

# Run (Windows) — launches all category agents from one script
cd agents/category-agent-portfolio
.\run-all.ps1
```

For `category-agent-portfolio`, the binary accepts a single argument: the path to a JSON metadata file. `run-all.ps1` iterates over every `agents/*.json` and spawns one process per file.

There are no test suites in this repository.

## Architecture

### Two agent patterns

**Pattern 1 — Command agent** (`agents/perpetual-market-strategist`)

`main.go` mints an NFT via `nft.Mint(metadataFile)`, loads config from env, then calls `agent.NewEnhancedAgent`. The `ProcessTask` method parses the first word as a command and dispatches on a `switch`. Market data comes from public Binance USDT-M Futures REST endpoints.

**Pattern 2 — NLP/portfolio agent** (`agents/category-agent-portfolio`, `perpetual-markets-strategist-ai`, `perpetual-markets-strategist-ai-v3`)

`main.go` reads a JSON metadata file, calls `deploy.DeployAgent` (registers/updates the agent on-chain), then calls `agent.NewEnhancedAgent`. The `ProcessTask` method either returns the `opening_line` for empty input or echoes a stub response. The real NLP processing is delegated to the Teneo platform.

### SDK entry points

| Package | Purpose |
|---|---|
| `github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent` | `DefaultConfig`, `NewEnhancedAgent`, `EnhancedAgentConfig` |
| `github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/deploy` | `DeployAgent`, `DeployConfig` — registers agent on-chain |
| `github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/nft` | `Mint` — mints agent NFT (used only by the original command agent) |
| `github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types` | `Capability` struct used in metadata |

### Metadata JSON

Every agent is defined by a JSON file. Key fields:

- `agent_id` — unique slug; used as the on-chain identifier and in state file names
- `agent_type` — `"command"` or `"nlp"`
- `health_port` — unique port per agent (existing: 8082–8088; next available: 8089+)
- `opening_line` — returned when `task` is empty
- `output_style` — hint used in stub `ProcessTask` responses

For category-agent-portfolio, the JSON path is passed at runtime via `AGENT_METADATA_FILE` env var or as `os.Args[1]`.

### State files

`.teneo-deploy-state-<agent-id>.json` and `.teneo-runtime-state-<agent-id>.json` are written at runtime to track deployed token IDs. They are gitignored and must not be committed.

## Adding a new category agent

1. Create `agents/category-agent-portfolio/agents/<slug>.json` following the existing agent JSON files.
2. Assign a unique `health_port` not already used by another agent.
3. The `run-all.ps1` launcher picks it up automatically — no code changes needed.

## Environment variables

| Variable | Used by | Purpose |
|---|---|---|
| `PRIVATE_KEY` | all agents | Wallet private key for Teneo network signing |
| `AGENT_METADATA_FILE` | category-agent-portfolio | Path to agent JSON (overridden by `os.Args[1]`) |
| `NFT_TOKEN_ID` | perpetual-markets-strategist-ai | Pre-minted token ID (no re-mint) |
| `BINANCE_FUTURES_BASE_URL` | perpetual-market-strategist | Defaults to `https://fapi.binance.com` |
| `ALLOW_LIVE_TRADING` | perpetual-market-strategist | Must be `"true"` to enable live orders |
| `MAX_ORDER_NOTIONAL_USD` | perpetual-market-strategist | Hard cap on live order notional |
| `BINANCE_FUTURES_API_KEY/SECRET` | perpetual-market-strategist | Required for live order placement |

Copy the relevant `.env.example` to `.env` before running.
