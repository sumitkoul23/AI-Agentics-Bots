"use client";

import { useState } from "react";
import { ArrowDownUp } from "lucide-react";

// SwapForm is a UI scaffold. Wire to lib/tx.ts::swap once the chain client
// is connected. v0 deliberately keeps the form local-state-only so it works
// before the chain is online — pasting addresses + amounts produces a
// signed tx blob in the console.

export default function SwapForm() {
  const [fromAmount, setFromAmount] = useState("");
  const [fromDenom, setFromDenom] = useState("SKY");
  const [toDenom, setToDenom] = useState("USDC");
  const [slippageBps, setSlippageBps] = useState(50); // 0.5 % default

  const flip = () => {
    setFromDenom(toDenom);
    setToDenom(fromDenom);
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        // TODO: call lib/tx.ts::swap with the offline signer from Cosmos Kit.
        console.log({ fromAmount, fromDenom, toDenom, slippageBps });
      }}
      className="rounded-2xl border border-white/10 bg-ink/40 p-5 max-w-md space-y-3"
    >
      <Box label="You pay">
        <input
          inputMode="decimal"
          placeholder="0.00"
          value={fromAmount}
          onChange={(e) => setFromAmount(e.target.value)}
          className="bg-transparent text-2xl font-mono outline-none flex-1"
        />
        <DenomPill denom={fromDenom} />
      </Box>

      <div className="flex justify-center -my-2 relative z-10">
        <button
          type="button"
          onClick={flip}
          className="p-2 rounded-full border border-white/10 bg-ink hover:bg-white/5"
          aria-label="flip"
        >
          <ArrowDownUp className="h-4 w-4" />
        </button>
      </div>

      <Box label="You receive">
        <span className="text-2xl font-mono text-ash flex-1">—</span>
        <DenomPill denom={toDenom} />
      </Box>

      <div className="flex items-center justify-between text-xs text-ash pt-2">
        <span>Max slippage</span>
        <select
          value={slippageBps}
          onChange={(e) => setSlippageBps(parseInt(e.target.value, 10))}
          className="bg-ink border border-white/10 rounded px-2 py-1 font-mono"
        >
          <option value={10}>0.10%</option>
          <option value={50}>0.50%</option>
          <option value={100}>1.00%</option>
          <option value={300}>3.00%</option>
        </select>
      </div>

      <button
        type="submit"
        className="w-full mt-2 py-3 rounded-xl bg-gen text-ink font-semibold hover:opacity-90 disabled:opacity-40"
        disabled={!fromAmount}
      >
        Swap
      </button>
    </form>
  );
}

function Box({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-xs uppercase tracking-widest text-ash">{label}</span>
      <div className="mt-1 rounded-xl bg-plum/40 border border-white/5 px-4 py-3 flex items-center gap-3">
        {children}
      </div>
    </label>
  );
}

function DenomPill({ denom }: { denom: string }) {
  return (
    <span className="px-3 py-1 rounded-full bg-ink/80 border border-white/10 font-mono text-sm">
      {denom}
    </span>
  );
}
