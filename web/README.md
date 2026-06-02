# 🌌 Chain Deployment Studio

A professional, guided web app for deploying a new **Genesis Protocol** chain —
the self-replicating AI agent civilization on Solana.

It walks you through configuring your chain identity, target network, genesis
agent mix, and economy, then "deploys" it with a live console and a verifiable
deployment summary (chain ID, authority, token mint, treasury, genesis
signature, and the genesis generation of agents).

## ✨ Features

- **Five-step deployment wizard** — Identity → Network → Agents → Economy → Review
- **Live deployment console** with animated, step-by-step progress
- **Real protocol integration** — genesis agents are born with genuine DNA from
  the project's `DNA` system and a `Treasury` is bootstrapped via the core
  `genesis` package (falls back gracefully if the package can't be imported)
- **Network selection** — Devnet / Testnet / Mainnet Beta with cost hints
- **Deployment artifacts** — copyable chain ID, authority, mint, treasury and
  signature, plus an Explorer link
- **Zero dependencies** — pure Python standard library backend + vanilla
  HTML/CSS/JS front-end (no build step)

## 🚀 Run it

From the repository root:

```bash
python web/server.py
# then open http://localhost:8000
```

Options:

```bash
python web/server.py --host 0.0.0.0 --port 9000
```

> Requires Python 3.8+. No `pip install` needed.

### Offline single-file build

Don't want to run a server? Build a fully self-contained HTML file (CSS + an
offline JS port of the deployer inlined) that runs the whole wizard from
`file://` with no network:

```bash
python web/build_standalone.py     # -> web/dist/chain-deployment-studio.html
```

Open the generated file in any browser — the five-step wizard runs live and
still generates/downloads the real `genesis-overrides.json`, `init-chain.sh`,
and `.env`.

## 🌐 Free hosting (Cloudflare Pages)

The same offline build doubles as the **static site** — the whole wizard runs
client-side, so it hosts anywhere with zero backend, **free hosting + a free
`*.pages.dev` domain**. This mirrors how [`genesis/site`](../genesis/site)
deploys.

Build the publishable directory locally:

```bash
python web/build_standalone.py --site   # -> web/dist/ (index.html + _headers + robots.txt)
```

Deploy in ~90 seconds:

1. <https://pages.cloudflare.com> → **Create a project** → **Connect to Git**
2. Pick `sumitkoul23/ai-agentics-bots`
3. **Build command:** `python web/build_standalone.py --site`
4. **Build output directory:** `web/dist`
5. **Save and Deploy**

Cloudflare hands you a live `https://<project>.pages.dev` URL immediately and
auto-redeploys on every push. To use a custom domain later, add it under the
project's **Custom domains** tab and point DNS at Cloudflare — still free.

`web/_headers` ships sensible security headers (CSP, `X-Frame-Options`, etc.)
that Cloudflare Pages applies automatically.

## 🧩 Architecture

```
web/
├── index.html            # Single-page app shell
├── server.py             # Zero-dependency HTTP server + JSON API + static serving
├── build_standalone.py   # Inlines everything -> offline file / Cloudflare Pages dist
├── _headers              # Security headers for Cloudflare Pages
├── robots.txt
├── assets/
│   ├── css/styles.css    # Cosmic dark theme
│   └── js/
│       ├── app.js         # Wizard + console controller (server-backed)
│       └── standalone.js  # Offline JS port of the deployer (no API calls)
└── backend/
    ├── __init__.py
    └── deployer.py        # Validates requests and emits real chain artifacts
```

### API

| Method | Endpoint                  | Description                          |
|--------|---------------------------|--------------------------------------|
| GET    | `/api/health`             | Liveness probe                       |
| GET    | `/api/networks`           | Supported networks                   |
| GET    | `/api/stats`              | Aggregate deployment stats           |
| GET    | `/api/deployments`        | List deployments (newest first)      |
| GET    | `/api/deployments/<id>`   | Single deployment record             |
| POST   | `/api/deploy`             | Validate + execute a deployment      |

#### Deploy request body

```json
{
  "chain_name": "Nebula Genesis",
  "symbol": "NEB",
  "network": "devnet",
  "agents": { "trader": 3, "governor": 2, "builder": 2 },
  "treasury_sol": 10,
  "token_supply": 1000000,
  "mutation_rate": 0.1
}
```

## 🔐 Note on on-chain deployment

Actual submission to Solana requires funded keypairs and live RPC access. The
final broadcast step is simulated deterministically, while validation, agent
birth, DNA generation, and treasury bootstrap use the genuine protocol classes —
so the artifacts you get back are real, reproducible, and ready to wire up to a
live `SolanaClient`.
