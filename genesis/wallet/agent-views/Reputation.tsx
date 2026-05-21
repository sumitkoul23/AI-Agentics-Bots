import { useReputation } from "./hooks";
import type { AgentViewProps } from "./Registry";

export function Reputation({ address, restUrl }: AgentViewProps) {
  const { data, isLoading, error } = useReputation(restUrl, address);

  return (
    <section className="rounded-2xl border border-white/10 bg-ink/40 p-5">
      <h2 className="font-display text-base mb-3">Reputation</h2>

      {!address && <Empty>Connect to view reputation.</Empty>}
      {address && isLoading && <Empty>Loading…</Empty>}
      {address && error && <Empty>Couldn't reach the chain.</Empty>}

      {data && (
        <>
          <div className="grid grid-cols-3 gap-3">
            <Stat label="Best score" value={data.best_score} />
            <Stat label="Settled" value={data.total_settled} tone="text-stake" />
            <Stat label="Slashed" value={data.total_slashed} tone={data.total_slashed > 0 ? "text-slash" : "text-ash"} />
          </div>

          <div className="mt-4 text-xs text-ash">
            Reputation is soul-bound — non-transferable. High-rep agents
            need less stake per task (see{" "}
            <a
              className="underline decoration-gen"
              href="https://github.com/sumitkoul23/AI-Agentics-Bots/blob/main/genesis/docs/01-architecture.md"
            >
              docs/01-architecture
            </a>
            ).
          </div>

          {data.nfts.length > 0 && (
            <div className="mt-4">
              <h3 className="text-xs uppercase tracking-widest text-ash mb-2">
                Reputation NFTs
              </h3>
              <ul className="grid grid-cols-3 gap-2">
                {data.nfts.map((id) => (
                  <li
                    key={id}
                    className="aspect-square rounded-xl border border-white/10 bg-ink/60 grid place-items-center font-mono text-xs text-ash"
                  >
                    #{id}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </>
      )}
    </section>
  );
}

function Stat({
  label,
  value,
  tone = "text-bone",
}: {
  label: string;
  value: number;
  tone?: string;
}) {
  return (
    <div className="rounded-xl bg-ink/60 border border-white/10 p-3 text-center">
      <div className={`font-mono text-xl ${tone}`}>{value}</div>
      <div className="text-[10px] uppercase tracking-widest text-ash mt-1">
        {label}
      </div>
    </div>
  );
}

function Empty({ children }: { children: React.ReactNode }) {
  return <div className="text-ash text-sm py-6 text-center">{children}</div>;
}
