//! Handler functions. Each `execute_*` mirrors one `Msg*` from the original
//! L1 module. The functions are deliberately small and side-effect-free
//! where possible to make unit testing easy.

use cosmwasm_std::{
    entry_point, to_json_binary, Addr, BankMsg, Binary, Coin, CosmosMsg, Decimal, Deps, DepsMut,
    Env, MessageInfo, Order, Response, StdResult, Uint128,
};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{
    require_funds, AgentResponse, BurnedTotalResponse, ExecuteMsg, FraudVoteCountResponse,
    InstantiateMsg, QueryMsg, TaskResponse,
};
use crate::state::{
    AgentRecord, Params, Task, AGENTS, BURNED_TOTAL, FRAUD_VOTES, PARAMS, TASKS, TASK_COUNTER,
};

const CONTRACT_NAME: &str = "crates.io:agentic-registry";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

// ───────────────────────── instantiate ─────────────────────────

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let split_sum = msg.split_agent + msg.split_treasury + msg.split_burn;
    if split_sum != Decimal::one() {
        return Err(ContractError::SplitMustSumToOne);
    }

    let params = Params {
        admin: msg.admin.unwrap_or(info.sender.clone()),
        stake_denom: msg.stake_denom,
        burn_sink: msg.burn_sink,
        treasury: msg.treasury,
        min_agent_stake: msg.min_agent_stake,
        min_agent_stake_floor: msg.min_agent_stake_floor,
        split_agent: msg.split_agent,
        split_treasury: msg.split_treasury,
        split_burn: msg.split_burn,
        fraud_proof_quorum: msg.fraud_proof_quorum,
        reputation_gain_per_task: msg.reputation_gain_per_task,
    };
    PARAMS.save(deps.storage, &params)?;
    TASK_COUNTER.save(deps.storage, &0u64)?;
    BURNED_TOTAL.save(deps.storage, &Uint128::zero())?;

    Ok(Response::new()
        .add_attribute("action", "instantiate")
        .add_attribute("admin", params.admin)
        .add_attribute("stake_denom", params.stake_denom))
}

// ───────────────────────── execute ─────────────────────────

#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::RegisterAgent { moniker, endpoint } => {
            execute_register_agent(deps, info, moniker, endpoint)
        }
        ExecuteMsg::CreateTask { agent, spec } => execute_create_task(deps, info, agent, spec),
        ExecuteMsg::SubmitResponse {
            task_id,
            response_cid,
        } => execute_submit_response(deps, info, task_id, response_cid),
        ExecuteMsg::SettleTask { task_id } => execute_settle_task(deps, env, info, task_id),
        ExecuteMsg::SubmitFraudProof { task_id, evidence } => {
            execute_submit_fraud_proof(deps, env, info, task_id, evidence)
        }
        ExecuteMsg::UpdateParams { params } => execute_update_params(deps, info, *params),
    }
}

fn execute_register_agent(
    deps: DepsMut,
    info: MessageInfo,
    moniker: String,
    endpoint: String,
) -> Result<Response, ContractError> {
    let params = PARAMS.load(deps.storage)?;

    if AGENTS.may_load(deps.storage, &info.sender)?.is_some() {
        return Err(ContractError::AgentAlreadyRegistered {
            operator: info.sender.to_string(),
        });
    }

    let funds = require_funds(&info.funds, &params.stake_denom)
        .ok_or_else(|| ContractError::WrongDenom {
            expected: params.stake_denom.clone(),
            got: info.funds.first().map(|c| c.denom.clone()).unwrap_or_default(),
        })?;

    if funds.amount < params.min_agent_stake {
        return Err(ContractError::StakeBelowMin {
            provided: funds.amount,
            required: params.min_agent_stake,
        });
    }

    let rec = AgentRecord {
        operator: info.sender.clone(),
        moniker: moniker.clone(),
        endpoint,
        stake: funds.amount,
        reputation: 0,
        jailed: false,
    };
    AGENTS.save(deps.storage, &info.sender, &rec)?;

    Ok(Response::new()
        .add_attribute("action", "register_agent")
        .add_attribute("operator", info.sender)
        .add_attribute("moniker", moniker)
        .add_attribute("stake", funds.amount.to_string()))
}

fn execute_create_task(
    deps: DepsMut,
    info: MessageInfo,
    agent: Addr,
    spec: String,
) -> Result<Response, ContractError> {
    let params = PARAMS.load(deps.storage)?;

    let agent_rec = AGENTS
        .may_load(deps.storage, &agent)?
        .ok_or_else(|| ContractError::AgentNotFound {
            operator: agent.to_string(),
        })?;
    if agent_rec.jailed {
        return Err(ContractError::AgentJailed {
            operator: agent.to_string(),
        });
    }

    let funds = require_funds(&info.funds, &params.stake_denom)
        .ok_or_else(|| ContractError::WrongDenom {
            expected: params.stake_denom.clone(),
            got: info.funds.first().map(|c| c.denom.clone()).unwrap_or_default(),
        })?;
    if funds.amount.is_zero() {
        return Err(ContractError::NoFunds);
    }

    let id = TASK_COUNTER.load(deps.storage)? + 1;
    TASK_COUNTER.save(deps.storage, &id)?;

    let task = Task {
        id,
        requester: info.sender.clone(),
        agent: agent.clone(),
        bounty: funds.amount,
        spec,
        response_cid: None,
        settled: false,
        slashed: false,
    };
    TASKS.save(deps.storage, id, &task)?;

    Ok(Response::new()
        .add_attribute("action", "create_task")
        .add_attribute("task_id", id.to_string())
        .add_attribute("requester", info.sender)
        .add_attribute("agent", agent)
        .add_attribute("bounty", funds.amount.to_string()))
}

fn execute_submit_response(
    deps: DepsMut,
    info: MessageInfo,
    task_id: u64,
    response_cid: String,
) -> Result<Response, ContractError> {
    let mut task = TASKS
        .may_load(deps.storage, task_id)?
        .ok_or(ContractError::TaskNotFound { id: task_id })?;

    if task.agent != info.sender {
        return Err(ContractError::Unauthorised {
            expected: task.agent.to_string(),
        });
    }
    if task.settled || task.slashed {
        return Err(ContractError::TaskClosed { id: task_id });
    }
    if task.response_cid.is_some() {
        return Err(ContractError::TaskAlreadyResponded { id: task_id });
    }
    task.response_cid = Some(response_cid.clone());
    TASKS.save(deps.storage, task_id, &task)?;

    Ok(Response::new()
        .add_attribute("action", "submit_response")
        .add_attribute("task_id", task_id.to_string())
        .add_attribute("response_cid", response_cid))
}

fn execute_settle_task(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    task_id: u64,
) -> Result<Response, ContractError> {
    let params = PARAMS.load(deps.storage)?;

    let mut task = TASKS
        .may_load(deps.storage, task_id)?
        .ok_or(ContractError::TaskNotFound { id: task_id })?;
    if task.requester != info.sender {
        return Err(ContractError::Unauthorised {
            expected: task.requester.to_string(),
        });
    }
    if task.settled || task.slashed {
        return Err(ContractError::TaskClosed { id: task_id });
    }
    if task.response_cid.is_none() {
        return Err(ContractError::TaskMissingResponse { id: task_id });
    }

    // Settlement math, identical to genesis/chain/x/agentic/keeper/msg_server.go.
    // Rounding dust is absorbed into the burn slice so agent + treasury + burn
    // == bounty exactly.
    let agent_cut = (Decimal::from_atomics(task.bounty, 0)
        .unwrap_or(Decimal::zero())
        * params.split_agent)
        .to_uint_floor();
    let treas_cut = (Decimal::from_atomics(task.bounty, 0)
        .unwrap_or(Decimal::zero())
        * params.split_treasury)
        .to_uint_floor();
    let burn_cut = task.bounty.checked_sub(agent_cut)?.checked_sub(treas_cut)?;

    // Reputation bump.
    let mut agent_rec = AGENTS
        .may_load(deps.storage, &task.agent)?
        .ok_or_else(|| ContractError::AgentNotFound {
            operator: task.agent.to_string(),
        })?;
    agent_rec.reputation = agent_rec
        .reputation
        .saturating_add(params.reputation_gain_per_task);
    AGENTS.save(deps.storage, &task.agent, &agent_rec)?;

    // Update burn counter.
    BURNED_TOTAL.update(deps.storage, |t| -> StdResult<_> {
        Ok(t.checked_add(burn_cut)?)
    })?;

    task.settled = true;
    TASKS.save(deps.storage, task_id, &task)?;

    let msgs = payout_msgs(
        &task.agent,
        &params.treasury,
        &params.burn_sink,
        &params.stake_denom,
        agent_cut,
        treas_cut,
        burn_cut,
    );

    Ok(Response::new()
        .add_messages(msgs)
        .add_attribute("action", "settle_task")
        .add_attribute("task_id", task_id.to_string())
        .add_attribute("agent_paid", agent_cut.to_string())
        .add_attribute("treasury_paid", treas_cut.to_string())
        .add_attribute("burned", burn_cut.to_string()))
}

fn execute_submit_fraud_proof(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    task_id: u64,
    evidence: String,
) -> Result<Response, ContractError> {
    let params = PARAMS.load(deps.storage)?;

    let mut task = TASKS
        .may_load(deps.storage, task_id)?
        .ok_or(ContractError::TaskNotFound { id: task_id })?;
    if task.settled || task.slashed {
        return Err(ContractError::TaskClosed { id: task_id });
    }

    // Deduplicate attestations.
    if FRAUD_VOTES.has(deps.storage, (task_id, &info.sender)) {
        return Err(ContractError::AlreadyVoted {
            id: task_id,
            attestor: info.sender.to_string(),
        });
    }
    FRAUD_VOTES.save(deps.storage, (task_id, &info.sender), &())?;

    // Count and check quorum.
    let count: u32 = FRAUD_VOTES
        .prefix(task_id)
        .keys(deps.storage, None, None, Order::Ascending)
        .count() as u32;

    if count < params.fraud_proof_quorum {
        return Ok(Response::new()
            .add_attribute("action", "fraud_proof_attestation")
            .add_attribute("task_id", task_id.to_string())
            .add_attribute("count", format!("{}/{}", count, params.fraud_proof_quorum)));
    }

    // Quorum reached → slash.
    let mut agent_rec = AGENTS
        .may_load(deps.storage, &task.agent)?
        .ok_or_else(|| ContractError::AgentNotFound {
            operator: task.agent.to_string(),
        })?;
    let stake_to_burn = agent_rec.stake;
    agent_rec.stake = Uint128::zero();
    agent_rec.reputation = 0;
    agent_rec.jailed = true;
    AGENTS.save(deps.storage, &task.agent, &agent_rec)?;

    BURNED_TOTAL.update(deps.storage, |t| -> StdResult<_> {
        Ok(t.checked_add(stake_to_burn)?)
    })?;

    task.slashed = true;
    TASKS.save(deps.storage, task_id, &task)?;

    let mut msgs = vec![];
    if !stake_to_burn.is_zero() {
        msgs.push(send_one(
            &params.burn_sink,
            &params.stake_denom,
            stake_to_burn,
        ));
    }
    // Refund bounty.
    msgs.push(send_one(
        &task.requester,
        &params.stake_denom,
        task.bounty,
    ));

    Ok(Response::new()
        .add_messages(msgs)
        .add_attribute("action", "slash")
        .add_attribute("task_id", task_id.to_string())
        .add_attribute("agent", task.agent)
        .add_attribute("stake_burned", stake_to_burn.to_string())
        .add_attribute("bounty_refunded", task.bounty.to_string())
        .add_attribute("evidence", evidence))
}

fn execute_update_params(
    deps: DepsMut,
    info: MessageInfo,
    new_params: Params,
) -> Result<Response, ContractError> {
    let current = PARAMS.load(deps.storage)?;
    if info.sender != current.admin {
        return Err(ContractError::Unauthorised {
            expected: current.admin.to_string(),
        });
    }
    let sum = new_params.split_agent + new_params.split_treasury + new_params.split_burn;
    if sum != Decimal::one() {
        return Err(ContractError::SplitMustSumToOne);
    }
    PARAMS.save(deps.storage, &new_params)?;
    Ok(Response::new().add_attribute("action", "update_params"))
}

// ───────────────────────── query ─────────────────────────

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Params {} => to_json_binary(&PARAMS.load(deps.storage)?),
        QueryMsg::Agent { operator } => {
            let agent = AGENTS.may_load(deps.storage, &operator)?;
            to_json_binary(&AgentResponse { agent })
        }
        QueryMsg::Task { task_id } => {
            let task = TASKS.may_load(deps.storage, task_id)?;
            to_json_binary(&TaskResponse { task })
        }
        QueryMsg::BurnedTotal {} => to_json_binary(&BurnedTotalResponse {
            total: BURNED_TOTAL.load(deps.storage)?,
        }),
        QueryMsg::FraudVoteCount { task_id } => {
            let params = PARAMS.load(deps.storage)?;
            let count = FRAUD_VOTES
                .prefix(task_id)
                .keys(deps.storage, None, None, Order::Ascending)
                .count() as u32;
            to_json_binary(&FraudVoteCountResponse {
                count,
                quorum: params.fraud_proof_quorum,
            })
        }
    }
}

// ───────────────────────── helpers ─────────────────────────

fn payout_msgs(
    agent: &Addr,
    treasury: &Addr,
    burn_sink: &Addr,
    denom: &str,
    agent_cut: Uint128,
    treas_cut: Uint128,
    burn_cut: Uint128,
) -> Vec<CosmosMsg> {
    let mut out = Vec::with_capacity(3);
    if !agent_cut.is_zero() {
        out.push(send_one(agent, denom, agent_cut));
    }
    if !treas_cut.is_zero() {
        out.push(send_one(treasury, denom, treas_cut));
    }
    if !burn_cut.is_zero() {
        out.push(send_one(burn_sink, denom, burn_cut));
    }
    out
}

fn send_one(to: &Addr, denom: &str, amount: Uint128) -> CosmosMsg {
    CosmosMsg::Bank(BankMsg::Send {
        to_address: to.to_string(),
        amount: vec![Coin {
            denom: denom.to_string(),
            amount,
        }],
    })
}

// ───────────────────────── tests ─────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{coins, Addr};

    // Tests build the InstantiateMsg without touching deps.api so the borrow
    // checker doesn't catch a simultaneous mutable + immutable borrow on
    // `deps`. MockApi.addr_make is otherwise the canonical way to mint test
    // addresses, but Addr::unchecked is acceptable in unit tests because we
    // never round-trip these through canonicalize.
    fn default_params() -> InstantiateMsg {
        InstantiateMsg {
            admin: None,
            stake_denom: "ugen".into(),
            burn_sink: Addr::unchecked("cosmos1burn"),
            treasury: Addr::unchecked("cosmos1treasury"),
            min_agent_stake: Uint128::new(100_000_000), // 100 GEN
            min_agent_stake_floor: Uint128::new(10_000_000),
            split_agent: Decimal::percent(50),
            split_treasury: Decimal::percent(30),
            split_burn: Decimal::percent(20),
            fraud_proof_quorum: 3,
            reputation_gain_per_task: 1,
        }
    }

    #[test]
    fn instantiate_sets_params() {
        let mut deps = mock_dependencies();
        let admin = deps.api.addr_make("admin");
        let info = mock_info(admin.as_str(), &[]);
        let msg = default_params();
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert!(res.attributes.iter().any(|a| a.key == "action"));

        let p = PARAMS.load(deps.as_ref().storage).unwrap();
        assert_eq!(p.stake_denom, "ugen");
        assert_eq!(p.split_agent + p.split_treasury + p.split_burn, Decimal::one());
    }

    #[test]
    fn instantiate_rejects_bad_split() {
        let mut deps = mock_dependencies();
        let admin = deps.api.addr_make("admin");
        let info = mock_info(admin.as_str(), &[]);
        let mut msg = default_params();
        msg.split_burn = Decimal::percent(25); // sum = 1.05
        let err = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap_err();
        assert!(matches!(err, ContractError::SplitMustSumToOne));
    }

    #[test]
    fn register_agent_escrows_stake() {
        let mut deps = mock_dependencies();
        let admin = deps.api.addr_make("admin");
        instantiate(
            deps.as_mut(),
            mock_env(),
            mock_info(admin.as_str(), &[]),
            default_params(),
        )
        .unwrap();

        let operator = deps.api.addr_make("operator");
        let info = mock_info(operator.as_str(), &coins(100_000_000, "ugen"));
        let res = execute(
            deps.as_mut(),
            mock_env(),
            info,
            ExecuteMsg::RegisterAgent {
                moniker: "priya".into(),
                endpoint: "https://priya.test".into(),
            },
        )
        .unwrap();
        assert!(res.attributes.iter().any(|a| a.key == "moniker"));

        let stored = AGENTS.load(deps.as_ref().storage, &operator).unwrap();
        assert_eq!(stored.stake.u128(), 100_000_000);
        assert_eq!(stored.reputation, 0);
        assert!(!stored.jailed);
    }

    #[test]
    fn register_agent_rejects_low_stake() {
        let mut deps = mock_dependencies();
        let admin = deps.api.addr_make("admin");
        instantiate(
            deps.as_mut(),
            mock_env(),
            mock_info(admin.as_str(), &[]),
            default_params(),
        )
        .unwrap();

        let operator = deps.api.addr_make("operator");
        let info = mock_info(operator.as_str(), &coins(1_000_000, "ugen")); // 1 GEN, below 100
        let err = execute(
            deps.as_mut(),
            mock_env(),
            info,
            ExecuteMsg::RegisterAgent {
                moniker: "priya".into(),
                endpoint: "x".into(),
            },
        )
        .unwrap_err();
        assert!(matches!(err, ContractError::StakeBelowMin { .. }));
    }

    #[test]
    fn settle_task_payouts_match_split() {
        let mut deps = mock_dependencies();
        let admin = deps.api.addr_make("admin");
        instantiate(
            deps.as_mut(),
            mock_env(),
            mock_info(admin.as_str(), &[]),
            default_params(),
        )
        .unwrap();

        let agent = deps.api.addr_make("agent");
        execute(
            deps.as_mut(),
            mock_env(),
            mock_info(agent.as_str(), &coins(100_000_000, "ugen")),
            ExecuteMsg::RegisterAgent {
                moniker: "a".into(),
                endpoint: "x".into(),
            },
        )
        .unwrap();

        // create task with bounty 1000
        let requester = deps.api.addr_make("requester");
        execute(
            deps.as_mut(),
            mock_env(),
            mock_info(requester.as_str(), &coins(1_000, "ugen")),
            ExecuteMsg::CreateTask {
                agent: agent.clone(),
                spec: "test".into(),
            },
        )
        .unwrap();

        execute(
            deps.as_mut(),
            mock_env(),
            mock_info(agent.as_str(), &[]),
            ExecuteMsg::SubmitResponse {
                task_id: 1,
                response_cid: "Qm123".into(),
            },
        )
        .unwrap();

        let res = execute(
            deps.as_mut(),
            mock_env(),
            mock_info(requester.as_str(), &[]),
            ExecuteMsg::SettleTask { task_id: 1 },
        )
        .unwrap();

        // 1000 * 0.50 = 500, 1000 * 0.30 = 300, burn = 200
        let agent_attr = res
            .attributes
            .iter()
            .find(|a| a.key == "agent_paid")
            .unwrap();
        let treas_attr = res
            .attributes
            .iter()
            .find(|a| a.key == "treasury_paid")
            .unwrap();
        let burn_attr = res.attributes.iter().find(|a| a.key == "burned").unwrap();
        assert_eq!(agent_attr.value, "500");
        assert_eq!(treas_attr.value, "300");
        assert_eq!(burn_attr.value, "200");

        // reputation bumped
        let stored = AGENTS.load(deps.as_ref().storage, &agent).unwrap();
        assert_eq!(stored.reputation, 1);

        // total burned recorded
        assert_eq!(BURNED_TOTAL.load(deps.as_ref().storage).unwrap().u128(), 200);
    }
}
