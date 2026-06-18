#!/usr/bin/env bash
# start-agent.sh — Run Deck Insight AI on Linux / macOS.
# By default it summarises every deck dropped into ./inbox; pass a file to do one.
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_DIR"

if ! command -v node &>/dev/null; then
    echo "ERROR: Node.js (v18+) is not installed. See https://nodejs.org/"
    exit 1
fi

# Optional: load .env for model API keys (OLLAMA_HOST, GROQ_API_KEY, ...).
if [[ -f ".env" ]]; then
    set -a; source .env; set +a
fi

# One-shot mode: a deck path was provided.
if [[ $# -ge 1 ]]; then
    exec node main.js "$@"
fi

# Watch mode: summarise any *.key / *.pptx that lands in ./inbox.
INBOX="${INBOX:-$PROJECT_DIR/inbox}"
OUTBOX="${OUTBOX:-$PROJECT_DIR/outbox}"
mkdir -p "$INBOX" "$OUTBOX"
echo "[$(date -Is)] Deck Insight AI watching $INBOX (drop .key/.pptx files here)"

while true; do
    shopt -s nullglob
    for f in "$INBOX"/*.key "$INBOX"/*.pptx; do
        base="$(basename "${f%.*}")"
        out="$OUTBOX/${base}.summary.md"
        echo "[$(date -Is)] Summarising $f"
        if node main.js "$f" --out "$out"; then
            mv "$f" "$INBOX/${base}.done.$(date +%s)"
            echo "[$(date -Is)] -> $out"
        fi
    done
    sleep 5
done
