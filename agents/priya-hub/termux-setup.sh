#!/data/data/com.termux/files/usr/bin/bash
# ── Bodhi + Ollama setup for Android via Termux ───────────────────────────────
# Run this once after installing Termux from F-Droid.
# https://f-droid.org/packages/com.termux/
#
# Usage:
#   bash termux-setup.sh
#
# After setup, start Bodhi:
#   ~/start-bodhi.sh

set -e

RED='\033[0;31m'; GREEN='\033[0;32m'; AMBER='\033[0;33m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓${NC} $1"; }
info() { echo -e "${AMBER}→${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; exit 1; }

echo ""
echo "  🌸  Bodhi — Android Setup"
echo "  ─────────────────────────────────────"
echo ""

# ── 1. Update Termux packages ─────────────────────────────────────────────────
info "Updating Termux packages…"
pkg update -y && pkg upgrade -y
ok "Packages updated"

# ── 2. Install dependencies ───────────────────────────────────────────────────
info "Installing wget, git, and storage permissions…"
pkg install -y wget curl git
termux-setup-storage 2>/dev/null || true
ok "Dependencies installed"

# ── 3. Download Ollama ARM64 ──────────────────────────────────────────────────
info "Downloading Ollama (Linux ARM64)…"
OLLAMA_DIR="$HOME/ollama"
mkdir -p "$OLLAMA_DIR"

# Use the official Ollama Linux ARM64 binary — works in Termux
OLLAMA_VERSION=$(curl -s https://api.github.com/repos/ollama/ollama/releases/latest | grep '"tag_name"' | cut -d'"' -f4 2>/dev/null || echo "v0.6.0")
OLLAMA_URL="https://github.com/ollama/ollama/releases/download/${OLLAMA_VERSION}/ollama-linux-arm64"

if [ -f "$OLLAMA_DIR/ollama" ]; then
  ok "Ollama already downloaded"
else
  wget -q --show-progress -O "$OLLAMA_DIR/ollama" "$OLLAMA_URL" || {
    fail "Failed to download Ollama. Check your internet connection."
  }
  chmod +x "$OLLAMA_DIR/ollama"
  ok "Ollama downloaded: $OLLAMA_DIR/ollama"
fi

# ── 4. Download Bodhi Hub binary ──────────────────────────────────────────────
BODHI_HUB_BIN="$HOME/bodhi-hub"
if [ -f "$BODHI_HUB_BIN" ]; then
  ok "bodhi-hub binary already present"
else
  info "Looking for bodhi-hub-android-arm64 in Downloads…"
  if [ -f "/sdcard/Download/bodhi-hub-android-arm64" ]; then
    cp /sdcard/Download/bodhi-hub-android-arm64 "$BODHI_HUB_BIN"
    chmod +x "$BODHI_HUB_BIN"
    ok "bodhi-hub copied from Downloads"
  else
    echo ""
    echo "  ⚠️  bodhi-hub binary not found."
    echo ""
    echo "  Download from GitHub Releases:"
    echo "    https://github.com/sumitkoul23/AI-Agentics-Bots/releases/latest"
    echo "    → bodhi-hub-android-arm64.tar.gz"
    echo ""
    echo "  Or build on your laptop:"
    echo "    cd agents/priya-hub && make android  # outputs dist/bodhi-hub-android-arm64"
    echo "    adb push dist/bodhi-hub-android-arm64 /sdcard/Download/"
    echo ""
    echo "  Re-run this script after copying."
    echo ""
  fi
fi

# Also check for bodhi-app
BODHI_APP_BIN="$HOME/bodhi-app"
if [ -f "/sdcard/Download/bodhi-app-android-arm64" ] && [ ! -f "$BODHI_APP_BIN" ]; then
  cp /sdcard/Download/bodhi-app-android-arm64 "$BODHI_APP_BIN"
  chmod +x "$BODHI_APP_BIN"
  ok "bodhi-app copied from Downloads"
fi

# ── 5. Pull a model (tiny for mobile) ────────────────────────────────────────
echo ""
echo "  Which model do you want to pull?"
echo "  1) llama3.2:1b   (~1.3 GB) — fastest, good quality  [RECOMMENDED]"
echo "  2) llama3.2:3b   (~2.0 GB) — better quality"
echo "  3) phi3:mini     (~2.3 GB) — great reasoning"
echo "  4) Skip (use template mode)"
echo ""
read -rp "  Choice [1]: " MODEL_CHOICE
MODEL_CHOICE="${MODEL_CHOICE:-1}"

case "$MODEL_CHOICE" in
  1) MODEL="llama3.2:1b" ;;
  2) MODEL="llama3.2:3b" ;;
  3) MODEL="phi3:mini"   ;;
  *) MODEL=""            ;;
esac

# ── 6. Create start script ────────────────────────────────────────────────────
START_SCRIPT="$HOME/start-bodhi.sh"
cat > "$START_SCRIPT" <<SCRIPT
#!/data/data/com.termux/files/usr/bin/bash
# Bodhi startup script

OLLAMA_DIR="\$HOME/ollama"
BODHI_HUB="\$HOME/bodhi-hub"
BODHI_APP="\$HOME/bodhi-app"

echo "🌸 Starting Bodhi…"

# Start Ollama in background
if [ -f "\$OLLAMA_DIR/ollama" ]; then
  echo "→ Starting Ollama…"
  OLLAMA_HOST=127.0.0.1:11434 "\$OLLAMA_DIR/ollama" serve > "\$HOME/ollama.log" 2>&1 &
  OLLAMA_PID=\$!
  sleep 3

$(if [ -n "$MODEL" ]; then
echo "  # Pull model if not already present"
echo "  echo '→ Pulling model $MODEL (this may take a while on first run)…'"
echo "  if ! \"\$OLLAMA_DIR/ollama\" pull $MODEL; then"
echo "    echo '✗ Failed to pull model $MODEL — check network and try again with: ollama pull $MODEL'"
echo "    exit 1"
echo "  fi"
fi)
  echo "✓ Ollama ready"
fi

# Start Bodhi Hub
if [ -f "\$BODHI_HUB" ]; then
  echo "→ Starting Bodhi Hub on :8080…"
  OLLAMA_HOST=http://127.0.0.1:11434 "\$BODHI_HUB" > "\$HOME/bodhi-hub.log" 2>&1 &
  sleep 1
  echo "✓ Hub running: http://localhost:8080"
fi

# Start Bodhi App (mobile UI)
if [ -f "\$BODHI_APP" ]; then
  echo "→ Starting Bodhi App on :9090…"
  PRIYA_HUB_URL=http://127.0.0.1:8080 "\$BODHI_APP" > "\$HOME/bodhi-app.log" 2>&1 &
  sleep 1
  echo "✓ App running: http://localhost:9090"
fi

echo ""
echo "  Open Chrome and visit:"
echo "  → http://localhost:9090  (mobile app)"
echo "  → http://localhost:8080  (full hub UI)"
echo ""
echo "  Tap ⋮ → 'Add to Home Screen' to install as PWA"
echo ""
echo "  Logs: ~/bodhi-hub.log  ~/bodhi-app.log  ~/ollama.log"
SCRIPT
chmod +x "$START_SCRIPT"
ok "Start script created: $START_SCRIPT"

# ── 7. Summary ────────────────────────────────────────────────────────────────
echo ""
echo "  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅  Setup complete!"
echo ""
echo "  Start everything with:"
echo "    ~/start-bodhi.sh"
echo ""
if [ -n "$MODEL" ]; then
  echo "  Model: $MODEL"
  echo "  (will be pulled on first start if not present)"
  echo ""
fi
echo "  Then open Chrome → http://localhost:9090"
echo "  Tap ⋮ → Add to Home Screen"
echo "  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
