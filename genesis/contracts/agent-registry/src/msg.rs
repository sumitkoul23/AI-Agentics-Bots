use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Addr, Coin, Decimal, Uint128};

use crate::state::{AgentRecord, Params, Task};

/// One-shot init.
#[cw_serde]
pub struct InstantiateMsg {
    /// Initial Params. If `admin` is None we use info.sender.
    pub admin: Option<Addr>,
    pub stake_denom: String,
    pub burn_sink: Addr,
    pub treasury: Addr,
    pub min_agent_stake: Uint128,
    pub min_agent_stake_floor: Uint128,
    pub split_agent: Decimal,
    pub split_treasury: Decimal,
    pub split_burn: Decimal,
    pub fraud_proof_quorum: u32,
    pub reputation_gain_per_task: u64,
}

#[cw_serde]
pub enum ExecuteMsg {
    /// Bond stake from `info.funds` and create an `AgentRecord` keyed on
    /// `info.sender`. Stake must be ≥ `params.min_agent_stake` and must be
    /// in `params.stake_denom`.
    RegisterAgent { moniker: String, endpoint: String },

    /// Escrow a bounty (from `info.funds`) and assign it to `agent`. Mints
    /// a new task id.
    CreateTask { agent: Addr, spec: String },

    /// Agent posts the response CID. Only callable by `task.agent`.
    SubmitResponse {
        task_id: u64,
        response_cid: String,
    },

    /// Requester closes the task. Splits the escrow 50/30/20 (agent /
    /// treasury / burn) and bumps agent reputation.
    SettleTask { task_id: u64 },

    /// Submit a fraud-proof attestation. Once the count reaches
    /// `params.fraud_proof_quorum`, the agent's stake is burned and the
    /// bounty is refunded.
    SubmitFraudProof {
        task_id: u64,
        evidence: String,
    },

    /// Admin-only.
    UpdateParams { params: Box<Params> },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Params)]
    Params {},

    #[returns(AgentResponse)]
    Agent { operator: Addr },

    #[returns(TaskResponse)]
    Task { task_id: u64 },

    #[returns(BurnedTotalResponse)]
    BurnedTotal {},

    #[returns(FraudVoteCountResponse)]
    FraudVoteCount { task_id: u64 },

    /// Tasks assigned to `agent` that are neither settled nor slashed and
    /// where `response_cid` is still `None`. Used by an agent operator's
    /// watch loop to find work without scanning every historical task.
    /// O(n) where n is the total number of tasks ever created; v1 swaps
    /// in a secondary `(agent, status) → task_id` index.
    #[returns(OpenTasksForAgentResponse)]
    OpenTasksForAgent { agent: Addr },

    /// Highest task ID ever issued. Cheap; useful for backfill scans by
    /// off-chain tooling.
    #[returns(LastTaskIdResponse)]
    LastTaskId {},
}

#[cw_serde]
pub struct AgentResponse {
    pub agent: Option<AgentRecord>,
}

#[cw_serde]
pub struct TaskResponse {
    pub task: Option<Task>,
}

#[cw_serde]
pub struct BurnedTotalResponse {
    pub total: Uint128,
}

#[cw_serde]
pub struct FraudVoteCountResponse {
    pub count: u32,
    pub quorum: u32,
}

#[cw_serde]
pub struct OpenTasksForAgentResponse {
    pub tasks: Vec<crate::state::Task>,
}

#[cw_serde]
pub struct LastTaskIdResponse {
    pub task_id: u64,
}

/// Internal payload describing what funds a tx sent. Used by handlers to
/// validate `info.funds` shape.
#[cw_serde]
pub struct ExpectedFunds<'a> {
    pub denom: &'a str,
    pub amount: Uint128,
}

/// Helper to find a single coin of the expected denom in `funds`.
pub fn require_funds<'a>(funds: &'a [Coin], denom: &str) -> Option<&'a Coin> {
    funds.iter().find(|c| c.denom == denom)
}
