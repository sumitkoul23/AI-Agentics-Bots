# 🟢 Chain Deployment Studio — Node.js backend

A real Node.js (Express) server for the Studio. Unlike the static GitHub Pages
build, this runs a live backend **and connects to a real blockchain node**.

## What it does

- **Serves** the `web/` front-end (clean-URL routes for every page)
- **Deployment API** with durable JSON-file persistence
- **Live chain proxy** — `/api/chain/*` forwards to a real CometBFT RPC +
  Cosmos REST node (defaults to a live public Cosmos Hub node, so the site is
  connected to a real chain the moment it boots). Point it at your own
  `skymetricd` node to go sovereign.

The browser talks to the backend, the backend talks to the node — so there are
no CORS problems and the active node is configured in one place.

## Run locally

```bash
cd server
npm install
npm start          # http://localhost:8000
```

Optional config (see `.env.example`):

```bash
CHAIN_RPC=https://rpc.skymetric.dev \
CHAIN_REST=https://rest.skymetric.dev \
CHAIN_ID=skymetric-1 \
npm start
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET  | `/api/health` | Liveness + active chain config |
| GET  | `/api/targets` | Supported deploy targets |
| GET  | `/api/stats` | Aggregate deployment stats |
| GET  | `/api/deployments` | List deployments (newest first) |
| GET  | `/api/deployments/:id` | One deployment |
| POST | `/api/deploy` | Validate + persist a deployment |
| GET  | `/api/chain/config` | Active node config |
| GET  | `/api/chain/status` | **Live** node status (real block height) |
| GET  | `/api/chain/balances/:address` | Live on-chain balances |
| GET  | `/api/chain/account/:address` | Account number + sequence |
| POST | `/api/chain/broadcast` | Broadcast a signed `TxRaw` (base64) |

## Deploy free

**Render** (one click): push the repo, then render.com → New → Blueprint → pick
the repo. The root [`render.yaml`](../render.yaml) provisions a free Node web
service with `/api/health` health checks.

**Docker / Fly.io**: the [`Dockerfile`](Dockerfile) builds a self-contained
image (Node 22 alpine). For Fly: `fly launch` from the repo root, then
`fly deploy`. Mount a volume at `/data` to persist deployments across redeploys.

## Live node — proof it's real

```bash
curl -s localhost:8000/api/chain/status
# { "ok": true, "network": "cosmoshub-4", "latest_block_height": "31535590", ... }
```

The height advances every block — it's a real, live chain node, not a mock.
