# Keplr → Skymetric Wallet fork plan

> The v1 production wallet. Apache-2.0 fork of [Keplr Wallet](https://github.com/chainapsis/keplr-wallet).
> Inherits Keplr's audited cryptography, keyring, hardware-wallet support,
> Chrome-Web-Store-passing codebase. Adds: SKYMETRIC defaults + agent-economy
> view layer + our branding.

## Upstream reference

- Repo: `github.com/chainapsis/keplr-wallet`
- License: Apache 2.0 (✅ permits forking + rebranding)
- Stack: TypeScript + React + Tendermint signing libs
- Architecture: monorepo with `apps/extension`, `apps/mobile`, `packages/*`

## Fork checklist — week-by-week

### Week 1 — Fork + rebrand

- [ ] `git clone https://github.com/chainapsis/keplr-wallet keplr-fork`
- [ ] Replace every occurrence of "Keplr" / "keplr" / "KEPLR" with
      "Skymetric" / "agentic" / "SKYMETRIC" in:
      - `apps/extension/manifest.json` (`name`, `short_name`, `description`)
      - `apps/extension/src/i18n/**/*.json` (all locales)
      - `packages/extension/src/**/*.tsx` (component copy + JSX text)
      - `apps/mobile/**/*` (display names, splash screens)
- [ ] Swap `apps/extension/public/icons/` and `apps/mobile/assets/icons/`
      with our brand icon (`../extension/icons/icon.svg`, exported to the
      sizes Keplr expects: 16, 32, 48, 64, 128, 256, 512)
- [ ] Update `apps/extension/src/styles/colors.ts` to our palette:
      `--ink #0A0E1A`, `--gen #7DF9FF`, etc. (full palette in
      `genesis/growth/social/brand.md`)
- [ ] Replace the Keplr brand fonts with Space Grotesk + Inter +
      JetBrains Mono (already in our repo's brand system)
- [ ] Verify `pnpm build` produces a `dist/` that loads in Chrome
      developer mode

### Week 2 — Hardcode SKYMETRIC defaults

- [ ] Move Skymetric chain entry to first position in
      `packages/extension/src/chains/embed-chains.ts`
- [ ] Set SKYMETRIC as the default chain on first-run onboarding
- [ ] Pre-populate the chain-registry list with our preferred IBC
      neighbours (Cosmos Hub, Osmosis, Neutron) so users have something
      to send to from day one
- [ ] Update the default RPC/REST endpoints to our infra
      (`rpc.skymetric.dev`, `rest.skymetric.dev`)
- [ ] Add the SKY logo to `packages/extension/src/icons/chains/`

### Week 3 — Inject agent-economy views

Reuse the components in `../agent-views/`. Insert them into Keplr's
existing tab structure:

- [ ] New tab in the home panel: "Agents" (loads `agent-views/Registry.tsx`)
- [ ] New tab: "Tasks" (`agent-views/Tasks.tsx`)
- [ ] New tab: "Reputation" (`agent-views/Reputation.tsx`)
- [ ] New tab: "Streams" (`agent-views/Streams.tsx`)
- [ ] Wire each tab to the SKYMETRIC REST endpoints
      (`/agentic.v1.Query/Agents`, `/agentic.v1.Query/Tasks`, etc.)

### Week 4 — Onboarding rewrite

Keplr's onboarding ("Welcome / Create Wallet / Import Wallet") is good but
generic. Rewrite for the agent-operator persona:

- [ ] Replace welcome copy with: "The wallet for AI-agent operators."
- [ ] Add an "I'm here to operate an agent" path that:
      1. Creates the wallet as usual.
      2. Immediately surfaces the "Register your first agent" CTA on
         the dashboard.
- [ ] Add an "I'm here to use agents" path that:
      1. Creates the wallet as usual.
      2. Surfaces the task-creation flow + the agent leaderboard.

### Week 5 — External audit

- [ ] Engage one of: Trail of Bits, NCC Group, Cure53, OtterSec.
- [ ] Scope: the diff vs upstream Keplr (everything we changed). Not the
      full Keplr codebase — Chainapsis already paid for that audit.
- [ ] Expect $30–80k for a diff-scoped audit, 2–3 weeks turnaround.

### Week 6 — Publish

- [ ] Chrome Web Store listing — $5 developer fee one-time. Submit
      manifest, screenshots, privacy policy.
- [ ] Firefox Add-ons (AMO) — free.
- [ ] Submit the wallet to:
      - `cosmos/chain-registry` (so other wallets show SKYMETRIC)
      - `keplr-wallet/keplr-wallet` releases page (we are a public fork
        and should be discoverable)
      - The Cosmos Discord `#wallets` channel
      - `awesome-cosmos` list

## Diff-keeping policy

We want to keep our fork *as close to upstream Keplr as possible* so we
can pull security patches with minimal merge pain.

Rules:
1. Branding changes (strings, colors, icons) live in a separate
   `branding/` directory layered on top of upstream, not interleaved.
2. Agent-economy views live in `packages/extension/src/pages/agentic/`
   — a new path, not an edit of an existing file.
3. Default-chain hardcoding lives in `packages/extension/src/chains/
   agentic.ts` and is imported from upstream's chain-loader by a single
   pulled patch.
4. Every two weeks we rebase `main` onto upstream `master` and resolve.

## Why this approach beats from-scratch

| Concern | From-scratch | Fork Keplr |
|---|---|---|
| Cryptography correctness | Months of work + an audit before launch | Already audited |
| Keyring + seed phrase UX | 4–6 weeks | Done |
| Hardware wallet (Ledger/Trezor) | 6–8 weeks per device | Done |
| Chrome Web Store listing acceptance | High risk for new wallets | Inherited reputation helps |
| Sign-doc handling for every Cosmos message type | Long tail of edge cases | Already covers all of them |
| dApp connector (`window.keplr`) | Need new dApp ecosystem | Aliased — zero migration |
| Time to ship | 4–6 months | 4–6 weeks |
| Risk of catastrophic key bug | Very high | Very low |

## What we ship between v0 (this PR) and v1 (the fork)

In the meantime users sign with **Keplr or Leap directly via Cosmos Kit**.
That's already wired in `genesis/frontend/dex/`. The scaffold extension
in `../extension/` is the brand surface — visible during this window so
the agent-economy UX is settled before the fork lands.

## Naming + brand notes

- Public name: **Skymetric Wallet**
- Wordmark style: same as the rest of our brand — Space Grotesk bold,
  widely letter-spaced
- Tagline (every store listing): *"The wallet built for AI-agent
  operators."*
- We do **not** name it "Skymetric Keplr" or any other dependent name —
  the upstream is a license obligation to disclose, not a brand to
  share.
