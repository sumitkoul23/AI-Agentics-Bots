# Mobile Assistant Bot

An all-in-one agentic bot for mobile users that handles three daily workflows in a single conversational interface:

| Workflow | What it does |
|---|---|
| **Social Media** | Draft posts for Twitter/X, LinkedIn & Instagram; suggest hashtags; surface trends |
| **Trading** | Crypto & stock trade ideas with entry/stop/target; portfolio P&L tracker |
| **Freelance & Jobs** | Search Upwork/Fiverr/Toptal/LinkedIn; draft proposals; track applications; skill-gap reports |

## Quick Start

```bash
# 1. Copy and fill in your keys
cp .env.example .env
# edit .env and set at least PRIVATE_KEY

# 2. Build & run
go build -o mobile-assistant-bot .
./mobile-assistant-bot

# Windows PowerShell
.\run-agent.ps1
```

## Commands

| Command | Description |
|---|---|
| `/social twitter <topic>` | Draft a tweet with hashtags |
| `/social linkedin <topic>` | Draft a LinkedIn post |
| `/social instagram <topic>` | Draft an Instagram caption |
| `/social trends` | Show trending topics |
| `/trade <symbol>` | Market analysis + trade idea |
| `/trade <symbol> long\|short` | Directional trade plan |
| `/portfolio` | Portfolio P&L snapshot |
| `/jobs <keywords>` | Search freelance platforms |
| `/apply <job title>` | Draft a proposal |
| `/track` | View application tracker |
| `/skills` | Skill-gap report |
| `/help` | Full command list |

You can also just type naturally — NLP fallback is enabled.

## Environment Variables

| Variable | Required | Purpose |
|---|---|---|
| `PRIVATE_KEY` | Yes | Teneo wallet key for agent deployment |
| `MARKET_DATA_KEY` | No | Enables live price data in trade plans |
| `EXCHANGE_API_KEY` + `EXCHANGE_API_SECRET` | No | Live order execution (advisory mode by default) |
| `JOBS_API_KEY` | No | Live freelance job search results |
| `TWITTER_BEARER_TOKEN` | No | Post directly to Twitter/X |
| `LINKEDIN_ACCESS_TOKEN` | No | Post directly to LinkedIn |
| `INSTAGRAM_ACCESS_TOKEN` | No | Post directly to Instagram |

## Safety

- **Trading**: advisory mode by default. Live execution requires `EXCHANGE_API_KEY`, a configured notional limit, and explicit per-trade confirmation.
- **Social posting**: drafts only unless platform tokens are set.
- **No secrets are logged** or transmitted beyond the configured API endpoints.
