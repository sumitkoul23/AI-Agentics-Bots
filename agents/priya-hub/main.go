package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mem := NewMemory(".priya-hub-memory.json")
	registry := NewRegistry(mem)
	swarm := NewSwarm(registry, mem)
	router := NewRouter(registry, swarm)

	swarm.Start()
	defer swarm.Stop()

	log.Printf("Priya Hub — %d agents | port %s", len(swarm.agents), port)

	if os.Getenv("CLI") == "1" {
		runCLI(router, registry, mem, swarm)
		return
	}

	mux := http.NewServeMux()
	registerRoutes(mux, router, registry, mem, swarm)

	log.Printf("Endpoints: POST /chat  GET /status  GET /agents  GET /memory")
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// ── HTTP handlers ─────────────────────────────────────────────────────────────

func registerRoutes(mux *http.ServeMux, router *Router, registry *Registry, mem *Memory, swarm *Swarm) {
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var req struct {
			Message string `json:"message"`
			Agent   string `json:"agent"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			req.Message = strings.TrimSpace(string(body))
		}

		input := strings.TrimSpace(req.Message)
		if input == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"reply": hubGreeting(swarm)})
			return
		}

		var reply string
		if req.Agent != "" {
			var ok bool
			reply, ok = router.Dispatch(req.Agent, input)
			if !ok {
				reply = fmt.Sprintf("Unknown agent %q. GET /agents to see available agents.", req.Agent)
			}
		} else {
			reply = handleInput(input, router, registry, mem, swarm)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"reply": reply})
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, swarm.Status())
	})

	mux.HandleFunc("/agents", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, registry.Catalog())
	})

	mux.HandleFunc("/memory", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mem.mu.RLock()
		json.NewEncoder(w).Encode(mem.Data)
		mem.mu.RUnlock()
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, hubGreeting(swarm))
	})
}

// ── CLI mode ──────────────────────────────────────────────────────────────────

func runCLI(router *Router, registry *Registry, mem *Memory, swarm *Swarm) {
	fmt.Println(hubGreeting(swarm))
	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		reply := handleInput(input, router, registry, mem, swarm)
		fmt.Printf("\nPriya: %s\n\n", reply)
	}
}

// ── Input dispatcher ──────────────────────────────────────────────────────────

func handleInput(input string, router *Router, registry *Registry, mem *Memory, swarm *Swarm) string {
	lower := strings.ToLower(input)

	switch {
	case lower == "/help":
		return helpText()

	case lower == "/agents" || lower == "/list":
		return registry.Catalog()

	case lower == "/status" || lower == "/swarm":
		return swarm.Status()

	case strings.HasPrefix(lower, "/learn voice "):
		sample := strings.TrimSpace(input[len("/learn voice "):])
		mem.AddVoice(sample)
		mem.Save()
		return "Got it — I'll write in your voice from now on."

	case strings.HasPrefix(lower, "/set "):
		parts := strings.SplitN(strings.TrimSpace(input[len("/set "):]), "=", 2)
		if len(parts) == 2 {
			mem.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			mem.Save()
			return "Saved."
		}
		return "Usage: /set key=value"

	case strings.HasPrefix(lower, "/learn "):
		parts := strings.SplitN(strings.TrimSpace(input[len("/learn "):]), "=", 2)
		if len(parts) == 2 {
			mem.Learn(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			mem.Save()
			return "Learned."
		}
		return "Usage: /learn key=value"

	case strings.HasPrefix(lower, "/use "):
		rest := strings.TrimSpace(input[len("/use "):])
		parts := strings.SplitN(rest, " ", 2)
		if len(parts) < 2 {
			return "Usage: /use <agent-id> <your message>"
		}
		reply, ok := router.Dispatch(parts[0], parts[1])
		if !ok {
			return fmt.Sprintf("Unknown agent %q. Type /agents for the full list.", parts[0])
		}
		return reply
	}

	return router.Route(input)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func hubGreeting(swarm *Swarm) string {
	engine := "Template mode (start Ollama for on-device AI)"
	if swarm.ollama != nil {
		engine = fmt.Sprintf("Ollama — %s (100%% on-device, no API keys)", swarm.ollama.Model)
	}
	return fmt.Sprintf(`Namaste! I'm Priya — your autonomous self-learning AI swarm 🌸

AI Engine : %s
Agents    : %d specialist agents active
Learning  : Every conversation makes me smarter

Just talk naturally — I'll route to the right specialist automatically.
Type /agents to see all specialists, /status for swarm health, /help for commands.`,
		engine, len(swarm.agents))
}

func helpText() string {
	return `Priya Swarm — Command Reference

━━ ROUTING ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
(automatic — just talk naturally)

/agents               List all specialist agents
/use <id> <message>   Force a specific agent
                        IDs: perp-markets, portfolio, social,
                             comms, organizer, finance, freelance, priya

━━ SWARM ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/status               Show swarm health + Ollama status

━━ MEMORY & LEARNING ━━━━━━━━━━━━━━━━━━━━━━━━
/set key=value        Save a preference
/learn key=value      Teach a fact
/learn voice <text>   Feed your writing style

━━ QUICK EXAMPLES ━━━━━━━━━━━━━━━━━━━━━━━━━━━
"BTC trade plan long"            → Perp Markets
"Rebalance my portfolio"         → Portfolio
"Draft a LinkedIn post about AI" → Social Media
"Write email to my client"       → Comms
"Brain dump all my tasks"        → Organizer
"Explain DeFi yields"            → Finance
"Find Upwork gigs for Go dev"    → Freelance

━━ API ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
POST /chat   {"message": "your text", "agent": "optional-id"}
GET  /status
GET  /agents
GET  /memory`
}
