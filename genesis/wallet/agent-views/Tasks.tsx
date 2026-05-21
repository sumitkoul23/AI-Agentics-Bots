import { useTasksFor, fmtGen, type Task } from "./hooks";
import type { AgentViewProps } from "./Registry";

export function Tasks({ address, restUrl }: AgentViewProps) {
  const { data, isLoading, error } = useTasksFor(restUrl, address);

  return (
    <section className="rounded-2xl border border-white/10 bg-ink/40 p-5">
      <h2 className="font-display text-base mb-3">Tasks</h2>

      {!address && <Empty>Connect to view your task history.</Empty>}
      {address && isLoading && <Empty>Loading…</Empty>}
      {address && error && <Empty>Couldn't reach the chain.</Empty>}
      {address && data && data.length === 0 && (
        <Empty>
          No tasks. Create one from the DEX, or wait for incoming work as
          an agent operator.
        </Empty>
      )}

      {data && data.length > 0 && (
        <ul className="divide-y divide-white/5">
          {data.map((t) => (
            <li key={t.id} className="py-3">
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <div className="font-mono text-xs text-ash">
                    task #{t.id}
                  </div>
                  <div className="truncate text-sm mt-1">{t.spec}</div>
                </div>
                <Status task={t} self={address!} />
              </div>
              <div className="mt-1 text-xs text-ash font-mono">
                bounty {fmtGen(t.bounty_ugen)} GEN
                {t.response_cid && (
                  <>
                    {" · "}cid <span className="truncate">{t.response_cid.slice(0, 14)}…</span>
                  </>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

function Status({ task, self }: { task: Task; self: string }) {
  const role = task.requester === self ? "requester" : "agent";
  let label: string, tone: string;
  if (task.slashed) {
    label = "SLASHED";
    tone = "bg-slash/10 text-slash";
  } else if (task.settled) {
    label = "settled";
    tone = "bg-stake/10 text-stake";
  } else if (task.response_cid) {
    label = "awaiting settle";
    tone = "bg-gen/10 text-gen";
  } else {
    label = "open";
    tone = "bg-violet/10 text-violet";
  }
  return (
    <div className="flex flex-col items-end gap-1">
      <span className={`text-[10px] px-2 py-0.5 rounded-full ${tone}`}>
        {label}
      </span>
      <span className="text-[10px] text-ash font-mono">{role}</span>
    </div>
  );
}

function Empty({ children }: { children: React.ReactNode }) {
  return <div className="text-ash text-sm py-6 text-center">{children}</div>;
}
