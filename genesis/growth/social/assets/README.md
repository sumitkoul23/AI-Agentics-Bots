# SKYMETRIC creative assets

SVG masters. Edit the `<text>` nodes for copy variations; the brand system
(colors, type, mark) stays identical.

| File | Dimensions | Use |
|---|---|---|
| [`banner-x.svg`](banner-x.svg) | 1500 × 500 | X / Twitter header, Farcaster cover |
| [`banner-linkedin.svg`](banner-linkedin.svg) | 1128 × 191 | LinkedIn company-page cover |
| [`banner-youtube.svg`](banner-youtube.svg) | 2560 × 1440 | YouTube channel art |
| [`og-image.svg`](og-image.svg) | 1200 × 630 | Open Graph + Twitter card for `skymetric.dev` link previews |
| [`tweet-graphic.svg`](tweet-graphic.svg) | 1200 × 675 | Attachment for non-thread tweets |
| [`post-square.svg`](post-square.svg) | 1080 × 1080 | IG / TikTok / Farcaster carousel covers; **template — edit the three text zones** |

## Exporting to PNG

Most platforms accept SVG uploads directly. For the ones that don't
(Telegram, Discord, some legacy clients), export to PNG once:

```bash
# Inkscape (free, all platforms)
inkscape banner-x.svg --export-type=png --export-filename=banner-x.png

# Or rsvg-convert (installs via brew / apt)
rsvg-convert -h 500 banner-x.svg > banner-x.png

# Or Figma: drop the SVG onto a frame, hit Export PNG @1x.
```

## Editing copy

The `<text>` nodes use the brand fonts (Space Grotesk for display, Inter for
body, JetBrains Mono for code/numerics). Open in any text editor — the SVG
is hand-readable.

For `post-square.svg`, three editable zones are marked with `⬇ EDITABLE ZONE`
comments. Each new social post should keep zones 1 and 3 stable (the hook
shape and the CTA) and vary zone 2 (the mechanism callout).

## Variants we haven't generated yet (add as needed)

- Story / vertical 1080 × 1920 (IG, TikTok stories)
- Mainnet-launch hero with countdown
- Per-validator celebration card (issued automatically when a validator
  joins the active set)
- Burn-milestone card (issued automatically by the Cloudflare Worker that
  watches the on-chain burn counter)
