#!/usr/bin/env bash
# start-agent.sh — Build and run FusionAI on Linux / macOS
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_DIR"

# ── Pre-flight checks ─────────────────────────────────────────────────────────
if ! command -v go &>/dev/null; then
    echo "ERROR: Go is not installed. See https://go.dev/dl/"
    exit 1
fi

if [[ ! -f ".env" ]]; then
    cp .env.example .env
    echo ""
    echo "  Created .env from .env.example"
    echo ""
    echo "  Fill in before starting:"
    echo "    PRIVATE_KEY     — Teneo wallet private key"
    echo "    GEMINI_API_KEY  — free at https://aistudio.google.com/app/apikey"
    echo "    GROQ_API_KEY    — free at https://console.groq.com  (optional)"
    echo ""
    exit 1
fi

# ── Build ─────────────────────────────────────────────────────────────────────
export GOCACHE="${GOCACHE:-$PROJECT_DIR/.gocache}"
export GOPATH="${GOPATH:-$PROJECT_DIR/.gopath}"
mkdir -p "$GOCACHE" "$GOPATH"

echo "[$(date -Is)] Building FusionAI..."
go build -o fusion-ai .
echo "[$(date -Is)] Build OK"

# ── Launch ────────────────────────────────────────────────────────────────────
echo "[$(date -Is)] Starting FusionAI on Teneo network..."
exec "$PROJECT_DIR/fusion-ai"
