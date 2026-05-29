// background.js — MV3 service worker.
//
// In v0 this is a thin message router between the popup and content
// scripts. v1 (after the Keplr fork) replaces it with the real keyring +
// signing logic that Keplr's `background/keyring/` already implements.
//
// The chrome.runtime.onMessage protocol below is a deliberate subset of
// Keplr's so the fork lands as a drop-in.

const STATE = {
  defaultChainId: "skymetric-1",
  chains: {
    "skymetric-1": {
      chainId: "skymetric-1",
      chainName: "SKYMETRIC",
      rpc: "https://rpc.skymetric.dev",
      rest: "https://rest.skymetric.dev",
      bech32Prefix: "agentic",
      coinType: 118,
      stakeCurrency: { coinDenom: "SKY", coinMinimalDenom: "usky", coinDecimals: 6 },
    },
    // Users can chrome.storage-persist additional chains; v0 hardcodes
    // a couple of well-known Cosmos chains as defaults so the wallet
    // doesn't look single-chain at first launch.
    "cosmoshub-4": {
      chainId: "cosmoshub-4",
      chainName: "Cosmos Hub",
      rpc: "https://rpc.cosmos.network",
      rest: "https://lcd.cosmos.network",
      bech32Prefix: "cosmos",
      coinType: 118,
      stakeCurrency: { coinDenom: "ATOM", coinMinimalDenom: "uatom", coinDecimals: 6 },
    },
    "osmosis-1": {
      chainId: "osmosis-1",
      chainName: "Osmosis",
      rpc: "https://rpc.osmosis.zone",
      rest: "https://lcd.osmosis.zone",
      bech32Prefix: "osmo",
      coinType: 118,
      stakeCurrency: { coinDenom: "OSMO", coinMinimalDenom: "uosmo", coinDecimals: 6 },
    },
  },
};

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  switch (msg.type) {
    case "get-chain":
      sendResponse(STATE.chains[msg.chainId] ?? null);
      return;

    case "list-chains":
      sendResponse(Object.values(STATE.chains));
      return;

    case "experimental-suggest-chain":
      // dApps can suggest a chain. In v0 we just stash it; v1 routes
      // through the user-approval modal the Keplr fork inherits.
      if (msg.chain?.chainId) {
        STATE.chains[msg.chain.chainId] = msg.chain;
        chrome.storage.local.set({ chains: STATE.chains });
      }
      sendResponse({ ok: true });
      return;

    case "enable":
      // v0 stub: pretend the user already approved this origin.
      sendResponse({ ok: true });
      return;

    case "get-key":
      // v0 STUB — returns a hardcoded testnet address. The Keplr fork
      // replaces this with the real keyring lookup.
      sendResponse({
        name: "agent-operator",
        algo: "secp256k1",
        bech32Address: "agentic1scaffold0000000000000000000000000000000",
        pubKey: new Uint8Array(33),
      });
      return;

    case "sign-direct":
    case "sign-amino":
      // v0: emphatically refuse to sign anything. Real signing lands in
      // v1.
      sendResponse({
        error: "Skymetric Wallet v0 scaffold cannot sign. Install Keplr or Leap until the v1 fork ships.",
      });
      return;

    default:
      sendResponse({ error: `unknown message type: ${msg.type}` });
  }
});

// Hydrate any persisted chains on extension boot.
chrome.storage.local.get("chains").then(({ chains }) => {
  if (chains) Object.assign(STATE.chains, chains);
});
