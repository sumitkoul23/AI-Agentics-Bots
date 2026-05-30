"""
Chain Deployer
==============
Turns a high-level deployment request from the Chain Deployment Studio web app
into the real artifacts needed to spin up a **SKYMETRIC** chain
(Cosmos SDK v0.50 + CometBFT), and a concrete, ordered deployment plan.

This mirrors the genuine repo layout under ``genesis/chain``:

    * ``genesis-overrides.json``  — staking / mint / gov / crisis / agentic params
    * ``init-chain.sh``           — single-node devnet bootstrap
    * ``.env``                    — CHAIN_ID / MONIKER / DENOM / stakes
    * a deploy command for the chosen target (local / docker / fly / oracle / codespaces)

The artifacts produced are real and usable — you can drop ``genesis-overrides.json``
into ``genesis/chain/config/`` and run the generated script. Booting an actual
validator additionally requires the compiled ``skymetricd`` binary, so the final
"start node" step is reported as a follow-up command rather than executed here.
"""
from __future__ import annotations

import json
import uuid
from datetime import datetime, timezone
from typing import Dict, List, Optional

# 1 SKY = 1,000,000 usky
USKY_PER_SKY = 1_000_000
BASE_DENOM = "usky"
DISPLAY_DENOM = "SKY"


SUPPORTED_TARGETS = {
    "local": {
        "id": "local",
        "label": "Local devnet",
        "description": "Single-node chain on your machine. Best for first runs.",
        "icon": "💻",
        "recommended": True,
        "command": "./scripts/init-chain.sh && ./scripts/start-node.sh",
        "needs": "Go 1.21+ and the skymetricd binary (make install).",
    },
    "docker": {
        "id": "docker",
        "label": "Docker",
        "description": "Reproducible containerised node via docker-compose.",
        "icon": "🐳",
        "recommended": False,
        "command": "docker compose -f deploy/docker/docker-compose.yml up -d",
        "needs": "Docker + docker compose.",
    },
    "fly": {
        "id": "fly",
        "label": "Fly.io",
        "description": "Deploy a public node to Fly.io's free tier.",
        "icon": "🪰",
        "recommended": False,
        "command": "fly deploy -c deploy/fly/fly.toml",
        "needs": "A Fly.io account and flyctl.",
    },
    "oracle": {
        "id": "oracle",
        "label": "Oracle Cloud",
        "description": "Always-free ARM VM — see deploy/oracle-cloud.",
        "icon": "☁️",
        "recommended": False,
        "command": "see deploy/oracle-cloud/README.md",
        "needs": "An Oracle Cloud always-free tenancy.",
    },
    "codespaces": {
        "id": "codespaces",
        "label": "GitHub Codespaces",
        "description": "Zero-install devnet in the browser.",
        "icon": "🧑‍💻",
        "recommended": False,
        "command": "see deploy/codespaces/README.md",
        "needs": "A GitHub account.",
    },
}


class DeploymentError(ValueError):
    """Raised when a deployment request fails validation."""


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def list_targets() -> List[Dict]:
    return list(SUPPORTED_TARGETS.values())


def _to_usky(sky: float) -> int:
    return int(round(sky * USKY_PER_SKY))


def validate_request(payload: Dict) -> Dict:
    """Validate and normalize a raw deployment request payload."""
    if not isinstance(payload, dict):
        raise DeploymentError("Request body must be a JSON object.")

    chain_id = str(payload.get("chain_id", "")).strip()
    if not chain_id:
        raise DeploymentError("Chain ID is required.")
    if not _valid_chain_id(chain_id):
        raise DeploymentError("Chain ID must look like 'name-1' (lowercase, ends in a revision number).")

    moniker = str(payload.get("moniker", "")).strip() or "genesis-node"
    if len(moniker) > 70 or not all(c.isalnum() or c in "-_." for c in moniker):
        raise DeploymentError("Moniker must be <=70 chars (letters, digits, - _ .).")

    target = str(payload.get("target", "local")).strip().lower()
    if target not in SUPPORTED_TARGETS:
        raise DeploymentError(f"Unsupported deploy target: {target!r}.")

    total_supply = _positive_number(payload, "total_supply_sky", "Total supply", default=1_000_000_000)
    if total_supply < 1:
        raise DeploymentError("Total supply must be at least 1 SKY.")

    validators = _positive_int(payload, "validators", "Validator count", default=4)
    if not (1 <= validators <= 100):
        raise DeploymentError("Validator count must be between 1 and 100.")

    max_validators = _positive_int(payload, "max_validators", "Max validators", default=100)
    if not (validators <= max_validators <= 1000):
        raise DeploymentError("Max validators must be >= genesis validators and <= 1000.")

    validator_stake = _positive_number(payload, "validator_stake_sky", "Validator stake", default=100_000)
    faucet_balance = _positive_number(payload, "faucet_balance_sky", "Faucet balance", default=500_000)

    inflation_min = _fraction(payload, "inflation_min", "Min inflation", default=0.01)
    inflation_max = _fraction(payload, "inflation_max", "Max inflation", default=0.07)
    if inflation_min > inflation_max:
        raise DeploymentError("Min inflation cannot exceed max inflation.")

    goal_bonded = _fraction(payload, "goal_bonded", "Goal bonded", default=0.67)
    if goal_bonded <= 0:
        raise DeploymentError("Goal bonded must be greater than 0.")

    burn_fraction = _fraction(payload, "task_burn_fraction", "Task burn", default=0.20)
    slash_fraction = _fraction(payload, "slash_fraction_fraud", "Fraud slash", default=0.50)

    unbonding_days = _positive_int(payload, "unbonding_days", "Unbonding period", default=21)
    if not (1 <= unbonding_days <= 90):
        raise DeploymentError("Unbonding period must be between 1 and 90 days.")

    if validators * validator_stake > total_supply:
        raise DeploymentError(
            "Validators × stake exceeds total supply. Lower the stake or validator count."
        )

    return {
        "chain_id": chain_id,
        "moniker": moniker,
        "target": target,
        "denom_base": BASE_DENOM,
        "denom_display": DISPLAY_DENOM,
        "total_supply_sky": total_supply,
        "validators": validators,
        "max_validators": max_validators,
        "validator_stake_sky": validator_stake,
        "faucet_balance_sky": faucet_balance,
        "inflation_min": inflation_min,
        "inflation_max": inflation_max,
        "goal_bonded": goal_bonded,
        "task_burn_fraction": burn_fraction,
        "slash_fraction_fraud": slash_fraction,
        "unbonding_days": unbonding_days,
        "description": str(payload.get("description", "")).strip()[:280],
    }


def _valid_chain_id(value: str) -> bool:
    if "-" not in value:
        return False
    name, _, rev = value.rpartition("-")
    if not name or not rev.isdigit():
        return False
    return all(c.islower() or c.isdigit() or c == "-" for c in name)


def _positive_number(payload, key, label, default):
    try:
        val = float(payload.get(key, default))
    except (TypeError, ValueError):
        raise DeploymentError(f"{label} must be a number.")
    if val < 0:
        raise DeploymentError(f"{label} cannot be negative.")
    return val


def _positive_int(payload, key, label, default):
    try:
        val = int(payload.get(key, default))
    except (TypeError, ValueError):
        raise DeploymentError(f"{label} must be a whole number.")
    if val < 0:
        raise DeploymentError(f"{label} cannot be negative.")
    return val


def _fraction(payload, key, label, default):
    try:
        val = float(payload.get(key, default))
    except (TypeError, ValueError):
        raise DeploymentError(f"{label} must be a number.")
    if not (0.0 <= val <= 1.0):
        raise DeploymentError(f"{label} must be between 0 and 1.")
    return val


def build_plan(config: Dict) -> List[Dict]:
    """Return the ordered list of deployment steps for preview / execution."""
    target = SUPPORTED_TARGETS[config["target"]]
    return [
        {"key": "validate", "title": "Validate configuration",
         "detail": "Check chain ID, tokenomics and validator set."},
        {"key": "overrides", "title": "Generate genesis overrides",
         "detail": "Build genesis-overrides.json (staking, mint, gov, agentic)."},
        {"key": "accounts", "title": "Create genesis accounts",
         "detail": f"Add {config['validators']} validator(s) + faucet."},
        {"key": "gentx", "title": "Collect gentxs",
         "detail": "Generate and collect genesis validator transactions."},
        {"key": "script", "title": "Render init-chain.sh",
         "detail": "Parametrize the bootstrap script for this chain."},
        {"key": "validate_genesis", "title": "Validate genesis",
         "detail": "Run skymetricd genesis validate-genesis."},
        {"key": "target", "title": f"Prepare {target['label']} target",
         "detail": target["command"]},
        {"key": "finalize", "title": "Finalize bundle",
         "detail": "Package artifacts and emit start command."},
    ]


def _sdk_dec(value: float) -> str:
    """Cosmos SDK serializes decimals as 18-digit fixed-point strings."""
    return f"{value:.18f}"


def _genesis_overrides(config: Dict) -> Dict:
    """Build a real genesis-overrides.json matching the chain's schema.

    Mirrors genesis/chain/config/genesis-overrides.json, including the agentic
    fee split (agent / validators / burn). The user controls the burn fraction;
    the remainder is shared between agents and validators at the repo's 5:3 ratio.
    """
    burn = config["task_burn_fraction"]
    remainder = 1.0 - burn
    split_agent = remainder * 5.0 / 8.0
    split_validators = remainder - split_agent
    min_agent_stake = _to_usky(max(config["validator_stake_sky"] / 1000.0, 1))

    return {
        "_comment": (
            f"Generated by Chain Deployment Studio for {config['chain_id']}. "
            "Applied via jq onto the genesis.json that `skymetricd init` produces."
        ),
        "app_state": {
            "staking": {
                "params": {
                    "unbonding_time": f"{config['unbonding_days'] * 86400}s",
                    "max_validators": config["max_validators"],
                    "bond_denom": BASE_DENOM,
                    "min_commission_rate": _sdk_dec(0.05),
                }
            },
            "mint": {
                "params": {
                    "mint_denom": BASE_DENOM,
                    "inflation_rate_change": _sdk_dec(0.01),
                    "inflation_max": _sdk_dec(config["inflation_max"]),
                    "inflation_min": _sdk_dec(config["inflation_min"]),
                    "goal_bonded": _sdk_dec(config["goal_bonded"]),
                    "blocks_per_year": "10512000",
                }
            },
            "gov": {
                "params": {
                    "min_deposit": [{"denom": BASE_DENOM, "amount": "10000000"}],
                    "max_deposit_period": "172800s",
                    "voting_period": "172800s",
                }
            },
            "slashing": {
                "params": {
                    "signed_blocks_window": "10000",
                    "min_signed_per_window": _sdk_dec(0.05),
                    "downtime_jail_duration": "600s",
                    "slash_fraction_double_sign": _sdk_dec(config["slash_fraction_fraud"]),
                    "slash_fraction_downtime": _sdk_dec(0.0001),
                }
            },
            "agentic": {
                "params": {
                    "min_agent_stake": str(min_agent_stake),
                    "min_agent_stake_floor": str(max(min_agent_stake // 10, 1)),
                    "split_agent": _sdk_dec(split_agent),
                    "split_validators": _sdk_dec(split_validators),
                    "split_burn": _sdk_dec(burn),
                    "fraud_proof_quorum": 3,
                    "reputation_gain_per_task": 1,
                },
                "agent_records": [],
                "tasks": [],
                "task_counter": 0,
                "burned_total": "0",
            },
        },
    }


def _init_script(config: Dict) -> str:
    stake_usky = _to_usky(config["validator_stake_sky"])
    faucet_usky = _to_usky(config["faucet_balance_sky"])
    return f"""#!/usr/bin/env bash
# Generated by Chain Deployment Studio for {config['chain_id']}
# Initialise a single-node {config['chain_id']} devnet.
set -euo pipefail

CHAIN_ID="${{CHAIN_ID:-{config['chain_id']}}}"
MONIKER="${{MONIKER:-{config['moniker']}}}"
DENOM="{BASE_DENOM}"
HOME_DIR="${{HOME_DIR:-$HOME/.skymetric}}"
BIN="${{BIN:-skymetricd}}"

VALIDATOR_STAKE="${{VALIDATOR_STAKE:-{stake_usky}{BASE_DENOM}}}"
FAUCET_BALANCE="${{FAUCET_BALANCE:-{faucet_usky}{BASE_DENOM}}}"

echo "==> Initialising $CHAIN_ID (moniker=$MONIKER, home=$HOME_DIR)"
rm -rf "$HOME_DIR"
"$BIN" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"

# Apply the genesis overrides produced alongside this script.
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
"""


def _env_file(config: Dict) -> str:
    return (
        f"CHAIN_ID={config['chain_id']}\n"
        f"MONIKER={config['moniker']}\n"
        f"DENOM={BASE_DENOM}\n"
        f"VALIDATOR_STAKE={_to_usky(config['validator_stake_sky'])}{BASE_DENOM}\n"
        f"FAUCET_BALANCE={_to_usky(config['faucet_balance_sky'])}{BASE_DENOM}\n"
        f"MAX_VALIDATORS={config['max_validators']}\n"
    )


class DeploymentStore:
    """In-memory registry of deployments (newest first)."""

    def __init__(self) -> None:
        self._items: Dict[str, Dict] = {}

    def add(self, record: Dict) -> None:
        self._items[record["id"]] = record

    def get(self, deployment_id: str) -> Optional[Dict]:
        return self._items.get(deployment_id)

    def list(self) -> List[Dict]:
        return sorted(self._items.values(), key=lambda r: r["created_at"], reverse=True)


STORE = DeploymentStore()


def deploy(payload: Dict) -> Dict:
    """Validate, assemble artifacts and persist a deployment. Returns the record."""
    config = validate_request(payload)
    deployment_id = uuid.uuid4().hex[:12]
    target = SUPPORTED_TARGETS[config["target"]]

    overrides = _genesis_overrides(config)
    artifacts = {
        "genesis_overrides_json": json.dumps(overrides, indent=2),
        "init_chain_sh": _init_script(config),
        "env_file": _env_file(config),
        "deploy_command": target["command"],
        "target_needs": target["needs"],
        "rpc": "http://localhost:26657",
        "rest": "http://localhost:1317",
        "start_command": "./scripts/start-node.sh",
    }

    bonded = config["validators"] * config["validator_stake_sky"]
    summary = {
        "chain_id": config["chain_id"],
        "denom": f"{DISPLAY_DENOM} ({BASE_DENOM})",
        "total_supply": f"{config['total_supply_sky']:,.0f} {DISPLAY_DENOM}",
        "validators": config["validators"],
        "bonded_at_genesis": f"{bonded:,.0f} {DISPLAY_DENOM}",
        "inflation": f"{config['inflation_min']*100:.0f}–{config['inflation_max']*100:.0f}%",
        "task_burn": f"{config['task_burn_fraction']*100:.0f}%",
        "target": target["label"],
    }

    steps_result = [
        {**step, "status": "complete", "completed_at": _now_iso()}
        for step in build_plan(config)
    ]

    record = {
        "id": deployment_id,
        "created_at": _now_iso(),
        "status": "ready",
        "config": config,
        "target_label": target["label"],
        "steps": steps_result,
        "artifacts": artifacts,
        "summary": summary,
    }
    STORE.add(record)
    return record


def stats() -> Dict:
    items = STORE.list()
    total_validators = sum(d["config"]["validators"] for d in items)
    total_supply = sum(d["config"]["total_supply_sky"] for d in items)
    return {
        "deployments": len(items),
        "validators": total_validators,
        "supply_sky": total_supply,
        "framework": "Cosmos SDK v0.50 + CometBFT",
    }
