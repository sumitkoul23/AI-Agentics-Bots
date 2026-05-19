# Priya — Autonomous Mobile AI

> Your autonomous AI — social media expert, finance analyst, copywriter, organizer, and freelance specialist. Powered entirely by Claude (Anthropic).

## Avatar

Use this prompt in Midjourney or DALL-E 3 to generate Priya's avatar:

```
professional Indian woman AI assistant, warm confident smile, modern business casual,
clean gradient background, photorealistic, 8k, friendly and approachable
```

## Quick Start

```bash
cd agents/priya-bot

# 1. Configure keys (only two required)
cp .env.example .env
# Edit .env:
#   PRIVATE_KEY=<your Teneo wallet key>
#   ANTHROPIC_API_KEY=<your key from console.anthropic.com>

# 2. Fetch dependencies and build
go mod tidy
go build -o priya-bot .

# 3. Run
./priya-bot

# Windows PowerShell
.\run-agent.ps1
```

## Architecture

```
priya-bot/
├── main.go          — Teneo SDK entry + NLP intent router + nil-safe guards
├── persona.go       — Priya's personality, expertise, and system prompt
├── ai_core.go       — Claude API wrapper with enriched context + conversation memory
├── memory.go        — Persistent JSON state (preferences, history, jobs, posts)
├── social.go        — All 7 social platforms + calendar, strategy, image briefs
├── finance.go       — Trade plans, portfolio, DeFi, macro briefings, explainers
├── comms.go         — Email, DMs, negotiation, decline, onboarding, inbox triage
├── organizer.go     — Brain-dump→action list, daily/weekly plans, calendar blocking
├── freelance.go     — Job search, proposals (auto-tracked), rates, skills, clients
└── scheduler.go     — Autonomous cron: morning briefing, market scan, job scan, weekly plan
```

## How Priya Learns

Every interaction is stored in `.priya-memory.json`. Priya automatically loads:
- Your writing voice samples (`/learn voice`)
- Your preferences (`/set key=value`)
- Learned facts about you
- Conversation history (last 100 messages)

She gets smarter and more personalised with every conversation.

## Autonomous Background Tasks

| Task | Schedule |
|---|---|
| Morning briefing | Daily 8:00 AM |
| Social trend check + post queue | Daily 9:00 AM |
| Market scan (crypto/stocks) | Every 4 hours |
| Freelance job scan | Every 6 hours |
| Weekly planning | Monday 7:30 AM |

All results are stored in memory and available on your next interaction.

## Commands

| Command | Description |
|---|---|
| `/social <platform> <topic>` | Content for any platform |
| `/social all <topic>` | Full cross-platform content pack |
| `/social calendar` | 7-day content calendar |
| `/social strategy` | Growth playbook |
| `/social image <brief>` | Midjourney / DALL-E visual brief |
| `/trade <symbol> [long\|short]` | Trade plan with entry / stop / target |
| `/finance portfolio` | Portfolio analysis |
| `/finance defi` | DeFi yield opportunities |
| `/finance news` | Macro + market briefing |
| `/finance explain <concept>` | Plain-language explainer |
| `/email <context>` | Draft a professional email |
| `/dm <context>` | Draft a direct message |
| `/comms negotiate <context>` | Negotiation scripts |
| `/comms inbox` | Inbox triage plan |
| `/organize <brain dump>` | Turn chaos into an action list |
| `/plan daily` | Morning briefing |
| `/plan weekly` | Weekly work plan |
| `/jobs <keywords>` | Find freelance opportunities |
| `/apply <job title>` | Draft proposal (auto-tracked) |
| `/track` | Application tracker |
| `/skills` | Skill-gap report |
| `/freelance rate` | Pricing strategy |
| `/set key=value` | Save a preference |
| `/learn voice <sample>` | Teach Priya your writing style |
| `/help` | Full command reference |

## Environment Variables

| Variable | Required | Purpose |
|---|---|---|
| `PRIVATE_KEY` | Yes | Teneo wallet key for agent deployment |
| `ANTHROPIC_API_KEY` | Yes | Claude API — Priya's intelligence engine |

No other keys needed. Priya is fully self-contained using Claude.
