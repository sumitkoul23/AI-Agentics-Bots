// Shared React-Query hooks against AGENTIC's REST endpoints.
// Each hook accepts an explicit `restUrl` so the components stay
// wallet-agnostic (Keplr fork vs Cosmos Kit vs anything else).

import { useQuery } from "@tanstack/react-query";

export interface AgentRecord {
  operator: string;
  moniker: string;
  endpoint: string;
  stake_ugen: string;
  reputation: number;
  jailed: boolean;
}

export interface Task {
  id: string;
  requester: string;
  agent: string;
  bounty_ugen: string;
  spec: string;
  response_cid: string;
  settled: boolean;
  slashed: boolean;
}

export function useAgentsByOperator(restUrl: string, operator: string | undefined) {
  return useQuery({
    queryKey: ["agents", restUrl, operator],
    enabled: Boolean(operator),
    queryFn: async (): Promise<AgentRecord[]> => {
      // v0: full-scan + client-side filter. The next chain release adds
      // a secondary index keyed by operator address so this becomes a
      // direct lookup.
      const r = await fetch(`${restUrl}/agentic/v1/agents`);
      if (!r.ok) throw new Error(`agents fetch: ${r.status}`);
      const { agents } = await r.json();
      return (agents as AgentRecord[]).filter((a) => a.operator === operator);
    },
    staleTime: 30_000,
  });
}

export function useTasksFor(restUrl: string, address: string | undefined) {
  return useQuery({
    queryKey: ["tasks", restUrl, address],
    enabled: Boolean(address),
    queryFn: async (): Promise<Task[]> => {
      const r = await fetch(`${restUrl}/agentic/v1/tasks`);
      if (!r.ok) throw new Error(`tasks fetch: ${r.status}`);
      const { tasks } = await r.json();
      return (tasks as Task[]).filter(
        (t) => t.requester === address || t.agent === address,
      );
    },
    staleTime: 15_000,
  });
}

export interface ReputationSummary {
  best_score: number;
  total_settled: number;
  total_slashed: number;
  nfts: string[];
}

export function useReputation(restUrl: string, address: string | undefined) {
  return useQuery({
    queryKey: ["reputation", restUrl, address],
    enabled: Boolean(address),
    queryFn: async (): Promise<ReputationSummary> => {
      // v0 derives the summary client-side from the agents endpoint;
      // v1 adds a dedicated `/agentic/v1/reputation/{addr}` view.
      const r = await fetch(`${restUrl}/agentic/v1/agents`);
      if (!r.ok) throw new Error(`reputation fetch: ${r.status}`);
      const { agents } = await r.json();
      const own = (agents as AgentRecord[]).filter((a) => a.operator === address);
      return {
        best_score: own.reduce((m, a) => Math.max(m, a.reputation), 0),
        total_settled: 0, // requires task aggregation, v1
        total_slashed: own.filter((a) => a.jailed).length,
        nfts: [],         // populated when the reputation-NFT module ships
      };
    },
    staleTime: 60_000,
  });
}

// Convenience: format ugen → GEN string with 6 decimals.
export function fmtGen(ugen: string | undefined): string {
  if (!ugen) return "—";
  try {
    const n = BigInt(ugen);
    const whole = n / 1_000_000n;
    const frac = (n % 1_000_000n).toString().padStart(6, "0").slice(0, 3);
    return `${whole.toLocaleString()}.${frac}`;
  } catch {
    return "—";
  }
}
