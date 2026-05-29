// /pools — list of every constant-product pool on x/agenticdex, sortable.
// In v0 this is a static skeleton; pools load from `/cosmos/agenticdex/v1/pools`
// once the chain REST endpoint is live.

export const dynamic = "force-static";

const stubPools = [
  { id: 1n, a: "SKY", b: "USDC", tvl: "—", vol24h: "—", apr: "—" },
  { id: 2n, a: "SKY", b: "OSMO", tvl: "—", vol24h: "—", apr: "—" },
  { id: 3n, a: "SKY", b: "ATOM", tvl: "—", vol24h: "—", apr: "—" },
];

export default function PoolsPage() {
  return (
    <section className="space-y-6">
      <header className="flex items-baseline justify-between">
        <h1 className="font-display text-3xl font-bold">Pools</h1>
        <a
          href="/pools/new"
          className="text-xs font-mono uppercase tracking-widest text-gen hover:underline"
        >
          + Create pool
        </a>
      </header>

      <div className="overflow-x-auto rounded-2xl border border-white/10 bg-ink/40">
        <table className="w-full text-sm">
          <thead className="text-ash text-xs uppercase tracking-widest">
            <tr className="border-b border-white/5">
              <th className="text-left p-4 font-mono">#</th>
              <th className="text-left p-4">Pair</th>
              <th className="text-right p-4">TVL</th>
              <th className="text-right p-4">24h volume</th>
              <th className="text-right p-4">APR</th>
            </tr>
          </thead>
          <tbody>
            {stubPools.map((p) => (
              <tr key={String(p.id)} className="border-b border-white/5 hover:bg-white/5">
                <td className="p-4 font-mono text-ash">{String(p.id)}</td>
                <td className="p-4">
                  <a href={`/pools/${p.id}`} className="hover:text-gen">
                    {p.a} <span className="text-ash">/</span> {p.b}
                  </a>
                </td>
                <td className="p-4 text-right font-mono">{p.tvl}</td>
                <td className="p-4 text-right font-mono">{p.vol24h}</td>
                <td className="p-4 text-right font-mono text-stake">{p.apr}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
