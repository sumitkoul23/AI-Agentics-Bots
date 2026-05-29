// /portfolio — the user's positions across both x/agenticdex (LP shares)
// and x/agentic (agent operator earnings / reputation). Cross-references
// both keepers via REST in v1.

export const dynamic = "force-static";

export default function PortfolioPage() {
  return (
    <section className="space-y-8">
      <h1 className="font-display text-3xl font-bold">Portfolio</h1>

      <Section title="Liquidity positions" empty="Connect your wallet to view LP positions." />
      <Section title="Agent operator earnings" empty="No registered agents under this address. Register one via x/agentic on the testnet faucet." />
      <Section title="Reputation NFTs" empty="None yet. Reputation is earned through settled tasks — see docs/04-growth-strategy.md." />
    </section>
  );
}

function Section({ title, empty }: { title: string; empty: string }) {
  return (
    <div className="rounded-2xl border border-white/10 bg-ink/40 p-6">
      <h2 className="font-display text-lg mb-3">{title}</h2>
      <p className="text-ash text-sm">{empty}</p>
    </div>
  );
}
