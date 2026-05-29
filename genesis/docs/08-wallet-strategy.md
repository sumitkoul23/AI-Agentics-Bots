# Wallet strategy — Skymetric Wallet

> What we use today, what we ship next, and why "build a multi-chain wallet"
> is the wrong way to frame the work.

## What we use right now

`genesis/frontend/dex/` is wired to **Cosmos Kit**, which is *not a wallet*
— it's a wallet adapter that auto-detects whichever wallet the user has
installed. The current adapter list:

| Wallet | Strength | Coverage |
|---|---|---|
| **Keplr** | The default Cosmos wallet. ~5M users, every major Cosmos chain. Built by Chainapsis. | Cosmos SDK chains + EVM bridge |
| **Leap** | Slicker UX, mobile-first, faster sign flow. Growing share. | Cosmos SDK + EVM |
| **Cosmostation** | Mature, validator-friendly. | Cosmos SDK only |
| **Ledger / Trezor** | Hardware. Routed through Cosmos Kit. | Whatever chain the adapter exposes |
| **WalletConnect v2** | Any mobile wallet that speaks WC. | Universal protocol layer |

This setup is **good enough to ship the testnet and the mainnet swap UI**.
We do not need our own wallet for the chain to function.

## Why we still build our own

Three reasons, in priority order:

1. **Distribution leverage.** Every chain that won the last cycle has a
   canonical wallet — Solana shipped Phantom, Cosmos shipped Keplr, Sui
   shipped Sui Wallet. The wallet is the most-visited surface in the
   ecosystem and the single largest top-of-funnel for new users.
2. **The agent-economy UX is unique.** Generic wallets show balances and
   stake. Our users — agent operators — need to see *agent registry*,
   *task queue*, *reputation*, *streaming-payment status*, *open
   perp positions*. None of those are first-class in Keplr; they all
   live three menus deep behind a `cosmos.tx` viewer if at all.
3. **Brand consolidation.** "Connect to Skymetric" reads as a single
   experience; "Connect to Skymetric chain in Keplr" reads as a
   third-party hop. Long-tail conversion difference is measurable.

## Why we fork Keplr instead of building from scratch

A from-scratch wallet at $0 budget is the project's single largest source
of *technical and brand suicide risk*. The honest cost stack of a
production wallet:

- Professional security audit: **$50–200k** (the *only* thing that matters).
- Phishing infra (canary domains, alert pipelines, browser-extension
  store signing chains): $5–20k setup.
- Chrome Web Store / Firefox AMO review queues — slow and unpredictable
  every release.
- Customer support burden — "I lost my keys" is the #1 ticket, every
  month, forever.
- Ongoing per-chain maintenance — ~2 eng-weeks per new chain.

Keplr already paid all of those costs. It's **Apache 2.0**, it's already
in the Chrome Web Store, it's already audited, the dApp injection model
(`window.keplr`) is already what every Cosmos dApp expects.

Forking Keplr gets us:
- The hard work of cryptography, key storage, and seed-phrase UX — done.
- The hard work of getting through Chrome Web Store review — half-done
  (we still need our own listing, but the codebase passes the bar).
- The hard work of supporting Ledger — done.
- The hard work of WalletConnect compatibility — done.

What we bring on top of that:
- Skymetric chain pre-registered + hardcoded as the default network.
- The agent-economy view layer (see `genesis/wallet/agent-views/`).
- Our branding, brand voice, and onboarding copy.

## Roadmap

```
Phase  Target              Deliverable
─────  ─────────────────   ──────────────────────────────────────────
v0     this PR             • Strategy doc (this file)
                           • Browser-extension MV3 scaffold (genesis/wallet/extension/)
                           • Agent-economy view components (genesis/wallet/agent-views/)
                           • Keplr fork plan (genesis/wallet/keplr-fork/README.md)
v0.5   testnet +14d        • Submit SKYMETRIC to cosmos/chain-registry
                             (Keplr / Leap / Cosmostation auto-detect us)
v1     mainnet +90d        • Public Skymetric Wallet — Keplr fork, rebranded,
                             agent views shipped, Chrome Web Store listing
                           • Mandatory: one external audit
v2     mainnet +180d       • PWA mobile companion + WalletConnect bridge
v3     year +1             • Native iOS / Android via React Native
v4     year +2             • Hardware wallet adapter (Ledger app for SKYMETRIC)
```

## What we do NOT do

- **Custodial mode.** No "we hold your keys for you." Custody = MTL =
  CEX territory (see `genesis/exchange/cex/`).
- **In-wallet fiat ramps.** Hand off to Moonpay / Transak via embedded
  widget; we don't touch USD.
- **DeFi auto-yield.** Every "auto-yield" wallet feature has cost users
  money historically. We surface yield *opportunities*, we don't
  auto-deploy capital on behalf of users.
- **Telegram bots that hold keys.** This is the fastest-growing failure
  mode in crypto right now. Educate users away from it.

## Anti-checklist (mistakes other chain wallets made)

- ❌ Slack-channel customer support that asks for seed phrases. Use
  ticketed-only support; train every operator to **never** ask for keys.
- ❌ Auto-update without signing-key verification. Auto-updates are
  fine; an *opaque* auto-update is a supply-chain attack waiting to
  happen.
- ❌ Closed-source. The wallet itself is the most security-critical
  software in the ecosystem and *must* be auditable by users.
- ❌ Mandatory analytics / telemetry. Off by default; opt-in only.
