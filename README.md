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
| Campaign URL Analyzer 8F1A | Digital Marketing + Analytics | `campaign-url-analyzer-8f1a` |

## Example: Manappuram Gold Loan Campaign URL Analysis

URL analyzed by Campaign URL Analyzer 8F1A:

```
https://www.manappuram.com/campaigns/instant-mt-plus-5-en
  ?utm_source=google
  &utm_medium=branded
  &utm_objective=lead_generation
  &utm_type=search_ads
  &utm_campaign=instant-mt-plus-5-en
  &utm_keyword=manappuram%20gold%20loan
  &utm_content=hindi_states_scalable
  &utm_adgroupid=179971762014
  &gad_source=2
  &gad_campaignid=22698563720
  &gclid=CjwKCAjw8arQBhB9EiwAfIKd...
```

| Parameter | Value | Interpretation |
|---|---|---|
| `utm_source` | `google` | Traffic originates from Google Ads |
| `utm_medium` | `branded` | Branded search — bidding on own brand keywords |
| `utm_objective` | `lead_generation` | Custom param declaring campaign KPI: leads |
| `utm_type` | `search_ads` | Search network placement |
| `utm_campaign` | `instant-mt-plus-5-en` | Product: Instant MT Plus, variant 5, English creative |
| `utm_keyword` | `manappuram gold loan` | Exact branded keyword triggering the ad |
| `utm_content` | `hindi_states_scalable` | Creative set targeting Hindi-belt states, auto-scaling bid |
| `utm_adgroupid` | `179971762014` | Google Ads ad group ID (non-standard param, appended manually) |
| `gad_source` | `2` | Google Ads click source indicator |
| `gad_campaignid` | `22698563720` | Google Ads campaign ID for cross-channel reconciliation |
| `gclid` | `CjwKCAjw8arQ…` | Google Click ID — enables server-side conversion import |

**Key findings:**
- Branded-keyword defense strategy: capturing users already searching for the brand to prevent competitor poaching.
- Language/geo split: `hindi_states_scalable` content tag reveals a geo-segmented creative strategy for Hindi-speaking Indian states.
- Dual campaign ID tracking (`utm_adgroupid` + `gad_campaignid`): supports granular ad-group-level attribution alongside campaign-level Google Ads data.
- `gclid` present: server-side conversion tracking is configured; enhanced conversions or offline import is likely in use.
- Landing page slug (`instant-mt-plus-5-en`) mirrors `utm_campaign`, confirming a dedicated campaign page — good attribution hygiene.
- **Gap**: no `utm_term` populated at landing-page level (keyword passed via custom `utm_keyword`); any downstream tool expecting standard `utm_term` will miss the keyword dimension.

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
