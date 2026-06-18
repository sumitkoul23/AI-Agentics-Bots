#!/usr/bin/env bash
# sap-sage-train.sh — one daily learning/training cycle for SAP Sage AI.
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_DIR"

if ! command -v node &>/dev/null; then
    echo "ERROR: Node.js (v18+) is not installed."
    exit 1
fi

# Optional model keys (Ollama needs none).
if [[ -f ".env" ]]; then set -a; source .env; set +a; fi

echo "[$(date -Is)] SAP Sage AI — daily training cycle"
node main.js train
echo "[$(date -Is)] confidence snapshot:"
node main.js status | sed -n '/CONFIDENCE/,/Training maturity/p'
