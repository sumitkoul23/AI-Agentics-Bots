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

// HubAgent is the single Teneo entry point that routes to all specialized agents.
type HubAgent struct {
	router   *Router
	registry *Registry
	mem      *HubMemory
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

// ProcessTask is called by the Teneo SDK for every incoming message.
func (h *HubAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	input := strings.TrimSpace(task)
	if input == "" {
		return h.greeting(), nil
	}

	lower := strings.ToLower(input)

	// ── Built-in hub commands ─────────────────────────────────────────────────

	switch {
	case lower == "/help":
		return helpText(), nil

	case lower == "/agents" || lower == "/list":
		return h.registry.Catalog(), nil

	case strings.HasPrefix(lower, "/learn voice "):
		sample := strings.TrimSpace(input[len("/learn voice "):])
		h.mem.AddVoiceSample(sample)
		_ = h.mem.Save()
		return "Got it — I'll write in your voice from now on.", nil

	case strings.HasPrefix(lower, "/set "):
		parts := strings.SplitN(strings.TrimSpace(input[len("/set "):]), "=", 2)
		if len(parts) == 2 {
			h.mem.SetPreference(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			_ = h.mem.Save()
			return "Saved.", nil
		}
		return "Usage: /set key=value", nil

	case strings.HasPrefix(lower, "/learn "):
		parts := strings.SplitN(strings.TrimSpace(input[len("/learn "):]), "=", 2)
		if len(parts) == 2 {
			h.mem.Learn(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			_ = h.mem.Save()
			return "Learned.", nil
		}
		return "Usage: /learn key=value", nil

	// Explicit agent routing — user can force a specific agent
	case strings.HasPrefix(lower, "/use "):
		rest := strings.TrimSpace(input[len("/use "):])
		parts := strings.SplitN(rest, " ", 2)
		if len(parts) < 2 {
			return fmt.Sprintf("Usage: /use <agent-id> <your message>\n\nAgents: %s",
				h.agentIDs()), nil
		}
		agentID := parts[0]
		if h.registry.Get(agentID) == nil {
			return fmt.Sprintf("Unknown agent %q. Available: %s", agentID, h.agentIDs()), nil
		}
		return h.router.dispatch(ctx, agentID, parts[1])
	}

	// ── Auto-route via keyword + Claude classifier ────────────────────────────
	return h.router.Route(ctx, input)
}

func (h *HubAgent) greeting() string {
	ready := "✅ All agents online."
	if !h.router.IsReady() {
		ready = "⚠️  Set ANTHROPIC_API_KEY for full AI capability (keyword routing active)."
	}
	return fmt.Sprintf(`Namaste! I'm Priya Hub 🌸

One chat. Every expert. I route your message to the right specialist automatically.

%s

Specialists available:
  📈 Perpetual Markets Strategist  — crypto perp analysis + trade plans
  💼 Portfolio Strategist          — allocation, rebalancing, risk
  📣 Social Media Expert           — content for all 7 platforms
  💬 Communication Specialist      — emails, DMs, proposals, negotiation
  🗂  Personal Organizer            — tasks, plans, focus, delegation
  💰 Finance & Crypto Analyst      — markets, DeFi, macro, explainers
  🔍 Freelance & Jobs Advisor      — job search, proposals, rates
  🌸 Priya (General)               — everything else

Commands:
  /agents                Show all agents
  /use <id> <message>    Force a specific agent
  /set key=value         Save a preference
  /learn voice <sample>  Teach me your writing style
  /help                  Full command reference

Or just talk naturally — I'll figure out the rest.`, ready)
}

func (h *HubAgent) agentIDs() string {
	ids := []string{"perp-markets", "portfolio", "social", "comms", "organizer", "finance", "freelance", "priya"}
	return strings.Join(ids, ", ")
}

func helpText() string {
	return `Priya Hub — Command Reference

━━ ROUTING ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
(automatic — just talk naturally)

/agents                   List all available agents
/use <id> <message>       Force a specific agent:
                            perp-markets, portfolio, social,
                            comms, organizer, finance, freelance, priya

━━ EXAMPLES (auto-routed) ━━━━━━━━━━━━━━━━━━━━━━
"BTC trade plan long"           → Perpetual Markets Strategist
"Rebalance my portfolio"        → Portfolio Strategist
"Draft a LinkedIn post about AI"→ Social Media Expert
"Write email to my client"      → Communication Specialist
"Brain dump: [your tasks]"      → Personal Organizer
"DeFi yields this week"         → Finance & Crypto Analyst
"Find Upwork jobs for Go dev"   → Freelance & Jobs Advisor
"Anything else"                 → Priya (General)

━━ MEMORY & SETTINGS ━━━━━━━━━━━━━━━━━━━━━━━━━━━
/set niche=AI development       Save your niche
/set skills=Go, Python          Save your skills
/set holdings=BTC 0.5, ETH 2   Save portfolio holdings
/learn voice <writing sample>   Teach Priya your voice
/learn key=value                Save any fact about you

━━ GENERAL ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/help                           This menu`
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	raw, err := os.ReadFile("hub-metadata.json")
	if err != nil {
		log.Fatal(err)
	}
	var meta Metadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		log.Fatal(err)
	}

	mem := NewHubMemory(".priya-hub-memory.json")
	registry := NewRegistry(mem)
	router := NewRouter(registry, mem)

	hub := &HubAgent{
		router:   router,
		registry: registry,
		mem:      mem,
	}

	log.Printf("Priya Hub starting — %d agents registered", len(registry.agents))

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
		StateFilePath:    ".teneo-deploy-state-hub.json",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Priya Hub deployed — token_id=%d", result.TokenID)

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
		AgentHandler:    hub,
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
