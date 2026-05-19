package main

import (
	"fmt"
	"strings"
)

type Agent struct {
	ID       string
	Name     string
	Desc     string
	Keywords []string
	Handle   func(input string, mem *Memory) string
}

type Registry struct {
	list  []*Agent
	index map[string]*Agent
}

func NewRegistry(mem *Memory) *Registry {
	r := &Registry{index: make(map[string]*Agent)}
	for _, a := range agents(mem) {
		r.list = append(r.list, a)
		r.index[a.ID] = a
	}
	return r
}

func (r *Registry) Route(lower string) *Agent {
	for _, a := range r.list {
		for _, kw := range a.Keywords {
			if strings.Contains(lower, kw) {
				return a
			}
		}
	}
	return r.index["priya"] // default
}

func (r *Registry) Get(id string) *Agent { return r.index[id] }

func (r *Registry) Catalog() string {
	var sb strings.Builder
	sb.WriteString("All agents available in this chat:\n\n")
	for _, a := range r.list {
		sb.WriteString(fmt.Sprintf("  %-30s %s\n", a.Name, a.Desc))
	}
	sb.WriteString("\nJust talk naturally — I route automatically.\nOr: /use <id> <message>")
	return sb.String()
}

// ── Agent definitions ─────────────────────────────────────────────────────────

func agents(mem *Memory) []*Agent {
	return []*Agent{
		perpMarketsAgent(),
		portfolioAgent(),
		socialAgent(mem),
		commsAgent(mem),
		organizerAgent(),
		financeAgent(),
		freelanceAgent(mem),
		priyaAgent(),
	}
}

func perpMarketsAgent() *Agent {
	return &Agent{
		ID:   "perp-markets",
		Name: "Perpetual Markets Strategist",
		Desc: "Crypto perps: chart analysis, trade plans, funding, OI, sentiment",
		Keywords: []string{
			"perp", "perpetual", "funding rate", "open interest", "liquidat",
			"trade plan", "entry zone", "stop loss", "take profit", "invalidation",
			"technical analysis", "support resistance", "bollinger", "vwap", "atr",
			"chart setup", "higher timeframe", "long short ratio",
		},
		Handle: handlePerpMarketsAgent,
	}
}

func portfolioAgent() *Agent {
	return &Agent{
		ID:   "portfolio",
		Name: "Portfolio Strategist",
		Desc: "Portfolio allocation, rebalancing, risk management, asset screening",
		Keywords: []string{
			"portfolio", "allocation", "rebalanc", "diversif",
			"asset mix", "risk management", "hedge", "sharpe",
			"drawdown", "multi-asset", "holdings breakdown",
		},
		Handle: handlePortfolio,
	}
}

func socialAgent(mem *Memory) *Agent {
	return &Agent{
		ID:   "social",
		Name: "Social Media Expert",
		Desc: "Content for all platforms: Twitter, LinkedIn, Instagram, TikTok, YouTube…",
		Keywords: []string{
			"tweet", "twitter", "linkedin post", "instagram", "tiktok", "youtube",
			"facebook post", "pinterest", "reel ", "carousel", "thread ",
			"hashtag", "caption", "content calendar", "social media", "viral hook",
			"posting schedule", "trending topic", "growth hack", "design brief",
			"midjourney", "dall-e",
		},
		Handle: func(input string, m *Memory) string { return handleSocial(input, m) },
	}
}

func commsAgent(mem *Memory) *Agent {
	return &Agent{
		ID:   "comms",
		Name: "Communication Specialist",
		Desc: "Emails, DMs, proposals, negotiation, inbox triage",
		Keywords: []string{
			"draft email", "write email", "email to", "send email",
			"direct message", " dm ", "message to", "write to",
			"follow up", "followup", "follow-up",
			"negotiat", "decline ", "say no", "reject",
			"inbox", "triage", "onboard client", "cold outreach",
		},
		Handle: func(input string, m *Memory) string { return handleComms(input, m) },
	}
}

func organizerAgent() *Agent {
	return &Agent{
		ID:   "organizer",
		Name: "Personal Organizer",
		Desc: "Brain-dumps, priority lists, daily/weekly plans, time blocking, focus",
		Keywords: []string{
			"brain dump", "braindump", "my mess", "too much to do",
			"priorit", "todo", "to-do", "task list",
			"daily plan", "weekly plan", "time block",
			"delegate", "outsource",
			"stuck ", "procrastinat", "blocked on", "overwhelmed",
			"what should i", "morning brief",
		},
		Handle: handleOrganizer,
	}
}

func financeAgent() *Agent {
	return &Agent{
		ID:   "finance",
		Name: "Finance & Crypto Analyst",
		Desc: "Crypto/stock/forex analysis, DeFi yields, macro news, explainers",
		Keywords: []string{
			"btc ", "eth ", "sol ", "bitcoin", "ethereum", "solana",
			"crypto market", "stock ", "forex", "defi ", "yield farm",
			"macro", "inflation", "interest rate",
			"market analysis", "market news", "what is ", "explain ",
		},
		Handle: handleFinance,
	}
}

func freelanceAgent(mem *Memory) *Agent {
	return &Agent{
		ID:   "freelance",
		Name: "Freelance & Jobs Advisor",
		Desc: "Job search, proposals, application tracking, rates, client strategy",
		Keywords: []string{
			"freelance", "upwork", "fiverr", "toptal", "contra",
			"job search", "find work", "gig ", "contract work",
			"proposal", "cover letter", "bid on", "apply for",
			"application", "skill gap", "client rate", "pricing strategy",
			"interview prep",
		},
		Handle: func(input string, m *Memory) string { return handleFreelance(input, m) },
	}
}

func priyaAgent() *Agent {
	return &Agent{
		ID:       "priya",
		Name:     "Priya (General)",
		Desc:     "Default — warm, knowledgeable assistant for any topic",
		Keywords: []string{},
		Handle:   handlePriya,
	}
}
