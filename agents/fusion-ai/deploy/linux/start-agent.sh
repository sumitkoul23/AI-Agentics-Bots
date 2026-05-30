#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_DIR"

export GOCACHE="${GOCACHE:-$PROJECT_DIR/.gocache}"
export GOPATH="${GOPATH:-$PROJECT_DIR/.gopath}"
mkdir -p "$GOCACHE" "$GOPATH"

# Auto-create .env on first run
if [ ! -f "$PROJECT_DIR/.env" ]; then
    cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
    echo "Created .env from .env.example"
    echo "Fill in PRIVATE_KEY and at least one AI key, then re-run."
    echo ""
    echo "  Free: GEMINI_API_KEY  -> https://aistudio.google.com/app/apikey"
    echo "  Free: GROQ_API_KEY    -> https://console.groq.com"
    exit 0
fi

# Source the env file
set -a; . "$PROJECT_DIR/.env"; set +a

echo "[$(date -Is)] Building FusionAI..."
go build -o fusion-ai .

echo "[$(date -Is)] Starting FusionAI..."
exec "$PROJECT_DIR/fusion-ai"
