# Avatar + banner dimension spec

> Same brand system everywhere. The logo SVG ([`logo.svg`](logo.svg)) is the
> single source of truth — export at the dimensions below.

## Avatar (the square / circle profile pic)

| Platform | Pixel size | Format | Notes |
|---|---|---|---|
| X / Twitter | 400 × 400 | PNG | Crops to circle; keep mark centered in inner 80% |
| Telegram | 512 × 512 | JPG/PNG | Cropped to circle |
| Discord (server icon) | 512 × 512 | PNG / GIF | Square; rounded by client |
| Discord (server avatar) | 1024 × 1024 | PNG | If Boost Level 2+ |
| GitHub org | 500 × 500 | PNG | Square; rounded by GH |
| YouTube | 800 × 800 | PNG | Cropped to circle |
| LinkedIn (company logo) | 400 × 400 | PNG | Square |
| Farcaster | 512 × 512 | PNG | Cropped to circle |
| Medium | 500 × 500 | PNG | Cropped to circle |
| Mirror.xyz | 256 × 256 | PNG | Square |
| TikTok | 200 × 200 | PNG | Cropped to circle |
| Instagram | 320 × 320 | PNG | Cropped to circle |
| Reddit subreddit icon | 256 × 256 | PNG | Cropped to circle |

**Recipe:** open `logo.svg` in any SVG editor (Inkscape free, Figma free), set
the artboard to the target px size, export as PNG with `#0A0E1A` background.

## Banner / cover

Composition is identical across all platforms (rule: wordmark on the left
third, tagline beneath, faint chain-graph illustration on the right). Only
the canvas size changes:

| Platform | Pixel size | Safe-zone notes |
|---|---|---|
| X / Twitter header | 1500 × 500 | Avoid bottom ~80px (avatar overlaps) |
| Discord (server banner) | 960 × 540 | Only renders if Boost Level 2 |
| Discord (invite splash) | 1920 × 1080 | Only if Boost Level 1+ |
| YouTube banner | 2560 × 1440 | Safe area: 1546 × 423 centered |
| LinkedIn cover | 1128 × 191 | Mobile crops aggressively — keep critical content in center |
| Farcaster cover | 1500 × 500 | Same as X |
| GitHub org README banner | 1280 × 640 | Goes in the org's profile README |
| TikTok cover | Square 1080 × 1080 | TikTok shows a vertical crop on mobile |
| Reddit subreddit banner | 4000 × 256 (recommended), 1920 × 384 (mobile) | Cropped heavily on mobile |

## Open Graph / link preview cards

When someone pastes `skymetric.dev` or a tweet into Discord / Slack / Telegram,
this is what they see. **One image, used everywhere:**

| Slot | Size | Notes |
|---|---|---|
| `og:image` | 1200 × 630 | Reused as Twitter `summary_large_image` |
| `twitter:image:src` | 1200 × 675 | Same composition, taller crop |

## Mascot — optional, for v1+

Not in v0 scope. If we add a mascot later, it should be a stylised "agent
node" character, never a face — to keep the brand machine-coded rather than
personified.
