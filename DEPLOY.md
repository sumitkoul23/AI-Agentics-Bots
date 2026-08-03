# Deployment Guide — AI-Agentics-Bots

Everything you need to run the full stack, from local development to production.

---

## Stack overview

| Service | Port | Description |
|---|---|---|
| **Agentic OS Gateway** | `8001` | Python FastAPI — agent kernel, tools, browser, memory |
| **Chain Deployment Studio** | `8000` | Node.js — SKYMETRIC chain wizard + live Cosmos node proxy |

---

## Quick start (Docker — recommended)

### 1. Clone and configure

```bash
git clone https://github.com/sumitkoul23/AI-Agentics-Bots.git
cd AI-Agentics-Bots

# Copy the env template and fill in your API keys
cp agents/agentic-os/.env.example .env
# Edit .env — add OPENAI_API_KEY or ANTHROPIC_API_KEY at minimum
```

### 2. Start everything

```bash
docker compose up --build
```

- **Agentic OS** → http://localhost:8001
- **API docs (Swagger)** → http://localhost:8001/docs
- **Chain Studio** → http://localhost:8000

### 3. Stop

```bash
docker compose down
```

Data is persisted in Docker volumes (`agentic_data`, `studio_data`).

---

## Local development (no Docker)

### Agentic OS

```bash
cd agents/agentic-os
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
playwright install chromium   # one-time browser install

cp .env.example .env          # fill in keys
uvicorn gateway.main:app --reload --port 8001
```

### Chain Deployment Studio

```bash
cd server
npm install
cp .env.example .env          # adjust CHAIN_RPC if needed
npm start                     # http://localhost:8000
```

---

## Production deployment

### Agentic OS → Render.com (free tier)

1. Push to GitHub.
2. Render → **New** → **Web Service** → connect repo.
3. Set:
   - **Root directory**: `agents/agentic-os`
   - **Build command**: `pip install -r requirements.txt && playwright install chromium --with-deps`
   - **Start command**: `uvicorn gateway.main:app --host 0.0.0.0 --port $PORT`
4. Add env vars: `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, etc.
5. Add a **disk** at `/data` for persistence (optional on free tier).

### Chain Studio → Render.com (auto-configured)

`render.yaml` in the repo root is already configured.

1. Render → **New** → **Blueprint** → select this repo.
2. Render reads `render.yaml` and provisions the `skymetric-studio` service automatically.

### Frontend → Cloudflare Pages (free)

1. Cloudflare → **Pages** → **Connect to Git** → select repo.
2. Build settings:
   - **Build command**: `python web/build_standalone.py --site`
   - **Output directory**: `web/dist`
3. Deploy.

### Netlify (Priya bot / agent hub)

`netlify.toml` is pre-configured.

1. Netlify → **Add new site** → **Import from Git** → select repo.
2. Add env vars for the Netlify Functions (e.g. `OPENAI_API_KEY`).
3. Deploy.

---

## Agentic OS API reference

The full interactive API is at **http://localhost:8001/docs** (Swagger UI).

### Key endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe — shows agent count and scheduler stats |
| `GET` | `/agents` | List all registered agents |
| `POST` | `/agents` | Register a new agent |
| `POST` | `/agents/{id}/run` | Run a task on an agent |
| `GET` | `/agents/{id}/memory` | View an agent's episodic memory |
| `GET` | `/capabilities` | List all available tools/capabilities |
| `POST` | `/tools/search` | Web search (DuckDuckGo / Brave / Serper) |
| `POST` | `/tools/read` | Fetch and extract text from any URL |
| `POST` | `/tools/execute` | Execute Python code in a sandbox |
| `GET` | `/kernel/stats` | Registry + scheduler statistics |
| `POST` | `/kernel/publish` | Publish a message to the agent bus |

### Example: create an agent and run a task

```bash
# Create an agent
curl -X POST http://localhost:8001/agents \
  -H "Content-Type: application/json" \
  -d '{"name": "ResearchBot", "description": "Searches the web", "capabilities": ["search"]}'

# Run a task (use the id returned above)
curl -X POST http://localhost:8001/agents/<id>/run \
  -H "Content-Type: application/json" \
  -d '{"task": "What are the latest developments in AI agents?"}'

# Web search
curl -X POST http://localhost:8001/tools/search \
  -H "Content-Type: application/json" \
  -d '{"query": "agentic AI 2025", "num_results": 3}'

# Execute code
curl -X POST http://localhost:8001/tools/execute \
  -H "Content-Type: application/json" \
  -d '{"code": "print(2 + 2)"}'
```

---

## Architecture

```
AI-Agentics-Bots/
├── agents/agentic-os/        ← Agentic OS (Python)
│   ├── kernel/               ←   Registry, MessageBus, Scheduler, CapabilityManager
│   ├── memory/               ←   ShortTermMemory, LongTermMemory, EpisodicMemory
│   ├── tools/                ←   WebSearch, CodeExecutor, FileIO, HttpClient
│   ├── browser/              ←   BrowserAgent (Playwright), PageReader
│   ├── gateway/main.py       ←   FastAPI REST gateway
│   ├── agent.py              ←   BaseAgent (LLM loop)
│   └── Dockerfile
├── server/                   ← Chain Deployment Studio (Node.js)
├── web/                      ← Static frontend
├── genesis/                  ← SKYMETRIC chain (Cosmos SDK / CosmWasm)
├── docker-compose.yml        ← One-command full-stack local deploy
├── render.yaml               ← Render.com blueprint (studio)
└── netlify.toml              ← Netlify config (Priya bot)
```

---

## Running tests

```bash
# Agentic OS unit tests (no API keys needed)
cd agents/agentic-os
pip install pytest
python -m pytest tests/ -v

# Node server smoke tests
cd server
npm install
node --check index.js
```

CI runs automatically on every push via GitHub Actions.

---

## Environment variables reference

| Variable | Service | Required | Default | Description |
|---|---|---|---|---|
| `OPENAI_API_KEY` | agentic-os | For LLM | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | agentic-os | For Claude | — | Anthropic API key |
| `AGENT_DEFAULT_MODEL` | agentic-os | No | `gpt-4o-mini` | Default LLM model |
| `BRAVE_SEARCH_KEY` | agentic-os | No | — | Brave Search API key |
| `SERPER_API_KEY` | agentic-os | No | — | Serper.dev API key |
| `AGENTIC_DATA_DIR` | agentic-os | No | `data` | Data persistence directory |
| `AGENTIC_WORKERS` | agentic-os | No | `4` | Scheduler worker count |
| `CHAIN_RPC` | studio | No | Cosmos Hub public | CometBFT RPC endpoint |
| `CHAIN_REST` | studio | No | Cosmos Hub public | Cosmos REST endpoint |
| `CHAIN_BECH32` | studio | No | `agentic` | Address prefix |
| `CHAIN_DENOM` | studio | No | `usky` | Native coin denom |
