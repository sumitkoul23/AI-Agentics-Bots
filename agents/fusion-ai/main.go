package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static/index.html
var indexHTML []byte

//go:embed static/manifest.json
var manifestJSON []byte

//go:embed static/sw.js
var swJS []byte

//go:embed static/icon.svg
var iconSVG []byte

const systemPrompt = `You are FusionAI — a highly capable assistant combining the best strengths of multiple AI models:
- Deep reasoning, nuanced writing, and long-context understanding (like Claude)
- Broad world knowledge, multimodal awareness, and speed (like Gemini)
- Expert code generation, debugging, and software engineering (like OpenAI Codex/GPT)

You excel at all of these:
• Code generation, debugging, and refactoring in any language
• Complex analysis, research summaries, and logical reasoning
• Creative writing, storytelling, and ideation
• Mathematics, science, and technical explanations
• General knowledge and conversational assistance

Always be thorough and practical. For code, include working examples. For analysis, be structured. For creative tasks, be imaginative.`

// FusionAI orchestrates all AI models
type FusionAI struct {
	gemini *GeminiClient
	groq   *GroqClient
	claude *ClaudeClient
	openai *OpenAIClient
	ollama *OllamaClient
}

type chatRequest struct {
	Message string `json:"message"`
	Model   string `json:"model"` // "auto", "gemini", "groq", "claude", "openai", "ollama"
}

type chatResponse struct {
	Reply     string `json:"reply"`
	ModelID   string `json:"model_id"`
	ModelName string `json:"model_name"`
}

type modelInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Free      bool   `json:"free"`
	Note      string `json:"note"`
	Color     string `json:"color"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ai := &FusionAI{}

	log.Println("FusionAI — initialising models...")

	// Gemini 2.0 Flash (free tier: 15 req/min, ~1M tokens/day)
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		ai.gemini = NewGeminiClient(key)
		log.Println("  [✓] Gemini 2.0 Flash    — FREE tier (aistudio.google.com)")
	} else {
		log.Println("  [○] Gemini              — set GEMINI_API_KEY (free at aistudio.google.com)")
	}

	// Groq — fast free Llama 3.3 70B
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		ai.groq = NewGroqClient(key)
		log.Println("  [✓] Groq Llama 3.3 70B  — FREE tier (console.groq.com)")
	} else {
		log.Println("  [○] Groq                — set GROQ_API_KEY (free at console.groq.com)")
	}

	// Claude (paid, optional — best reasoning)
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		ai.claude = NewClaudeClient(key)
		log.Println("  [✓] Claude Sonnet       — paid API active")
	}

	// OpenAI GPT-4o mini (paid, optional — strong at code)
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		ai.openai = NewOpenAIClient(key)
		log.Println("  [✓] OpenAI GPT-4o mini  — paid API active")
	}

	// Ollama (local, completely free)
	oc := NewOllamaClient()
	if oc.IsAvailable() {
		oc.AutoModel()
		ai.ollama = oc
		log.Printf("  [✓] Ollama (%s)  — local FREE model", oc.Model)
	} else {
		log.Println("  [○] Ollama              — install from ollama.ai for 100% local AI")
	}

	if !ai.hasAnyModel() {
		log.Println()
		log.Println("  ⚠  No models configured!")
		log.Println("     → Get a free Gemini key: https://aistudio.google.com/app/apikey")
		log.Println("     → Get a free Groq key:   https://console.groq.com")
		log.Println("     → Copy .env.example → .env and fill in at least one key")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 128*1024))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var req chatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			req.Message = strings.TrimSpace(string(body))
		}
		input := strings.TrimSpace(req.Message)
		w.Header().Set("Content-Type", "application/json")

		if input == "" {
			json.NewEncoder(w).Encode(chatResponse{
				Reply: greeting(ai), ModelID: "system", ModelName: "FusionAI",
			})
			return
		}

		reply, mid, mname := ai.Chat(r.Context(), input, req.Model)
		json.NewEncoder(w).Encode(chatResponse{Reply: reply, ModelID: mid, ModelName: mname})
	})

	mux.HandleFunc("/models", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ai.ModelList())
	})

	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		w.Write(manifestJSON)
	})
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(swJS)
	})
	mux.HandleFunc("/icon.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(iconSVG)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(indexHTML)
	})

	log.Printf("FusionAI running → http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// ── Routing ───────────────────────────────────────────────────────────────────

type queryType int

const (
	queryGeneral  queryType = iota
	queryCode               // code, debugging, programming
	queryAnalysis           // deep analysis, research, comparison
	queryCreative           // writing, stories, essays
	queryMath               // math, calculations, equations
)

func classifyQuery(input string) queryType {
	lower := strings.ToLower(input)
	for _, kw := range []string{
		"code", "function", "debug", "error", "bug", "script", "program",
		"python", "golang", " go ", "javascript", "typescript", "java", "rust",
		"c++", "html", "css", "sql", "regex", "algorithm", "implement", "refactor",
		"compile", "syntax", "class", "method", "loop", "array", "api endpoint",
		"dockerfile", "kubernetes", "terraform", "bash script", "shell script",
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
		"breakdown", "deep dive", "detailed analysis",
	} {
		if strings.Contains(lower, kw) {
			return queryAnalysis
		}
	}
	for _, kw := range []string{
		"write a story", "write a poem", "creative writing", "fiction",
		"blog post", "essay", "narrative", "screenplay",
	} {
		if strings.Contains(lower, kw) {
			return queryCreative
		}
	}
	return queryGeneral
}

// modelPriority returns the ordered list of model IDs to try for a given query type.
// Free models (gemini, groq, ollama) are preferred first.
func modelPriority(qt queryType) []string {
	switch qt {
	case queryCode:
		// Gemini Flash is strong at code + free; OpenAI GPT-4o also excellent; Ollama CodeLlama local
		return []string{"gemini", "openai", "groq", "ollama", "claude"}
	case queryAnalysis, queryCreative:
		// Claude is best for nuanced reasoning; Gemini is strong + free
		return []string{"claude", "gemini", "groq", "openai", "ollama"}
	case queryMath:
		// Gemini is strong at math and it's free; Claude and OpenAI also good
		return []string{"gemini", "claude", "openai", "groq", "ollama"}
	default:
		// General: Gemini first (free), then Groq (free), then paid
		return []string{"gemini", "groq", "claude", "openai", "ollama"}
	}
}

// Chat selects the best model and returns a response
func (f *FusionAI) Chat(ctx context.Context, input, prefer string) (reply, modelID, modelName string) {
	qt := classifyQuery(input)

	if prefer != "" && prefer != "auto" {
		r, id, name := f.invoke(ctx, prefer, input)
		if id != "" {
			return r, id, name
		}
		// Requested model unavailable — fall through to auto
	}

	for _, mid := range modelPriority(qt) {
		r, id, name := f.invoke(ctx, mid, input)
		if id != "" {
			return r, id, name
		}
	}
	return "⚠ No AI models configured. Add GEMINI_API_KEY (free) or GROQ_API_KEY (free) to your .env file to get started. See .env.example for instructions.", "none", "None"
}

// invoke calls a specific model by ID
func (f *FusionAI) invoke(ctx context.Context, modelID, input string) (reply, id, name string) {
	switch modelID {
	case "gemini":
		if f.gemini == nil {
			return "", "", ""
		}
		r, err := f.gemini.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("gemini error: %v", err)
			return "", "", ""
		}
		return r, "gemini", "Gemini 2.0 Flash"

	case "groq":
		if f.groq == nil {
			return "", "", ""
		}
		r, err := f.groq.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("groq error: %v", err)
			return "", "", ""
		}
		return r, "groq", "Llama 3.3 70B (Groq)"

	case "claude":
		if f.claude == nil {
			return "", "", ""
		}
		r, err := f.claude.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("claude error: %v", err)
			return "", "", ""
		}
		return r, "claude", "Claude Sonnet"

	case "openai":
		if f.openai == nil {
			return "", "", ""
		}
		r, err := f.openai.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("openai error: %v", err)
			return "", "", ""
		}
		return r, "openai", "GPT-4o mini"

	case "ollama":
		if f.ollama == nil {
			return "", "", ""
		}
		r, err := f.ollama.Generate(ctx, systemPrompt, input)
		if err != nil {
			log.Printf("ollama error: %v", err)
			return "", "", ""
		}
		return r, "ollama", fmt.Sprintf("Ollama (%s)", f.ollama.Model)
	}
	return "", "", ""
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (f *FusionAI) hasAnyModel() bool {
	return f.gemini != nil || f.groq != nil || f.claude != nil || f.openai != nil || f.ollama != nil
}

func (f *FusionAI) ModelList() []modelInfo {
	ollamaName := "Ollama (local)"
	if f.ollama != nil {
		ollamaName = fmt.Sprintf("Ollama (%s)", f.ollama.Model)
	}
	return []modelInfo{
		{ID: "gemini", Name: "Gemini 2.0 Flash", Available: f.gemini != nil, Free: true,
			Note: "Google — best at code & math — FREE tier", Color: "#4285f4"},
		{ID: "groq", Name: "Llama 3.3 70B (Groq)", Available: f.groq != nil, Free: true,
			Note: "Fast open-source inference — FREE tier", Color: "#f97316"},
		{ID: "claude", Name: "Claude Sonnet", Available: f.claude != nil, Free: false,
			Note: "Anthropic — best reasoning & writing", Color: "#f59e0b"},
		{ID: "openai", Name: "GPT-4o mini", Available: f.openai != nil, Free: false,
			Note: "OpenAI Codex successor — strong at code", Color: "#10b981"},
		{ID: "ollama", Name: ollamaName, Available: f.ollama != nil, Free: true,
			Note: "Local model — 100% private & free", Color: "#6366f1"},
	}
}

func greeting(f *FusionAI) string {
	var active, inactive []string
	check := func(available bool, name string) {
		if available {
			active = append(active, name)
		} else {
			inactive = append(inactive, name)
		}
	}
	check(f.gemini != nil, "Gemini 2.0 Flash (free)")
	check(f.groq != nil, "Groq Llama 3.3 (free)")
	check(f.claude != nil, "Claude Sonnet")
	check(f.openai != nil, "GPT-4o mini")
	check(f.ollama != nil, fmt.Sprintf("Ollama %s (local)", func() string {
		if f.ollama != nil {
			return f.ollama.Model
		}
		return ""
	}()))

	activeStr := "none configured"
	if len(active) > 0 {
		activeStr = strings.Join(active, ", ")
	}

	msg := fmt.Sprintf("Welcome to **FusionAI** — combining Claude, Gemini, and Codex capabilities in one interface.\n\n**Active models:** %s\n\nSmart routing automatically picks the best model for your query:\n- **Code & Debugging** → Gemini Flash or GPT-4o (Codex-style)\n- **Analysis & Writing** → Claude or Gemini\n- **Fast Q&A** → Groq (Llama 3.3)\n- **Private/Offline** → Ollama (local)\n\nOr select a model manually from the dropdown. Just ask anything!", activeStr)

	if len(active) == 0 {
		msg += "\n\n⚠ **No models active.** Copy `.env.example` → `.env` and add at least one API key.\n- Free: `GEMINI_API_KEY` (aistudio.google.com) or `GROQ_API_KEY` (console.groq.com)"
	} else if len(inactive) > 0 {
		msg += fmt.Sprintf("\n\n💡 *Unlock more models: %s — see `.env.example`*", strings.Join(inactive, ", "))
	}
	return msg
}
