# Bodhi — Autonomous AI Swarm

On-device self-learning AI assistant. 12 specialist agents. No cloud. No API keys. Runs entirely on your device using [Ollama](https://ollama.ai).

```
bodhi-hub  →  12 specialist agents  →  SSE push notifications
bodhi-app  →  mobile chat UI (PWA — installs as Android / iOS app)
```

---

## Download

Latest release: **[Releases →](https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest)**

| Platform | Hub (AI engine) | App (chat UI) |
|----------|----------------|---------------|
| **macOS Apple Silicon** (M1/M2/M3/M4) | [bodhi-hub-macos-arm64.tar.gz][h-mac-arm] | [bodhi-app-macos-arm64.tar.gz][a-mac-arm] |
| **macOS Intel** | [bodhi-hub-macos-amd64.tar.gz][h-mac-x64] | [bodhi-app-macos-amd64.tar.gz][a-mac-x64] |
| **Windows** (64-bit) | [bodhi-hub-windows-amd64.zip][h-win] | [bodhi-app-windows-amd64.zip][a-win] |
| **Linux** (x86\_64) | [bodhi-hub-linux-amd64.tar.gz][h-lnx] | [bodhi-app-linux-amd64.tar.gz][a-lnx] |
| **Linux / Ubuntu** (ARM64) | [bodhi-hub-linux-arm64.tar.gz][h-lnx-arm] | [bodhi-app-linux-arm64.tar.gz][a-lnx-arm] |
| **Android** (ARM64 via Termux) | [bodhi-hub-android-arm64.tar.gz][h-and] | [bodhi-app-android-arm64.tar.gz][a-and] |
| **iOS** | PWA — see instructions below | — |

[h-mac-arm]: https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-macos-arm64.tar.gz
[h-mac-x64]: https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-macos-amd64.tar.gz
[h-win]:     https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-windows-amd64.zip
[h-lnx]:     https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-linux-amd64.tar.gz
[h-lnx-arm]: https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-linux-arm64.tar.gz
[h-and]:     https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-android-arm64.tar.gz
[a-mac-arm]: https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-macos-arm64.tar.gz
[a-mac-x64]: https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-macos-amd64.tar.gz
[a-win]:     https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-windows-amd64.zip
[a-lnx]:     https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-linux-amd64.tar.gz
[a-lnx-arm]: https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-linux-arm64.tar.gz
[a-and]:     https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-android-arm64.tar.gz

---

## Installation by Platform

### macOS

```bash
# 1. Install Ollama
brew install ollama
ollama pull llama3.2        # or: phi3:mini (smaller), llama3.2:1b (fastest)

# 2. Download and extract (Apple Silicon)
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-macos-arm64.tar.gz | tar xz
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-macos-arm64.tar.gz | tar xz
# Intel Mac: replace arm64 with amd64 in the URLs above

# 3. Run
ollama serve &              # start the AI engine
./bodhi-hub-macos-arm64 &   # start the swarm (port 8080)
./bodhi-app-macos-arm64     # start the chat app (port 9090)

# 4. Open
open http://localhost:9090
```

> **macOS security prompt**: If blocked by Gatekeeper, run:
> `xattr -d com.apple.quarantine bodhi-hub-macos-arm64 bodhi-app-macos-arm64`

---

### Windows

```powershell
# 1. Install Ollama
# Download from https://ollama.ai/download and run the installer, then:
ollama pull llama3.2

# 2. Download and extract (PowerShell)
Invoke-WebRequest -Uri "https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-windows-amd64.zip" -OutFile bodhi-hub.zip
Invoke-WebRequest -Uri "https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-windows-amd64.zip" -OutFile bodhi-app.zip
Expand-Archive bodhi-hub.zip .
Expand-Archive bodhi-app.zip .

# 3. Run (open two Command Prompt / PowerShell windows)
Start-Process .\bodhi-hub-windows-amd64.exe   # window 1 — port 8080
.\bodhi-app-windows-amd64.exe                 # window 2 — port 9090

# 4. Open browser
Start-Process http://localhost:9090
```

---

### Linux (x86\_64 / Ubuntu amd64)

```bash
# 1. Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull llama3.2

# 2. Download and extract
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-linux-amd64.tar.gz | tar xz
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-linux-amd64.tar.gz | tar xz

# 3. Run
ollama serve &
./bodhi-hub-linux-amd64 &
./bodhi-app-linux-amd64

# 4. Open
xdg-open http://localhost:9090   # or paste into any browser
```

---

### Linux (ARM64 — Raspberry Pi / ARM server / Ubuntu ARM)

```bash
# 1. Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull llama3.2:1b    # 1b model recommended for Raspberry Pi

# 2. Download
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-linux-arm64.tar.gz | tar xz
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-linux-arm64.tar.gz | tar xz

# 3. Run
ollama serve &
./bodhi-hub-linux-arm64 &
./bodhi-app-linux-arm64

# 4. Open
http://localhost:9090
```

---

### Android (via Termux)

> **Prerequisite:** Install [Termux from F-Droid](https://f-droid.org/packages/com.termux/) — the Play Store version is outdated and won't work.

**Option A — Automated setup (recommended)**

```bash
# In Termux:
pkg update && pkg install curl
curl -fsSL https://raw.githubusercontent.com/sumitkoul23/AI-Agentics-Bots/main/agents/priya-hub/termux-setup.sh | bash
~/start-bodhi.sh
```

The setup script downloads Ollama for Android, lets you pick a model, downloads the binaries, and generates `~/start-bodhi.sh`.

**Option B — Manual**

```bash
# In Termux:
pkg update && pkg install curl

# Download binaries
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-hub-android-arm64.tar.gz | tar xz
curl -L https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest/download/bodhi-app-android-arm64.tar.gz | tar xz

# Download Ollama for Android (Linux ARM64 binary)
curl -L https://ollama.ai/download/ollama-linux-arm64 -o ollama && chmod +x ollama
./ollama pull llama3.2:1b   # ~1.3 GB — fits on 6 GB RAM phones

# Run
./ollama serve &
./bodhi-hub-android-arm64 &
./bodhi-app-android-arm64
```

**Install as app:** Open Chrome → `http://localhost:9090` → tap **⋮** → **"Add to Home screen"**

---

### iOS

Bodhi runs as a **Progressive Web App (PWA)** on iPhone and iPad — no App Store required.

1. Run `bodhi-hub` and `bodhi-app` on any device on the same Wi-Fi network (Mac, PC, Linux, or Android)
2. Find the host device's local IP:
   - **Mac/Linux:** `ip route get 1 | awk '{print $7}'` or `ifconfig | grep "inet "`
   - **Windows:** `ipconfig` → look for IPv4 Address
3. On iPhone/iPad, open **Safari** and navigate to `http://<ip>:9090`
4. Tap the **Share** button → **"Add to Home Screen"** → **Add**

Bodhi appears on your home screen with its icon, launches full-screen, and caches for offline use.

> Safari is required for "Add to Home Screen" — Chrome on iOS cannot install PWAs.

---

## CLI Reference

### bodhi-hub

```
USAGE
  ./bodhi-hub

ENVIRONMENT VARIABLES
  PORT           HTTP port             (default: 8080)
  OLLAMA_HOST    Ollama API URL        (default: http://localhost:11434)

EXAMPLES
  ./bodhi-hub
  PORT=3000 ./bodhi-hub
  OLLAMA_HOST=http://192.168.1.5:11434 ./bodhi-hub

API ENDPOINTS
  POST /chat        send a message to an agent
  GET  /status      swarm status + training score
  GET  /agents      list all 12 specialist agents
  GET  /memory      learned facts + preferences (JSON)
  GET  /events      SSE stream of autonomous notifications

CHAT PAYLOAD
  {"message": "your message", "agent": "auto"}

  agent values:
    auto | bodhi | perp-markets | portfolio | social | comms |
    organizer | finance | freelance | code | health | research | news

SLASH COMMANDS (send as the message field)
  /status     training score, queue depths, per-agent confidence
  /agents     list all specialists with descriptions
  /memory     dump learned facts and preferences
  /help       show all commands
  /use <id>   force next message to a specific agent
```

### bodhi-app

```
USAGE
  ./bodhi-app

ENVIRONMENT VARIABLES
  PORT              HTTP port          (default: 9090)
  PRIYA_HUB_URL     bodhi-hub URL      (default: http://localhost:8080)

EXAMPLES
  ./bodhi-app
  PORT=8888 PRIYA_HUB_URL=http://192.168.1.10:8080 ./bodhi-app
```

### curl / API usage

```bash
# Start a conversation (auto-routes to the best agent)
curl -s -X POST http://localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"What is BTC funding rate telling us right now?","agent":"auto"}' | jq .reply

# Send to a specific agent
curl -s -X POST http://localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Debug this Go nil pointer dereference","agent":"code"}' | jq .reply

# Check swarm status
curl -s http://localhost:8080/status

# List agents
curl -s http://localhost:8080/agents

# Stream live notifications (SSE)
curl -N http://localhost:8080/events
```

---

## Agents

| ID | Specialty |
|----|-----------|
| `auto` | Bodhi picks the best agent automatically |
| `bodhi` | General coordinator — answers directly or delegates |
| `perp-markets` | Crypto perpetuals: funding rates, OI, entry/SL/TP trade plans |
| `portfolio` | Asset allocation, rebalancing, Sharpe/drawdown risk metrics |
| `social` | Content for LinkedIn, Twitter/X, Instagram, TikTok, YouTube |
| `comms` | Cold emails, proposals, negotiations, client communication |
| `organizer` | Time-blocking, MIT planning, energy management, deep work |
| `finance` | Macro, DeFi, crypto fundamentals, equities, sector rotation |
| `freelance` | Job hunting, proposal writing, rate strategy, client pipeline |
| `code` | Debugging, architecture, code review — all languages |
| `health` | Training programs, nutrition, sleep, recovery (evidence-based) |
| `research` | Deep dives, synthesis, fact-checking, confidence-rated output |
| `news` | Signal vs noise, market events, trend analysis |

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   bodhi-hub :8080                   │
│                                                     │
│   Router → Swarm ──► 12 Agent goroutines            │
│                │         ↕ Ollama (local LLM)       │
│            NotifBus ──► SSE /events                 │
│                │                                    │
│            Memory  (.bodhi-memory.json)             │
│            Trainer (self-eval every 8 turns)        │
│            Learner (fact extraction per turn)       │
│            DecisionEngine (per-agent confidence)    │
│            Mesh    (UDP multicast LAN discovery)    │
└─────────────────────────────────────────────────────┘
              ↑ proxy (same device or LAN)
┌─────────────────────────────────────────────────────┐
│                   bodhi-app :9090                   │
│                                                     │
│   Mobile PWA ──► /api/* proxy ──► hub              │
│   EventSource   /api/events    (SSE passthrough)    │
└─────────────────────────────────────────────────────┘
```

- **Zero external dependencies** — pure Go stdlib
- **Self-learning** — training score 0–100; behaviour adapts at 20 / 45 / 70 / 100
- **Autonomous background tasks** — agents push insights every 4h / 6h / 8h / 12h
- **Deep eval loop** — weakest agent targeted for improvement every 3h
- **Mesh networking** — UDP multicast peer discovery; multiple devices share memory
- **Single binary** — all assets embedded via `//go:embed`

---

## Build from Source

**Requirements:** Go 1.24+ · Ollama

```bash
git clone https://github.com/sumitkoul23/AI-Agentics-Bots.git
cd AI-Agentics-Bots

# Hub
cd agents/priya-hub && go build -o bodhi-hub . && ./bodhi-hub

# App (separate terminal)
cd agents/priya-app && go build -o bodhi-app .
PRIYA_HUB_URL=http://localhost:8080 ./bodhi-app
```

**Cross-compile:**

```bash
# All platforms
cd agents/priya-hub && make all

# Individual targets
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o bodhi-hub-linux    .
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o bodhi-hub-macos    .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bodhi-hub.exe      .
GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o bodhi-hub-android  .
```

**Publish a release** (GitHub Actions builds all 14 binaries automatically):

```bash
git tag v1.0.0 && git push origin v1.0.0
```

---

## Chain Deployment Studio

A professional, guided web app for spinning up a new **SKYMETRIC** chain
(Cosmos SDK + CometBFT) lives in [`web/`](web/README.md). It walks you through
chain identity, tokenomics, validators and a deploy target, then generates
ready-to-run artifacts (`genesis-overrides.json`, `init-chain.sh`, `.env`) for
the real [`genesis/chain`](genesis/chain) toolchain — with zero external
dependencies.

```bash
python web/server.py        # then open http://localhost:8000
```

See [`web/README.md`](web/README.md) for full details.

---

## Safety

- No API keys required — Ollama is fully local
- `.env` files and memory state are git-ignored
- Finance / trading: advisory only — no live execution by default
- SSE `/events` has no wildcard CORS — safe from cross-origin reads
