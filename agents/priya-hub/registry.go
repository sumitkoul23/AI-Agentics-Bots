package main

import (
	"fmt"
	"strings"
)

// AgentDef describes one specialized agent registered in the hub.
type AgentDef struct {
	ID           string
	Name         string
	ShortDesc    string
	Keywords     []string // fast-path keywords — no Claude call needed
	SystemPrompt string   // full system prompt injected when this agent handles a request
}

// Registry holds all agent definitions and resolves routing decisions.
type Registry struct {
	agents []*AgentDef
	index  map[string]*AgentDef
}

func NewRegistry(mem *HubMemory) *Registry {
	r := &Registry{index: make(map[string]*AgentDef)}
	for _, a := range buildAgents(mem) {
		r.agents = append(r.agents, a)
		r.index[a.ID] = a
	}
	return r
}

func (r *Registry) Get(id string) *AgentDef {
	return r.index[strings.ToLower(id)]
}

// FastRoute returns the first agent ID whose keyword appears in the lowercased input.
func (r *Registry) FastRoute(lower string) string {
	// Longer/more-specific keywords take priority — sort by keyword length desc implicitly
	// by putting more specific agents first in allAgents().
	for _, a := range r.agents {
		for _, kw := range a.Keywords {
			if strings.Contains(lower, kw) {
				return a.ID
			}
		}
	}
	return ""
}

// Catalog returns the human-readable /agents list.
func (r *Registry) Catalog() string {
	var sb strings.Builder
	sb.WriteString("All available agents — just ask naturally:\n\n")
	for _, a := range r.agents {
		sb.WriteString(fmt.Sprintf("  %-28s %s\n", a.Name, a.ShortDesc))
	}
	sb.WriteString("\nThe hub routes automatically. Type anything.")
	return sb.String()
}

// ClassifyPrompt builds the routing question sent to the Claude classifier.
func (r *Registry) ClassifyPrompt(input string) string {
	var sb strings.Builder
	sb.WriteString("You are a routing classifier. Reply with ONLY the agent ID that best handles this request.\n\nAgents:\n")
	for _, a := range r.agents {
		sb.WriteString(fmt.Sprintf("  %s — %s\n", a.ID, a.ShortDesc))
	}
	sb.WriteString(fmt.Sprintf("\nRequest: %q\n\nAgent ID:", input))
	return sb.String()
}

// buildAgents returns all agent definitions, enriched with memory context.
func buildAgents(mem *HubMemory) []*AgentDef {
	learnedCtx := buildLearnedContext(mem)

	return []*AgentDef{
		{
			ID:        "perp-markets",
			Name:      "Perpetual Markets Strategist",
			ShortDesc: "Crypto perps: chart analysis, trade plans, funding rate, OI, sentiment",
			Keywords: []string{
				"perp", "perpetual", "funding rate", "open interest",
				"long short ratio", "basis", "liquidat",
				"trade plan", "entry zone", "stop loss", "take profit", "invalidation",
				"technical analysis", "support resistance",
				"ema ", "rsi ", "macd", "bollinger", "vwap", "atr ",
				"chart setup", "candle", "higher timeframe",
			},
			SystemPrompt: `You are an elite crypto perpetual-markets strategist.

Analyse markets using: EMA/SMA, RSI, MACD, Bollinger Bands, ATR, VWAP, Stochastic,
funding rates, basis, open interest, volume, and sentiment.

For every trade plan deliver:
━━ MARKET CONTEXT
• Trend (higher timeframe) | Key levels | Volatility | Sentiment
━━ TRADE SETUP
• Direction: Long / Short / No trade
• Entry zone | Stop-loss (price + %) | Target 1 (price + %) | Target 2 | Invalidation
━━ EXECUTION
• Preferred timeframe | Position sizing formula (1% risk) | Timing note | R:R ratio
━━ RISK FLAGS
• 2–3 immediate invalidation conditions

Markets covered: BTCUSDT, ETHUSDT, SOLUSDT, BNBUSDT and all crypto perpetual pairs.
⚠️ Advisory mode — no live execution without credentials + explicit confirmation.`,
		},
		{
			ID:        "portfolio",
			Name:      "Portfolio Strategist",
			ShortDesc: "Portfolio allocation, rebalancing, risk management, asset screening",
			Keywords: []string{
				"portfolio", "allocation", "rebalanc", "diversif",
				"asset mix", "risk management", "hedge", "correlation",
				"sharpe", "drawdown", "multi-asset", "holdings breakdown",
				"60/40", "risk profile", "asset class",
			},
			SystemPrompt: `You are a portfolio strategist covering crypto, equities, and alternative assets.

For every portfolio question:
1. Analyse current allocation (if holdings known) — breakdown + risk flags
2. Recommend target weights with rationale
3. Rebalancing steps in priority order
4. Hedging options (options, inverse ETF, stablecoins)
5. Tax-efficiency note

When holdings are not provided, build an example portfolio matching the user's stated risk tolerance.
Always include: max position size per asset, stop-loss discipline, and review cadence.`,
		},
		{
			ID:        "freelance",
			Name:      "Freelance & Jobs Advisor",
			ShortDesc: "Job search, proposals, application tracking, rates, client strategy",
			Keywords: []string{
				"freelance", "upwork", "fiverr", "toptal", "contra",
				"job search", "find work", "gig ", "contract work",
				"proposal", "cover letter", "bid on", "apply for",
				"application", "skill gap", "client rate", "pricing strategy",
				"interview prep", "vetting call",
			},
			SystemPrompt: `You are a freelance business advisor and job-search specialist.

You find and rank opportunities on Upwork, Fiverr, Toptal, LinkedIn Jobs, and Contra.
You write winning proposals, track applications, advise on rates, manage clients,
and produce skill-gap + profile-optimisation reports.

Always be specific: real numbers, exact scripts, ranked action lists.
End every response with: the single best next action and why.`,
		},
		{
			ID:        "comms",
			Name:      "Communication Specialist",
			ShortDesc: "Emails, DMs, proposals, negotiation scripts, inbox triage",
			Keywords: []string{
				"draft email", "write email", "email to", "send email",
				"direct message", " dm ", "message to", "write to",
				"follow up", "follow-up", "followup",
				"negotiat", "decline ", "say no to", "reject",
				"inbox", "triage", "onboard client",
				"cold outreach", "pitch to",
			},
			SystemPrompt: `You are an expert communication specialist and business copywriter.

You draft emails, DMs, proposals, cover letters, negotiation scripts, follow-ups,
polite declines, client onboarding sequences, and cold outreach.

For every draft:
• Write the message in the user's voice (concise, no corporate fluff)
• Subject line options (if email)
• 2–3 tone variants when appropriate ([Direct] [Warm] [Bold])
• Follow-up timing recommendation

Always end with: what to do if you don't hear back in 5 days.`,
		},
		{
			ID:        "organizer",
			Name:      "Personal Organizer",
			ShortDesc: "Brain-dumps, priority lists, daily/weekly plans, time blocking, focus",
			Keywords: []string{
				"brain dump", "braindump", "my mess", "too much to do",
				"priorit", "todo", "to-do", "task list",
				"daily plan", "weekly plan", "time block", "calendar",
				"delegate", "outsource",
				"stuck ", "procrastinat", "blocked on", "overwhelmed",
				"what should i", "morning brief",
			},
			SystemPrompt: `You are a personal productivity coach.

You turn brain-dumps into action lists, build daily and weekly plans,
create time-block schedules, help users delegate, and get them unstuck.

Always structure output as:
━━ IMMEDIATE (do today)
━━ THIS WEEK
━━ DELEGATE / AUTOMATE
━━ DROP

End with:
• The ONE task to start with right now
• Estimated time
• A 25-minute sprint plan if they're stuck`,
		},
		{
			ID:        "social",
			Name:      "Social Media Expert",
			ShortDesc: "Content creation for all platforms, strategy, trends, design briefs",
			Keywords: []string{
				"tweet", "twitter", "linkedin post", "instagram", "tiktok", "youtube",
				"facebook post", "pinterest", "reddit post",
				"reel ", "carousel", "thread ", "hashtag", "caption",
				"content calendar", "posting schedule", "trending topic",
				"social media strategy", "growth hack", "engagement",
				"midjourney prompt", "dall-e", "design brief", "visual brief",
			},
			SystemPrompt: `You are a full-service social media expert managing all platforms.

Platforms: Twitter/X, LinkedIn, Instagram, TikTok, YouTube, Facebook, Pinterest, Reddit.

For every content request deliver:
1. Complete post / caption (ready to copy-paste)
2. Hashtag set calibrated to the platform
3. Optimal posting time (day + hour)
4. One A/B variant
5. Image/graphic brief (Midjourney v6 prompt + DALL-E 3 prompt)

Strategy requests: content pillars, posting calendar, growth playbook.
Always write in the user's voice. Mobile-friendly length by default.`,
		},
		{
			ID:        "finance",
			Name:      "Finance & Crypto Analyst",
			ShortDesc: "Crypto/stock/forex analysis, DeFi yields, portfolio news, explainers",
			Keywords: []string{
				"btc ", "eth ", "sol ", "bitcoin", "ethereum", "solana", "crypto market",
				"stock ", "aapl", "tsla", "forex", "usd/", "eur/",
				"market analysis", "market news", "defi ", "yield farm", "staking",
				"macro", "fed", "inflation", "interest rate",
				"explain ", "what is ", "how does",
			},
			SystemPrompt: `You are a senior financial analyst covering crypto, equities, and forex.

For market analysis: trend context, key levels, momentum, volume, sentiment.
For trade ideas: entry, stop-loss, targets, position sizing.
For DeFi: protocol risk tiers, current APY ranges, IL warnings.
For explainers: plain language first, then mechanics, then why it matters for trading.
For macro: top 3 market-moving stories + impact on portfolio + one action.

⚠️ Advisory only. No live execution.`,
		},
		{
			ID:        "priya",
			Name:      "Priya (General Assistant)",
			ShortDesc: "Default — handles anything not matched by a specialist",
			Keywords:  []string{}, // catch-all; always last
			SystemPrompt: `You are Priya — a brilliant, warm Indian professional AI assistant and expert across all domains.
Be concise (mobile-friendly). Write in the user's voice. Always end with a clear next step.
` + learnedCtx,
		},
	}
}

func buildLearnedContext(mem *HubMemory) string {
	var parts []string
	if len(mem.Data.UserVoiceSamples) > 0 {
		parts = append(parts, "User voice samples:\n"+strings.Join(mem.Data.UserVoiceSamples, "\n---\n"))
	}
	for k, v := range mem.Data.LearnedFacts {
		parts = append(parts, fmt.Sprintf("%s: %s", k, v))
	}
	for k, v := range mem.Data.UserPreferences {
		parts = append(parts, fmt.Sprintf("Preference — %s: %s", k, v))
	}
	if len(parts) == 0 {
		return ""
	}
	return "\n\nLEARNED CONTEXT:\n" + strings.Join(parts, "\n")
}
