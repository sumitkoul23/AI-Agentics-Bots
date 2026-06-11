/**
 * Chain — a thin client to a REAL, live blockchain node.
 *
 * Talks to a CometBFT RPC endpoint (/status, broadcast) and a Cosmos SDK REST
 * endpoint (auth, bank). Configured via env (CHAIN_RPC / CHAIN_REST); defaults
 * to a live public Cosmos Hub node so the site connects to a real chain out of
 * the box. Point these at your own `skymetricd` node to go fully sovereign.
 *
 * The backend proxies these so the browser never hits the node directly —
 * avoids CORS and keeps a single source of truth for the active node.
 */

const RPC  = (process.env.CHAIN_RPC  || "https://cosmos-rpc.publicnode.com").replace(/\/+$/, "");
const REST = (process.env.CHAIN_REST || "https://cosmos-rest.publicnode.com").replace(/\/+$/, "");
const CHAIN_ID    = process.env.CHAIN_ID    || "";          // optional override / display
const BECH32      = process.env.CHAIN_BECH32 || "agentic";  // address prefix
const COIN_DENOM  = process.env.CHAIN_DENOM  || "usky";
const COIN_SYMBOL = process.env.CHAIN_SYMBOL || "SKY";

const TIMEOUT_MS = 12_000;

async function fetchJson(url, opts = {}) {
  const ctrl = new AbortController();
  const t = setTimeout(() => ctrl.abort(), TIMEOUT_MS);
  try {
    const r = await fetch(url, { ...opts, signal: ctrl.signal });
    const body = await r.json().catch(() => ({}));
    if (!r.ok) {
      const msg = body?.message || body?.error || `${r.status} ${r.statusText}`;
      const e = new Error(msg);
      e.status = r.status;
      throw e;
    }
    return body;
  } finally {
    clearTimeout(t);
  }
}

export const chain = {
  config() {
    return { rpc: RPC, rest: REST, chain_id: CHAIN_ID, bech32_prefix: BECH32, denom: COIN_DENOM, symbol: COIN_SYMBOL };
  },

  /** Live node status from CometBFT /status — proves a real connection. */
  async status() {
    const j = await fetchJson(`${RPC}/status`);
    const r = j.result || j;
    const ni = r.node_info || {};
    const si = r.sync_info || {};
    return {
      ok: true,
      rpc: RPC,
      rest: REST,
      network: ni.network || CHAIN_ID || null,
      moniker: ni.moniker || null,
      node_version: ni.version || null,
      latest_block_height: si.latest_block_height || null,
      latest_block_time: si.latest_block_time || null,
      catching_up: si.catching_up ?? null,
    };
  },

  /** Bank balances for an address from the Cosmos REST API. */
  async balances(address) {
    const j = await fetchJson(`${REST}/cosmos/bank/v1beta1/balances/${encodeURIComponent(address)}`);
    return j.balances || [];
  },

  /** Auth account (account_number, sequence) — needed to build a signed tx. */
  async account(address) {
    try {
      const j = await fetchJson(`${REST}/cosmos/auth/v1beta1/accounts/${encodeURIComponent(address)}`);
      const a = j.account || {};
      return { account_number: a.account_number ?? "0", sequence: a.sequence ?? "0", exists: true };
    } catch (e) {
      if (e.status === 404) return { account_number: "0", sequence: "0", exists: false };
      throw e;
    }
  },

  /** Broadcast a signed TxRaw (base64) through the node's REST endpoint. */
  async broadcast(txBytesB64, mode = "BROADCAST_MODE_SYNC") {
    const j = await fetchJson(`${REST}/cosmos/tx/v1beta1/txs`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tx_bytes: txBytesB64, mode }),
    });
    return j.tx_response || j;
  },
};
