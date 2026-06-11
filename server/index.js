/**
 * Chain Deployment Studio — real Node.js backend.
 *
 *  • Serves the existing web/ front-end (clean-URL routes for the multi-page app)
 *  • JSON API for deployments (durable JSON-file store)
 *  • LIVE chain proxy: /api/chain/* forwards to a real CometBFT RPC + Cosmos REST
 *    node (configured via env; defaults to a live public Cosmos Hub node)
 *
 *  Run:   npm install && npm start      ->  http://localhost:8000
 *  Env:   PORT, CHAIN_RPC, CHAIN_REST, CHAIN_ID, CHAIN_BECH32, CHAIN_DENOM, CHAIN_SYMBOL
 */
import express from "express";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

import { buildRecord, TARGETS, DeploymentError } from "./lib/deployer.js";
import { Store } from "./lib/store.js";
import { chain } from "./lib/chain.js";

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEB_ROOT = join(__dirname, "..", "web");
const DATA_FILE = process.env.DATA_FILE || join(__dirname, "data", "deployments.json");
const PORT = parseInt(process.env.PORT || "8000", 10);

const store = new Store(DATA_FILE);
const app = express();
app.disable("x-powered-by");
app.use(express.json({ limit: "256kb" }));

// ── Clean-URL routes for the multi-page app ────────────────────────────
const APP_ROUTES = {
  "/": "index.html",
  "/deploy": "index.html",
  "/chains": "app/chains.html",
  "/dashboard": "app/dashboard.html",
  "/admin": "app/admin.html",
  "/wallet": "app/wallet.html",
  "/dapp-demo": "app/dapp-demo.html",
};
for (const [route, file] of Object.entries(APP_ROUTES)) {
  app.get(route, (_req, res) => res.sendFile(join(WEB_ROOT, file)));
}

// ── Deployment API ─────────────────────────────────────────────────────
app.get("/api/health", (_req, res) => res.json({ status: "ok", backend: "node", chain: chain.config() }));
app.get("/api/targets", (_req, res) => res.json({ targets: TARGETS.map(({ command, ...t }) => t) }));
app.get("/api/stats", (_req, res) => res.json(store.stats()));
app.get("/api/deployments", (_req, res) => res.json({ deployments: store.all() }));
app.get("/api/deployments/:id", (req, res) => {
  const d = store.get(req.params.id);
  if (!d) return res.status(404).json({ error: "Deployment not found." });
  res.json(d);
});
app.post("/api/deploy", (req, res) => {
  try {
    const cfg = chain.config();
    const record = buildRecord(req.body || {}, { rpc: cfg.rpc, rest: cfg.rest });
    store.add(record);
    res.status(201).json(record);
  } catch (e) {
    if (e instanceof DeploymentError) return res.status(400).json({ error: e.message });
    console.error("deploy error:", e);
    res.status(500).json({ error: "Internal error while building the deployment." });
  }
});

// ── LIVE chain proxy (real node) ───────────────────────────────────────
app.get("/api/chain/config", (_req, res) => res.json(chain.config()));
app.get("/api/chain/status", async (_req, res) => {
  try { res.json(await chain.status()); }
  catch (e) { res.status(502).json({ ok: false, error: e.message, rpc: chain.config().rpc }); }
});
app.get("/api/chain/balances/:address", async (req, res) => {
  try { res.json({ balances: await chain.balances(req.params.address) }); }
  catch (e) { res.status(502).json({ error: e.message }); }
});
app.get("/api/chain/account/:address", async (req, res) => {
  try { res.json(await chain.account(req.params.address)); }
  catch (e) { res.status(502).json({ error: e.message }); }
});
app.post("/api/chain/broadcast", async (req, res) => {
  const { tx_bytes, mode } = req.body || {};
  if (!tx_bytes) return res.status(400).json({ error: "tx_bytes (base64) is required." });
  try { res.json(await chain.broadcast(tx_bytes, mode)); }
  catch (e) { res.status(502).json({ error: e.message }); }
});

// ── Static assets (CSS/JS/design) ──────────────────────────────────────
app.use(express.static(WEB_ROOT, { extensions: ["html"], index: false }));

// ── 404 ────────────────────────────────────────────────────────────────
app.use((req, res) => {
  if (req.path.startsWith("/api/")) return res.status(404).json({ error: "Not found." });
  res.status(404).sendFile(join(WEB_ROOT, "index.html"));
});

app.listen(PORT, () => {
  const c = chain.config();
  console.log(`Chain Deployment Studio (Node) — http://localhost:${PORT}`);
  console.log(`Live chain node: RPC ${c.rpc} · REST ${c.rest}`);
});
