# Agentic Wallet

> The native wallet for the AGENTIC chain. Forked from Keplr (Apache 2.0),
> rebranded, with agent-economy UX layered on top.

| Folder | Purpose |
|---|---|
| [`extension/`](extension/) | Minimal MV3 browser-extension scaffold — manifest, popup, content script, background service worker. Demonstrates the dApp-injection model and the agent-economy view shell. |
| [`keplr-fork/`](keplr-fork/) | Fork plan + checklist: clone, rebrand, hardcode AGENTIC, ship to Chrome Web Store |
| [`agent-views/`](agent-views/) | React components for the agent-economy surfaces — registry, reputation, tasks, streams. Drop-in for the wallet popup and the DEX frontend. |

## Status

v0 — scaffold only. The extension popup loads, displays the agent-economy
shell, and exposes `window.agentic` to web pages. Real signing logic lands
in v1 (the Keplr fork) — see `genesis/docs/08-wallet-strategy.md`.

## Why this is structured as three folders

We're shipping the **brand surface** (extension scaffold) and the
**differentiated views** (agent-economy components) ahead of the heavy
crypto plumbing (the Keplr fork). That ordering is deliberate:

1. The brand surface is the cheapest thing to validate — a hundred
   developers see the popup before any signing risk exists.
2. The differentiated views are what makes "Agentic Wallet" worth
   installing over Keplr — those need to be ready *before* we ship the
   fork, otherwise we ship a worse Keplr.
3. The Keplr fork itself is mostly *configuration* once the views and
   brand are settled. Doing it last means the fork lands with zero
   throwaway UI work.

## Security posture

**Do not install or run this scaffold against real keys.** It is a
demonstration of structure, not a production wallet. The Keplr fork — once
it lands — inherits Keplr's audited cryptography. Until then, treat every
file in this folder as a UX prototype.

A line we will write into the v1 release notes and onto every distribution
page:

> Agentic Wallet is a fork of Keplr (Chainapsis, Apache 2.0). The
> cryptographic primitives are unchanged. Changes are: branding, default
> chain selection, and the agent-economy view layer. Diff is public at
> github.com/agentic-chain/wallet.
