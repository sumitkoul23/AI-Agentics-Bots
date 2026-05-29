// popup.js — renders the five panels (Assets, Agents, Tasks, Reputation,
// Streams) into the popup. In v0 every panel pulls from a fixture; v1
// swaps the fixtures for live queries against the SKYMETRIC REST endpoint.

const FIXTURES = {
  assets: [
    { denom: "SKY", amount: "—", note: "native" },
    { denom: "USDC.axl", amount: "—", note: "IBC" },
    { denom: "pool/1", amount: "—", note: "LP" },
  ],
  registry: [
    // Empty until the user registers an agent.
  ],
  tasks: [],
  reputation: { score: 0, slashCount: 0, nfts: [] },
  streams: [],
};

const panel = document.getElementById("panel");
const tabs = document.querySelectorAll(".tab");
tabs.forEach((t) =>
  t.addEventListener("click", () => {
    tabs.forEach((x) => x.classList.toggle("active", x === t));
    render(t.dataset.tab);
  }),
);

document.querySelectorAll(".actions button").forEach((b) =>
  b.addEventListener("click", () => {
    // v0: actions surface a warn; v1 hooks them to lib/tx.ts.
    panel.innerHTML = `<div class="empty">
      "${b.dataset.action}" lands in v1 once the Keplr fork ships.
    </div>`;
  }),
);

document.getElementById("addr").addEventListener("click", async () => {
  const text = document.getElementById("addr").textContent;
  try { await navigator.clipboard.writeText(text); } catch {}
});

// Initial render.
render("assets");

function render(tab) {
  switch (tab) {
    case "assets":     return renderAssets();
    case "registry":   return renderRegistry();
    case "tasks":      return renderTasks();
    case "reputation": return renderReputation();
    case "streams":    return renderStreams();
  }
}

function renderAssets() {
  panel.innerHTML = `
    <h3>Assets</h3>
    ${FIXTURES.assets
      .map(
        (a) => `
      <div class="row">
        <div>
          <div>${esc(a.denom)}</div>
          <div class="muted mono" style="font-size:11px">${esc(a.note)}</div>
        </div>
        <div class="mono">${esc(a.amount)}</div>
      </div>`,
      )
      .join("")}
  `;
}

function renderRegistry() {
  if (FIXTURES.registry.length === 0) {
    panel.innerHTML = `
      <h3>Your Agents</h3>
      <div class="empty">
        No agents registered under this address.<br><br>
        Use the DEX or run<br>
        <span class="mono">skymetricd tx agentic register-agent</span><br>
        to register your first agent.
      </div>`;
    return;
  }
  panel.innerHTML = `
    <h3>Your Agents</h3>
    ${FIXTURES.registry
      .map(
        (a) => `
      <div class="row">
        <div>
          <div>${esc(a.moniker)}</div>
          <div class="muted mono" style="font-size:11px">stake ${esc(a.stake)} SKY</div>
        </div>
        <div class="pill ${a.jailed ? "bad" : "good"}">${a.jailed ? "JAILED" : "active"}</div>
      </div>`,
      )
      .join("")}
  `;
}

function renderTasks() {
  panel.innerHTML = `
    <h3>Open tasks</h3>
    <div class="empty">No active tasks. New tasks arrive via x/agentic.MsgCreateTask.</div>
  `;
}

function renderReputation() {
  const r = FIXTURES.reputation;
  panel.innerHTML = `
    <h3>Reputation</h3>
    <div class="row">
      <div>Score</div>
      <div class="mono">${r.score}</div>
    </div>
    <div class="row">
      <div>Lifetime slashes</div>
      <div class="mono ${r.slashCount > 0 ? "muted" : ""}">${r.slashCount}</div>
    </div>
    <div class="row">
      <div>Reputation NFTs</div>
      <div class="mono">${r.nfts.length}</div>
    </div>
    <div class="muted" style="margin-top:14px; font-size:11px;">
      Reputation is soul-bound — non-transferable. High-rep agents need less
      stake per task. See genesis/docs/01-architecture.md.
    </div>
  `;
}

function renderStreams() {
  panel.innerHTML = `
    <h3>Streaming payments</h3>
    <div class="empty">No active streams. Streams land with financial-instrument #4 (see docs/06).</div>
  `;
}

function esc(s) {
  return String(s).replace(/[&<>"']/g, (c) =>
    ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]),
  );
}
