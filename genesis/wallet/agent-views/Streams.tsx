import type { AgentViewProps } from "./Registry";

// Streams are financial-instrument #4 (genesis/docs/06-financial-instruments.md).
// The on-chain module doesn't ship until mainnet day 0; until then this
// component renders the explainer + a waitlist hook.

export function Streams({ address }: AgentViewProps) {
  return (
    <section className="rounded-2xl border border-white/10 bg-ink/40 p-5">
      <h2 className="font-display text-base mb-3">Streaming payments</h2>

      <p className="text-sm text-ash">
        Continuous money streams from a requester to an agent — pay an
        agent <code className="font-mono text-bone">0.01 GEN per minute</code>{" "}
        for retainer-style work without one-off task creation. Settles
        on every block tick.
      </p>

      <div className="mt-4 rounded-xl border border-white/10 bg-ink/60 p-4 text-xs text-ash">
        <div className="font-mono uppercase tracking-widest text-gen mb-1">
          Status
        </div>
        Streams land with the v0.5 chain release — the post-testnet
        upgrade. Once live, this panel shows your active subscriptions.
      </div>

      {address && (
        <div className="mt-3 text-xs text-ash">
          Your address is{" "}
          <code className="font-mono text-bone">{address}</code> —
          subscriptions will appear here automatically.
        </div>
      )}
    </section>
  );
}
