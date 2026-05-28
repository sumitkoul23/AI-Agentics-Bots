# SKYMETRIC brand system

## Colors

| Role | Hex | Use |
|---|---|---|
| **Primary — Skymetric Black** | `#0A0E1A` | Backgrounds, dark UI, dominant brand surface |
| **Accent — SKY Electric** | `#7DF9FF` | Primary CTAs, links, the highlight color in every graphic |
| **Secondary — Validator Violet** | `#A78BFA` | Validators, governance, "trust" surfaces |
| **Stake Green** | `#22C55E` | Positive on-chain actions: stake, settle, mint |
| **Slash Red** | `#EF4444` | Negative on-chain actions: burn, slash, jail |
| **Mono — Ash** | `#B0B7C3` | Body text on dark, dividers |
| **Mono — Bone** | `#F5F7FA` | Off-white backgrounds, light-mode UI |

Palette is consistent with the on-chain action vocabulary in `x/agentic` —
i.e. when someone sees Slash Red they should already associate it with the
slashing primitive.

## Typography

| Use | Family | Where to get it (free) |
|---|---|---|
| Display / wordmark | **Space Grotesk** | Google Fonts (OFL license) |
| Body | **Inter** | Google Fonts (OFL license) |
| Code / numerics | **JetBrains Mono** | JetBrains (free, open-source) |

Both display and body are deliberately geometric / readable — they read well
at 12pt on a phone Discord notification *and* at 200pt on a conference
banner.

## Voice

Three rules:

1. **Technical respect.** The audience builds chains and ships AI agents.
   Don't dumb things down; explain mechanisms.
2. **No hype words.** "Revolutionary," "game-changing," "to the moon," "the
   future of X" — all banned. The mechanism is the marketing.
3. **First-person plural ("we"), not "I".** Even though the Genesis System
   is autonomous, the public-facing identity is collective.

### Examples

| ❌ Don't | ✅ Do |
|---|---|
| "SKYMETRIC is the future of AI 🚀🚀🚀" | "SKYMETRIC settles AI-agent work on-chain. Code is in the repo." |
| "WAGMI 💎🙌 $SKY MOONING SOON 🌕" | "$SKY year-1 emission is ~70M. Burn rate at break-even is ~800k tasks/yr." |
| "Our incredible team is so excited to announce" | "We shipped the x/agentic keeper today. Diff is in PR #7." |
| "Don't miss the airdrop!!!" | "Airdrop snapshot is 2025-08-01. Eligibility is computable from public on-chain data." |

## Logo

Minimal vector mark — see [`logo.svg`](logo.svg). The mark is a circular
"agent node" with three radiating links — representing the agent ↔ user ↔
validator triangle that powers every task settlement.

Construction:
```
   ◯ ─── ◯
    \   /
     ◉      ← the "agent" node (filled, SKY Electric)
    /   \
   ◯ ─── ◯
```

Allowed variants:
- **Mono-light** — for dark backgrounds (default)
- **Mono-dark** — for light backgrounds
- **Wordmark** — `SKYMETRIC` set in Space Grotesk Bold + the mark, locked-up to the left
- **Coin glyph** — just the central filled node, used as the `SKY` token icon

Do not skew, recolor outside the palette, place on busy photographs, or add
drop shadows.

## Banner / cover image direction

Every banner shares the same composition: a dark gradient background (Skymetric
Black → Validator Violet at 30% opacity), the wordmark left-aligned at 1/3
horizontal, and one line of tagline (always from `bios.md`) in Ash.

Right-third of the banner: a faint chain-graph illustration in Stake Green at
8% opacity. No photographs, ever.

## Iconography

When we need icons (in docs, in the explorer, in social posts) we use
[Lucide](https://lucide.dev/) — free, MIT, 1,400+ icons, single SVG style
that matches the Inter/Space Grotesk pairing.

Specific assignments:
- `box` — block
- `link` — chain
- `flame` — burn
- `shield-check` — validator
- `bot` — agent
- `coin` — SKY
- `gavel` — governance / slashing
