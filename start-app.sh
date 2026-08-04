#!/usr/bin/env bash
# start-app.sh — Build and start SkyBot.app (port 9090)
# SkyAgents-hub must be running on port 8080 first.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$SCRIPT_DIR/agents/priya-app"

echo "==> SkyBot.app"
echo "    Building..."
cd "$APP_DIR"
go build -o skybot-app .

echo "    Loading env from .env"
set -a
# shellcheck source=/dev/null
[ -f .env ] && source .env
set +a

HUB="${PRIYA_HUB_URL:-http://localhost:8080}"
echo "    Starting on port ${PORT:-9090}  →  hub: $HUB"
echo "    Open: http://localhost:${PORT:-9090}"
echo ""
exec ./skybot-app
