# `agentic.dev` — landing page

Single-file static site. **Zero build step.** Drop into Cloudflare Pages,
connect to this repo, set the build output to `genesis/site/` — done.

## Deploy in 90 seconds

1. https://pages.cloudflare.com → "Create a project" → "Connect to Git"
2. Pick `sumitkoul23/ai-agentics-bots`
3. Branch: `claude/genesis-blockchain-agents-Q6BEg` (or whatever you merge it to)
4. Build command: *(blank)*
5. Build output directory: `genesis/site`
6. Click "Save and Deploy"

Cloudflare will give you a `*.pages.dev` URL immediately. Point
`agentic.dev` at it via the Cloudflare DNS dashboard once you register the
domain.

## Files

| File | Purpose |
|---|---|
| [`index.html`](index.html) | The whole site. Tailwind via CDN, no JS build. |
| [`og-image.svg`](og-image.svg) | Copy of `../growth/social/assets/og-image.svg` — referenced from `<meta>` |
| [`favicon.svg`](favicon.svg) | The mark, simplified for 32×32 |
| [`robots.txt`](robots.txt) | Standard allow-all |
| [`_headers`](_headers) | Cloudflare Pages security headers config |

## Editing copy

Open `index.html` — every editable string is in the obvious place. No
framework, no component tree to learn.

## Why not Vitepress / Nextra / Docusaurus?

Each one is a fine choice once we have ≥ 20 doc pages. At v0 we have four
markdown files in `genesis/docs/` that GitHub already renders beautifully —
the landing page just needs to link to them. A static site beats a
build-step site at this scale.

When we cross 20 pages, swap this folder for a Vitepress project that pulls
the markdown directly from `genesis/docs/`. The CI workflow goes:
`vitepress build → publish dist/ to Pages`.
