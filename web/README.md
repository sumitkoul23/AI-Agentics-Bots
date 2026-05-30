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

## 🧩 Architecture

```
web/
├── index.html            # Single-page app shell
├── server.py             # Zero-dependency HTTP server + JSON API + static serving
├── assets/
│   ├── css/styles.css    # Cosmic dark theme
│   └── js/app.js         # Wizard + console controller
└── backend/
    ├── __init__.py
    └── deployer.py        # Bridges the web app to the core Genesis Protocol
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
