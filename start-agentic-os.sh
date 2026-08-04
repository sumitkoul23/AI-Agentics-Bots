#!/usr/bin/env bash
# start-agentic-os.sh — Start SKYagentic-os FastAPI gateway (port 8001)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OS_DIR="$SCRIPT_DIR/agents/agentic-os"

echo "==> SKYagentic-os"
cd "$OS_DIR"

# Activate virtualenv if present
if [ -f .venv/bin/activate ]; then
    echo "    Activating .venv"
    # shellcheck source=/dev/null
    source .venv/bin/activate
elif command -v python3 &>/dev/null; then
    echo "    No .venv found — using system Python"
    echo "    Tip: python -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt"
else
    echo "ERROR: python3 not found. Install Python 3.10+ first." >&2
    exit 1
fi

echo "    Loading env from .env"
set -a
# shellcheck source=/dev/null
[ -f .env ] && source .env
set +a

PORT="${PORT:-8001}"
echo "    Starting on port $PORT"
echo "    API docs: http://localhost:$PORT/docs"
echo ""
exec uvicorn gateway.main:app --host 0.0.0.0 --port "$PORT" --reload
