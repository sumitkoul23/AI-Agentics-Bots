# AI Agentics Bots

Monorepo for the current Teneo agent portfolio.

## Included agents

| Agent | Purpose | Agent ID |
| --- | --- | --- |
| Perpetual Market Strategist | Command-based perpetual futures research and execution workflow | `perp-strategist-7fb31d` |
| Perpetual Markets Strategist AI | NLP strategy assistant for perpetual markets | `perp-strategist-7fb31d-v2` |
| Perpetual Markets Strategist AI Public | Public NLP variant prepared for Agent Console publication | `perp-strategist-7fb31d-v3` |
| Social Signal Strategist 8F1A | Social Media + AI | `social-signal-strategist-8f1a` |
| Lead Intelligence Qualifier 8F1A | Lead Generation + Automation | `lead-intelligence-qualifier-8f1a` |
| Commerce Pricing Analyst 8F1A | E-Commerce + Price Lists | `commerce-pricing-analyst-8f1a` |
| Travel Opportunity Planner 8F1A | Travel + AI | `travel-opportunity-planner-8f1a` |
| Developer Workflow Auditor 8F1A | Developer Tools + Automation | `developer-workflow-auditor-8f1a` |
| News Impact Briefing 8F1A | News + AI | `news-impact-briefing-8f1a` |

## Layout

- `agents/perpetual-market-strategist` - original command-based trading agent
- `agents/perpetual-markets-strategist-ai` - NLP v2 agent
- `agents/perpetual-markets-strategist-ai-v3` - public NLP v3 agent
- `agents/category-agent-portfolio` - shared runtime plus six category-based NLP agents

## Safety

- Real wallet keys stay local and are never committed.
- Runtime `.env` files, logs, binaries, caches, and local deploy-state files are ignored.
- Use the checked-in `.env.example` files as templates for local setup.

## Local workflow

1. Copy the relevant `.env.example` file to `.env`.
2. Fill only the values needed for that agent.
3. Run the agent-specific PowerShell launcher from its folder.

The Windows launch scripts assume Go is installed at `C:\Program Files\Go\bin\go.exe`.
