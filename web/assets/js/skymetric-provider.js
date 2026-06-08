/**
 * SKYMETRIC Provider — window.skymetric for dApps
 *
 * Any web page can `<script src="/assets/js/skymetric-provider.js">` and then call:
 *
 *   await window.skymetric.connect()
 *     -> { address, chainId, pubkey }
 *
 *   await window.skymetric.signAndBroadcast({ chainId, msgs, fee, memo })
 *     -> { txhash, code, raw_log }
 *
 *   await window.skymetric.getKey(chainId)
 *     -> { name, bech32Address, pubkey, algo }
 *
 *   skymetric.on("accountsChanged", (newAddr) => ...)
 *
 * Under the hood:
 *   - Opens /wallet?dapp=<origin> in a popup if not connected
 *   - postMessage protocol matches a small subset of Keplr's API
 *   - Each request gets a unique id; responses are matched by id
 */
(() => {
  "use strict";
  if (window.skymetric) return; // already injected

  const WALLET_URL = "/wallet?dapp=" + encodeURIComponent(location.origin);
  const PROVIDER_NAME = "SKYMETRIC";

  let walletWindow = null;
  let connectedAccount = null;
  const pending = new Map();    // id -> { resolve, reject }
  const listeners = { accountsChanged: [], chainChanged: [], disconnect: [] };
  let nextId = 1;
  let walletReady = false;
  const readyWaiters = [];   // resolved when the wallet posts skymetric:ready

  /* ── postMessage transport ───────────────────────────────────────── */
  function openWallet() {
    if (walletWindow && !walletWindow.closed) { walletWindow.focus(); return walletWindow; }
    walletReady = false;
    walletWindow = window.open(WALLET_URL, "skymetric-wallet", "width=480,height=720,popup=yes");
    return walletWindow;
  }

  /** Resolve once the wallet popup has announced it's ready (or time out). */
  function waitForReady(timeoutMs = 15000) {
    if (walletReady && walletWindow && !walletWindow.closed) return Promise.resolve();
    return new Promise((resolve, reject) => {
      const t = setTimeout(() => {
        const i = readyWaiters.indexOf(entry);
        if (i >= 0) readyWaiters.splice(i, 1);
        // Fall through anyway — older wallet builds may not send "ready".
        resolve();
      }, timeoutMs);
      const entry = () => { clearTimeout(t); resolve(); };
      readyWaiters.push(entry);
    });
  }

  function call(method, params = {}, opts = {}) {
    return new Promise(async (resolve, reject) => {
      const id = nextId++;
      pending.set(id, { resolve, reject });
      const needsUI = opts.needsUI ?? true;
      const w = needsUI ? openWallet() : walletWindow;
      if (!w) { pending.delete(id); return reject(new Error("Wallet is not open. Call connect() first.")); }

      try {
        if (needsUI) await waitForReady();
        w.postMessage({ type: "skymetric:request", id, method, params, origin: location.origin }, "*");
      } catch (e) {
        pending.delete(id);
        return reject(new Error("Wallet not reachable: " + e.message));
      }

      setTimeout(() => {
        if (pending.has(id)) { pending.delete(id); reject(new Error("Wallet request timed out")); }
      }, 60000);
    });
  }

  // Wallet announces readiness once its message handler is live.
  window.addEventListener("message", e => {
    if (e.data && e.data.type === "skymetric:ready") {
      walletReady = true;
      while (readyWaiters.length) { try { readyWaiters.shift()(); } catch {} }
    }
  });

  window.addEventListener("message", e => {
    const msg = e.data;
    if (!msg || msg.type !== "skymetric:response") return;
    const p = pending.get(msg.id);
    if (!p) return;
    pending.delete(msg.id);
    if (msg.error) p.reject(new Error(msg.error));
    else           p.resolve(msg.result);
  });

  window.addEventListener("message", e => {
    const msg = e.data;
    if (!msg || msg.type !== "skymetric:event") return;
    const list = listeners[msg.event] || [];
    if (msg.event === "accountsChanged") connectedAccount = msg.payload;
    list.forEach(fn => { try { fn(msg.payload); } catch {} });
  });

  /* ── Public API ──────────────────────────────────────────────────── */
  const skymetric = {
    name: PROVIDER_NAME,
    version: "0.1.0",

    isSkymetric: true,

    async connect() {
      const res = await call("connect");
      connectedAccount = res;
      return res;
    },

    async disconnect() {
      connectedAccount = null;
      try { walletWindow?.close(); } catch {}
      walletWindow = null;
    },

    /** Returns the connected key without prompting (returns null if not connected). */
    async getKey(chainId) {
      return call("getKey", { chainId }, { needsUI: false }).catch(() => null);
    },

    /**
     * Sign + broadcast a single transaction.
     * @param {object} args - { chainId, msgs: [{typeUrl, value}], fee, memo }
     */
    async signAndBroadcast(args) {
      return call("signAndBroadcast", args);
    },

    /** Sign-only (caller broadcasts). Returns TxRaw bytes (base64). */
    async sign(args) {
      return call("sign", args);
    },

    /** Suggest a chain — wallet may prompt the user to add it. */
    async experimentalSuggestChain(chainConfig) {
      return call("suggestChain", chainConfig);
    },

    /** Event subscription. Supported: accountsChanged, chainChanged, disconnect. */
    on(event, fn) {
      if (!listeners[event]) listeners[event] = [];
      listeners[event].push(fn);
    },
    off(event, fn) {
      if (!listeners[event]) return;
      listeners[event] = listeners[event].filter(f => f !== fn);
    },
  };

  Object.defineProperty(window, "skymetric", { value: skymetric, writable: false, configurable: false });

  // Announce to dApps that listen for an injection event (matches EIP-6963 pattern loosely)
  window.dispatchEvent(new Event("skymetric#initialized"));
})();
