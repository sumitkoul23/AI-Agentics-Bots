# FusionAI — Linux deployment

## 1. Prepare the server

```bash
sudo useradd --system --create-home --shell /usr/sbin/nologin teneo
sudo mkdir -p /opt/fusion-ai
sudo chown -R teneo:teneo /opt/fusion-ai
sudo apt-get update && sudo apt-get install -y git golang-go
```

## 2. Copy the project

Copy this project folder to `/opt/fusion-ai`, then on the server:

```bash
cd /opt/fusion-ai
cp deploy/linux/.env.example .env
chmod +x deploy/linux/start-agent.sh
```

Edit `.env` and set:
- `PRIVATE_KEY` — your Teneo signer wallet (no `0x` prefix)
- At least one AI key: `GEMINI_API_KEY` (free) or `GROQ_API_KEY` (free)

## 3. Install the systemd service

```bash
sudo cp /opt/fusion-ai/deploy/linux/fusion-ai.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now fusion-ai
sudo systemctl status fusion-ai
```

## 4. Verify

```bash
journalctl -u fusion-ai -f
```

Expected output on a healthy start:
```
[✓] Gemini 2.0 Flash    — FREE tier
FusionAI live — models: Gemini, Groq
```

## 5. Updating

```bash
cd /opt/fusion-ai
git pull
sudo systemctl restart fusion-ai
```
