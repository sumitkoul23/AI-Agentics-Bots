# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## What This Repo Is

Monorepo for a portfolio of Teneo Protocol AI agents. Each agent is an autonomous Go program that connects to the Teneo decentralized agent network, registers its identity on-chain (as an NFT), and handles user tasks through the Teneo Agent SDK. Agents are monetized via the x402 payment protocol — the platform verifies on-chain payment before routing a task to `ProcessTask`.

---

## Build & Run

Each agent is an independent Go module. All commands run from inside the agent's own directory.

```bash
# Build
cd agents/<agent-dir>
go build .

# Run (Linux/Mac)
PRIVATE_KEY=<hex-key> ./agent-binary <metadata-file.json>

# Run (Windows) — category-agent-portfolio only
cd agents/category-agent-portfolio
.\run-all.ps1          # builds binary, then spawns one process per agents/*.json
```

Go 1.24+ required. No test suite exists in this repo.

---

## Architecture

### Two agent patterns

**Pattern 1 — Command agent** (`agents/perpetual-market-strategist`)

Calls `nft.Mint("metadata.json")` on every startup (gasless; safe to re-run). Parses `ProcessTask` input as whitespace-separated tokens: first token = command name, rest = args. All market data fetches public Binance USDT-M Futures REST (`https://fapi.binance.com`). Live order placement is off by default; enabled only when `ALLOW_LIVE_TRADING=true` + API keys + `confirm=EXECUTE_LIVE_ORDER` are all present.

**Pattern 2 — NLP/portfolio agent** (`category-agent-portfolio`, `perpetual-markets-strategist-ai`, `perpetual-markets-strategist-ai-v3`)

Calls `deploy.DeployAgent(DeployConfig{...})` on startup, which creates or updates the on-chain agent registration and writes `.teneo-deploy-state-<agent-id>.json` locally. Then calls `agent.NewEnhancedAgent`. `ProcessTask` is a stub — empty input returns `opening_line`; non-empty input returns a formatted echo. Real NLP handling is done by the Teneo platform, not in this code.

**Pattern 2b — Pre-minted variant** (`perpetual-markets-strategist-ai`)

Same as Pattern 2 but skips minting. Reads `NFT_TOKEN_ID` from env and passes it directly to `EnhancedAgentConfig`. Use this when the token already exists and re-deploying would create a duplicate.

---

## Teneo Agent SDK — Deep Reference

### Required interface

```go
type AgentHandler interface {
    ProcessTask(ctx context.Context, task string) (string, error)
}
```

### Optional interfaces the agent struct can implement

| Interface | Method | When it runs |
|---|---|---|
| `AgentInitializer` | `Initialize(ctx)` | Once at startup before first task |
| `AgentCleaner` | `Cleanup(ctx)` | On graceful shutdown |
| `TaskResultHandler` | `OnTaskResult(ctx, task, result)` | After each successful task |
| `StreamingTaskHandler` | `ProcessTaskStreaming(ctx, task, sender)` | Replace ProcessTask for multi-step responses |

`StreamingTaskHandler` receives a `types.MessageSender` with methods: `SendMessage`, `SendTaskUpdate`, `SendMessageAsJSON`, `SendMessageAsMD`, `TriggerWalletTx`.

### `EnhancedAgentConfig` fields

```go
&agent.EnhancedAgentConfig{
    Config:          cfg,            // from agent.DefaultConfig() + cfg.LoadFromEnv()
    AgentHandler:    &MyAgent{},
    AgentID:         meta.AgentID,
    TokenID:         result.TokenID,
    SubmitForReview: true,           // auto-submits for public listing after deploy
    StateFilePath:   ".teneo-runtime-state-<id>.json",
}
```

### `deploy.DeployConfig` fields

```go
deploy.DeployConfig{
    PrivateKey:       os.Getenv("PRIVATE_KEY"),  // hex, 0x prefix optional
    AgentID:          meta.AgentID,
    AgentName:        meta.Name,
    Description:      meta.Description,
    AgentType:        meta.AgentType,            // "command", "nlp", or "commandless"
    Capabilities:     capabilitiesJSON,           // []byte from json.Marshal
    Commands:         meta.Commands,              // json.RawMessage
    NlpFallback:      meta.NLPFallback,
    Categories:       categoriesJSON,
    ShortDescription: meta.ShortDescription,
    FAQItems:         meta.FAQItems,
    MetadataVersion:  "2.4.0",                   // always use this version
    StateFilePath:    ".teneo-deploy-state-<id>.json",
}
```

### Agent visibility lifecycle

```
private → in_review → public   (or declined)
```

- `SubmitForReview: true` in `EnhancedAgentConfig` auto-submits after deploy.
- Any change to capabilities or commands resets status to `private` — resubmit required.
- Review takes up to 72 hours; agent must be **online** during review.
- `review_status` is tracked in `agent-identity.json` (see below).

### Health endpoints (built into SDK)

Default port: `8080`. Overridden by `health_port` in metadata JSON (or `HEALTH_PORT` env var).
- `GET /health` — liveness
- `GET /status` — operational status
- `GET /info` — agent metadata

Each agent in the portfolio must have a **unique** health port:

| Agent | Port |
|---|---|
| social-signal-strategist-8f1a | 8082 |
| lead-intelligence-qualifier-8f1a | 8083 |
| commerce-pricing-analyst-8f1a | 8084 |
| travel-opportunity-planner-8f1a | 8085 |
| developer-workflow-auditor-8f1a | 8086 |
| news-impact-briefing-8f1a | 8087 |
| campaign-url-analyzer-8f1a | 8088 |
| **next available** | **8089+** |

---

## Metadata JSON — Two Schemas

### Runtime metadata (used by Go binary, snake_case keys)

Used by `category-agent-portfolio/main.go` and `perpetual-markets-strategist-ai-v3`. Fields:

```json
{
  "name": "...",
  "agent_id": "slug-with-dashes",
  "short_description": "...",
  "description": "...",
  "agent_type": "nlp",
  "categories": ["Category 1", "Category 2"],
  "capabilities": [{"name": "snake_case_name", "description": "..."}],
  "commands": [],
  "nlp_fallback": false,
  "faq_items": [],
  "health_port": 8088,
  "opening_line": "...",
  "output_style": "..."
}
```

### Agent Console template (camelCase keys, for manual re-registration)

Files like `perpetual-markets-strategist-ai-template.json` use camelCase (`agentId`, `shortDescription`, `nlpFallback`, `faqItems`). These are **not** read by the Go binary directly — they are reference templates for the Agent Console UI or future tooling.

### `agent-identity.json`

Written by the SDK (or manually maintained) to record the live deployed state:

```json
{
  "agent_id": "perp-strategist-7fb31d-v3",
  "token_id": 1134,
  "wallet_address": "0x...",
  "agent_type": "nlp",
  "categories": ["Crypto", "Trading"],
  "review_status": "submitted"
}
```

`token_id` from this file feeds `NFT_TOKEN_ID` for pre-minted agent variants.

---

## Environment Variables

| Variable | Required | Agent | Notes |
|---|---|---|---|
| `PRIVATE_KEY` | Yes | all | Hex Ethereum key; `0x` prefix optional |
| `ACCEPT_EULA` | Recommended | all | Set `true` to suppress interactive prompt |
| `AGENT_METADATA_FILE` | Yes* | category-agent-portfolio | Path to JSON; overridden by `os.Args[1]` |
| `NFT_TOKEN_ID` | Yes* | perpetual-markets-strategist-ai | Pre-minted token; skips re-mint |
| `HEALTH_PORT` | No | all | Overrides `health_port` in JSON; default 8080 |
| `RATE_LIMIT_PER_MINUTE` | No | all | `0` = disabled |
| `REDIS_ENABLED` | No | all | `true` enables response caching |
| `BINANCE_FUTURES_BASE_URL` | No | perp-market-strategist | Defaults to `https://fapi.binance.com` |
| `ALLOW_LIVE_TRADING` | No | perp-market-strategist | Must be `"true"` for live Binance orders |
| `MAX_ORDER_NOTIONAL_USD` | If live | perp-market-strategist | Hard cap per order |
| `BINANCE_FUTURES_API_KEY` | If live | perp-market-strategist | Binance Futures API key |
| `BINANCE_FUTURES_API_SECRET` | If live | perp-market-strategist | Binance Futures API secret |
| `OPENAI_API_KEY` | For LLM agents | future | Required for `SimpleOpenAIAgent` pattern |
| `OPENCLAW_API_TOKEN` | For OpenClaw | future | Required for `openclaw.NewOpenClawAgent` |

Copy the relevant `.env.example` to `.env` before running. `.env` is gitignored.

---

## Adding a New Agent

**Category agent (NLP):** Add one JSON file to `agents/category-agent-portfolio/agents/<slug>.json`. Pick a health port ≥ 8089. `run-all.ps1` launches it automatically — no Go code changes needed.

**New standalone agent:** Create a new directory under `agents/`, add `go.mod`, `main.go`, a metadata JSON, `.env.example`, and a launcher script. Follow the Pattern 2 structure from `perpetual-markets-strategist-ai-v3`.

**`agent_id` rules:** lowercase, hyphen-separated, permanent. Changing it creates a new on-chain identity. Suffix `-8f1a` is a convention for the current category cohort (wallet `0x8f1a…`).

---

## State Files (gitignored)

| File | Written by | Purpose |
|---|---|---|
| `.teneo-deploy-state-<id>.json` | `deploy.DeployAgent` | Cached deployment result; prevents redundant on-chain calls |
| `.teneo-runtime-state-<id>.json` | `agent.NewEnhancedAgent` | Live runtime token/session state |
| `agent-identity.json` | manually / SDK | Deployed token_id + review_status reference |

Delete deploy-state files to force a fresh on-chain registration on next run.

---

## Pricing (x402 Protocol)

Commands declare price in the metadata JSON `commands` array:

```json
{
  "trigger": "analyze",
  "pricePerUnit": 0.5,
  "priceType": "task-transaction",
  "taskUnit": "per-query"
}
```

The platform verifies payment before calling `ProcessTask`. `ProcessTask` logic is unchanged regardless of pricing — no payment code needed in the handler.

First-query-free is declared via `"first_query_free": true` in the metadata `pricing_policy` block (command agent pattern). Platform-layer enforcement only; the Go handler cannot differentiate paid vs free calls.

---

## Session Intelligence — Campaign URL Analyzer

This section records knowledge built during the Manappuram campaign analysis engagement. Future sessions working on `campaign-url-analyzer-8f1a` or similar digital marketing agents should use this as baseline.

### UTM Parameter Standard vs. What Manappuram Uses

| Standard GA4 Field | Manappuram Param | Gap |
|---|---|---|
| `utm_source` | `utm_source=google` | None |
| `utm_medium` | `utm_medium=branded` | None |
| `utm_campaign` | `utm_campaign=instant-mt-plus-5-en` | None |
| `utm_content` | `utm_content=hindi_states_scalable` | None |
| `utm_term` ← GA4 reads this | `utm_keyword=manappuram gold loan` | **Keyword dimension blank in GA4** |
| *(custom — register in GA4)* | `utm_objective=lead_generation` | Custom dimension, not auto-read |
| *(custom — register in GA4)* | `utm_type=search_ads` | Custom dimension, not auto-read |
| *(custom — register in GA4)* | `utm_adgroupid=179971762014` | Custom dimension, not auto-read |

**Fix:** append `&utm_term={keyword}` (ValueTrack macro) to all final URLs. One hour of work, zero media cost, restores keyword reporting in GA4 immediately.

**Fix 2:** Register `utm_objective`, `utm_type`, `utm_adgroupid` as custom dimensions in GA4 Admin → Custom Definitions → Custom Dimensions. ~30 minutes.

### India Digital Marketing Platform Rates (Verified, FY2026)

These were sourced from itsaugust.com, sociolabs.in, themediaant.com, vgraple.com, upgrowth.in, sovran.ai, atomcomm.in, jigsawkraft.com, tring.co.in during this engagement.

**Google Search (Financial Services)**
- Branded CPC: ₹8–₹25 | Branded CPL: ₹400–₹700
- Non-branded CPC: ₹30–₹80 | Non-branded CPL: ₹800–₹1,800
- Competitor conquest CPC: ₹15–₹45

**YouTube**
- 6s Bumper: CPM ₹50–₹100
- 15s Non-Skip: CPV ₹0.70–₹1.50
- 30s TrueView: CPV ₹1.50–₹3.00
- Discovery: CPV ₹2.00–₹3.50

**Meta (Facebook + Instagram, Financial Services)**
- Feed Video CPM: ₹200–₹400
- Stories CPM: ₹150–₹250
- Lead Ads CPL: ₹280–₹750
- Reels CPM: ₹120–₹200

**OTT**
- JioHotstar CPM: ₹100–₹360
- ZEE5 CPM: ₹80–₹120
- SonyLIV CPM: ₹100–₹180
- MX Player CPM: ₹60–₹100

**Influencer (per Reel)**
- Nano (1K–10K): ₹2,000–₹8,000
- Micro (10K–100K): ₹8,000–₹80,000
- Mid-tier (100K–500K): ₹50,000–₹3,50,000
- Mega (1M+): ₹5,00,000–₹25,00,000

### India Gold Loan Market Context (FY2026)
- Market portfolio: ₹15.6 lakh crore (+41.9% YoY)
- Key players: Manappuram, Muthoot Finance, IIFL Gold Loan, Bajaj Finance, SBI/HDFC
- IIFL risk: RBI ban lifted Sept 2024 — re-entering digital aggressively. Manappuram gained ~99.1% AUM YoY during ban period. RLSA retargeting essential to defend those customers.
- Manappuram differentiator: Instant MT Plus = 45-min disbursal. No competitor matches this speed claim.
- Hindi-belt creative gap: `utm_content=hindi_states_scalable` tag indicates geo-targeting to Hindi-speaking states (UP, Bihar, MP, Rajasthan) but English-only creative. Native Hindi creative expected to improve CTR 15–25%.

### Deliverables Created (branch: `claude/analyze-campaign-url-orMHV`)
- **Gamma presentation** (10 slides, Aurum theme): https://gamma.app/docs/9vxccysgyxwpzav
- **Campaign Intelligence Sheet**: https://docs.google.com/spreadsheets/d/1sQ72838s_A7fdMHOYZeNIjTeTctMFjaiy7HCeAcuP3g
- **Full-Funnel Media Plan Sheet**: https://docs.google.com/spreadsheets/d/1dWtDIhHa2JzKva9_Fzko0sby2nagzagcgxDsSmb0KSA
- **Verified Platform Rate Card**: https://docs.google.com/spreadsheets/d/1c_x24vsjeGWnTq_HZ7NPqPniLpN4lnoviEr7fypUfls

---

## Repo Conventions

- No celebrity or brand ambassador references in any agent output, analysis, or deliverable.
- Real wallet keys never committed — always in `.env` (gitignored).
- When generating real rates for Indian digital media, use verified sources (not estimates) and note the source year.
- `metadata_version` stays `"2.4.0"` until the SDK changelog says otherwise.
- Capability names in runtime JSON use `snake_case`; Agent Console templates use `camelCase` — keep them in sync when updating both.
