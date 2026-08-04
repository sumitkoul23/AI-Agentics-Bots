#!/usr/bin/env bash
# start-studio.sh — Start SKy Chain Deployment Studio (port 8000)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STUDIO_DIR="$SCRIPT_DIR/server"

echo "==> SKy Chain Studio"
cd "$STUDIO_DIR"

if [ ! -d node_modules ]; then
    echo "    Installing dependencies..."
    npm install
fi

echo "    Loading env from .env"
set -a
# shellcheck source=/dev/null
[ -f .env ] && source .env
set +a

PORT="${PORT:-8000}"
echo "    Starting on port $PORT"
echo "    Open: http://localhost:$PORT"
echo ""
exec npm start
