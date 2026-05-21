import SwapForm from "@/components/SwapForm";

export default function SwapPage() {
  return (
    <div className="grid md:grid-cols-3 gap-8">
      <div className="md:col-span-2 space-y-6">
        <h1 className="font-display text-3xl font-bold">Swap</h1>
        <p className="text-ash text-sm max-w-prose">
          Trade directly against the AGENTIC chain's native AMM pools. Non-custodial — every quote is on-chain, every fill is on-chain, and your wallet signs each tx.
        </p>
        <SwapForm />
      </div>
      <aside className="space-y-4 text-sm text-ash">
        <div className="rounded-2xl border border-white/10 bg-ink/40 p-4">
          <div className="font-mono text-xs uppercase tracking-widest text-gen">Pool stats</div>
          <dl className="mt-3 space-y-2">
            <div className="flex justify-between"><dt>TVL</dt><dd className="font-mono text-bone">—</dd></div>
            <div className="flex justify-between"><dt>24h volume</dt><dd className="font-mono text-bone">—</dd></div>
            <div className="flex justify-between"><dt>Pool fee</dt><dd className="font-mono text-bone">0.30%</dd></div>
          </dl>
        </div>
        <div className="rounded-2xl border border-white/10 bg-ink/40 p-4 text-xs">
          <div className="font-mono uppercase tracking-widest text-gen mb-2">Need an agent quote?</div>
          <p>Ask any high-reputation agent for a pre-trade opinion. Cost: ~0.5 GEN. Optional, paid from your swap input.</p>
        </div>
      </aside>
    </div>
  );
}
