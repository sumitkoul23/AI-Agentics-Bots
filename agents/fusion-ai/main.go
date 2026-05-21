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

// ── Teneo handler ─────────────────────────────────────────────────────────────

// FusionAIHandler implements the Teneo agent handler interface.
// Each incoming task is routed to the best available AI model.
type FusionAIHandler struct {
	ai *FusionAI
}

// ProcessTask classifies the query and dispatches it to the best AI model.
func (h *FusionAIHandler) ProcessTask(ctx context.Context, task string) (string, error) {
	input := strings.TrimSpace(task)
	lower := strings.ToLower(input)

	// Command dispatch
	switch {
	case input == "" || lower == "/help":
		return helpText(h.ai), nil

	case lower == "/models":
		return modelStatus(h.ai), nil

	case strings.HasPrefix(lower, "/model "):
		// /model <id> <message>
		rest := strings.TrimSpace(input[7:])
		parts := strings.SplitN(rest, " ", 2)
		if len(parts) < 2 {
			return "Usage: /model gemini|groq|claude|openai|ollama <your message>", nil
		}
		modelID, msg := parts[0], parts[1]
		reply, _, name := h.ai.Chat(ctx, msg, modelID)
		return fmt.Sprintf("**[%s]**\n\n%s", name, reply), nil

	case strings.HasPrefix(lower, "/code "):
		return h.dispatch(ctx, input[6:], "code"), nil

	case strings.HasPrefix(lower, "/analyze "):
		return h.dispatch(ctx, input[9:], "analysis"), nil

	case strings.HasPrefix(lower, "/write "):
		return h.dispatch(ctx, input[7:], "creative"), nil

	case strings.HasPrefix(lower, "/math "):
		return h.dispatch(ctx, input[6:], "math"), nil
	}

	// Auto-route
	reply, _, modelName := h.ai.Chat(ctx, input, "")
	return fmt.Sprintf("**[%s]**\n\n%s", modelName, reply), nil
}

func (h *FusionAIHandler) dispatch(ctx context.Context, input, hint string) string {
	// Map hint to a forced query type by prefixing context
	prefixes := map[string]string{
		"code":     "Write code for: ",
		"analysis": "Analyze in detail: ",
		"creative": "Write creatively: ",
		"math":     "Solve this math problem: ",
	}
	msg := input
	if p, ok := prefixes[hint]; ok && !strings.Contains(strings.ToLower(input), hint) {
		msg = p + input
	}
	reply, _, modelName := h.ai.Chat(ctx, msg, "")
	return fmt.Sprintf("**[%s]**\n\n%s", modelName, reply)
}

// ── Metadata ─────────────────────────────────────────────────────────────────

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

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	// Load metadata
	raw, err := os.ReadFile("fusion-ai-metadata.json")
	if err != nil {
		log.Fatal("Cannot read fusion-ai-metadata.json: ", err)
	}
	var meta Metadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		log.Fatal("Cannot parse metadata: ", err)
	}

	// Initialise AI models
	ai := initModels()

	// Deploy on-chain via Teneo
	capJSON, _ := json.Marshal(meta.Capabilities)
	catJSON, _ := json.Marshal(meta.Categories)

	result, err := deploy.DeployAgent(deploy.DeployConfig{
		PrivateKey:       os.Getenv("PRIVATE_KEY"),
		AgentID:          meta.AgentID,
		AgentName:        meta.Name,
		Description:      meta.Description,
		AgentType:        meta.AgentType,
		Capabilities:     capJSON,
		Commands:         meta.Commands,
		NlpFallback:      meta.NLPFallback,
		Categories:       catJSON,
		ShortDescription: meta.ShortDescription,
		FAQItems:         meta.FAQItems,
		MetadataVersion:  "2.4.0",
		StateFilePath:    ".teneo-deploy-state.json",
	})
	if err != nil {
		log.Fatal("Deploy failed: ", err)
	}
	log.Printf("FusionAI on-chain — token_id=%d", result.TokenID)

	// Configure and start the Teneo agent
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
		AgentHandler:    &FusionAIHandler{ai: ai},
		AgentID:         meta.AgentID,
		TokenID:         result.TokenID,
		SubmitForReview: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("FusionAI live — models: %s", activeModelList(ai))
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// initModels reads API keys from env and initialises each model client.
func initModels() *FusionAI {
	ai := &FusionAI{}

	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		ai.gemini = NewGeminiClient(key)
		log.Println("  [✓] Gemini 2.0 Flash    — FREE tier")
	} else {
		log.Println("  [○] Gemini              — set GEMINI_API_KEY (free: aistudio.google.com)")
	}

	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		ai.groq = NewGroqClient(key)
		log.Println("  [✓] Groq Llama 3.3 70B  — FREE tier")
	} else {
		log.Println("  [○] Groq                — set GROQ_API_KEY (free: console.groq.com)")
	}

	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		ai.claude = NewClaudeClient(key)
		log.Println("  [✓] Claude Sonnet       — paid API")
	}

	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		ai.openai = NewOpenAIClient(key)
		log.Println("  [✓] OpenAI GPT-4o mini  — paid API")
	}

	oc := NewOllamaClient()
	if oc.IsAvailable() {
		oc.AutoModel()
		ai.ollama = oc
		log.Printf("  [✓] Ollama (%s)  — local free model", oc.Model)
	} else {
		log.Println("  [○] Ollama              — install from ollama.ai for local AI")
	}

	if !ai.hasAnyModel() {
		log.Println("  ⚠  No AI models configured — add GEMINI_API_KEY or GROQ_API_KEY to .env")
	}
	return ai
}

// ── Routing ───────────────────────────────────────────────────────────────────

const systemPrompt = `You are FusionAI — a highly capable assistant combining the strengths of multiple AI models:
- Expert code generation, debugging, and software engineering (like OpenAI Codex/GPT)
- Deep reasoning, nuanced writing, and analysis (like Claude)
- Broad knowledge, mathematics, and multimodal understanding (like Gemini)

For code: provide working examples with clear explanations.
For analysis: be structured, thorough, and cite reasoning.
For writing: be creative, clear, and audience-aware.
For math: show working steps and verify results.`

type queryType int

const (
	queryGeneral  queryType = iota
	queryCode
	queryAnalysis
	queryCreative
	queryMath
)

func classifyQuery(input string) queryType {
	lower := strings.ToLower(input)
	for _, kw := range []string{
		"code", "function", "debug", "error", "bug", "script", "program",
		"python", "golang", " go ", "javascript", "typescript", "java", "rust",
		"c++", "html", "css", "sql", "regex", "algorithm", "implement", "refactor",
		"compile", "syntax", "class", "method", "loop", "array", "api endpoint",
		"dockerfile", "kubernetes", "terraform", "bash script", "shell script",
		"write code", "fix this", "why is this failing",
	} {
		if strings.Contains(lower, kw) {
			return queryCode
		}
	}
	for _, kw := range []string{
		"calculate", " math", "equation", "formula", "solve", "derivative",
		"integral", "probability", "statistics", "compute", "proof",
	} {
		if strings.Contains(lower, kw) {
			return queryMath
		}
	}
	for _, kw := range []string{
		"analyze", "analyse", "explain why", "how does", "compare", "difference between",
		"research", "summarize", "summarise", "review", "evaluate", "pros and cons",
		"breakdown", "deep dive", "detailed analysis", "what are the implications",
	} {
		if strings.Contains(lower, kw) {
			return queryAnalysis
		}
	}
	for _, kw := range []string{
		"write a story", "write a poem", "creative writing", "fiction",
		"blog post", "essay", "narrative", "screenplay", "write me a",
	} {
		if strings.Contains(lower, kw) {
			return queryCreative
		}
	}
	return queryGeneral
}

// modelPriority returns ordered model IDs for a given query type (free models first).
func modelPriority(qt queryType) []string {
	switch qt {
	case queryCode:
		return []string{"gemini", "openai", "groq", "ollama", "claude"}
	case queryAnalysis, queryCreative:
		return []string{"claude", "gemini", "groq", "openai", "ollama"}
	case queryMath:
		return []string{"gemini", "claude", "openai", "groq", "ollama"}
	default:
		return []string{"gemini", "groq", "claude", "openai", "ollama"}
	}
}

// Chat routes a message to the best available model.
func (f *FusionAI) Chat(ctx context.Context, input, prefer string) (reply, modelID, modelName string) {
	qt := classifyQuery(input)

	if prefer != "" && prefer != "auto" {
		r, id, name := f.invoke(ctx, prefer, input)
		if id != "" {
			return r, id, name
		}
	}

	for _, mid := range modelPriority(qt) {
		r, id, name := f.invoke(ctx, mid, input)
		if id != "" {
			return r, id, name
		}
	}
	return "⚠ No AI models configured. Add GEMINI_API_KEY (free) or GROQ_API_KEY (free) to your .env file.", "none", "None"
}

// invoke calls a specific model by ID.
func (f *FusionAI) invoke(ctx context.Context, modelID, input string) (reply, id, name string) {
	switch modelID {
	case "gemini":
		if f.gemini == nil {
			return "", "", ""
		}
		r, err := f.gemini.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("gemini: %v", err)
			return "", "", ""
		}
		return r, "gemini", "Gemini 2.0 Flash"

	case "groq":
		if f.groq == nil {
			return "", "", ""
		}
		r, err := f.groq.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("groq: %v", err)
			return "", "", ""
		}
		return r, "groq", "Llama 3.3 70B (Groq)"

	case "claude":
		if f.claude == nil {
			return "", "", ""
		}
		r, err := f.claude.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("claude: %v", err)
			return "", "", ""
		}
		return r, "claude", "Claude Sonnet"

	case "openai":
		if f.openai == nil {
			return "", "", ""
		}
		r, err := f.openai.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("openai: %v", err)
			return "", "", ""
		}
		return r, "openai", "GPT-4o mini"

	case "ollama":
		if f.ollama == nil {
			return "", "", ""
		}
		r, err := f.ollama.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("ollama: %v", err)
			return "", "", ""
		}
		return r, "ollama", fmt.Sprintf("Ollama (%s)", f.ollama.Model)
	}
	return "", "", ""
}

func (f *FusionAI) hasAnyModel() bool {
	return f.gemini != nil || f.groq != nil || f.claude != nil || f.openai != nil || f.ollama != nil
}

// ── Display helpers ───────────────────────────────────────────────────────────

func activeModelList(f *FusionAI) string {
	var parts []string
	if f.gemini != nil {
		parts = append(parts, "Gemini")
	}
	if f.groq != nil {
		parts = append(parts, "Groq")
	}
	if f.claude != nil {
		parts = append(parts, "Claude")
	}
	if f.openai != nil {
		parts = append(parts, "GPT-4o")
	}
	if f.ollama != nil {
		parts = append(parts, "Ollama("+f.ollama.Model+")")
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func modelStatus(f *FusionAI) string {
	status := func(ok bool) string {
		if ok {
			return "✓ active"
		}
		return "○ not configured"
	}
	ollamaModel := ""
	if f.ollama != nil {
		ollamaModel = " (" + f.ollama.Model + ")"
	}
	return fmt.Sprintf(`FusionAI — Model Status

FREE models:
  Gemini 2.0 Flash    %s
  Groq Llama 3.3 70B  %s
  Ollama (local)%s %s

PAID models (optional):
  Claude Sonnet       %s
  OpenAI GPT-4o mini  %s

Use /model <id> <message> to force a specific model.
IDs: gemini  groq  claude  openai  ollama`,
		status(f.gemini != nil),
		status(f.groq != nil),
		ollamaModel,
		status(f.ollama != nil),
		status(f.claude != nil),
		status(f.openai != nil),
	)
}

func helpText(f *FusionAI) string {
	return fmt.Sprintf(`FusionAI — Multi-Model Intelligence

Active models: %s

COMMANDS
  /model <id> <msg>   Force a model (gemini|groq|claude|openai|ollama)
  /models             Show all model statuses
  /code <task>        Code mode  → routes to Gemini/GPT
  /analyze <topic>    Analysis   → routes to Claude/Gemini
  /write <request>    Creative   → routes to Claude/Gemini
  /math <problem>     Math mode  → routes to Gemini/Claude
  /help               Show this message

AUTO ROUTING
  Code & debugging    → Gemini Flash → GPT-4o mini
  Analysis & writing  → Claude Sonnet → Gemini
  Mathematics         → Gemini Flash → Claude
  General / fast      → Gemini Flash → Groq Llama

Just type naturally — I route automatically.`, activeModelList(f))
}
