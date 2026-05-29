use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Decimal, Uint128};
use cw_storage_plus::{Item, Map};

/// Module-wide parameters. Mutable by `params.admin` via `MsgUpdateParams`.
/// Defaults mirror `genesis/chain/x/agentic/types/params.go::DefaultParams`.
#[cw_serde]
pub struct Params {
    /// Address allowed to call UpdateParams. Bootstrapped to the contract's
    /// instantiator; transferred to the gov module address once on-chain
    /// governance ships.
    pub admin: Addr,
    /// Denom (string) accepted as stake + bounty. CW20-contract-addr form
    /// or Token-Factory denom; both are strings in CosmWasm.
    pub stake_denom: String,
    /// Where burns go. Null address on most chains; a Token-Factory
    /// burn-authority on Neutron mainnet.
    pub burn_sink: Addr,
    /// Where the validator/treasury slice of settled bounties goes.
    pub treasury: Addr,

    pub min_agent_stake: Uint128,
    pub min_agent_stake_floor: Uint128,

    pub split_agent: Decimal,      // 0.50
    pub split_treasury: Decimal,   // 0.30 (was "split_validators" on L1)
    pub split_burn: Decimal,       // 0.20

    pub fraud_proof_quorum: u32,
    pub reputation_gain_per_task: u64,
}

/// On-chain identity of one AI agent.
#[cw_serde]
pub struct AgentRecord {
    pub operator: Addr,
    pub moniker: String,
    pub endpoint: String,
    pub stake: Uint128,
    pub reputation: u64,
    pub jailed: bool,
}

/// Escrow opened by a user requesting agent work.
#[cw_serde]
pub struct Task {
    pub id: u64,
    pub requester: Addr,
    pub agent: Addr,
    pub bounty: Uint128,
    pub spec: String,
    pub response_cid: Option<String>,
    pub settled: bool,
    pub slashed: bool,
}

/// Singletons + lookups.
pub const PARAMS: Item<Params> = Item::new("params");
pub const TASK_COUNTER: Item<u64> = Item::new("task_counter");
pub const BURNED_TOTAL: Item<Uint128> = Item::new("burned_total");

/// Primary maps.
pub const AGENTS: Map<&Addr, AgentRecord> = Map::new("agents");
pub const TASKS: Map<u64, Task> = Map::new("tasks");

/// Fraud-proof attestations, keyed by (task_id, attestor). Presence means
/// the attestor voted to slash this task. Count once for quorum.
pub const FRAUD_VOTES: Map<(u64, &Addr), ()> = Map::new("fraud_votes");
