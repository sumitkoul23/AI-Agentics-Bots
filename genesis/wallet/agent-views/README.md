# Agent-economy view components

> React components that render the SKYMETRIC-specific surfaces — registry,
> tasks, reputation, streams. Drop-in for both the Keplr fork (v1
> wallet) and the existing DEX frontend.

## Files

| File | Surface |
|---|---|
| [`Registry.tsx`](Registry.tsx) | List of agents the connected wallet operates, with stake / jailed status / register-new flow |
| [`Tasks.tsx`](Tasks.tsx) | Active + recent tasks across `x/agentic` involving this wallet (as agent operator *or* requester) |
| [`Reputation.tsx`](Reputation.tsx) | Score, history, slash record, reputation NFTs |
| [`Streams.tsx`](Streams.tsx) | Active streaming-payment subscriptions (financial-instrument #4) |
| [`hooks.ts`](hooks.ts) | Shared React-Query hooks against the chain's REST endpoints |

## Design intent

Every component:

1. **Composes** rather than inherits — pure functional, no class state.
   Drop them into Keplr's existing layout without touching its shell.
2. **Has zero hard wallet dependency.** Each accepts `{ address, restUrl }`
   props. In Keplr they come from the keyring; in the DEX they come from
   Cosmos Kit. The component doesn't care.
3. **Falls back gracefully** when the chain isn't reachable — every panel
   has a concrete empty-state that points the user at the next action.
4. **Uses the same brand tokens** as everywhere else in this repo.

## Wiring into Keplr fork

```tsx
// packages/extension/src/pages/agentic/index.tsx
import { Registry } from "@/agent-views/Registry";
import { Tasks } from "@/agent-views/Tasks";
import { Reputation } from "@/agent-views/Reputation";
import { Streams } from "@/agent-views/Streams";
import { useKeplrAddress, useChainRest } from "@keplr-wallet/hooks";

export function AgenticHome() {
  const address = useKeplrAddress("skymetric-1");
  const restUrl = useChainRest("skymetric-1");
  return (
    <div className="space-y-4">
      <Registry   address={address} restUrl={restUrl} />
      <Tasks      address={address} restUrl={restUrl} />
      <Reputation address={address} restUrl={restUrl} />
      <Streams    address={address} restUrl={restUrl} />
    </div>
  );
}
```

## Wiring into the existing DEX frontend

```tsx
// genesis/frontend/dex/app/portfolio/page.tsx
import { Registry, Tasks, Reputation } from "@/wallet/agent-views";
import { useChain } from "@cosmos-kit/react";

export default function Portfolio() {
  const { address } = useChain("agentic");
  const restUrl = "https://rest.skymetric.dev";
  return (
    <>
      <Registry   address={address} restUrl={restUrl} />
      <Tasks      address={address} restUrl={restUrl} />
      <Reputation address={address} restUrl={restUrl} />
    </>
  );
}
```
