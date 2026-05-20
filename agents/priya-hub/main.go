package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed static/index.html
var indexHTML []byte

//go:embed static/manifest.json
var manifestJSON []byte

//go:embed static/sw.js
var swJS []byte

//go:embed static/icon.svg
var iconSVG []byte

//go:embed static/icon-maskable.svg
var iconMaskableSVG []byte

//go:embed static/icon-192.png
var icon192PNG []byte

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mem := NewMemory(".bodhi-memory.json")
	registry := NewRegistry(mem)
	swarm := NewSwarm(registry, mem)
	router := NewRouter(registry, swarm)
	mesh := NewMesh(port, mem, swarm)

	swarm.Start()
	mesh.Start()
	defer swarm.Stop()
	defer mesh.Stop()

	log.Printf("Bodhi Hub — %d agents | port %s | mesh discovery active", len(swarm.agents), port)

	if os.Getenv("CLI") == "1" {
		runCLI(router, registry, mem, swarm, mesh)
		return
	}

	mux := http.NewServeMux()
	registerRoutes(mux, router, registry, mem, swarm, mesh)

	log.Printf("UI:        http://localhost:%s", port)
	log.Printf("API:       POST /chat  GET /status /agents /memory /mesh /peers")
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// ── HTTP handlers ─────────────────────────────────────────────────────────────

func registerRoutes(mux *http.ServeMux, router *Router, registry *Registry, mem *Memory, swarm *Swarm, mesh ...*Mesh) {
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
			var m *Mesh
			if len(mesh) > 0 {
				m = mesh[0]
			}
			reply = handleInput(input, router, registry, mem, swarm, m)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"reply": reply})
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		s := swarm.Status()
		if len(mesh) > 0 && mesh[0] != nil {
			s += "\n" + mesh[0].Status()
		}
		fmt.Fprint(w, s)
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

	if len(mesh) > 0 && mesh[0] != nil {
		m := mesh[0]
		mux.HandleFunc("/mesh", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprint(w, m.Status())
		})
		mux.HandleFunc("/peers", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m.Peers())
		})
	}

	// SSE — real-time push notifications from autonomous agents
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
		flusher.Flush()

		subID, ch := swarm.Notifs.Subscribe()
		defer swarm.Notifs.Unsubscribe(subID)

		tick := time.NewTicker(25 * time.Second)
		defer tick.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case n := <-ch:
				payload, _ := json.Marshal(map[string]string{
					"type": "notification",
					"from": n.From,
					"text": n.Text,
					"at":   n.At.Format(time.RFC3339),
				})
				fmt.Fprintf(w, "data: %s\n\n", payload)
				flusher.Flush()
			case <-tick.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	})

	mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(manifestJSON)
	})
	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(swJS)
	})
	mux.HandleFunc("/icon.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(iconSVG)
	})
	mux.HandleFunc("/icon-maskable.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(iconMaskableSVG)
	})
	mux.HandleFunc("/icon-192.png", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(icon192PNG)
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
}

// ── CLI mode ──────────────────────────────────────────────────────────────────

func runCLI(router *Router, registry *Registry, mem *Memory, swarm *Swarm, mesh *Mesh) {
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
		reply := handleInput(input, router, registry, mem, swarm, mesh)
		fmt.Printf("\nBodhi: %s\n\n", reply)
	}
}

// ── Input dispatcher ──────────────────────────────────────────────────────────

func handleInput(input string, router *Router, registry *Registry, mem *Memory, swarm *Swarm, mesh *Mesh) string {
	lower := strings.ToLower(input)

	switch {
	case lower == "/help":
		return helpText()

	case lower == "/agents" || lower == "/list":
		return registry.Catalog()

	case lower == "/status" || lower == "/swarm":
		s := swarm.Status()
		if mesh != nil {
			s += "\n" + mesh.Status()
		}
		return s

	case lower == "/mesh" || lower == "/peers":
		if mesh != nil {
			return mesh.Status()
		}
		return "Mesh not initialised."

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
	return fmt.Sprintf(`Namaste! I'm Bodhi — your autonomous self-learning AI swarm 🌸

AI Engine : %s
Agents    : %d specialist agents active
Learning  : Every conversation makes me smarter

Just talk naturally — I'll route to the right specialist automatically.
Type /agents to see all specialists, /status for swarm health, /help for commands.`,
		engine, len(swarm.agents))
}

func helpText() string {
	return `Bodhi Swarm — Command Reference

━━ ROUTING ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
(automatic — just talk naturally)

/agents               List all specialist agents
/use <id> <message>   Force a specific agent
                        IDs: auto, bodhi,
                             perp-markets, portfolio, finance,
                             social, comms, organizer, freelance,
                             code, health, research, news

━━ SWARM ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/status               Show swarm health + Ollama status

━━ MEMORY & LEARNING ━━━━━━━━━━━━━━━━━━━━━━━━
/set key=value        Save a preference
/learn key=value      Teach a fact
/learn voice <text>   Feed your writing style

━━ QUICK EXAMPLES ━━━━━━━━━━━━━━━━━━━━━━━━━━━
"BTC trade plan long"            → perp-markets
"Rebalance my portfolio"         → portfolio
"Draft a LinkedIn post about AI" → social
"Write email to my client"       → comms
"Brain dump all my tasks"        → organizer
"Explain DeFi yields"            → finance
"Find Upwork gigs for Go dev"    → freelance
"Debug this nil pointer"         → code
"Create a workout plan"          → health
"Research best LLM frameworks"   → research
"What's moving crypto today"     → news

━━ MESH ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
/mesh                 Show connected device peers
/peers                (same)

━━ API ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
POST /chat   {"message": "your text", "agent": "optional-id"}
GET  /status  (swarm + mesh)
GET  /agents
GET  /memory
GET  /mesh
GET  /peers`
}
