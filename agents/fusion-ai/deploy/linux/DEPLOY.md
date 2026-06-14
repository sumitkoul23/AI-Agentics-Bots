# FusionAI — Linux Deployment Guide

## Prerequisites

- Go 1.24+ (`go version`)
- A Teneo wallet private key (`PRIVATE_KEY`)
- At least one free AI key — Gemini or Groq

---

## Quick start (any Linux / macOS)

```bash
git clone https://github.com/sumitkoul23/AI-Agentics-Bots
cd AI-Agentics-Bots/agents/fusion-ai

cp .env.example .env
# Edit .env:  PRIVATE_KEY + GEMINI_API_KEY (free at aistudio.google.com)

chmod +x deploy/linux/start-agent.sh
./deploy/linux/start-agent.sh
```

---

## Production — systemd service

### 1. Create a dedicated user

```bash
sudo useradd -r -s /bin/false teneo
```

### 2. Deploy the agent

```bash
sudo mkdir -p /opt/fusion-ai
sudo cp -r . /opt/fusion-ai/
sudo chown -R teneo:teneo /opt/fusion-ai
sudo chmod +x /opt/fusion-ai/deploy/linux/start-agent.sh
```

### 3. Configure environment

```bash
sudo cp /opt/fusion-ai/deploy/linux/.env.example /opt/fusion-ai/.env
sudo nano /opt/fusion-ai/.env   # fill in PRIVATE_KEY + API keys
sudo chmod 600 /opt/fusion-ai/.env
sudo chown teneo:teneo /opt/fusion-ai/.env
```

### 4. Install and enable the service

```bash
sudo cp /opt/fusion-ai/deploy/linux/fusion-ai.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable fusion-ai
sudo systemctl start fusion-ai
```

### 5. Verify

```bash
sudo systemctl status fusion-ai
sudo journalctl -u fusion-ai -f
```

---

## Updating the agent

```bash
cd /opt/fusion-ai
sudo -u teneo git pull
sudo systemctl restart fusion-ai
```

---

## Model priority (auto-routing)

| Query type | Model order |
|---|---|
| Code / debugging | Gemini → GPT-4o → Groq → Ollama → Claude |
| Analysis / writing | Claude → Gemini → Groq → GPT-4o → Ollama |
| Mathematics | Gemini → Claude → GPT-4o → Groq → Ollama |
| General / fast | Gemini → Groq → Claude → GPT-4o → Ollama |

Free models (Gemini, Groq, Ollama) are always tried first.

---

## Slash commands (via Teneo chat)

```
/help               Command reference + active models
/models             Status of all configured models
/model <id> <msg>   Force a specific model
/code <task>        Code generation mode
/analyze <topic>    Deep analysis mode
/write <request>    Creative writing mode
/math <problem>     Math / science mode
```
