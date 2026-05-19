package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/deploy"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/joho/godotenv"
)

// PriyaBot is the root handler that orchestrates all modules.
type PriyaBot struct {
	ai        *AICore
	social    *SocialModule
	finance   *FinanceModule
	comms     *CommsModule
	organizer *OrganizerModule
	freelance *FreelanceModule
	scheduler *Scheduler
	mem       *Memory
}

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

// ProcessTask is the entry point called by the Teneo SDK for every incoming message.
func (p *PriyaBot) ProcessTask(ctx context.Context, task string) (string, error) {
	input := strings.TrimSpace(task)
	if input == "" {
		return p.greeting(), nil
	}

	// If AI core is not available, return a setup message.
	if p.ai == nil {
		return "Priya is not fully initialised yet. Please set ANTHROPIC_API_KEY in your .env file and restart.", nil
	}

	lower := strings.ToLower(input)

	// ── Built-in commands that don't need module routing ──────────────────────

	if strings.HasPrefix(lower, "/help") {
		return p.helpText(), nil
	}

	if strings.HasPrefix(lower, "/learn voice") {
		sample := extractArgs(input, "/learn voice")
		if sample == "" {
			return "Usage: /learn voice <sample of your writing>", nil
		}
		p.mem.AddVoiceSample(sample)
		_ = p.mem.Save()
		return "Got it — I'll write in your voice from now on. The more samples you share, the more accurate I get.", nil
	}

	if strings.HasPrefix(lower, "/set ") {
		parts := strings.SplitN(extractArgs(input, "/set"), "=", 2)
		if len(parts) == 2 {
			p.mem.SetPreference(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			_ = p.mem.Save()
			return "Saved. I'll remember that.", nil
		}
		return "Usage: /set key=value  (e.g. /set niche=AI development)", nil
	}

	// ── Domain routing ────────────────────────────────────────────────────────

	switch {

	// Social media
	case strings.HasPrefix(lower, "/social") ||
		containsAny(lower,
			"tweet", "thread", "linkedin post", "instagram", "tiktok", "youtube",
			"facebook post", "pinterest", "reel", "carousel", "content calendar",
			"hashtag", "caption", "social media", "post about", "content for"):
		return p.social.Handle(ctx, input)

	// Trading & finance
	case strings.HasPrefix(lower, "/trade") || strings.HasPrefix(lower, "/finance") ||
		containsAny(lower,
			"btc", "eth", "sol", "bitcoin", "ethereum", "crypto", "stock", "forex",
			"trade plan", "trade idea", "portfolio", "defi", "yield", "market analysis",
			"price of", "market cap", "technical analysis", "chart", "rsi", "macd"):
		return p.finance.Handle(ctx, input)

	// Communication
	case strings.HasPrefix(lower, "/comms") || strings.HasPrefix(lower, "/email") || strings.HasPrefix(lower, "/dm") ||
		containsAny(lower,
			"draft email", "write email", "write a message", "send message",
			"follow up", "follow-up", "direct message", "inbox", "triage",
			"negotiat", "decline offer", "onboarding message", "write to client"):
		return p.comms.Handle(ctx, input)

	// Organizer / planner
	case strings.HasPrefix(lower, "/organise") || strings.HasPrefix(lower, "/organize") || strings.HasPrefix(lower, "/plan") ||
		containsAny(lower,
			"brain dump", "my mess", "prioritis", "todo list", "to-do", "daily plan",
			"weekly plan", "calendar block", "stuck", "procrastinat", "morning brief",
			"time block", "delegate", "what should i work on"):
		return p.organizer.Handle(ctx, input)

	// Freelance & jobs
	case strings.HasPrefix(lower, "/jobs") || strings.HasPrefix(lower, "/apply") ||
		strings.HasPrefix(lower, "/track") || strings.HasPrefix(lower, "/skills") ||
		strings.HasPrefix(lower, "/freelance") ||
		containsAny(lower,
			"freelance", "upwork", "fiverr", "toptal", "proposal", "cover letter",
			"job search", "skill gap", "client rate", "interview prep", "find work",
			"find a job", "gig", "contract work"):
		return p.freelance.Handle(ctx, input)

	// Pure NLP — let Priya think freely
	default:
		return p.ai.Think(ctx, input)
	}
}

// ── Greeting & help ───────────────────────────────────────────────────────────

func (p *PriyaBot) greeting() string {
	return `Namaste! I'm Priya.

Your autonomous AI — social media expert, finance analyst, copywriter, organizer, and freelance specialist. All in one.

Quick start:
  /social <platform> <topic>   Content for any platform
  /trade <symbol>              Trade plan + market analysis
  /email <context>             Draft any email or message
  /organize <brain dump>       Turn chaos into action
  /jobs <keywords>             Find freelance work
  /apply <job title>           Draft a winning proposal
  /plan daily                  Morning briefing
  /help                        Full command list

Or just talk to me — I understand plain language.`
}

func (p *PriyaBot) helpText() string {
	return `Priya — Full Command Reference

━━ SOCIAL MEDIA ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/social twitter <topic>       Tweet or thread
/social linkedin <topic>      LinkedIn post
/social instagram <topic>     Caption + reel script
/social tiktok <topic>        Hook + TikTok script
/social youtube <topic>       Title, description, SEO tags
/social facebook <topic>      Facebook post
/social all <topic>           Full cross-platform content pack
/social trends                Trending topics in your niche
/social calendar              7-day content calendar
/social strategy              Growth playbook + content pillars
/social image <brief>         Visual brief (Midjourney / DALL-E)
/social reply <comment>       Engagement reply options

━━ TRADING & FINANCE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/trade <symbol>               Full trade plan (e.g. /trade BTC long)
/trade <symbol> short         Bearish trade plan
/finance portfolio            Portfolio analysis + rebalancing
/finance defi                 DeFi yield opportunities
/finance news                 Macro + crypto market briefing
/finance explain <concept>    Plain-language financial explainer
/finance screen <criteria>    Find investment opportunities

━━ COMMUNICATION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/email <context>              Professional email draft
/dm <context>                 Direct message (LinkedIn, Twitter, Upwork)
/comms followup <context>     Follow-up message
/comms negotiate <context>    Negotiation scripts
/comms decline <context>      Polite decline
/comms onboard <context>      Client onboarding sequence
/comms inbox                  Inbox triage plan

━━ ORGANIZER ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/organize <brain dump>        Turn chaos into action list
/plan daily                   Morning briefing
/plan weekly                  Weekly work plan
/plan calendar <details>      Time-block schedule
/organize delegate <tasks>    What to outsource + brief template
/organize stuck <situation>   Get unstuck

━━ FREELANCE & JOBS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/jobs <keywords>              Find freelance opportunities
/apply <job title>            Draft proposal (auto-tracked)
/track                        Application tracker
/skills                       Skill-gap report
/freelance rate               Pricing strategy + rate scripts
/freelance client             Client management scripts
/freelance niche              Niche strategy

━━ SETTINGS & LEARNING ━━━━━━━━━━━━━━━━━━━━━━━━━━━
/set niche=AI development     Save your niche
/set skills=Go, Python        Save your skills
/set portfolio_holdings=...   Save your holdings
/learn voice <sample>         Teach Priya your writing style
/help                         This menu

Tip: Just talk naturally — no slash commands needed.`
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	raw, err := os.ReadFile("priya-bot-metadata.json")
	if err != nil {
		log.Fatal(err)
	}
	var meta Metadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		log.Fatal(err)
	}

	mem := NewMemory(".priya-memory.json")

	ai, err := NewAICore(mem)
	if err != nil {
		log.Printf("Warning: AI core unavailable (%v) — set ANTHROPIC_API_KEY to enable full intelligence", err)
	}

	var priyaBot *PriyaBot

	if ai != nil {
		social := NewSocialModule(ai, mem)
		finance := NewFinanceModule(ai, mem)
		comms := NewCommsModule(ai, mem)
		organizer := NewOrganizerModule(ai, mem)
		freelance := NewFreelanceModule(ai, mem)
		sched := NewScheduler(social, finance, freelance, organizer, mem)

		priyaBot = &PriyaBot{
			ai:        ai,
			social:    social,
			finance:   finance,
			comms:     comms,
			organizer: organizer,
			freelance: freelance,
			scheduler: sched,
			mem:       mem,
		}

		sched.Start()
		defer sched.Stop()
		log.Println("Priya AI core is online. All modules active.")
	} else {
		priyaBot = &PriyaBot{mem: mem}
		log.Println("Priya running without AI core — set ANTHROPIC_API_KEY for full capability.")
	}

	capabilitiesJSON, _ := json.Marshal(meta.Capabilities)
	categoriesJSON, _ := json.Marshal(meta.Categories)

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
	log.Printf("Priya deployed — token_id=%d", result.TokenID)

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
		AgentHandler:    priyaBot,
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

// ── Shared helpers (used across modules) ─────────────────────────────────────

// extractArgs strips the command prefix and returns the remaining text.
func extractArgs(input, cmd string) string {
	idx := strings.Index(strings.ToLower(input), strings.ToLower(cmd))
	if idx == -1 {
		return strings.TrimSpace(input)
	}
	return strings.TrimSpace(input[idx+len(cmd):])
}

// containsAny reports whether s contains any of the given keywords.
func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
