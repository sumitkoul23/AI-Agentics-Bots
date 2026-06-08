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
│       ├── shell.js       # Shared app nav + mobile tab bar + API helper
│       └── standalone.js  # Offline JS port of the deployer (no API calls)
├── app/                  # Multi-page app (clean routes, live data from the API)
│   ├── chains.html        # /chains    — ecosystem home (grid + stats + search)
│   ├── dashboard.html     # /dashboard — per-chain console (?id=…)
│   └── admin.html         # /admin     — members, audit log, targets, danger zone
├── design/               # Research + design system + Stitch/AI Studio handoff
│   ├── competitor-research.md
│   ├── design-system.css
│   └── stitch-aistudio-handoff.md
├── mockups/index.html    # Static design gallery (all surfaces, switchable)
└── backend/
    ├── __init__.py
    └── deployer.py        # Validates requests and emits real chain artifacts
```

### Routes (clean URLs → pages)

| Route | Page |
|-------|------|
| `/` or `/deploy` | Deployment wizard (`index.html`) |
| `/chains` | Ecosystem home — live grid of deployments |
| `/dashboard` (`?id=…`) | Per-chain console — metrics, endpoints, artifacts |
| `/wallet` | **SKYMETRIC Wallet** — send/receive SKY, manage agents |
| `/admin` | Admin panel — members, audit log, targets |

## 👛 SKYMETRIC Wallet

A browser-native Cosmos wallet, embedded in the Studio at `/wallet`. Real
cryptography (no shortcuts):

| Layer | Implementation |
|-------|----------------|
| Mnemonic | 24-word BIP39 (192-bit entropy) |
| Seed | PBKDF2-HMAC-SHA512, 2048 rounds |
| HD path | BIP32, Cosmos default `m/44'/118'/0'/0/0` |
| Signing | secp256k1 ECDSA (low-s) via `@noble/secp256k1@2.1.0` |
| Address | `bech32("agentic", RIPEMD-160(SHA-256(pubkey)))` |
| TX encoding | Hand-rolled protobuf (byte-exact match with CosmJS) |
| Storage | `localStorage`, AES-GCM (PBKDF2 100k rounds for key) |

**Verified interoperable** — bit-perfect match against CosmJS's
`DirectSecp256k1HdWallet` for address derivation and `MsgSend` / `TxBody`
encoding.

The wallet's crypto stack lives in three modules:
`assets/js/wallet-crypto.js` (keygen + signing), `assets/js/wallet-proto.js`
(protobuf encoder), and `assets/js/wallet-rpc.js` (REST client + sign
pipeline). The Send tab attempts a real broadcast against the selected
network's REST endpoint and falls back to "signed locally, RPC unreachable"
when offline.

### dApp integration — `window.skymetric`

Any web page can talk to the wallet by loading the provider script:

```html
<script src="https://your-studio-host/assets/js/skymetric-provider.js"></script>
<script>
  // Request connection (opens the wallet popup on first call)
  const { address, chainId } = await window.skymetric.connect();

  // Send tokens
  const result = await window.skymetric.signAndBroadcast({
    chainId: "skymetric-1",
    toAddress: "agentic1…",
    amount: [{ denom: "usky", amount: "1000000" }],
    memo: "hello from my dApp",
  });

  window.skymetric.on("disconnect", () => console.log("user locked wallet"));
</script>
```

The API is a small subset of Keplr's protocol so existing Cosmos dApps can
adapt with minimal changes.

### Browser extension

A Chrome MV3 extension version lives at
[`genesis/wallet/extension/`](../genesis/wallet/extension/) with the same
keygen + storage stack (using `chrome.storage.local` instead of
`localStorage`). Load it unpacked: chrome://extensions → Developer mode →
Load unpacked → select `genesis/wallet/extension/`.

### API

| Method | Endpoint                  | Description                          |
|--------|---------------------------|--------------------------------------|
| GET    | `/api/health`             | Liveness probe                       |
| GET    | `/api/targets`            | Supported deploy targets             |
| GET    | `/api/stats`              | Aggregate deployment stats           |
| GET    | `/api/deployments`        | List deployments (newest first)      |
| GET    | `/api/deployments/<id>`   | Single deployment record             |
| POST   | `/api/deploy`             | Validate + execute a deployment      |

#### Deploy request body

```json
{
  "chain_id": "skymetric-1",
  "moniker": "genesis-node",
  "target": "local",
  "total_supply_sky": 1000000000,
  "validators": 4,
  "max_validators": 100,
  "validator_stake_sky": 100000,
  "faucet_balance_sky": 500000,
  "inflation_min": 0.01,
  "inflation_max": 0.07,
  "goal_bonded": 0.67,
  "task_burn_fraction": 0.2,
  "slash_fraction_fraud": 0.5,
  "unbonding_days": 21,
  "authority_address": "agentic1…"
}
```

`authority_address` is optional — when present it's validated as an
`agentic1…` bech32 address and recorded as the chain's owner (the Studio's
`/chains` view can then filter "Only mine" by the connected wallet address).

## 🔐 Note on on-chain deployment

Booting an actual validator requires the compiled `skymetricd` binary
(`cd genesis/chain && make install`) plus funded keys and live RPC. The
Studio generates a validated genesis (`genesis-overrides.json`), a
single-node bootstrap script (`init-chain.sh`), and the `.env` — everything
up to the point where you run the node yourself. The final "start node" step
is surfaced as a follow-up command rather than executed in-browser.

The **wallet**, by contrast, performs real cryptography end-to-end: it signs
genuine secp256k1 transactions and will broadcast them to a live REST
endpoint when one is reachable (falling back to local-sign-only when not).
