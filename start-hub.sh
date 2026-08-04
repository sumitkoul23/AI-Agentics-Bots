#!/usr/bin/env bash
# start-hub.sh — Build and start SkyAgents-hub (port 8080)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HUB_DIR="$SCRIPT_DIR/agents/priya-hub"

echo "==> SkyAgents-hub"
echo "    Building..."
cd "$HUB_DIR"
go build -o skyagents-hub .

echo "    Loading env from .env"
set -a
# shellcheck source=/dev/null
[ -f .env ] && source .env
set +a

echo "    Starting on port ${PORT:-8080}..."
echo "    UI:  http://localhost:${PORT:-8080}"
echo "    API: POST /chat  GET /status /agents /memory /events"
echo ""
exec ./skyagents-hub
