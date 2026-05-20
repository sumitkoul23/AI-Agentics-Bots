# AI Agentics Bots

Monorepo for the current Teneo agent portfolio.

## ⭐ Start here: Priya Hub

`agents/priya-hub` is the **universal entry point** — one chat gives you access to every specialist agent automatically.

```bash
cd agents/priya-hub
cp .env.example .env   # add PRIVATE_KEY + ANTHROPIC_API_KEY
go mod tidy && go build -o priya-hub .
./priya-hub
```

The hub auto-routes every message to the right specialist using keyword matching + Claude AI.
You never have to pick an agent — just talk naturally.

---

## Agents

### 🌸 Priya Hub (universal router)
| Agent ID | `priya-hub-z7w4v9` |
|---|---|
| Path | `agents/priya-hub/` |
| Purpose | Routes every chat to the right specialist — one entry point for all agents |

**Built-in specialists** (all accessible from the hub):

| Specialist | Trigger keywords | Description |
|---|---|---|
| Perpetual Markets Strategist | perp, funding rate, trade plan, RSI, OI… | Crypto perp analysis + trade plans |
| Portfolio Strategist | portfolio, allocation, rebalance, hedge… | Multi-asset portfolio management |
| Social Media Expert | tweet, linkedin post, instagram, tiktok… | Content for all 7 platforms |
| Communication Specialist | email, dm, follow up, negotiate… | Emails, proposals, scripts |
| Personal Organizer | brain dump, todo, daily plan, stuck… | Tasks, planning, focus |
| Finance & Crypto Analyst | btc, eth, defi, macro, stock… | Markets, yields, explainers |
| Freelance & Jobs Advisor | freelance, upwork, proposal, skill gap… | Jobs, bids, rates, clients |
| Priya (General) | *(catch-all)* | Default for everything else |

Force a specific agent: `/use perp-markets BTC trade plan long`

---

### Other standalone agents

| Agent | Agent ID | Purpose |
|---|---|---|
| Priya Bot | `priya-mobile-bot-x9k2m7` | Autonomous mobile AI — social, trading, comms, freelance |
| Mobile Assistant Bot | `mobile-assistant-bot-a1b2c3` | Quick mobile-first assistant |
| Perpetual Market Strategist | `perp-strategist-7fb31d` | Command-based perp trading workflow |
| Perpetual Markets Strategist AI v2 | `perp-strategist-7fb31d-v2` | NLP perp strategy assistant |
| Perpetual Markets Strategist AI v3 | `perp-strategist-7fb31d-v3` | Public NLP variant |
| Category Agent Portfolio | *(per metadata)* | Shared runtime for category-based agents |

---

## Layout

```
agents/
├── priya-hub/                  ← universal entry point (start here)
├── priya-bot/                  ← full autonomous Priya agent
├── mobile-assistant-bot/       ← lightweight mobile agent
├── perpetual-market-strategist/
├── perpetual-markets-strategist-ai/
├── perpetual-markets-strategist-ai-v3/
└── category-agent-portfolio/
```

## Safety

- Wallet keys stay local — never committed.
- `.env` files, binaries, and deploy-state files are git-ignored.
- Use `.env.example` in each agent folder as the template.

## Local workflow

1. `cd agents/priya-hub`
2. `cp .env.example .env` and fill in `PRIVATE_KEY` + `ANTHROPIC_API_KEY`
3. `go mod tidy && go build -o priya-hub . && ./priya-hub`
