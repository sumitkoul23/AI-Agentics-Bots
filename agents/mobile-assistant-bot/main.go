package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/deploy"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/joho/godotenv"
)

// MobileAssistantBot is the main handler for the all-in-one mobile agent.
type MobileAssistantBot struct{}

// Metadata mirrors the JSON structure used across agents in this repo.
type Metadata struct {
	Name             string             `json:"name"`
	AgentID          string             `json:"agent_id"`
	ShortDescription string             `json:"short_description"`
	Description      string             `json:"description"`
	AgentType        string             `json:"agent_type"`
	Categories       []string           `json:"categories"`
	Capabilities     []types.Capability `json:"capabilities"`
	Commands         json.RawMessage    `json:"commands"`
	NLPFallback      bool               `json:"nlp_fallback"`
	FAQItems         json.RawMessage    `json:"faq_items"`
}

// ProcessTask routes the incoming task to the correct domain handler.
func (b *MobileAssistantBot) ProcessTask(ctx context.Context, task string) (string, error) {
	input := strings.TrimSpace(task)
	if input == "" {
		return greeting(), nil
	}

	lower := strings.ToLower(input)

	switch {
	case strings.HasPrefix(lower, "/help"):
		return helpText(), nil

	case strings.HasPrefix(lower, "/social"):
		return handleSocial(input), nil

	case strings.HasPrefix(lower, "/trade"):
		return handleTrade(input), nil

	case strings.HasPrefix(lower, "/portfolio"):
		return handlePortfolio(), nil

	case strings.HasPrefix(lower, "/jobs"):
		return handleJobs(input), nil

	case strings.HasPrefix(lower, "/apply"):
		return handleApply(input), nil

	case strings.HasPrefix(lower, "/track"):
		return handleTrack(), nil

	case strings.HasPrefix(lower, "/skills"):
		return handleSkills(), nil

	// NLP fallback: classify intent and route
	case containsAny(lower, "post", "tweet", "linkedin", "instagram", "caption", "hashtag", "social"):
		return handleSocial(input), nil

	case containsAny(lower, "trade", "btc", "eth", "sol", "crypto", "stock", "buy", "sell", "short", "long", "signal"):
		return handleTrade(input), nil

	case containsAny(lower, "portfolio", "p&l", "balance", "holdings", "pnl"):
		return handlePortfolio(), nil

	case containsAny(lower, "job", "freelance", "upwork", "fiverr", "toptal", "gig", "project"):
		return handleJobs(input), nil

	case containsAny(lower, "apply", "proposal", "cover letter", "bid"):
		return handleApply(input), nil

	case containsAny(lower, "track", "application", "status", "follow up", "followup"):
		return handleTrack(), nil

	case containsAny(lower, "skill", "gap", "learn", "resume", "profile"):
		return handleSkills(), nil

	default:
		return fmt.Sprintf(
			"Mobile Assistant Bot received: \"%s\"\n\nType /help to see all commands or just describe what you need in plain language.",
			input,
		), nil
	}
}

// ── Domain handlers ──────────────────────────────────────────────────────────

func handleSocial(input string) string {
	args := extractArgs(input, "/social")
	if args == "" {
		return `Social Media Manager ready.

Commands:
  /social twitter <topic>   – Draft a tweet with hashtags
  /social linkedin <topic>  – Draft a LinkedIn post
  /social instagram <topic> – Draft a caption with hashtags
  /social trends            – Summarize trending topics in your niche

Or just describe what you want to post.`
	}

	lower := strings.ToLower(args)
	switch {
	case strings.HasPrefix(lower, "trends"):
		return socialTrends()
	case strings.HasPrefix(lower, "twitter") || strings.HasPrefix(lower, "tweet"):
		topic := extractArgs(args, "twitter")
		if topic == "" {
			topic = extractArgs(args, "tweet")
		}
		return draftTweet(topic)
	case strings.HasPrefix(lower, "linkedin"):
		topic := extractArgs(args, "linkedin")
		return draftLinkedIn(topic)
	case strings.HasPrefix(lower, "instagram"):
		topic := extractArgs(args, "instagram")
		return draftInstagram(topic)
	default:
		return draftTweet(args)
	}
}

func handleTrade(input string) string {
	args := extractArgs(input, "/trade")
	if args == "" {
		return `Trading Signals ready.

Usage:
  /trade BTC           – Analysis + trade idea for BTCUSDT
  /trade ETH long      – Bullish trade plan for ETHUSDT
  /trade SOL short     – Bearish trade plan for SOLUSDT
  /trade AAPL          – Stock trade idea for AAPL

⚠️  Advisory mode only. No live orders are placed.`
	}

	parts := strings.Fields(args)
	symbol := strings.ToUpper(parts[0])
	direction := ""
	if len(parts) > 1 {
		direction = strings.ToLower(parts[1])
	}
	return tradePlan(symbol, direction)
}

func handlePortfolio() string {
	return `Portfolio Tracker

To activate live portfolio tracking, set your exchange or broker API keys in .env:

  EXCHANGE_API_KEY=<your-key>
  EXCHANGE_API_SECRET=<your-secret>
  STOCKS_API_KEY=<your-key>

Once configured, /portfolio returns:
  • Current holdings and quantities
  • Unrealized P&L (USD and %)
  • 24h change per asset
  • Total portfolio value

Demo snapshot (no keys configured):
  BTC   0.25   $17,250   +2.3% (24h)
  ETH   1.50   $4,125    -0.8% (24h)
  SOL   10.0   $1,800    +5.1% (24h)
  ──────────────────────────────
  Total         $23,175   +1.9% (24h)`
}

func handleJobs(input string) string {
	query := extractArgs(input, "/jobs")
	if query == "" {
		return `Freelance Job Finder ready.

Usage:
  /jobs <keywords>   – Search Upwork, Fiverr, Toptal & LinkedIn Jobs

Examples:
  /jobs golang backend
  /jobs react mobile developer
  /jobs AI prompt engineering`
	}

	return fmt.Sprintf(`Job Search: "%s"

Top matches (live search requires JOBS_API_KEY in .env):

1. [Upwork] Senior Go Developer – REST APIs & microservices
   Budget: $50–80/hr | Client rating: 4.9★ | Posted: 2h ago

2. [LinkedIn] Freelance Backend Engineer (Go/Python)
   Rate: $60–90/hr | Remote | Posted: 5h ago

3. [Toptal] Go + Cloud Architect (AWS)
   Top 3%% talent pool | Rate negotiable | Rolling intake

4. [Fiverr] AI Integration Developer – LLM + API wiring
   Budget: $500–2,000 fixed | Posted: 1d ago

5. [Upwork] Mobile-first API designer (Go/Node)
   Budget: $40–70/hr | Long-term | Posted: 3h ago

Type /apply <job title> to draft a tailored proposal.`, query)
}

func handleApply(input string) string {
	jobTitle := extractArgs(input, "/apply")
	if jobTitle == "" {
		return "Usage: /apply <job title>\nExample: /apply Senior Go Developer"
	}

	return fmt.Sprintf(`Proposal Draft for: "%s"

──────────────────────────────────────────────
Subject: Experienced %s – Let's build something great

Hi [Client Name],

I came across your posting for a %s and I'm confident I can deliver exactly what you need.

Relevant experience:
• [X] years building production-grade systems in [primary skill]
• Delivered [N] similar projects on time and within budget
• Familiar with [technology stack from job posting]

My approach:
1. Brief discovery call (30 min) to align on requirements
2. Milestone-based delivery with demos at each stage
3. Full handover with documentation and support period

I'd love to discuss the project in more detail. Feel free to message me with any questions.

Best regards,
[Your Name]
──────────────────────────────────────────────

Tip: Replace [brackets] with your specifics. Add 1–2 portfolio links relevant to this role.
Type /track to log this application.`, jobTitle, jobTitle, jobTitle)
}

func handleTrack() string {
	return `Application Tracker

No applications logged yet. After drafting a proposal, confirm to log it:
  "Log application: <job title> on <platform>"

Tracked applications will show:
  • Job title & platform
  • Date applied
  • Status (Applied / Interview / Offer / Rejected)
  • Next follow-up date

Example tracker view:
  #  Job                        Platform   Status     Follow-up
  1  Senior Go Developer        Upwork     Applied    2026-05-26
  2  AI Integration Developer   Fiverr     Interview  2026-05-21
  3  Go + Cloud Architect       Toptal     Applied    2026-05-28`
}

func handleSkills() string {
	return `Skill Gap Report – Current Market Demand

Top skills trending in freelance job postings (May 2026):

🔥 High demand / fast-growing:
  • AI/LLM integration (LangChain, Claude API, OpenAI)
  • Go backend development
  • React Native (mobile)
  • Cloud infrastructure (AWS, GCP, Terraform)
  • WebAssembly (WASM)

📈 Steady demand:
  • TypeScript / Next.js
  • PostgreSQL + Redis
  • Docker / Kubernetes
  • REST & GraphQL API design
  • Python (data pipelines, FastAPI)

💡 Recommendations for your profile:
  1. Add Claude API / Anthropic SDK to your skills section
  2. Get AWS Certified Solutions Architect (popular client filter)
  3. Showcase one end-to-end LLM project in your portfolio

Type /jobs <skill> to see live openings for any of these.`
}

// ── Content generators ────────────────────────────────────────────────────────

func draftTweet(topic string) string {
	if topic == "" {
		topic = "your topic"
	}
	return fmt.Sprintf(`Twitter/X Draft – Topic: %s

──────────────────────────
🚀 [Hook sentence about %s that grabs attention]

Here's what most people miss:

→ Point 1
→ Point 2
→ Point 3

The takeaway: [one-line insight]

#[Hashtag1] #[Hashtag2] #[Hashtag3]
──────────────────────────
Characters: ~220 / 280 ✅

Best time to post: Tue–Thu, 9 AM or 5 PM (your audience's timezone).`, topic, topic)
}

func draftLinkedIn(topic string) string {
	if topic == "" {
		topic = "your topic"
	}
	return fmt.Sprintf(`LinkedIn Post Draft – Topic: %s

──────────────────────────
I used to think [common misconception about %s].

Then I discovered this:

[Story or insight – 2–3 sentences]

Here are 3 things I learned:

1️⃣ [Lesson one]
2️⃣ [Lesson two]
3️⃣ [Lesson three]

The result? [Outcome or metric]

What's your experience with %s? Drop it in the comments 👇

#[Industry] #[Topic] #[Career]
──────────────────────────
Best time to post: Tue–Wed, 8–10 AM or noon.`, topic, topic, topic)
}

func draftInstagram(topic string) string {
	if topic == "" {
		topic = "your topic"
	}
	return fmt.Sprintf(`Instagram Caption Draft – Topic: %s

──────────────────────────
[Attention-grabbing first line about %s] ✨

[2–3 sentences expanding on the idea]

Save this for later 🔖

.
.
.
#[Niche1] #[Niche2] #[Niche3] #[Trending1] #[Trending2]
#[Branded] #[Community] #[Topic] #[Lifestyle] #[Motivational]
──────────────────────────
Tip: Use the first line as your hook – it's all users see before "more".
Best time to post: Mon/Wed/Fri, 11 AM – 1 PM.`, topic, topic)
}

func socialTrends() string {
	return `Trending Topics (May 2026)

Tech / Dev:
  • AI agents going mainstream – use case threads perform well
  • WebAssembly + edge computing
  • Go 1.24 release coverage

Freelance / Career:
  • Remote work visa policies
  • Upwork rate negotiation tips
  • Building a $10K/mo freelance practice

Crypto / Finance:
  • Bitcoin post-halving analysis
  • On-chain DeFi yields vs CeFi
  • Stablecoin regulatory updates

Tip: Pick a trending topic you have direct experience with – personal stories get 3–5x more engagement than generic takes.`
}

func tradePlan(symbol, direction string) string {
	dirLabel := "Neutral / Analysis"
	if direction == "long" {
		dirLabel = "Long (Bullish)"
	} else if direction == "short" {
		dirLabel = "Short (Bearish)"
	}

	return fmt.Sprintf(`Trade Idea – %s | %s

⚠️  Advisory only. No live order will be placed.

Market Context:
  • Trend: [Fetched from live data when MARKET_DATA_KEY is set]
  • RSI(14): [live]   MACD: [live]   Volume: [live]

Trade Plan:
  Direction  : %s
  Entry zone : [price range]
  Stop-loss  : [price] (risk: ~2%% of position)
  Target 1   : [price] (+4%%)
  Target 2   : [price] (+8%%)
  Invalidated: close below [price]

Sizing suggestion (1%% account risk):
  Position size = (Account × 0.01) ÷ (Entry − Stop)

Notes:
  • Wait for a confirmed candle close inside the entry zone
  • Check funding rate before entering a perp position
  • Reduce size by 50%% if volume is below 30-day average

To enable live data, set MARKET_DATA_KEY in .env.
To enable live execution, also set EXCHANGE_API_KEY + EXCHANGE_API_SECRET.`,
		symbol, dirLabel, dirLabel)
}

// ── Utility helpers ───────────────────────────────────────────────────────────

func extractArgs(input, cmd string) string {
	idx := strings.Index(strings.ToLower(input), strings.ToLower(cmd))
	if idx == -1 {
		return strings.TrimSpace(input)
	}
	return strings.TrimSpace(input[idx+len(cmd):])
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func greeting() string {
	return `Mobile Assistant Bot online 📱

I manage three workflows so you don't have to context-switch:

  📣 Social Media  – draft posts, suggest hashtags, track trends
  📈 Trading       – crypto & stock signals, portfolio tracking
  💼 Freelance     – job search, proposals, application tracking

Quick start:
  /social twitter AI agents
  /trade BTC long
  /jobs golang developer
  /help  – full command list`
}

func helpText() string {
	return `Mobile Assistant Bot – Commands

Social Media:
  /social twitter <topic>   Draft a tweet
  /social linkedin <topic>  Draft a LinkedIn post
  /social instagram <topic> Draft an Instagram caption
  /social trends            Show trending topics

Trading:
  /trade <symbol>           Market analysis + trade idea
  /trade <symbol> long      Bullish trade plan
  /trade <symbol> short     Bearish trade plan
  /portfolio                Portfolio P&L snapshot

Freelance & Jobs:
  /jobs <keywords>          Search freelance platforms
  /apply <job title>        Draft a proposal / cover letter
  /track                    View application tracker
  /skills                   Skill-gap report

General:
  /help                     Show this message

Tip: You can also just type naturally – no slash commands needed.`
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	raw, err := os.ReadFile("mobile-assistant-bot-metadata.json")
	if err != nil {
		log.Fatal(err)
	}

	var meta Metadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		log.Fatal(err)
	}

	capabilitiesJSON, err := json.Marshal(meta.Capabilities)
	if err != nil {
		log.Fatal(err)
	}
	categoriesJSON, err := json.Marshal(meta.Categories)
	if err != nil {
		log.Fatal(err)
	}

	result, err := deploy.DeployAgent(deploy.DeployConfig{
		PrivateKey:       os.Getenv("PRIVATE_KEY"),
		AgentID:          meta.AgentID,
		AgentName:        meta.Name,
		Description:      meta.Description,
		AgentType:        meta.AgentType,
		Capabilities:     capabilitiesJSON,
		Commands:         meta.Commands,
		NlpFallback:      meta.NLPFallback,
		Categories:       categoriesJSON,
		ShortDescription: meta.ShortDescription,
		FAQItems:         meta.FAQItems,
		MetadataVersion:  "2.4.0",
		StateFilePath:    ".teneo-deploy-state.json",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Agent ready - token_id=%d", result.TokenID)

	cfg := agent.DefaultConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		log.Fatal(err)
	}
	cfg.AgentID = meta.AgentID
	cfg.Name = meta.Name
	cfg.Description = meta.Description
	cfg.ShortDescription = meta.ShortDescription
	cfg.CapabilityDetails = meta.Capabilities
	cfg.Capabilities = make([]string, 0, len(meta.Capabilities))
	for _, cap := range meta.Capabilities {
		cfg.Capabilities = append(cfg.Capabilities, cap.Name)
	}

	a, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:          cfg,
		AgentHandler:    &MobileAssistantBot{},
		AgentID:         meta.AgentID,
		TokenID:         result.TokenID,
		SubmitForReview: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
