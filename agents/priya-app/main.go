package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed static/index.html
var appHTML []byte

//go:embed static/manifest.json
var manifestJSON []byte

//go:embed static/sw.js
var swJS []byte

//go:embed static/icon.svg
var iconSVG []byte

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	hubURL := os.Getenv("PRIYA_HUB_URL")
	if hubURL == "" {
		hubURL = "http://localhost:8080"
	}

	mux := http.NewServeMux()

	// ── Proxy to Priya Hub ────────────────────────────────────────────────────
	proxy := &hubProxy{baseURL: hubURL, client: &http.Client{Timeout: 100 * time.Second}}

	mux.HandleFunc("/api/chat", proxy.chat)
	mux.HandleFunc("/api/status", proxy.status)
	mux.HandleFunc("/api/agents", proxy.agents)
	mux.HandleFunc("/api/memory", proxy.memory)

	// ── Static assets ─────────────────────────────────────────────────────────
	mux.HandleFunc("/manifest.json", staticHandler(manifestJSON, "application/manifest+json", "public, max-age=3600"))
	mux.HandleFunc("/sw.js", staticHandler(swJS, "application/javascript", "no-cache"))
	mux.HandleFunc("/icon.svg", staticHandler(iconSVG, "image/svg+xml", "public, max-age=86400"))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(appHTML)
	})

	log.Printf("Priya App — port %s  →  hub: %s", port, hubURL)
	log.Printf("Open: http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func staticHandler(data []byte, contentType, cacheControl string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", cacheControl)
		w.Write(data)
	}
}

// ── Hub proxy ─────────────────────────────────────────────────────────────────

type hubProxy struct {
	baseURL string
	client  *http.Client
}

func (p *hubProxy) chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Allow client-supplied hub URL override via header
	target := r.Header.Get("X-Hub-URL")
	if target == "" {
		target = p.baseURL
	}
	if _, err := url.ParseRequestURI(target); err != nil {
		target = p.baseURL
	}

	resp, err := p.client.Post(target+"/chat", "application/json", strings.NewReader(string(body)))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"reply": fmt.Sprintf("Cannot reach Priya Hub at %s — is it running?\n\nStart it with: `./priya-hub`\n\nOr update the Hub URL in ⚙️ Settings.", target),
		})
		return
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (p *hubProxy) status(w http.ResponseWriter, r *http.Request) {
	p.proxyGET(w, r, "/status", "text/plain")
}

func (p *hubProxy) agents(w http.ResponseWriter, r *http.Request) {
	p.proxyGET(w, r, "/agents", "text/plain")
}

func (p *hubProxy) memory(w http.ResponseWriter, r *http.Request) {
	p.proxyGET(w, r, "/memory", "application/json")
}

func (p *hubProxy) proxyGET(w http.ResponseWriter, r *http.Request, path, ct string) {
	target := r.Header.Get("X-Hub-URL")
	if target == "" {
		target = p.baseURL
	}
	if _, err := url.ParseRequestURI(target); err != nil {
		target = p.baseURL
	}

	resp, err := p.client.Get(target + path)
	if err != nil {
		http.Error(w, "hub unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", ct)
	w.Write(out)
}
