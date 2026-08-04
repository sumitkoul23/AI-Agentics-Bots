#!/usr/bin/env bash
# start-hub.sh — Build and start SkyAgents-hub (port 8080)
# Enables: IP tunnel (cloudflared), attached LLM model, NLP training pipeline
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HUB_DIR="$SCRIPT_DIR/agents/priya-hub"
TRAIN_DIR="$SCRIPT_DIR/training"

# ── 1. Bootstrap env from .env.example if .env is missing ─────────────────────
if [ ! -f "$HUB_DIR/.env" ]; then
    echo "    No .env found — copying from .env.example"
    cp "$HUB_DIR/.env.example" "$HUB_DIR/.env"
fi

# ── 2. Load env ────────────────────────────────────────────────────────────────
echo "    Loading env from agents/priya-hub/.env"
set -a
# shellcheck source=/dev/null
source "$HUB_DIR/.env"
set +a

# ── 3. Auto-install cloudflared if tunnel is requested but binary is missing ───
TUNNEL="${TUNNEL:-cloudflared}"
if [ "$TUNNEL" = "cloudflared" ] && ! command -v cloudflared &>/dev/null; then
    echo "    cloudflared not found — attempting install..."
    if command -v brew &>/dev/null; then
        brew install cloudflared
    elif command -v apt-get &>/dev/null; then
        curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
        echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/cloudflared.list
        sudo apt-get update -q && sudo apt-get install -y cloudflared
    else
        echo "    WARNING: cannot auto-install cloudflared. Install manually from https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/"
        echo "    Continuing without tunnel..."
        export TUNNEL=off
    fi
fi

# ── 4. Auto-detect best available Ollama model if not explicitly set ───────────
if [ -z "${OLLAMA_MODEL:-}" ]; then
    OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
    detected=$(curl -sf "$OLLAMA_HOST/api/tags" 2>/dev/null | \
        python3 -c "
import sys,json
data=json.load(sys.stdin)
preferred=['llama3.2','llama3.1','llama3','mistral','gemma2','phi3','qwen']
models=[m['name'] for m in data.get('models',[])]
for p in preferred:
    for m in models:
        if m.startswith(p):
            print(m); sys.exit(0)
if models:
    print(models[0])
" 2>/dev/null || true)
    if [ -n "$detected" ]; then
        export OLLAMA_MODEL="$detected"
        echo "    Auto-detected LLM model: $OLLAMA_MODEL"
    else
        echo "    WARNING: Ollama not reachable at ${OLLAMA_HOST}. Running in template mode."
        echo "    Install Ollama: https://ollama.ai — then: ollama pull llama3.2"
    fi
fi

# ── 5. Set NLP_TRAIN_DIR to the repo training/ directory if not already set ───
if [ -z "${NLP_TRAIN_DIR:-}" ] && [ -d "$TRAIN_DIR" ]; then
    export NLP_TRAIN_DIR="$TRAIN_DIR"
    echo "    NLP training data: $TRAIN_DIR"
fi

# Enable the /train endpoint by default
export NLP_TRAIN_ENDPOINT="${NLP_TRAIN_ENDPOINT:-1}"

# ── 6. Build ───────────────────────────────────────────────────────────────────
echo "==> SkyAgents-hub"
echo "    Building..."
cd "$HUB_DIR"
go build -o skyagents-hub .

# ── 7. Start ───────────────────────────────────────────────────────────────────
PORT="${PORT:-8080}"
echo ""
echo "    Starting on port $PORT"
echo "    UI:     http://localhost:$PORT"
echo "    API:    POST /chat  GET /status /agents /memory /events"
echo "    Train:  POST /train  GET /train/status"
if [ "${TUNNEL:-off}" != "off" ]; then
    echo "    Tunnel: TUNNEL=$TUNNEL (public URL will print below)"
fi
echo ""
exec ./skyagents-hub

