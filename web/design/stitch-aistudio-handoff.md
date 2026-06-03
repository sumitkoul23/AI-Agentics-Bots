# Stitch → AI Studio — Handoff Prompts (SKYMETRIC Studio)

Copy-paste workflow. **Step 1** briefs Google Stitch to generate the UI.
**Step 2** carries those designs into Google AI Studio to turn them into
production code. Everything is grounded in `competitor-research.md` and
`design-system.css`.

---

## STEP 1 — Paste into Google Stitch (one prompt per screen)

Stitch works best one screen at a time. Set mode to **Web** for the first four,
**Mobile** for the last two. Reuse the same Design/Theme block each time so the
system stays consistent.

### 🎨 Shared theme block (prepend to every Stitch prompt)

```
THEME — "Cosmic" dark, professional, web3.
Background: near-black #06070f with subtle radial violet (#1a1740) top-right and
cyan (#0b2740) top-left aurora glows. Surfaces: translucent dark panels
(rgba(20,23,43,.72)) with 14px backdrop blur, 1px border rgba(130,140,200,.16),
16px radius, soft shadow. Primary action = violet→cyan gradient (#7c5cff→#18d3ff).
Status colors: live=green #3ddc97, deploying=amber #ffb454, error=red #ff6b81.
Text: #e8eafc primary, #a3a8c8 dim, #767ba0 muted. Fonts: Inter (UI), JetBrains
Mono (addresses/code). Pill-shaped status chips with a glowing dot. WCAG 4.5:1
contrast. Rounded cards, generous spacing, tasteful gradients — not flashy.
Product: SKYMETRIC — a tool to deploy a Cosmos-SDK / CometBFT app-chain (coin SKY).
```

### Screen 1 — Deploy Wizard (Web)
```
[THEME BLOCK]
Design a 5-step chain-deployment wizard, desktop, two columns.
Left column: a pill stepper (Identity · Tokenomics · Validators · Target ·
Review) above a form card. Show STEP 3 "Validators": a number stepper for
"Genesis validators" (value 4), inputs for "Validator stake (SKY)" 100000 and
"Faucet balance (SKY)" 500000, a "Fraud slash · 50%" slider, and an info callout
"Bonded at genesis: 400,000 SKY". Bottom: secondary "Back" + gradient "Next →".
Right column: a live "Deployment console" card titled with a status chip, showing
an animated step log (validate config, generate genesis overrides, create genesis
accounts…) with check/spinner icons. Every field pre-filled with sane defaults.
```

### Screen 2 — Chains Home / Ecosystem (Web)
```
[THEME BLOCK]
Design an ecosystem home. Top bar: logo "SKYMETRIC Studio", search field,
"Connect Keplr" ghost button, gradient "+ New chain", avatar. Heading "Your
chains" with a segmented filter (All / Live / Devnet / Draft). A 3-card grid of
chain cards — each card: chain icon + id (e.g. skymetric-1), status chip
(live/deploying/draft), one line "Mainnet · SKY · 4 validators", a "Bonded 67%"
progress bar. Below: a 4-up stat row (Chains, Validators, SKY planned, Uptime).
```

### Screen 3 — Chain Dashboard (Web)
```
[THEME BLOCK]
Design a post-launch chain dashboard with a left sidebar (Overview, Validators,
Endpoints, Economy, Governance, "← All chains"). Main: title "skymetric-1" with a
green "live" chip and a mono chain-id, plus "Explorer" and "Faucet" buttons.
A 4-up metric grid (Block height 1,284,902; TPS 42; Bonded 67%; Inflation 6.2%).
Below: a 2/3 "Block production" card with a 24H bar chart + 1H/24H/7D toggle, and
a 1/3 "Treasury" card (10 SKY, burn note, progress bar). Bottom: an "Endpoints"
card listing RPC :26657, REST :1317, gRPC :9090 each with a copy button.
```

### Screen 4 — Admin Panel (Web)
```
[THEME BLOCK]
Design an admin panel, left sidebar (Members, Deployments, Targets, API keys,
Settings). Main: "Members & roles" with "+ Invite" and a table (avatar+name+email,
role chip Owner/Maintainer/Viewer, chains count, last active, ⋯). Below: a
"Deployment audit log" table (chain, action, by, target, when, status chip). Then
a 4-up "Connected targets" row (Local, Docker, Fly.io, Oracle) each with an ok/link
chip. Finally a red-bordered "Danger zone" card with a "Teardown…" button.
```

### Screen 5 — Mobile: Wizard step (Mobile)
```
[THEME BLOCK]
Design the mobile deploy-wizard, "Step 3 of 5 · Validators". Top: title + thin
progress bar at 60%. A number stepper (− 4 +) for validators, "Validator stake"
input 100000, a "Fraud slash · 50%" slider, and a violet-tinted callout "Bonded at
genesis: 400,000 SKY". Sticky bottom bar: small "Back" + wide gradient "Next →".
Thumb-reachable primary action.
```

### Screen 6 — Mobile: Chain dashboard (Mobile)
```
[THEME BLOCK]
Design the mobile chain dashboard. Header: "skymetric-1" + green live chip. A 2×2
metric grid (Height 1.28M, TPS 42, Bonded 67%, Inflation 6.2%). A small "Block
production · 24H" card with a mini bar chart. A copyable "RPC :26657" row. Fixed
bottom tab bar: Home, Vals, RPC, Econ, More.
```

---

## STEP 2 — Carry the designs from Stitch into Google AI Studio

Stitch lets you **export to code** or copy the design. Two routes:

**Route A — Figma bridge (cleanest):** In Stitch, click **Export → Figma** (paste
to Figma), then in AI Studio use the prompt below referencing the screenshots.

**Route B — Direct (fastest):** In Stitch, **copy the generated HTML/CSS** for
each screen (Stitch shows a Code panel). Then paste into AI Studio with this:

### Paste into Google AI Studio (Gemini)
```
You are turning Stitch UI designs into a production front-end for "SKYMETRIC
Studio", a tool to deploy a Cosmos-SDK / CometBFT app-chain.

Inputs: I will paste, one screen at a time, the HTML/CSS Stitch generated (and/or
a screenshot). For each screen, produce a clean, responsive React + Tailwind
component that:
- matches the cosmic dark theme exactly (violet #7c5cff → cyan #18d3ff gradient,
  near-black #06070f bg, translucent blurred panels, pill status chips);
- uses semantic, accessible markup (WCAG 4.5:1), keyboard-navigable;
- is fully responsive (desktop 2-col → mobile single-col with bottom tab bar);
- wires to this JSON API (already built): GET /api/targets, GET /api/stats,
  GET /api/deployments, GET /api/deployments/:id, POST /api/deploy.
  POST /api/deploy body: {chain_id, moniker, target, total_supply_sky, validators,
  max_validators, validator_stake_sky, faucet_balance_sky, inflation_min,
  inflation_max, goal_bonded, task_burn_fraction, slash_fraction_fraud,
  unbonding_days}. Response: {id, status, summary, artifacts:{genesis_overrides_json,
  init_chain_sh, env_file, deploy_command, rpc, rest}}.
- keeps copy/download actions for the generated artifacts.

Screens to generate: Deploy Wizard (5 steps), Chains Home, Chain Dashboard,
Admin Panel, plus mobile variants. Start with the Deploy Wizard. Output one
component file per screen, plus a shared theme/tokens file. Ask me for the next
screen's Stitch export when ready.
```

### Optional — also bring the artifacts back into this repo
When AI Studio gives you the React/Tailwind code, drop it into a new
`web/app/` folder (or paste it back to me) and I'll wire it to the live
`web/server.py` API, validate it, and commit it to PR #11.

---

## Reference values (so designs match the real product)

| Thing | Value |
|---|---|
| Chain framework | Cosmos SDK v0.50 + CometBFT |
| Coin / denom | SKY / `usky` (1 SKY = 1,000,000 usky) |
| Default chain id | `skymetric-1` |
| Default validators | 4 (max 100) |
| Default stake | 100,000 SKY → 400,000 bonded |
| Inflation | 1–7% (taper), goal bonded 67% |
| Task burn | 20% (50/30 agent/validator split of remainder) |
| Deploy targets | Local · Docker · Fly.io · Oracle · Codespaces |
| Endpoints | RPC :26657 · REST :1317 · gRPC :9090 |
