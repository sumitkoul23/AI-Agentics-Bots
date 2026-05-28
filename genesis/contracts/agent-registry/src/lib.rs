//! SKYMETRIC agent registry — the CosmWasm successor to `genesis/chain/x/agentic/`.
//!
//! State model:
//!  - `AgentRecord` keyed by operator address. Holds bonded stake (escrowed
//!    on the contract), reputation score, and jailed flag.
//!  - `Task` keyed by auto-incrementing u64. Records requester, agent,
//!    bounty (escrowed), spec, and the agent's response CID once submitted.
//!  - `FraudVote` keyed by (task_id, attestor). Validator-quorum slashing.
//!
//! Settlement math matches the Cosmos SDK module exactly:
//!  - Bounty splits 50/30/20 — agent / treasury / burn
//!  - Burns are routed to a configured null address (or Token Factory burn
//!    permission once we move off CW20)
//!  - Reputation increments by `params.reputation_gain_per_task` on settle
//!  - Slashing burns entire agent stake + refunds the task bounty
//!
//! The five user-facing operations are the same five `Msg*` types from the
//! L1 module:
//!   RegisterAgent · CreateTask · SubmitResponse · SettleTask · SubmitFraudProof
//!
//! Plus admin-only:
//!   UpdateParams (gated by `params.admin`; will be a gov module address once
//!   on-chain governance ships)
//!
//! This file is the public entry; the meat lives in:
//!   contract.rs — instantiate / execute / query handlers
//!   msg.rs      — ExecuteMsg, QueryMsg, *Response types
//!   state.rs    — cw-storage-plus Item / Map definitions
//!   error.rs    — ContractError variants

pub mod contract;
pub mod error;
pub mod msg;
pub mod state;

pub use crate::error::ContractError;
