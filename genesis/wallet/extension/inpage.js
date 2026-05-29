// inpage.js — runs in the page's MAIN world. Exposes `window.agentic`.
//
// The shape mirrors `window.keplr` so any dApp that already speaks the
// Keplr protocol works against us with zero changes. The v1 Keplr fork
// retains this exact interface — that's the migration story.

(function () {
  if (window.agentic) return;

  let nextId = 1;
  const pending = new Map();

  function send(payload) {
    return new Promise((resolve, reject) => {
      const id = nextId++;
      pending.set(id, { resolve, reject });
      window.postMessage(
        { target: "agentic-content", id, payload },
        "*"
      );
    });
  }

  window.addEventListener("message", (ev) => {
    if (ev.source !== window) return;
    const data = ev.data;
    if (!data || data.target !== "agentic-inpage") return;
    const cb = pending.get(data.id);
    if (!cb) return;
    pending.delete(data.id);
    if (data.response?.error) cb.reject(new Error(data.response.error));
    else cb.resolve(data.response);
  });

  const provider = {
    version: "0.0.1-scaffold",
    isAgentic: true,
    defaultOptions: {},

    async enable(chainId) {
      return send({ type: "enable", chainId });
    },

    async getKey(chainId) {
      return send({ type: "get-key", chainId });
    },

    async getChainInfo(chainId) {
      return send({ type: "get-chain", chainId });
    },

    async experimentalSuggestChain(chain) {
      return send({ type: "experimental-suggest-chain", chain });
    },

    async signDirect(chainId, signer, signDoc) {
      return send({ type: "sign-direct", chainId, signer, signDoc });
    },

    async signAmino(chainId, signer, signDoc) {
      return send({ type: "sign-amino", chainId, signer, signDoc });
    },
  };

  Object.defineProperty(window, "agentic", {
    value: provider,
    writable: false,
    configurable: false,
  });

  // Compatibility shim: until the broader Cosmos ecosystem auto-detects
  // `window.agentic`, also expose ourselves as a Keplr-compatible
  // fallback when no other Cosmos wallet is present.
  if (!window.keplr) {
    Object.defineProperty(window, "keplr", {
      value: provider,
      writable: false,
      configurable: false,
    });
  }
})();
