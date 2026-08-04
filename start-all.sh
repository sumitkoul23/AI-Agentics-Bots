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

# ── 1. SkyAgents-hub ──────────────────────────────────────────────────────────
if $START_HUB; then
    echo "==> [1/4] SkyAgents-hub (port 8080)"
    cd "$SCRIPT_DIR/agents/priya-hub"
    echo "    Building..."
    go build -o skyagents-hub . 2>&1 | sed 's/^/    /'
    set -a; [ -f .env ] && source .env; set +a
    ./skyagents-hub &
    PIDS+=($!)
    echo "    PID $! started"
    sleep 2
fi

# ── 2. SkyBot.app ─────────────────────────────────────────────────────────────
echo "==> [2/4] SkyBot.app (port 9090)"
cd "$SCRIPT_DIR/agents/priya-app"
echo "    Building..."
go build -o skybot-app . 2>&1 | sed 's/^/    /'
set -a; [ -f .env ] && source .env; set +a
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
[ ! -d node_modules ] && npm install --silent
set -a; [ -f .env ] && source .env; set +a
npm start &
PIDS+=($!)
echo "    PID $! started"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "============================================"
echo "  All services running:"
if $START_HUB; then
    echo "    SkyAgents-hub    http://localhost:8080"
fi
echo "    SkyBot.app       http://localhost:9090"
echo "    SKYagentic-os    http://localhost:8001"
echo "    Chain Studio     http://localhost:8000"
echo ""
echo "  Press Ctrl+C to stop all services."
echo "============================================"

# Wait for all background processes
wait
