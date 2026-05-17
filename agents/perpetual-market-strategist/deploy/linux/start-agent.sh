#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_DIR"

export GOCACHE="${GOCACHE:-$PROJECT_DIR/.gocache}"
export GOPATH="${GOPATH:-$PROJECT_DIR/.gopath}"
mkdir -p "$GOCACHE" "$GOPATH"

echo "[$(date -Is)] Building agent..."
go build -o perpetual-market-strategist .

echo "[$(date -Is)] Starting agent..."
exec "$PROJECT_DIR/perpetual-market-strategist"
