#!/usr/bin/env bash
# start-all.sh — Start the full AI-Agentics-Bots stack locally
#
# Services started:
#   SkyAgents-hub    http://localhost:8080   (Go — AI swarm engine)
#   SkyBot.app       http://localhost:9090   (Go — chat PWA)
#   SKYagentic-os    http://localhost:8001   (Python FastAPI — agent OS)
#   SKy Chain Studio http://localhost:8000   (Node.js — chain wizard)
#
# Usage:
#   ./start-all.sh          # start all four services
#   ./start-all.sh --no-hub # skip SkyAgents-hub (if already running)
#
# Stop: Ctrl+C  (sends SIGTERM to all child processes)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TRAIN_DIR="$SCRIPT_DIR/training"
PIDS=()

cleanup() {
    echo ""
    echo "==> Shutting down all services..."
    for pid in "${PIDS[@]}"; do
        kill "$pid" 2>/dev/null || true
    done
    wait 2>/dev/null || true
    echo "    Done."
}
trap cleanup EXIT INT TERM

START_HUB=true
for arg in "$@"; do
    [[ "$arg" == "--no-hub" ]] && START_HUB=false
done

echo "============================================"
echo "  AI-Agentics-Bots — Full Stack Launcher"
echo "============================================"
echo ""

# ── 1. SkyAgents-hub (tunnel + LLM + NLP training) ───────────────────────────
if $START_HUB; then
    echo "==> [1/4] SkyAgents-hub (port 8080 + cloudflared tunnel + NLP training)"
    cd "$SCRIPT_DIR/agents/priya-hub"
    [ ! -f .env ] && cp .env.example .env
    set -a; source .env; set +a
    TUNNEL="${TUNNEL:-cloudflared}"
    if [ "$TUNNEL" = "cloudflared" ] && ! command -v cloudflared &>/dev/null; then
        echo "    WARNING: cloudflared not found. Install: brew install cloudflared (mac) or see https://developers.cloudflare.com/"
    fi
    [ -z "${NLP_TRAIN_DIR:-}" ] && [ -d "$TRAIN_DIR" ] && export NLP_TRAIN_DIR="$TRAIN_DIR"
    export NLP_TRAIN_ENDPOINT="${NLP_TRAIN_ENDPOINT:-1}"
    export TUNNEL="${TUNNEL:-cloudflared}"
    echo "    Building..."
    go build -o skyagents-hub . 2>&1 | sed 's/^/    /'
    ./skyagents-hub &
    PIDS+=($!)
    echo "    PID $! started — tunnel + NLP training active"
    sleep 2
fi

# ── 2. SkyBot.app ─────────────────────────────────────────────────────────────
echo "==> [2/4] SkyBot.app (port 9090)"
cd "$SCRIPT_DIR/agents/priya-app"
[ ! -f .env ] && cp .env.example .env
set -a; source .env; set +a
echo "    Building..."
go build -o skybot-app . 2>&1 | sed 's/^/    /'
./skybot-app &
PIDS+=($!)
echo "    PID $! started"

# ── 3. SKYagentic-os ─────────────────────────────────────────────────────────
echo "==> [3/4] SKYagentic-os (port 8001)"
cd "$SCRIPT_DIR/agents/agentic-os"
set -a; [ -f .env ] && source .env; set +a
if [ -f .venv/bin/activate ]; then
    source .venv/bin/activate
fi
uvicorn gateway.main:app --host 0.0.0.0 --port "${PORT:-8001}" &
PIDS+=($!)
echo "    PID $! started"

# ── 4. SKy Chain Studio ───────────────────────────────────────────────────────
echo "==> [4/4] SKy Chain Studio (port 8000)"
cd "$SCRIPT_DIR/server"
[ ! -f .env ] && cp .env.example .env
[ ! -d node_modules ] && npm install --silent
set -a; source .env; set +a
npm start &
PIDS+=($!)
echo "    PID $! started"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "============================================"
echo "  All services running:"
if $START_HUB; then
    echo "    SkyAgents-hub    http://localhost:8080  (+ cloudflared tunnel)"
  echo "      LLM: Ollama auto-detect (set OLLAMA_MODEL to override)"
  echo "      NLP: training/ corpus loaded; POST /train for remote push"
fi
echo "    SkyBot.app       http://localhost:9090"
echo "    SKYagentic-os    http://localhost:8001"
echo "    Chain Studio     http://localhost:8000"
echo ""
echo "  Press Ctrl+C to stop all services."
echo "============================================"

# Wait for all background processes
wait
