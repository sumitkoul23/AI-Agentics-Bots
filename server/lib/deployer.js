/**
 * Deployer — validates deploy requests and emits real SKYMETRIC chain artifacts.
 * Ported from web/assets/js/standalone.js (which mirrors web/backend/deployer.py),
 * so the Node backend produces byte-identical genesis-overrides.json / init-chain.sh.
 */

export const BASE_DENOM = "usky";
export const DISPLAY_DENOM = "SKY";
export const USKY_PER_SKY = 1_000_000;

export const TARGETS = [
  { id: "local",      label: "Local devnet",      icon: "💻",  recommended: true,  needs: "skymetricd binary",
    command: "./init-chain.sh && skymetricd start --home ~/.skymetric" },
  { id: "docker",     label: "Docker",            icon: "🐳",  recommended: false, needs: "Docker",
    command: "docker compose up -d" },
  { id: "fly",        label: "Fly.io",            icon: "🪰",  recommended: false, needs: "flyctl + account",
    command: "fly launch && fly deploy" },
  { id: "oracle",     label: "Oracle Cloud",      icon: "☁️", recommended: false, needs: "OCI free tier",
    command: "ssh oracle 'bash -s' < init-chain.sh" },
  { id: "codespaces", label: "GitHub Codespaces", icon: "🧑‍💻", recommended: false, needs: "GitHub account",
    command: "gh codespace create && ./init-chain.sh" },
];

export class DeploymentError extends Error {}

const toUsky = (sky) => Math.round(sky * USKY_PER_SKY);
const sdkDec = (v) => Number(v).toFixed(18);

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
function validChainId(v) {
  return /^[a-z0-9]+(-[a-z0-9]+)*-\d+$/.test(v);
}

export function validateRequest(p) {
  const chain_id = String(p.chain_id || "").trim();
  if (!chain_id) throw new DeploymentError("Chain ID is required.");
  if (!validChainId(chain_id))
    throw new DeploymentError("Chain ID must look like 'name-1' (lowercase, ends in a revision number).");

  const moniker = String(p.moniker || "").trim() || "genesis-node";
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

  const authority_address = String(p.authority_address || "").trim();
  if (authority_address && !(authority_address.startsWith("agentic1") && authority_address.length >= 39 && authority_address.length <= 90))
    throw new DeploymentError("Authority address must be a bech32 'agentic1…' address.");

  return {
    chain_id, moniker, target,
    total_supply_sky: total_supply, validators, max_validators,
    validator_stake_sky: validator_stake, faucet_balance_sky: faucet_balance,
    inflation_min, inflation_max, goal_bonded,
    task_burn_fraction, slash_fraction_fraud, unbonding_days,
    description: String(p.description || "").trim().slice(0, 280),
    denom_display: DISPLAY_DENOM,
    authority_address: authority_address || null,
  };
}

export function genesisOverrides(c) {
  const burn = c.task_burn_fraction;
  const remainder = 1.0 - burn;
  const split_agent = (remainder * 5.0) / 8.0;
  const split_validators = remainder - split_agent;
  const minAgentStake = toUsky(Math.max(c.validator_stake_sky / 1000.0, 1));
  return {
    _comment: `Generated by Chain Deployment Studio for ${c.chain_id}. Applied via jq onto the genesis.json that \`skymetricd init\` produces.`,
    app_state: {
      staking: { params: { unbonding_time: `${c.unbonding_days * 86400}s`, max_validators: c.max_validators, bond_denom: BASE_DENOM, min_commission_rate: sdkDec(0.05) } },
      mint: { params: { mint_denom: BASE_DENOM, inflation_rate_change: sdkDec(0.01), inflation_max: sdkDec(c.inflation_max), inflation_min: sdkDec(c.inflation_min), goal_bonded: sdkDec(c.goal_bonded), blocks_per_year: "10512000" } },
      gov: { params: { min_deposit: [{ denom: BASE_DENOM, amount: "10000000" }], max_deposit_period: "172800s", voting_period: "172800s" } },
      slashing: { params: { signed_blocks_window: "10000", min_signed_per_window: sdkDec(0.05), downtime_jail_duration: "600s", slash_fraction_double_sign: sdkDec(c.slash_fraction_fraud), slash_fraction_downtime: sdkDec(0.0001) } },
      agentic: {
        params: {
          min_agent_stake: String(minAgentStake),
          min_agent_stake_floor: String(Math.max(Math.floor(minAgentStake / 10), 1)),
          split_agent: sdkDec(split_agent), split_validators: sdkDec(split_validators), split_burn: sdkDec(burn),
          fraud_proof_quorum: 3, reputation_gain_per_task: 1,
        },
        agent_records: [], tasks: [], task_counter: 0, burned_total: "0",
      },
    },
  };
}

export function initScript(c) {
  const stake = `${toUsky(c.validator_stake_sky)}${BASE_DENOM}`;
  const faucet = `${toUsky(c.faucet_balance_sky)}${BASE_DENOM}`;
  return `#!/usr/bin/env bash
# Generated by Chain Deployment Studio for ${c.chain_id}
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
for KEY in validator faucet; do
  "$BIN" keys add "$KEY" --keyring-backend test --home "$HOME_DIR" 2>/dev/null || true
done
VAL_ADDR=$("$BIN" keys show validator -a --keyring-backend test --home "$HOME_DIR")
FAUCET_ADDR=$("$BIN" keys show faucet -a --keyring-backend test --home "$HOME_DIR")
"$BIN" genesis add-genesis-account "$VAL_ADDR" "$VALIDATOR_STAKE" --home "$HOME_DIR"
"$BIN" genesis add-genesis-account "$FAUCET_ADDR" "$FAUCET_BALANCE" --home "$HOME_DIR"
"$BIN" genesis gentx validator "$VALIDATOR_STAKE" --chain-id "$CHAIN_ID" --keyring-backend test --home "$HOME_DIR"
"$BIN" genesis collect-gentxs --home "$HOME_DIR"
"$BIN" genesis validate-genesis --home "$HOME_DIR"
echo "==> Done. Start with: skymetricd start --home $HOME_DIR"
`;
}

export function envFile(c) {
  return (
    `CHAIN_ID=${c.chain_id}\n` +
    `MONIKER=${c.moniker}\n` +
    `DENOM=${BASE_DENOM}\n` +
    `VALIDATOR_STAKE_SKY=${c.validator_stake_sky}\n` +
    `FAUCET_BALANCE_SKY=${c.faucet_balance_sky}\n`
  );
}

const fmt = (n) => Number(n).toLocaleString("en-US");

export function buildRecord(payload, { rpc, rest } = {}) {
  const c = validateRequest(payload);
  const target = TARGETS.find((t) => t.id === c.target);
  const id = Math.random().toString(16).slice(2, 14);
  const bonded = c.validators * c.validator_stake_sky;
  return {
    id,
    created_at: new Date().toISOString(),
    status: "ready",
    config: c,
    target_label: target.label,
    artifacts: {
      genesis_overrides_json: JSON.stringify(genesisOverrides(c), null, 2),
      init_chain_sh: initScript(c),
      env_file: envFile(c),
      deploy_command: target.command,
      target_needs: target.needs,
      rpc: rpc || "http://localhost:26657",
      rest: rest || "http://localhost:1317",
      start_command: `skymetricd start --home ~/.skymetric`,
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
}
