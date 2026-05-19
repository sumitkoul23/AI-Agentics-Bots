# Priya — Autonomous Mobile AI

> Your autonomous AI expert — social media, trading, comms, organizing & freelance, with the warmth and confidence of a brilliant Indian professional.

## Avatar

Priya's image generation prompt (use in Midjourney or DALL-E 3):

```
professional Indian woman AI assistant, warm confident smile, modern business casual,
clean gradient background, photorealistic, 8k, friendly and approachable
```

## Quick Start

```bash
cd agents/priya-bot

# 1. Set up environment
cp .env.example .env
# Edit .env — at minimum set PRIVATE_KEY and ANTHROPIC_API_KEY

# 2. Fetch dependencies and build
go mod tidy
go build -o priya-bot .

# 3. Run
./priya-bot

# Windows
.\run-agent.ps1
```

## Architecture

```
priya-bot/
├── main.go          # Teneo SDK entry + intent router
├── persona.go       # Priya's personality + system prompt
├── ai_core.go       # Claude API wrapper with conversation memory
├── memory.go        # Persistent JSON state (preferences, history, jobs)
├── social.go        # All social media platforms
├── finance.go       # Trading, crypto, stocks, portfolio
├── comms.go         # Email, DM, negotiation, inbox triage
├── organizer.go     # Task management, planning, getting unstuck
├── freelance.go     # Job search, proposals, application tracker
└── scheduler.go     # Autonomous background tasks (cron)
```

## Autonomous Tasks (no user action needed)

| Task | Schedule |
|---|---|
| Morning briefing | Daily 8:00 AM |
| Social trend check + post queue | Daily 9:00 AM |
| Market scan (crypto/stocks) | Every 4 hours |
| Freelance job scan | Every 6 hours |
| Weekly plan | Monday 7:30 AM |

## Key Commands

| Command | Description |
|---|---|
| `/social <platform> <topic>` | Content for any platform |
| `/social all <topic>` | Full cross-platform content pack |
| `/social calendar` | 7-day content calendar |
| `/trade <symbol> [long\|short]` | Trade plan with entry/stop/target |
| `/finance portfolio` | Portfolio analysis |
| `/email <context>` | Draft a professional email |
| `/dm <context>` | Draft a direct message |
| `/organize <brain dump>` | Turn chaos into action list |
| `/plan daily` | Morning briefing |
| `/plan weekly` | Weekly work plan |
| `/jobs <keywords>` | Search freelance opportunities |
| `/apply <job title>` | Draft proposal (auto-tracked) |
| `/track` | Application tracker |
| `/skills` | Skill-gap report |
| `/set key=value` | Save a preference |
| `/learn voice <sample>` | Teach Priya your writing style |

## Teaching Priya Your Voice

```
/learn voice Here's how I write: "I spent 3 months building this feature and 
almost gave up twice. Here's what I learned..."
```

The more samples you share, the more accurately Priya mirrors your style.

## Environment Variables

| Variable | Required | Purpose |
|---|---|---|
| `PRIVATE_KEY` | Yes | Teneo wallet key |
| `ANTHROPIC_API_KEY` | Yes | Claude API (Priya's brain) |
| `MARKET_DATA_KEY` | No | Live price data |
| `EXCHANGE_API_KEY/SECRET` | No | Live order execution |
| `JOBS_API_KEY` | No | Live job search |
| `TWITTER_*` | No | Direct Twitter posting |
| `LINKEDIN_ACCESS_TOKEN` | No | Direct LinkedIn posting |
| `INSTAGRAM_*` | No | Direct Instagram posting |
| `GMAIL_*` | No | Direct email sending |

## Safety

- **Trading**: advisory-only by default. Live execution requires exchange credentials + explicit per-trade confirmation.
- **Social posting**: Priya drafts content; you approve before it goes live (unless posting tokens are set and you explicitly ask her to post).
- **No device access**: Priya is a conversational agent — she does not access your phone hardware, contacts, camera, or messages without a companion app you install separately.
