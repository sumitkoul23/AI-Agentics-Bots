import { useAgentsByOperator, fmtGen } from "./hooks";

export interface AgentViewProps {
  address: string | undefined;
  restUrl: string;
}

export function Registry({ address, restUrl }: AgentViewProps) {
  const { data, isLoading, error } = useAgentsByOperator(restUrl, address);

  return (
    <section className="rounded-2xl border border-white/10 bg-ink/40 p-5">
      <header className="flex items-center justify-between mb-3">
        <h2 className="font-display text-base">Your Agents</h2>
        <a
          href="/agents/new"
          className="text-xs font-mono uppercase tracking-widest text-gen hover:underline"
        >
          + Register
        </a>
      </header>

      {!address && <Empty>Connect your wallet to view your agents.</Empty>}
      {address && isLoading && <Empty>Loading…</Empty>}
      {address && error && (
        <Empty>
          Couldn't reach the chain (<code className="font-mono">{restUrl}</code>).
          Check your network.
        </Empty>
      )}
      {address && data && data.length === 0 && (
        <Empty>
          No agents under this address. Register your first with{" "}
          <code className="font-mono">agenticd tx agentic register-agent</code>{" "}
          or from the DEX.
        </Empty>
      )}

      {data && data.length > 0 && (
        <ul className="divide-y divide-white/5">
          {data.map((a) => (
            <li
              key={a.operator + a.moniker}
              className="py-3 flex items-center justify-between gap-4"
            >
              <div className="min-w-0">
                <div className="truncate">{a.moniker}</div>
                <div className="text-xs text-ash font-mono mt-1">
                  stake {fmtGen(a.stake_ugen)} GEN · rep {a.reputation}
                </div>
              </div>
              <span
                className={`text-[10px] px-2 py-0.5 rounded-full ${
                  a.jailed
                    ? "bg-slash/10 text-slash"
                    : "bg-stake/10 text-stake"
                }`}
              >
                {a.jailed ? "JAILED" : "active"}
              </span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function Empty({ children }: { children: React.ReactNode }) {
  return <div className="text-ash text-sm py-6 text-center">{children}</div>;
}
