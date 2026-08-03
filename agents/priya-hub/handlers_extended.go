package main

// handlers_extended.go — extra HTTP endpoints for voice, vision, tunnel, MCP,
// reasoning, and capabilities. Registered via registerExtRoutes() called from
// main() after the standard registerRoutes() call.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func registerExtRoutes(mux *http.ServeMux, swarm *Swarm, tunnel *Tunnel) {
	// ── Voice ────────────────────────────────────────────────────────────────
	mux.HandleFunc("/transcribe", handleTranscribe)
	mux.HandleFunc("/speak", handleSpeak)
	mux.HandleFunc("/voice/status", handleVoiceStatus)

	// ── Vision ───────────────────────────────────────────────────────────────
	var ollama *OllamaClient
	if swarm != nil {
		ollama = swarm.ollama
	}
	mux.HandleFunc("/vision", func(w http.ResponseWriter, r *http.Request) {
		handleVision(w, r, ollama)
	})

	// ── Tunnel ───────────────────────────────────────────────────────────────
	mux.HandleFunc("/tunnel", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		url := ""
		provider := ""
		if tunnel != nil {
			url = tunnel.URL()
			provider = tunnel.Provider
		}
		json.NewEncoder(w).Encode(map[string]string{
			"url":      url,
			"provider": provider,
			"status":   tunnelStatus(tunnel),
		})
	})

	mux.HandleFunc("/tunnel/qr", func(w http.ResponseWriter, _ *http.Request) {
		url := ""
		if tunnel != nil {
			url = tunnel.URL()
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "no-cache")
		fmt.Fprint(w, TunnelQRSVG(url))
	})

	// ── MCP tools list ───────────────────────────────────────────────────────
	mux.HandleFunc("/mcp/tools", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		type toolInfo struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Enabled     bool   `json:"enabled"`
		}
		var list []toolInfo
		for _, t := range mcpTools {
			list = append(list, toolInfo{
				Name:        t.Name,
				Description: t.Description,
				Enabled:     t.Enabled(),
			})
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": list,
			"hint":  "Enable with env vars: MCP_SHELL=1 MCP_FILESYSTEM=1 MCP_FETCH=1",
		})
	})

	// ── Capabilities summary ─────────────────────────────────────────────────
	mux.HandleFunc("/capabilities", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		caps := MCPCapabilities()
		caps["reasoning_info"] = ReasoningInfo()
		if tunnel != nil {
			caps["tunnel_url"] = tunnel.URL()
		}
		json.NewEncoder(w).Encode(caps)
	})

	// ── Chat guide ───────────────────────────────────────────────────────────
	mux.HandleFunc("/guide", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, chatGuide(tunnel))
	})
}

func tunnelStatus(t *Tunnel) string {
	if t == nil {
		return "disabled"
	}
	if t.URL() == "" {
		return "starting"
	}
	return "active"
}

func chatGuide(tunnel *Tunnel) string {
	tunnelSection := "  Not configured (set TUNNEL=cloudflared to enable)"
	if tunnel != nil {
		url := tunnel.URL()
		if url == "" {
			url = "(starting...)"
		}
		tunnelSection = fmt.Sprintf("  Provider : %s\n  URL      : %s\n  QR code  : GET /tunnel/qr", tunnel.Provider, url)
	}

	return strings.Join([]string{
		`╔══════════════════════════════════════════════════════════════╗`,
		`║         BODHI — Complete Chat & Capability Guide            ║`,
		`╚══════════════════════════════════════════════════════════════╝`,
		``,
		`━━ 1. WEB CHAT (Browser) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  Open: http://localhost:8080`,
		`  • Type a message and press Enter`,
		`  • Click an agent in the left panel to target it directly`,
		`  • 🎤 Mic button — speak your question (browser STT or Whisper)`,
		`  • 📎 Attach an image to analyse it with the vision agent`,
		`  • 💭 Reasoning traces appear below answers when enabled`,
		``,
		`━━ 2. API (curl / script) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  # Chat (auto-route)`,
		`  curl -X POST http://localhost:8080/chat \`,
		`    -H 'Content-Type: application/json' \`,
		`    -d '{"message":"Explain quantum entanglement","agent":"auto"}'`,
		``,
		`  # Force a specific agent`,
		`  curl -X POST http://localhost:8080/chat \`,
		`    -d '{"message":"debug this nil pointer","agent":"code"}'`,
		``,
		`  # Speech → text (Whisper)`,
		`  curl -X POST http://localhost:8080/transcribe \`,
		`    -F 'audio=@recording.wav'`,
		``,
		`  # Text → speech (Piper)`,
		`  curl "http://localhost:8080/speak?text=Hello+Bodhi" -o reply.wav`,
		``,
		`  # Analyse an image`,
		`  curl -X POST http://localhost:8080/vision \`,
		`    -F 'image=@photo.jpg' -F 'prompt=What is in this image?'`,
		``,
		`  # Live SSE notifications`,
		`  curl -N http://localhost:8080/events`,
		``,
		`━━ 3. SLASH COMMANDS (send as message) ━━━━━━━━━━━━━━━━━━━━━━━`,
		`  /agents               — list all 35 specialist agents`,
		`  /use code debug this  — force a specific agent`,
		`  /status               — swarm + mesh + tunnel status`,
		`  /memory               — learned facts & preferences`,
		`  /set name=Alice       — save a preference`,
		`  /learn city=Tokyo     — teach a fact`,
		`  /learn voice <sample> — feed your writing style`,
		`  /mesh                 — show LAN peer devices`,
		`  /help                 — full command reference`,
		``,
		`━━ 4. AGENT IDs (35 specialists) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  auto | bodhi | perp-markets | portfolio | finance | tax`,
		`  real-estate | startup | sales | marketing | legal | hr`,
		`  ecommerce | supply-chain | organizer | freelance | writing`,
		`  tutor | language | mindset | code | research | devops`,
		`  data | security | web3 | social | comms | news | video`,
		`  design | health | food | travel | consulting | medical`,
		``,
		`━━ 5. MCP TOOLS (system control) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  Enable:  MCP_SHELL=1 MCP_FILESYSTEM=1 MCP_FETCH=1`,
		`  List:    GET /mcp/tools`,
		`  The LLM can then call tools inline:`,
		`    <tool name="shell"><cmd>ls -la</cmd></tool>`,
		`    <tool name="read_file"><path>/etc/hosts</path></tool>`,
		`    <tool name="fetch"><url>https://example.com</url></tool>`,
		``,
		`━━ 6. VOICE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  Install:  brew install whisper-cpp piper   (mac)`,
		`            sudo apt install whisper-cpp       (linux)`,
		`  Config:   WHISPER_BIN=/path/to/whisper-cli`,
		`            WHISPER_MODEL=/path/to/model.gguf`,
		`            PIPER_BIN=/path/to/piper`,
		`            PIPER_MODEL=/path/to/voice.onnx`,
		`  Status:   GET /voice/status`,
		``,
		`━━ 7. VISION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  Install:  ollama pull llava   (or moondream, bakllava)`,
		`  Config:   VISION_MODEL=llava`,
		`  Use:      POST /vision -F image=@photo.jpg -F prompt=Describe`,
		`            Or drag-drop an image into the web UI`,
		``,
		`━━ 8. REASONING ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  Config:   REASONING_MODE=chain-of-thought`,
		`  Effect:   Every answer is preceded by a step-by-step thinking`,
		`            pass. Trace stored in memory for self-improvement.`,
		``,
		`━━ 9. IP TUNNEL (remote access) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		tunnelSection,
		`  Install:  brew install cloudflared   (mac)`,
		`            sudo apt install cloudflared (linux)`,
		`  Start:    TUNNEL=cloudflared ./bodhi-hub`,
		`  QR code:  GET /tunnel/qr  (scan from phone)`,
		``,
		`━━ 10. MOBILE PWA ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  Android:  Chrome → http://localhost:8080 → Install App`,
		`  iPhone:   Safari → http://<ip>:8080 → Share → Add to Home`,
		`  Remote:   Use tunnel URL above`,
		``,
		`━━ FULL API REFERENCE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		`  POST /chat          main chat endpoint`,
		`  GET  /status        swarm health`,
		`  GET  /agents        list all agents`,
		`  GET  /memory        learned knowledge`,
		`  GET  /events        SSE live notifications`,
		`  GET  /mesh          LAN peer devices`,
		`  POST /transcribe    audio → text (Whisper)`,
		`  GET  /speak         text → audio (Piper)`,
		`  GET  /voice/status  voice capability status`,
		`  POST /vision        image → description (llava)`,
		`  GET  /tunnel        tunnel URL + status`,
		`  GET  /tunnel/qr     QR code SVG for mobile`,
		`  GET  /mcp/tools     list MCP tools`,
		`  GET  /capabilities  all capabilities at a glance`,
		`  GET  /guide         this guide`,
	}, "\n")
}
