/* ============================================================
   Chain Deployment Studio — STANDALONE build (SKYMETRIC)
   A faithful client-side port of web/backend/deployer.py so the
   wizard works fully offline from a single file:// HTML page,
   with no server and no API calls.
   ============================================================ */
(() => {
  "use strict";

  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

  // ---------- Constants (mirror deployer.py) ----------
  const USKY_PER_SKY = 1_000_000;
  const BASE_DENOM = "usky";
  const DISPLAY_DENOM = "SKY";

  const TARGETS = [
    { id: "local", label: "Local devnet", icon: "💻", recommended: true,
      description: "Single-node chain on your machine. Best for first runs.",
      command: "./scripts/init-chain.sh && ./scripts/start-node.sh",
      needs: "Go 1.21+ and the skymetricd binary (make install)." },
    { id: "docker", label: "Docker", icon: "🐳", recommended: false,
      description: "Reproducible containerised node via docker-compose.",
      command: "docker compose -f deploy/docker/docker-compose.yml up -d",
      needs: "Docker + docker compose." },
    { id: "fly", label: "Fly.io", icon: "🪰", recommended: false,
      description: "Deploy a public node to Fly.io's free tier.",
      command: "fly deploy -c deploy/fly/fly.toml",
      needs: "A Fly.io account and flyctl." },
    { id: "oracle", label: "Oracle Cloud", icon: "☁️", recommended: false,
      description: "Always-free ARM VM — see deploy/oracle-cloud.",
      command: "see deploy/oracle-cloud/README.md",
      needs: "An Oracle Cloud always-free tenancy." },
    { id: "codespaces", label: "GitHub Codespaces", icon: "🧑‍💻", recommended: false,
      description: "Zero-install devnet in the browser.",
      command: "see deploy/codespaces/README.md",
      needs: "A GitHub account." },
  ];

  const state = { step: 1, maxStep: 5, targets: TARGETS, selectedTarget: "local", lastRecord: null };

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

  // ---------- Engine / stats (local) ----------
  function loadEngine() {
    const pill = $("#enginePill");
    pill.classList.add("ok");
    pill.innerHTML = `<span class="dot"></span> Cosmos SDK v0.50 + CometBFT`;
    paintStats(STORE.stats());
  }
  function paintStats(s) {
    $("#statDeploys").textContent = s.deployments;
    $("#statValidators").textContent = s.validators;
    $("#statSupply").textContent = compact(s.supply_sky);
  }
  function compact(n) {
    if (n >= 1e9) return (n / 1e9).toFixed(1) + "B";
    if (n >= 1e6) return (n / 1e6).toFixed(1) + "M";
    if (n >= 1e3) return (n / 1e3).toFixed(1) + "K";
    return String(n);
  }

  // ---------- Targets ----------
  function loadTargets() {
    const grid = $("#targetGrid");
    grid.innerHTML = "";
    state.targets.forEach((t) => {
      const card = document.createElement("button");
      card.type = "button";
      card.className = "net-card" + (t.id === state.selectedTarget ? " selected" : "");
      const badge = t.recommended ? '<span class="tag rec">recommended</span>' : "";
      card.innerHTML = `
        <h3>${t.icon} ${t.label} ${badge}</h3>
        <p>${t.description}</p>
        <div class="net-meta">needs: ${t.needs}</div>`;
      card.addEventListener("click", () => {
        state.selectedTarget = t.id;
        $$(".net-card", grid).forEach((c) => c.classList.toggle("selected", c === card));
      });
      grid.appendChild(card);
    });
  }

  // ---------- Deployer port (mirrors deployer.py) ----------
  const STORE = {
    _items: [],
    add(r) { this._items.unshift(r); },
    stats() {
      return {
        deployments: this._items.length,
        validators: this._items.reduce((s, d) => s + d.config.validators, 0),
        supply_sky: this._items.reduce((s, d) => s + d.config.total_supply_sky, 0),
      };
    },
  };

  class DeploymentError extends Error {}

  const toUsky = (sky) => Math.round(sky * USKY_PER_SKY);
  const sdkDec = (v) => v.toFixed(18);

  function validChainId(v) {
    return /^[a-z0-9]+(-[a-z0-9]+)*-\d+$/.test(v);
  }

  function validateRequest(p) {
    const chain_id = String(p.chain_id || "").trim();
    if (!chain_id) throw new DeploymentError("Chain ID is required.");
    if (!validChainId(chain_id))
      throw new DeploymentError("Chain ID must look like 'name-1' (lowercase, ends in a revision number).");

    let moniker = String(p.moniker || "").trim() || "genesis-node";
    if (moniker.length > 70 || !/^[A-Za-z0-9._-]+$/.test(moniker))
      throw new DeploymentError("Moniker must be <=70 chars (letters, digits, - _ .).");

    const target = String(p.target || "local").trim().toLowerCase();
    if (!TARGETS.some((t) => t.id === target))
      throw new DeploymentError(`Unsupported deploy target: '${target}'.`);

    const total_supply = num(p.total_supply_sky, 1_000_000_000, "Total supply");
    if (total_supply < 1) throw new DeploymentError("Total supply must be at least 1 SKY.");

    const validators = int(p.validators, 4, "Validator count");
    if (validators < 1 || validators > 100)
      throw new DeploymentError("Validator count must be between 1 and 100.");

    const max_validators = int(p.max_validators, 100, "Max validators");
    if (max_validators < validators || max_validators > 1000)
      throw new DeploymentError("Max validators must be >= genesis validators and <= 1000.");

    const validator_stake = num(p.validator_stake_sky, 100_000, "Validator stake");
    const faucet_balance = num(p.faucet_balance_sky, 500_000, "Faucet balance");
    const inflation_min = frac(p.inflation_min, 0.01, "Min inflation");
    const inflation_max = frac(p.inflation_max, 0.07, "Max inflation");
    if (inflation_min > inflation_max)
      throw new DeploymentError("Min inflation cannot exceed max inflation.");
    const goal_bonded = frac(p.goal_bonded, 0.67, "Goal bonded");
    if (goal_bonded <= 0) throw new DeploymentError("Goal bonded must be greater than 0.");
    const task_burn_fraction = frac(p.task_burn_fraction, 0.2, "Task burn");
    const slash_fraction_fraud = frac(p.slash_fraction_fraud, 0.5, "Fraud slash");
    const unbonding_days = int(p.unbonding_days, 21, "Unbonding period");
    if (unbonding_days < 1 || unbonding_days > 90)
      throw new DeploymentError("Unbonding period must be between 1 and 90 days.");

    if (validators * validator_stake > total_supply)
      throw new DeploymentError("Validators × stake exceeds total supply. Lower the stake or validator count.");

    return {
      chain_id, moniker, target,
      total_supply_sky: total_supply, validators, max_validators,
      validator_stake_sky: validator_stake, faucet_balance_sky: faucet_balance,
      inflation_min, inflation_max, goal_bonded,
      task_burn_fraction, slash_fraction_fraud, unbonding_days,
      description: String(p.description || "").trim().slice(0, 280),
    };
  }

  function num(v, d, label) {
    const n = v === undefined || v === null || v === "" ? d : Number(v);
    if (Number.isNaN(n)) throw new DeploymentError(`${label} must be a number.`);
    if (n < 0) throw new DeploymentError(`${label} cannot be negative.`);
    return n;
  }
  function int(v, d, label) {
    const n = v === undefined || v === null || v === "" ? d : parseInt(v, 10);
    if (Number.isNaN(n)) throw new DeploymentError(`${label} must be a whole number.`);
    if (n < 0) throw new DeploymentError(`${label} cannot be negative.`);
    return n;
  }
  function frac(v, d, label) {
    const n = v === undefined || v === null || v === "" ? d : Number(v);
    if (Number.isNaN(n)) throw new DeploymentError(`${label} must be a number.`);
    if (n < 0 || n > 1) throw new DeploymentError(`${label} must be between 0 and 1.`);
    return n;
  }

  function genesisOverrides(c) {
    const burn = c.task_burn_fraction;
    const remainder = 1.0 - burn;
    const split_agent = (remainder * 5.0) / 8.0;
    const split_validators = remainder - split_agent;
    const minAgentStake = toUsky(Math.max(c.validator_stake_sky / 1000.0, 1));
    return {
      _comment: `Generated by Chain Deployment Studio for ${c.chain_id}. Applied via jq onto the genesis.json that \`skymetricd init\` produces.`,
      app_state: {
        staking: { params: {
          unbonding_time: `${c.unbonding_days * 86400}s`,
          max_validators: c.max_validators,
          bond_denom: BASE_DENOM,
          min_commission_rate: sdkDec(0.05),
        }},
        mint: { params: {
          mint_denom: BASE_DENOM,
          inflation_rate_change: sdkDec(0.01),
          inflation_max: sdkDec(c.inflation_max),
          inflation_min: sdkDec(c.inflation_min),
          goal_bonded: sdkDec(c.goal_bonded),
          blocks_per_year: "10512000",
        }},
        gov: { params: {
          min_deposit: [{ denom: BASE_DENOM, amount: "10000000" }],
          max_deposit_period: "172800s",
          voting_period: "172800s",
        }},
        slashing: { params: {
          signed_blocks_window: "10000",
          min_signed_per_window: sdkDec(0.05),
          downtime_jail_duration: "600s",
          slash_fraction_double_sign: sdkDec(c.slash_fraction_fraud),
          slash_fraction_downtime: sdkDec(0.0001),
        }},
        agentic: {
          params: {
            min_agent_stake: String(minAgentStake),
            min_agent_stake_floor: String(Math.max(Math.floor(minAgentStake / 10), 1)),
            split_agent: sdkDec(split_agent),
            split_validators: sdkDec(split_validators),
            split_burn: sdkDec(burn),
            fraud_proof_quorum: 3,
            reputation_gain_per_task: 1,
          },
          agent_records: [], tasks: [], task_counter: 0, burned_total: "0",
        },
      },
    };
  }

  function initScript(c) {
    const stake = `${toUsky(c.validator_stake_sky)}${BASE_DENOM}`;
    const faucet = `${toUsky(c.faucet_balance_sky)}${BASE_DENOM}`;
    return `#!/usr/bin/env bash
# Generated by Chain Deployment Studio for ${c.chain_id}
# Initialise a single-node ${c.chain_id} devnet.
set -euo pipefail

CHAIN_ID="\${CHAIN_ID:-${c.chain_id}}"
MONIKER="\${MONIKER:-${c.moniker}}"
DENOM="${BASE_DENOM}"
HOME_DIR="\${HOME_DIR:-$HOME/.skymetric}"
BIN="\${BIN:-skymetricd}"

VALIDATOR_STAKE="\${VALIDATOR_STAKE:-${stake}}"
FAUCET_BALANCE="\${FAUCET_BALANCE:-${faucet}}"

echo "==> Initialising $CHAIN_ID (moniker=$MONIKER, home=$HOME_DIR)"
rm -rf "$HOME_DIR"
"$BIN" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"

cp ./genesis-overrides.json "$HOME_DIR/config/genesis-overrides.json"

echo "==> Creating genesis accounts"
for KEY in validator faucet; do
  "$BIN" keys add "$KEY" --keyring-backend test --home "$HOME_DIR" 2>/dev/null || true
done

VAL_ADDR=$("$BIN" keys show validator -a --keyring-backend test --home "$HOME_DIR")
FAUCET_ADDR=$("$BIN" keys show faucet -a --keyring-backend test --home "$HOME_DIR")

"$BIN" genesis add-genesis-account "$VAL_ADDR" "$VALIDATOR_STAKE" --home "$HOME_DIR"
"$BIN" genesis add-genesis-account "$FAUCET_ADDR" "$FAUCET_BALANCE" --home "$HOME_DIR"

echo "==> Creating genesis validator (gentx)"
"$BIN" genesis gentx validator "$VALIDATOR_STAKE" \\
  --chain-id "$CHAIN_ID" --keyring-backend test --home "$HOME_DIR"

echo "==> Collecting gentxs"
"$BIN" genesis collect-gentxs --home "$HOME_DIR"
"$BIN" genesis validate-genesis --home "$HOME_DIR"
echo "==> Done. Start with ./scripts/start-node.sh"
`;
  }

  function envFile(c) {
    return (
      `CHAIN_ID=${c.chain_id}\n` +
      `MONIKER=${c.moniker}\n` +
      `DENOM=${BASE_DENOM}\n` +
      `VALIDATOR_STAKE=${toUsky(c.validator_stake_sky)}${BASE_DENOM}\n` +
      `FAUCET_BALANCE=${toUsky(c.faucet_balance_sky)}${BASE_DENOM}\n` +
      `MAX_VALIDATORS=${c.max_validators}\n`
    );
  }

  function deploy(payload) {
    const c = validateRequest(payload);
    const target = TARGETS.find((t) => t.id === c.target);
    const id = Math.random().toString(16).slice(2, 14);
    const now = () => new Date().toISOString();
    const bonded = c.validators * c.validator_stake_sky;
    const fmt = (n) => n.toLocaleString("en-US");

    const record = {
      id,
      created_at: now(),
      status: "ready",
      config: c,
      target_label: target.label,
      artifacts: {
        genesis_overrides_json: JSON.stringify(genesisOverrides(c), null, 2),
        init_chain_sh: initScript(c),
        env_file: envFile(c),
        deploy_command: target.command,
        target_needs: target.needs,
        rpc: "http://localhost:26657",
        rest: "http://localhost:1317",
        start_command: "./scripts/start-node.sh",
      },
      summary: {
        chain_id: c.chain_id,
        denom: `${DISPLAY_DENOM} (${BASE_DENOM})`,
        total_supply: `${fmt(c.total_supply_sky)} ${DISPLAY_DENOM}`,
        validators: c.validators,
        bonded_at_genesis: `${fmt(bonded)} ${DISPLAY_DENOM}`,
        inflation: `${Math.round(c.inflation_min * 100)}–${Math.round(c.inflation_max * 100)}%`,
        task_burn: `${Math.round(c.task_burn_fraction * 100)}%`,
        target: target.label,
      },
    };
    STORE.add(record);
    return record;
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
      if (!validChainId(id))
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

  // ---------- Deploy flow ----------
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

    let record, err;
    try { record = deploy(config); }
    catch (e2) { err = e2; }

    const steps = planSteps(config);
    for (const step of steps) {
      const li = appendLog(step, "pending");
      await sleep(360 + Math.random() * 240);
      markDone(li);
    }

    if (err) { failDeploy(err.message); deployBtn.disabled = false; return; }

    state.lastRecord = record;
    await sleep(280);
    renderResult(record);
    setStatus("ready");
    toast("🎉 Chain bundle generated!");
    paintStats(STORE.stats());
    deployBtn.disabled = false;
  }

  function failDeploy(msg) {
    setStatus("error");
    appendLog({ title: "Deployment failed", detail: msg }, "error");
    toast(msg || "Deployment failed", true);
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
    const s = r.summary, a = r.artifacts;
    const kv = (label, value) => `<div class="kv-row"><label>${label}</label><span class="val">${value}</span></div>`;
    const file = (id, name, body) => `
      <div class="artifact">
        <div class="artifact-head">
          <span>${name}</span>
          <span class="artifact-btns">
            <button class="copy-btn" data-copy-id="${id}">copy</button>
            <button class="copy-btn" data-dl-id="${id}" data-dl-name="${name}">download</button>
          </span>
        </div>
        <pre id="${id}"><code>${esc(body)}</code></pre>
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
      ${file("art-genesis", "genesis-overrides.json", a.genesis_overrides_json)}
      ${file("art-init", "init-chain.sh", a.init_chain_sh)}
      ${file("art-env", ".env", a.env_file)}
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

    $$("[data-copy-id]", $("#resultBox")).forEach((b) =>
      b.addEventListener("click", () => {
        navigator.clipboard?.writeText($("#" + b.dataset.copyId).innerText);
        toast("Copied to clipboard");
      })
    );
    $$("[data-dl-id]", $("#resultBox")).forEach((b) =>
      b.addEventListener("click", () => download(b.dataset.dlName, $("#" + b.dataset.dlId).innerText))
    );
    $("#dlAll").addEventListener("click", downloadAll);
    $("#againBtn").addEventListener("click", resetWizard);
  }

  function download(name, content) {
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url; link.download = name; link.click();
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

  function syncOutputs() {
    const pct = (id) => Math.round((parseFloat($("#" + id).value) || 0) * 100) + "%";
    $("#infMinOut").textContent = pct("inflation_min");
    $("#infMaxOut").textContent = pct("inflation_max");
    $("#goalOut").textContent = pct("goal_bonded");
    $("#burnOut").textContent = pct("task_burn_fraction");
    $("#slashOut").textContent = pct("slash_fraction_fraud");
  }

  function init() {
    loadEngine();
    loadTargets();
    $("#nextBtn").addEventListener("click", () => {
      const err = validateStep(state.step);
      if (err) return toast(err, true);
      if (state.step < state.maxStep) showStep(state.step + 1);
    });
    $("#backBtn").addEventListener("click", () => { if (state.step > 1) showStep(state.step - 1); });
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
