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

// ProcessTask is the main entry point called by the Teneo SDK for every message.
func (p *PriyaBot) ProcessTask(ctx context.Context, task string) (string, error) {
	input := strings.TrimSpace(task)
	if input == "" {
		return p.greeting(), nil
	}

	lower := strings.ToLower(input)

	// Learn user voice if flagged
	if strings.HasPrefix(lower, "/learn voice") {
		sample := extractArgs(input, "/learn voice")
		p.mem.AddVoiceSample(sample)
		_ = p.mem.Save()
		return "Got it, Priya will write in your voice from now on. Keep sharing samples any time.", nil
	}

	// Set preference
	if strings.HasPrefix(lower, "/set ") {
		parts := strings.SplitN(extractArgs(input, "/set"), "=", 2)
		if len(parts) == 2 {
			p.mem.SetPreference(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			_ = p.mem.Save()
			return "Saved. I'll remember that.", nil
		}
	}

	// Help
	if strings.HasPrefix(lower, "/help") {
		return p.helpText(), nil
	}

	// Route by domain
	switch {
	// Social media
	case strings.HasPrefix(lower, "/social") ||
		containsAny(lower, "tweet", "linkedin post", "instagram", "tiktok", "youtube video", "facebook post", "pinterest", "reel", "carousel", "content calendar", "hashtag", "caption"):
		return p.social.Handle(ctx, input)

	// Trading & finance
	case strings.HasPrefix(lower, "/trade") || strings.HasPrefix(lower, "/finance") ||
		containsAny(lower, "btc", "eth", "sol", "bitcoin", "crypto", "stock", "forex", "trade plan", "portfolio", "defi", "yield", "market analysis"):
		return p.finance.Handle(ctx, input)

	// Communication
	case strings.HasPrefix(lower, "/comms") || strings.HasPrefix(lower, "/email") || strings.HasPrefix(lower, "/dm") ||
		containsAny(lower, "draft email", "write email", "send message", "follow up", "follow-up", "direct message", "inbox", "triage", "negotiat", "decline offer"):
		return p.comms.Handle(ctx, input)

	// Organizer
	case strings.HasPrefix(lower, "/organise") || strings.HasPrefix(lower, "/organize") || strings.HasPrefix(lower, "/plan") ||
		containsAny(lower, "brain dump", "my mess", "prioritise", "todo list", "daily plan", "weekly plan", "calendar block", "stuck", "procrastinat", "morning brief"):
		return p.organizer.Handle(ctx, input)

	// Freelance & jobs
	case strings.HasPrefix(lower, "/jobs") || strings.HasPrefix(lower, "/apply") || strings.HasPrefix(lower, "/track") || strings.HasPrefix(lower, "/skills") ||
		containsAny(lower, "freelance", "upwork", "fiverr", "toptal", "proposal", "cover letter", "job search", "skill gap", "client rate", "interview prep"):
		return p.freelance.Handle(ctx, input)

	default:
		// Pure NLP — let Priya decide and respond intelligently
		return p.ai.Think(ctx, input)
	}
}

func (p *PriyaBot) greeting() string {
	return `Namaste! I'm Priya 🌸

Your autonomous AI expert — social media, trading, comms, organizing & freelance.

What can I do for you today?

  📣 /social <platform> <topic>  — content creation for any platform
  📈 /trade <symbol>             — trade plan + market analysis
  💬 /email or /dm               — draft any message or communication
  🗂  /organize or /plan         — clear the chaos, set priorities
  💼 /jobs <keywords>            — find freelance opportunities
  ✍️  /apply <job title>         — draft a winning proposal
  📊 /skills                     — skill-gap analysis
  🌅 /plan daily                 — morning briefing
  ❓  /help                       — full command list

Or just talk to me naturally — I'll figure out what you need.`
}

func (p *PriyaBot) helpText() string {
	return `Priya — Full Command Reference

━━━ SOCIAL MEDIA ━━━
/social twitter <topic>       Draft a tweet or thread
/social linkedin <topic>      Draft a LinkedIn post
/social instagram <topic>     Caption + reel script
/social tiktok <topic>        TikTok hook + script
/social youtube <topic>       Title, description, tags
/social facebook <topic>      Facebook post
/social all <topic>           Full cross-platform content pack
/social trends                Trending topics in your niche
/social calendar              7-day content calendar
/social strategy              Growth strategy + content pillars
/social image <brief>         Visual design brief (Midjourney/DALL-E)

━━━ TRADING & FINANCE ━━━
/trade <symbol>               Full trade plan (e.g. /trade BTC long)
/finance portfolio            Portfolio analysis + recommendations
/finance defi                 DeFi yield opportunities
/finance news                 Macro + crypto market briefing
/finance explain <concept>    Plain-language explainer
/finance screen <criteria>    Find investment opportunities

━━━ COMMUNICATION ━━━
/email <context>              Draft a professional email
/dm <context>                 Draft a direct message (LinkedIn, Twitter)
/comms followup <context>     Write a follow-up message
/comms negotiate <context>    Negotiation scripts
/comms decline <context>      Polite decline message
/comms onboard <context>      Client onboarding sequence
/comms inbox                  Inbox triage + clean-up plan

━━━ ORGANIZER ━━━
/organize <brain dump>        Turn chaos into action list
/plan daily                   Morning briefing
/plan weekly                  Weekly work plan
/plan calendar <details>      Time-block schedule
/organize delegate <tasks>    What to outsource + briefing template
/organize stuck <situation>   Help me get unstuck

━━━ FREELANCE & JOBS ━━━
/jobs <keywords>              Search freelance opportunities
/apply <job title>            Draft a winning proposal (auto-tracked)
/track                        View application tracker
/skills                       Skill-gap report + profile tips
/freelance rate               Pricing strategy
/freelance client             Client management scripts
/freelance niche              Niche strategy

━━━ SETTINGS ━━━
/set niche=<your niche>       e.g. /set niche=AI development
/set skills=<your skills>     e.g. /set skills=Go, Python, React
/set portfolio_holdings=<...> e.g. /set portfolio_holdings=BTC 0.5, ETH 2
/learn voice <sample text>    Teach Priya your writing style

Type anything naturally — Priya understands plain language too.`
}

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

	// Initialize persistent memory
	mem := NewMemory(".priya-memory.json")

	// Initialize AI core (Claude)
	ai, err := NewAICore(mem)
	if err != nil {
		log.Printf("Warning: AI core unavailable (%v) — running in template mode", err)
	}

	// Build modules
	var priyaBot *PriyaBot
	if ai != nil {
		social := NewSocialModule(ai, mem)
		finance := NewFinanceModule(ai, mem)
		comms := NewCommsModule(ai, mem)
		organizer := NewOrganizerModule(ai, mem)
		freelance := NewFreelanceModule(ai, mem)
		scheduler := NewScheduler(social, finance, freelance, organizer, mem)

		priyaBot = &PriyaBot{
			ai:        ai,
			social:    social,
			finance:   finance,
			comms:     comms,
			organizer: organizer,
			freelance: freelance,
			scheduler: scheduler,
			mem:       mem,
		}

		// Start autonomous background tasks
		scheduler.Start()
		defer scheduler.Stop()
	} else {
		priyaBot = &PriyaBot{mem: mem}
	}

	// Deploy to Teneo network
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
	log.Printf("Priya is online — token_id=%d", result.TokenID)

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

// extractArgs strips the command prefix and returns remaining text.
func extractArgs(input, cmd string) string {
	idx := strings.Index(strings.ToLower(input), strings.ToLower(cmd))
	if idx == -1 {
		return strings.TrimSpace(input)
	}
	return strings.TrimSpace(input[idx+len(cmd):])
}

// containsAny checks if s contains any of the keywords.
func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
