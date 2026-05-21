use cosmwasm_std::{OverflowError, StdError, Uint128};
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("{0}")]
    Overflow(#[from] OverflowError),

    #[error("unauthorised: only {expected} may call this")]
    Unauthorised { expected: String },

    #[error("agent {operator} already registered")]
    AgentAlreadyRegistered { operator: String },

    #[error("agent {operator} not found")]
    AgentNotFound { operator: String },

    #[error("agent {operator} is jailed")]
    AgentJailed { operator: String },

    #[error("task {id} not found")]
    TaskNotFound { id: u64 },

    #[error("task {id} is already closed")]
    TaskClosed { id: u64 },

    #[error("task {id} has no response yet")]
    TaskMissingResponse { id: u64 },

    #[error("task {id} already has a response")]
    TaskAlreadyResponded { id: u64 },

    #[error("stake {provided} below minimum {required}")]
    StakeBelowMin { provided: Uint128, required: Uint128 },

    #[error("expected stake denom {expected}, got {got}")]
    WrongDenom { expected: String, got: String },

    #[error("split shares must sum to 1.0")]
    SplitMustSumToOne,

    #[error("attestor {attestor} already voted on task {id}")]
    AlreadyVoted { id: u64, attestor: String },

    #[error("no funds provided")]
    NoFunds,
}
