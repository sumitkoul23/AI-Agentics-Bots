# Chain Deployment Studio — Competitor Design Research & UX Teardown

> Foundation document for the GUI/UX redesign of the SKYMETRIC Chain Deployment
> Studio (web app, mobile app, and admin panel). Researched June 2026.
> Everything below feeds the design system, the Figma frames, and the coded
> mockups that follow.

---

## 1. Who we benchmarked

The "deploy your own chain" product category splits into two lineages. We pulled
UX patterns from both because SKYMETRIC sits in the middle (a Cosmos-SDK
sovereign chain with an agent economy).

| Lineage | Products studied | Why relevant |
|---|---|---|
| **Rollup-as-a-Service (EVM)** | Caldera, Conduit, AltLayer, Gelato, Dymension, Arbitrum Orbit | Best-in-class **no-code deploy wizards** + post-launch **chain dashboards** |
| **Cosmos / appchain L1** | Ignite CLI, Spawn, Initia, Avalanche CLI/Cogitus | Closest to SKYMETRIC's actual stack (Cosmos SDK + CometBFT, validators, staking) |

---

## 2. Per-product UX teardown

### 2.1 Arbitrum Orbit — "form pre-filled with sane defaults"
- **Pattern:** a single long configuration **form pre-filled with default
  values**; users escape a shared environment to get "dedicated throughput,
  predictable fees, superior UX."
- **Fields exposed:** chain ID, validator/sequencer set, gas token, data
  availability mode, challenge/security parameters.
- **Takeaway for us:** *defaults are the product.* Never show an empty field.
  Our wizard already pre-fills `skymetric-1`, 4 validators, 1–7% inflation — keep
  that and make the "why this default" legible via inline hints.

### 2.2 Caldera — "deploy, then a real telemetry console"
- **Deploy:** choose framework → configure params → Caldera handles sequencing,
  DA routing, bridge UI, monitoring.
- **Post-launch dashboard exposes:** block-production metrics, sequencer
  latency, L1 submission cost, bridge activity; every chain ships an **RPC node,
  block explorer, data indexer, faucet, bridge UI**.
- **Takeaway:** the deploy flow is only half the product. The **post-launch
  dashboard** (status, metrics, endpoints, explorer, faucet) is what users live
  in. This justifies our "deploy wizard + dashboard + admin" scope.

### 2.3 AltLayer — "no-code dashboard, minutes to launch"
- **Pattern:** explicit **no-code dashboard** for non-programmers; launch in
  minutes; customize sequencer count, gas limit, network + chain-level params.
- **Takeaway:** a non-developer must be able to finish without docs. Plain-
  language labels, not Cosmos jargon ("Validators" not "CometBFT validator set
  cardinality").

### 2.4 Conduit — "self-serve, enterprise reliability, metrics built in"
- Self-serve OP-Stack deploys; **performance metrics to monitor your chain**
  baked into the console.
- **Takeaway:** monitoring is table stakes; the admin panel needs live charts.

### 2.5 Dymension — "RollApp Dev Kit + ecosystem view"
- Modular PoS ecosystem; **RDK** toolkit; a portal that lists RollApps.
- **Takeaway:** a **chain list / ecosystem home** is expected once a user has >1
  chain. Include it in the full product surface.

### 2.6 Initia — "wallet-first onboarding, username as identity"
- Onboarding = connect wallet → fund account (INIT or USDC gas) → **register a
  username (ENS-style identity)**. App tabs: Swap, Stake, Governance. Wallets:
  Keplr, Leap, Compass; "Connect Wallet" top-right.
- **Takeaway:** Cosmos users expect **Connect Wallet (Keplr/Leap) top-right**,
  and **Stake / Governance** as first-class destinations. Our admin panel should
  speak this dialect.

### 2.7 Ignite CLI & Spawn — "wizard prompts, choose your modules"
- Spawn = "custom network in a few clicks"; prompts for **consensus type
  (PoA / PoS / interchain-security)** and **module selection**.
- **Takeaway:** a **module/feature picker** step is a known pattern. Our agentic
  economy params (fee split, fraud slash, burn) map naturally onto this.

### 2.8 Avalanche CLI / Cogitus (Zeeve) — "wizard asks VM, validators, chain ID, gas token"
- Wizard asks **VM type, validator model, chain ID, gas token**; Cogitus =
  "no-code, zero learning curve, production-optimized."
- **Takeaway:** validates our exact step order (identity → token → validators →
  target).

---

## 3. Cross-product pattern synthesis

What virtually every winner does, distilled into rules we will design to:

| # | Pattern | How we apply it |
|---|---|---|
| 1 | **Defaults over blanks** | Every field pre-filled; hints explain the default |
| 2 | **Progressive disclosure** | Beginner view (5 steps) + "Advanced" reveals raw params |
| 3 | **Deploy → live console** | Animated step log → result with copyable artifacts (we have this) |
| 4 | **Post-launch dashboard** | Status, validators, RPC/REST, explorer, faucet, metrics |
| 5 | **Connect Wallet top-right** | Keplr / Leap affordance in the admin shell |
| 6 | **Stake + Governance as destinations** | First-class nav in admin panel |
| 7 | **Chain list / ecosystem home** | Grid of deployed chains with status chips |
| 8 | **Metrics = table stakes** | Block time, TPS, bonded %, inflation, treasury charts |
| 9 | **Plain language, jargon on hover** | "Validators", tooltip explains gentx |
| 10 | **Dark, saturated, gradient web3 aesthetic** | Keep our cosmic theme; WCAG 4.5:1 |
| 11 | **Trust signals** | Show what's simulated vs. on-chain; explicit warnings on mainnet |
| 12 | **One-click recovery on errors** | Failed step → retry with adjusted settings |

---

## 4. Design principles (our north star)

1. **Defaults are the product.** A non-developer reaches "deploy" without docs.
2. **Two altitudes.** Beginner (guided, 5 steps) and Advanced (every param,
   raw `genesis-overrides.json` editable) — toggle, never a fork.
3. **The dashboard is where users live.** Treat post-launch as the main app, the
   wizard as the on-ramp.
4. **Honest by design.** Clearly mark what is generated vs. broadcast on-chain
   (SKYMETRIC needs the compiled `skymetricd` binary to truly boot — we never
   imply otherwise).
5. **Responsive per device, not a shrunk desktop.** Mobile = bottom-sheet wizard
   + thumb-reachable primary action; admin charts reflow to stacked cards.
6. **Accessible cosmic theme.** Dark, gradient, but 4.5:1 text contrast and full
   keyboard nav.

---

## 5. Information architecture (full product surface + admin)

```
SKYMETRIC Studio
├─ Landing (marketing)            public
├─ Deploy Wizard                  Identity → Tokenomics → Validators → Target → Review
│   └─ Advanced toggle            raw params + live genesis-overrides.json editor
├─ Deployment Console             animated steps → artifacts (copy / download)
├─ Chains (ecosystem home)        grid of chains, status chips, search
├─ Chain Dashboard  ┐
│   ├─ Overview      │ status, block time, TPS, bonded %, inflation, treasury
│   ├─ Validators    │ set, stake, commission, uptime, jail status
│   ├─ Endpoints     │ RPC / REST / gRPC, explorer, faucet, copy buttons
│   ├─ Economy       │ supply, inflation curve, task burn, fee split
│   └─ Governance    │ proposals, voting, params
└─ Admin Panel  ┐
    ├─ Org / Members      roles, invites, API keys
    ├─ Deployments        audit log, redeploy, teardown
    ├─ Billing / Targets  connected targets (local/docker/fly/oracle/codespaces)
    └─ Settings           theme, defaults, danger zone
```

---

## 6. Device matrix (responsive intent)

| Surface | Desktop (≥1024px) | Tablet (640–1024) | Mobile (<640) |
|---|---|---|---|
| Wizard | 2-col: form + live console | stacked, console below | full-screen steps, sticky CTA, bottom-sheet review |
| Chains home | 3–4 card grid + sidebar | 2-col grid | 1-col list, FAB "New chain" |
| Chain dashboard | sidebar + metric grid | collapsible sidebar | tab bar, stacked metric cards |
| Admin | left nav + data tables | drawer nav + tables | drawer + card-ified tables |

---

## 7. What we will produce next (staged, each committed)

1. **Design system** — tokens (color, type, space, radius, elevation),
   components — as code (`web/design/design-system.*`) and mirrored into Figma.
2. **Figma frames** — web + mobile + admin, built from the design system via the
   Figma MCP (live, editable files).
3. **Coded mockups** — high-fidelity, responsive HTML screens in `web/mockups/`
   reusing the cosmic theme, runnable in any browser immediately.

---

## Sources

- [Exploring the RaaS Landscape — Nansen](https://research.nansen.ai/articles/exploring-the-rollups-as-a-service-raas-landscape)
- [The New Modular Rollup Toolkit — Gate Learn](https://www.gate.com/learn/articles/the-new-modular-rollup-toolkit/2091)
- [RaaS comparison: Caldera, AltLayer, Dymension, Eclipse — AiCoin](https://www.aicoin.com/en/article/378265)
- [What Is Caldera? — Eco](https://eco.com/support/en/articles/10370689-what-is-caldera-rollup-as-a-service-explained)
- [Rollup Tooling and Features — Conduit](https://www.conduit.xyz/features)
- [Launch a Chain — Arbitrum](https://arbitrum.io/launch-chain)
- [Overview of Arbitrum chains — Arbitrum Docs](https://docs.arbitrum.io/launch-orbit-chain/orbit-quickstart)
- [Meet Spawn — Interchain Stack](https://rollchains.github.io/spawn/v0.50/)
- [Ignite CLI](https://ignite.com/)
- [Getting Started with Initia — Bankless](https://www.bankless.com/read/getting-started-with-initia)
- [Initia App](https://app.initia.xyz/)
- [Avalanche L1s with Zeeve](https://www.zeeve.io/appchains/avalanche-l1s/)
- [Blockchain UX Best Practices — Aufait UX](https://www.aufaitux.com/blog/blockchain-ux-design-best-practices/)
- [Design and UX in web3 — ethereum.org](https://ethereum.org/developers/docs/design-and-ux/)
- [20 Dashboard UI/UX Principles for 2025 — Medium](https://medium.com/@allclonescript/20-best-dashboard-ui-ux-design-principles-you-need-in-2025-30b661f2f795)
