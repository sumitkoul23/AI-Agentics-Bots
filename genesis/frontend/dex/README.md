# SKYMETRIC DEX frontend

Next.js 14 static-export site for `dex.skymetric.dev`. Deploys to Cloudflare
Pages free tier in 90 seconds.

## Local dev

```bash
cd genesis/frontend/dex
npm install
npm run dev          # http://localhost:3000
```

## Production build (static)

```bash
npm run build        # produces out/ (static html + assets)
```

## Deploy to Cloudflare Pages

1. https://pages.cloudflare.com → "Create a project" → "Connect to Git"
2. Repo: `sumitkoul23/ai-agentics-bots`
3. Branch: `main` (or the genesis branch)
4. Build command: `cd genesis/frontend/dex && npm install && npm run build`
5. Build output: `genesis/frontend/dex/out`

## What's here vs what's next

| File | Status |
|---|---|
| `package.json`, `next.config.mjs`, `tailwind.config.ts` | ✅ ready |
| `app/layout.tsx`, `app/globals.css` | ✅ ready |
| `app/page.tsx` (Swap) | ✅ UI scaffold; `lib/tx.ts::swap` wires it to the chain |
| `app/pools/page.tsx` | ✅ static table; live data via REST in v1 |
| `app/portfolio/page.tsx` | ✅ empty-states; live data via REST in v1 |
| `components/SwapForm.tsx` | ✅ UI scaffold; sets local state, console-logs intent |
| `lib/chain.ts` | ✅ chain-registry config matching `app/config.go` |
| `lib/tx.ts` | ✅ MsgSwap encoder + signing helper |
| Wallet hookup (Cosmos Kit `<ChainProvider>`) | 🟡 next batch — provider config + actual signer flow |
| Pool detail page `/pools/[id]` | 🟡 next batch — add/remove liquidity forms |
| TradingView chart panel | 🟡 next batch — needs the price-history indexer |
| Agent-quote panel | 🟡 next batch — calls into `x/agentic.MsgCreateTask` |

## Why static export (no SSR)

- Cloudflare Pages free tier serves static assets globally, no cold starts.
- Every page is wallet-driven — no server-side rendering needed for any
  data that isn't already public on-chain.
- A static export means the entire frontend is auditable: zero hidden
  backend, every request goes either to the chain or to the user's wallet.
