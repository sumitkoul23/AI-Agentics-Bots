/* ============================================================
   Chain Deployment Studio — front-end controller (SKYMETRIC)
   Vanilla JS, no build step. Talks to the Python API in server.py.
   ============================================================ */
(() => {
  "use strict";

  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

  const state = {
    step: 1,
    maxStep: 5,
    targets: [],
    selectedTarget: "local",
    lastRecord: null,
  };

  // ---------- API helpers ----------
  const api = {
    async get(path) {
      const res = await fetch(path);
      if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || res.statusText);
      return res.json();
    },
    async post(path, body) {
      const res = await fetch(path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) throw new Error(data.error || res.statusText);
      return data;
    },
  };

  // ---------- Toast ----------
  let toastTimer;
  function toast(msg, isError = false) {
    const el = $("#toast");
    el.textContent = msg;
    el.classList.toggle("error", isError);
    el.hidden = false;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => (el.hidden = true), 3400);
  }

  // ---------- Engine / stats ----------
  async function loadEngine() {
    const pill = $("#enginePill");
    try {
      const stats = await api.get("/api/stats");
      pill.classList.add("ok");
      pill.innerHTML = `<span class="dot"></span> ${stats.framework}`;
      paintStats(stats);
    } catch {
      pill.classList.add("warn");
      pill.innerHTML = `<span class="dot"></span> offline`;
    }
  }

  function paintStats(s) {
    $("#statDeploys").textContent = s.deployments ?? 0;
    $("#statValidators").textContent = s.validators ?? 0;
    $("#statSupply").textContent = compact(s.supply_sky ?? 0);
  }

  function compact(n) {
    if (n >= 1e9) return (n / 1e9).toFixed(1) + "B";
    if (n >= 1e6) return (n / 1e6).toFixed(1) + "M";
    if (n >= 1e3) return (n / 1e3).toFixed(1) + "K";
    return String(n);
  }

  // ---------- Targets ----------
  async function loadTargets() {
    const grid = $("#targetGrid");
    try {
      const { targets } = await api.get("/api/targets");
      state.targets = targets;
    } catch {
      state.targets = [
        { id: "local", label: "Local devnet", description: "Single-node chain.", icon: "💻", recommended: true, needs: "skymetricd binary" },
      ];
    }
    grid.innerHTML = "";
    state.targets.forEach((t) => {
      const card = document.createElement("button");
      card.type = "button";
      card.className = "net-card" + (t.id === state.selectedTarget ? " selected" : "");
      card.dataset.target = t.id;
      const badge = t.recommended ? '<span class="tag rec">recommended</span>' : "";
      card.innerHTML = `
        <h3>${t.icon || "📦"} ${t.label} ${badge}</h3>
        <p>${t.description}</p>
        <div class="net-meta">needs: ${t.needs}</div>`;
      card.addEventListener("click", () => {
        state.selectedTarget = t.id;
        $$(".net-card", grid).forEach((c) => c.classList.toggle("selected", c === card));
      });
      grid.appendChild(card);
    });
  }

  // ---------- Stepper navigation ----------
  function showStep(n) {
    state.step = n;
    $$(".step").forEach((s) => s.classList.toggle("active", +s.dataset.step === n));
    $$("#stepper li").forEach((li) => {
      const s = +li.dataset.step;
      li.classList.toggle("active", s === n);
      li.classList.toggle("done", s < n);
    });
    $("#backBtn").disabled = n === 1;
    const last = n === state.maxStep;
    $("#nextBtn").hidden = last;
    $("#deployBtn").hidden = !last;
    if (last) buildReview();
  }

  function validateStep(n) {
    if (n === 1) {
      const id = $("#chain_id").value.trim();
      if (!/^[a-z0-9]+(-[a-z0-9]+)*-\d+$/.test(id))
        return "Chain ID must be lowercase and end in a revision number, e.g. mychain-1.";
      if (!$("#moniker").value.trim()) return "Please enter a validator moniker.";
    }
    if (n === 2) {
      if ((parseFloat($("#total_supply_sky").value) || 0) < 1) return "Total supply must be at least 1 SKY.";
      if (parseFloat($("#inflation_min").value) > parseFloat($("#inflation_max").value))
        return "Min inflation cannot exceed max inflation.";
    }
    if (n === 3) {
      const v = parseInt($("#validators").value, 10) || 0;
      const max = parseInt($("#max_validators").value, 10) || 0;
      if (v < 1) return "You need at least one genesis validator.";
      if (max < v) return "Max validators must be >= genesis validators.";
      const bonded = v * (parseFloat($("#validator_stake_sky").value) || 0);
      if (bonded > (parseFloat($("#total_supply_sky").value) || 0))
        return "Validators × stake exceeds total supply.";
    }
    return null;
  }

  // ---------- Gather + review ----------
  function gather() {
    return {
      chain_id: $("#chain_id").value.trim(),
      moniker: $("#moniker").value.trim(),
      description: $("#description").value.trim(),
      target: state.selectedTarget,
      total_supply_sky: parseFloat($("#total_supply_sky").value) || 0,
      validators: parseInt($("#validators").value, 10) || 0,
      max_validators: parseInt($("#max_validators").value, 10) || 0,
      validator_stake_sky: parseFloat($("#validator_stake_sky").value) || 0,
      faucet_balance_sky: parseFloat($("#faucet_balance_sky").value) || 0,
      inflation_min: parseFloat($("#inflation_min").value) || 0,
      inflation_max: parseFloat($("#inflation_max").value) || 0,
      goal_bonded: parseFloat($("#goal_bonded").value) || 0,
      task_burn_fraction: parseFloat($("#task_burn_fraction").value) || 0,
      slash_fraction_fraud: parseFloat($("#slash_fraction_fraud").value) || 0,
      unbonding_days: parseInt($("#unbonding_days").value, 10) || 0,
    };
  }

  function buildReview() {
    const c = gather();
    const target = state.targets.find((t) => t.id === c.target);
    const pct = (x) => `${Math.round(x * 100)}%`;
    const rows = [
      ["Chain ID", c.chain_id || "—"],
      ["Moniker", c.moniker || "—"],
      ["Native coin", "SKY (usky)"],
      ["Total supply", `${c.total_supply_sky.toLocaleString()} SKY`],
      ["Inflation", `${pct(c.inflation_min)} – ${pct(c.inflation_max)}`],
      ["Goal bonded", pct(c.goal_bonded)],
      ["Task burn", pct(c.task_burn_fraction)],
      ["Genesis validators", `${c.validators} (max ${c.max_validators})`],
      ["Validator stake", `${c.validator_stake_sky.toLocaleString()} SKY`],
      ["Unbonding", `${c.unbonding_days} days`],
      ["Deploy target", target ? target.label : c.target],
    ];
    $("#reviewBox").innerHTML = rows
      .map(([k, v]) => `<div class="review-row"><label>${k}</label><span>${v}</span></div>`)
      .join("");
  }

  // ---------- Deploy ----------
  async function runDeploy(e) {
    e.preventDefault();
    if (!$("#confirm").checked) return toast("Please confirm before deploying.", true);

    const config = gather();
    const deployBtn = $("#deployBtn");
    deployBtn.disabled = true;

    $("#emptyState").hidden = true;
    $("#resultBox").hidden = true;
    const log = $("#deployLog");
    log.hidden = false;
    log.innerHTML = "";
    setStatus("deploying");

    const steps = planSteps(config);
    const kickoff = api.post("/api/deploy", config);

    for (const step of steps) {
      const li = appendLog(step, "pending");
      await sleep(360 + Math.random() * 240);
      markDone(li);
    }

    try {
      const record = await kickoff;
      state.lastRecord = record;
      await sleep(280);
      renderResult(record);
      setStatus("ready");
      toast("🎉 Chain bundle generated!");
      loadEngine();
    } catch (err) {
      failDeploy(err.message);
    } finally {
      deployBtn.disabled = false;
    }
  }

  function failDeploy(msg) {
    setStatus("error");
    appendLog({ title: "Deployment failed", detail: msg }, "error");
    toast(msg || "Deployment failed", true);
    $("#deployBtn").disabled = false;
  }

  function planSteps(c) {
    const target = state.targets.find((t) => t.id === c.target) || { label: c.target };
    return [
      { title: "Validate configuration", detail: "Checking chain ID, tokenomics and validators." },
      { title: "Generate genesis overrides", detail: "Building genesis-overrides.json." },
      { title: "Create genesis accounts", detail: `Adding ${c.validators} validator(s) + faucet.` },
      { title: "Collect gentxs", detail: "Genesis validator transactions." },
      { title: "Render init-chain.sh", detail: "Parametrizing the bootstrap script." },
      { title: "Validate genesis", detail: "skymetricd genesis validate-genesis." },
      { title: `Prepare ${target.label} target`, detail: "Assembling deploy recipe." },
      { title: "Finalize bundle", detail: "Packaging artifacts." },
    ];
  }

  function appendLog(step, kind) {
    const li = document.createElement("li");
    li.className = kind;
    const ico = kind === "pending" ? '<span class="spinner"></span>' : kind === "error" ? "✕" : "✓";
    li.innerHTML = `<span class="log-ico">${ico}</span>
      <span class="log-text"><strong>${step.title}</strong><small>${step.detail || ""}</small></span>`;
    $("#deployLog").appendChild(li);
    li.scrollIntoView({ behavior: "smooth", block: "nearest" });
    return li;
  }
  function markDone(li) {
    li.className = "done";
    li.querySelector(".log-ico").innerHTML = "✓";
  }

  function esc(s) {
    return String(s).replace(/[&<>]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;" }[c]));
  }

  function renderResult(r) {
    const s = r.summary;
    const a = r.artifacts;
    const kv = (label, value) => `<div class="kv-row"><label>${label}</label><span class="val">${value}</span></div>`;

    const file = (id, name, body, lang) => `
      <div class="artifact">
        <div class="artifact-head">
          <span>${name}</span>
          <span class="artifact-btns">
            <button class="copy-btn" data-copy-id="${id}">copy</button>
            <button class="copy-btn" data-dl-id="${id}" data-dl-name="${name}">download</button>
          </span>
        </div>
        <pre id="${id}" class="${lang}"><code>${esc(body)}</code></pre>
      </div>`;

    $("#resultBox").innerHTML = `
      <div class="result-banner">
        <span class="big">🌌</span>
        <div>
          <strong>${r.config.chain_id} is ready to launch</strong>
          <span>${s.total_supply} · ${s.validators} validators · ${s.target}</span>
        </div>
      </div>
      <div class="kv">
        ${kv("Chain ID", s.chain_id)}
        ${kv("Native coin", s.denom)}
        ${kv("Total supply", s.total_supply)}
        ${kv("Bonded at genesis", s.bonded_at_genesis)}
        ${kv("Inflation", s.inflation)}
        ${kv("Task burn", s.task_burn)}
      </div>

      <h4 class="artifacts-title">Generated artifacts</h4>
      ${file("art-genesis", "genesis-overrides.json", a.genesis_overrides_json, "json")}
      ${file("art-init", "init-chain.sh", a.init_chain_sh, "bash")}
      ${file("art-env", ".env", a.env_file, "ini")}

      <div class="deploy-cmd">
        <label>Deploy command (${esc(r.target_label)})</label>
        <pre><code>${esc(a.deploy_command)}</code></pre>
        <small>${esc(a.target_needs)}</small>
        <div class="endpoints">RPC <code>${a.rpc}</code> · REST <code>${a.rest}</code></div>
      </div>

      <div class="result-actions">
        <button class="link-btn" id="dlAll">⬇ Download all artifacts</button>
        <button class="again-btn" id="againBtn">Deploy another</button>
      </div>`;
    $("#resultBox").hidden = false;

    // copy buttons
    $$("[data-copy-id]", $("#resultBox")).forEach((b) =>
      b.addEventListener("click", () => {
        navigator.clipboard?.writeText($("#" + b.dataset.copyId).innerText);
        toast("Copied to clipboard");
      })
    );
    // per-file download
    $$("[data-dl-id]", $("#resultBox")).forEach((b) =>
      b.addEventListener("click", () =>
        download(b.dataset.dlName, $("#" + b.dataset.dlId).innerText)
      )
    );
    $("#dlAll").addEventListener("click", downloadAll);
    $("#againBtn").addEventListener("click", resetWizard);
  }

  function download(name, content) {
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = name;
    link.click();
    URL.revokeObjectURL(url);
  }

  function downloadAll() {
    const r = state.lastRecord;
    if (!r) return;
    download("genesis-overrides.json", r.artifacts.genesis_overrides_json);
    download("init-chain.sh", r.artifacts.init_chain_sh);
    download(".env", r.artifacts.env_file);
    toast("Downloading 3 files…");
  }

  function resetWizard() {
    $("#wizardForm").reset();
    $("#confirm").checked = false;
    state.selectedTarget = "local";
    state.lastRecord = null;
    loadTargets();
    syncOutputs();
    showStep(1);
    $("#deployLog").hidden = true;
    $("#deployLog").innerHTML = "";
    $("#resultBox").hidden = true;
    $("#emptyState").hidden = false;
    setStatus("idle");
  }

  function setStatus(s) {
    const chip = $("#statusChip");
    chip.className = "status-chip " + (s === "idle" ? "" : s);
    chip.textContent = s;
  }

  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

  // ---------- Range outputs ----------
  function syncOutputs() {
    const pct = (id) => Math.round((parseFloat($("#" + id).value) || 0) * 100) + "%";
    $("#infMinOut").textContent = pct("inflation_min");
    $("#infMaxOut").textContent = pct("inflation_max");
    $("#goalOut").textContent = pct("goal_bonded");
    $("#burnOut").textContent = pct("task_burn_fraction");
    $("#slashOut").textContent = pct("slash_fraction_fraud");
  }

  // ---------- Wiring ----------
  function init() {
    loadEngine();
    loadTargets();

    $("#nextBtn").addEventListener("click", () => {
      const err = validateStep(state.step);
      if (err) return toast(err, true);
      if (state.step < state.maxStep) showStep(state.step + 1);
    });
    $("#backBtn").addEventListener("click", () => {
      if (state.step > 1) showStep(state.step - 1);
    });
    $("#wizardForm").addEventListener("submit", runDeploy);

    ["inflation_min", "inflation_max", "goal_bonded", "task_burn_fraction", "slash_fraction_fraud"].forEach(
      (id) => $("#" + id).addEventListener("input", syncOutputs)
    );

    $("#chain_id").addEventListener("input", (e) => {
      e.target.value = e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, "");
    });

    syncOutputs();
    showStep(1);
  }

  document.addEventListener("DOMContentLoaded", init);
})();
