# Skymetric Wallet — extension scaffold (v0)

> **This is a UX prototype, not a wallet.** It loads, displays the
> agent-economy popup, exposes `window.agentic` to web pages, and
> emphatically refuses to sign anything. Real signing lands when the
> Keplr fork ships — see `../keplr-fork/`.

## What works in v0

- MV3 manifest loads in Chrome/Brave/Edge/Firefox-with-MV3-support.
- Popup renders with the brand surface + 5 tabs (Assets, Agents, Tasks,
  Reputation, Streams).
- Content script injects `window.agentic` into every page; the in-page
  provider matches the Keplr API shape, so any Cosmos dApp that already
  speaks Keplr works against us.
- Background service worker maintains a chain registry; users can
  `experimentalSuggestChain(...)` from any dApp to add new chains.

## What doesn't work in v0

- **Signing.** Every `signDirect` / `signAmino` call returns an explicit
  refusal error. Use Keplr or Leap until v1.
- **Seed-phrase onboarding.** The address is hardcoded to a scaffold
  bech32. v1 ships the real keyring (inherited from Keplr's
  `background/keyring/`).
- **Real balances.** The popup pulls from a fixture array; v1 swaps for
  live REST queries.

## Local install (for testing the UX only)

1. Open Chrome → `chrome://extensions/`.
2. Toggle "Developer mode" (top right).
3. Click "Load unpacked" and pick this directory.
4. Pin the extension. Click the icon to open the popup.

Generate the missing PNG icons from `icons/icon.svg`:

```bash
# Inkscape
for size in 16 48 128; do
  inkscape icons/icon.svg --export-type=png \
    --export-filename="icons/icon-${size}.png" \
    --export-width=${size} --export-height=${size}
done

# Or rsvg-convert
for size in 16 48 128; do
  rsvg-convert -w ${size} -h ${size} icons/icon.svg > icons/icon-${size}.png
done
```

## Files

| File | What |
|---|---|
| `manifest.json` | MV3 manifest. Permissions kept minimal — `storage` + the two `host_permissions` for the chain's RPC/REST. |
| `background.js` | Service worker. Message-routes between popup and content script. Holds the chain registry. **No keyring in v0.** |
| `content-script.js` | Bridges the page world ↔ extension world. Pattern matches Keplr verbatim. |
| `inpage.js` | Runs in the page's MAIN world. Exposes `window.agentic`. Also self-aliases to `window.keplr` when no other Cosmos wallet is present (compatibility shim). |
| `popup/index.html` | Popup shell — tabs, balance card, action buttons. |
| `popup/popup.css` | Same brand tokens as the rest of the repo. |
| `popup/popup.js` | Renders the five panels from fixtures. |
| `icons/icon.svg` | Master icon — export to PNG at 16/48/128. |

## Why MV3 and not MV2

Manifest V2 will be removed from Chrome by January 2025. Every new wallet
ships MV3 from day one — service workers, no persistent background pages,
strict CSP. Keplr itself shipped MV3 in 2024; our fork inherits that.

## Why we expose `window.agentic` *and* alias to `window.keplr`

The Cosmos dApp ecosystem assumes `window.keplr` exists. The alias is the
no-friction migration path: dApps that haven't heard of us yet still
work, while dApps that do know about us can prefer `window.agentic` for
agent-economy-specific calls we'll add later (e.g.
`window.agentic.getAgentRecord()`).
