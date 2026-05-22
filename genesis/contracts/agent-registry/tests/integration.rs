//! End-to-end integration tests for the agent-registry contract using
//! cw-multi-test. These tests run the *full lifecycle* against a simulated
//! multi-contract environment:
//!
//!   1. Instantiate the contract with realistic params.
//!   2. Two simulated users: `operator` (the AI agent) and `requester`
//!      (the human paying for work). Each is funded with native GEN.
//!   3. Operator calls RegisterAgent — verifies on-chain that stake was
//!      escrowed from the operator's wallet into the contract's wallet.
//!   4. Requester calls CreateTask — bounty moves requester → contract.
//!   5. Operator calls SubmitResponse — task state advances.
//!   6. Requester calls SettleTask — verifies the 50/30/20 split moves
//!      coins to (agent / treasury / burn_sink) at the exact expected
//!      amounts, and the contract's balance returns to (only) the
//!      original stake.
//!   7. Verifies reputation incremented + burned-total counter updated.
//!
//! This is the closest thing to "running the chain" that's possible
//! without deploying to Neutron testnet. Every step uses real CosmWasm
//! storage, real BankMsg routing, real address rules.

use cosmwasm_std::{coin, coins, Addr, Decimal, Uint128};
use cw_multi_test::{App, AppBuilder, ContractWrapper, Executor};

use agentic_registry::contract;
use agentic_registry::msg::{
    AgentResponse, BurnedTotalResponse, ExecuteMsg, InstantiateMsg, QueryMsg, TaskResponse,
};

const DENOM: &str = "ugen";

// Helper: build the contract wrapper for cw-multi-test.
fn registry_code() -> Box<dyn cw_multi_test::Contract<cosmwasm_std::Empty>> {
    Box::new(ContractWrapper::new(
        contract::execute,
        contract::instantiate,
        contract::query,
    ))
}

// Set up an App with two pre-funded test accounts. cw-multi-test 2.x
// requires addresses to be valid bech32 — we mint them via `api.addr_make`
// instead of `Addr::unchecked`, then thread the resulting addresses out.
fn setup_app() -> (App, Addr, Addr, Addr, Addr) {
    // Pre-compute the addresses the same way MockApi.addr_make will: the
    // identifier is what gets bech32-encoded.
    let mut operator = Addr::unchecked("placeholder");
    let mut requester = Addr::unchecked("placeholder");
    let mut treasury = Addr::unchecked("placeholder");
    let mut burn_sink = Addr::unchecked("placeholder");

    let app = AppBuilder::new().build(|router, api, storage| {
        operator = api.addr_make("operator");
        requester = api.addr_make("requester");
        treasury = api.addr_make("treasury");
        burn_sink = api.addr_make("burn_sink");

        router
            .bank
            .init_balance(storage, &operator, coins(1_000_000_000_000, DENOM))
            .unwrap();
        router
            .bank
            .init_balance(storage, &requester, coins(10_000_000_000, DENOM))
            .unwrap();
    });

    (app, operator, requester, treasury, burn_sink)
}

fn instantiate_registry(
    app: &mut App,
    sender: &Addr,
    treasury: &Addr,
    burn_sink: &Addr,
) -> Addr {
    let code_id = app.store_code(registry_code());
    let msg = InstantiateMsg {
        admin: None,
        stake_denom: DENOM.into(),
        burn_sink: burn_sink.clone(),
        treasury: treasury.clone(),
        min_agent_stake: Uint128::new(100_000_000),
        min_agent_stake_floor: Uint128::new(10_000_000),
        split_agent: Decimal::percent(50),
        split_treasury: Decimal::percent(30),
        split_burn: Decimal::percent(20),
        fraud_proof_quorum: 3,
        reputation_gain_per_task: 1,
    };
    app.instantiate_contract(code_id, sender.clone(), &msg, &[], "agent-registry", None)
        .unwrap()
}

// Balance helper.
fn bal(app: &App, addr: &Addr) -> u128 {
    app.wrap()
        .query_balance(addr, DENOM)
        .map(|c| c.amount.u128())
        .unwrap_or(0)
}

#[test]
fn full_happy_path_register_create_submit_settle() {
    let (mut app, operator, requester, treasury, burn_sink) = setup_app();
    let admin = Addr::unchecked("cosmos1admin");

    // Sanity: pre-flight balances.
    assert_eq!(bal(&app, &operator), 1_000_000_000_000);
    assert_eq!(bal(&app, &requester), 10_000_000_000);
    assert_eq!(bal(&app, &treasury), 0);
    assert_eq!(bal(&app, &burn_sink), 0);

    let registry = instantiate_registry(&mut app, &admin, &treasury, &burn_sink);
    assert_eq!(bal(&app, &registry), 0);

    // 1. Operator registers — bonds 250 GEN.
    let stake = 250_000_000u128;
    app.execute_contract(
        operator.clone(),
        registry.clone(),
        &ExecuteMsg::RegisterAgent {
            moniker: "pr-reviewer".into(),
            endpoint: "https://reviewer.example.com".into(),
        },
        &coins(stake, DENOM),
    )
    .unwrap();

    // Operator wallet ↓ by stake; contract wallet ↑ by stake.
    assert_eq!(bal(&app, &operator), 1_000_000_000_000 - stake);
    assert_eq!(bal(&app, &registry), stake);

    // 2. Requester creates a task with 1000 ugen bounty (small but exact
    //    for the split math).
    let bounty = 1_000u128;
    app.execute_contract(
        requester.clone(),
        registry.clone(),
        &ExecuteMsg::CreateTask {
            agent: operator.clone(),
            spec: "Review PR #7 on github.com/foo/bar".into(),
        },
        &coins(bounty, DENOM),
    )
    .unwrap();

    assert_eq!(bal(&app, &requester), 10_000_000_000 - bounty);
    assert_eq!(bal(&app, &registry), stake + bounty);

    // 3. Operator submits the response CID.
    app.execute_contract(
        operator.clone(),
        registry.clone(),
        &ExecuteMsg::SubmitResponse {
            task_id: 1,
            response_cid: "QmExampleReviewCid12345".into(),
        },
        &[],
    )
    .unwrap();

    // No coin movement on response submission.
    assert_eq!(bal(&app, &registry), stake + bounty);

    // 4. Requester settles the task — splits 1000 into 500/300/200.
    app.execute_contract(
        requester.clone(),
        registry.clone(),
        &ExecuteMsg::SettleTask { task_id: 1 },
        &[],
    )
    .unwrap();

    // Operator's wallet went up by 500 (agent slice).
    assert_eq!(
        bal(&app, &operator),
        1_000_000_000_000 - stake + 500,
        "agent slice"
    );
    // Treasury got 300.
    assert_eq!(bal(&app, &treasury), 300, "treasury slice");
    // Burn sink got 200.
    assert_eq!(bal(&app, &burn_sink), 200, "burn slice");
    // Requester is down by full bounty.
    assert_eq!(bal(&app, &requester), 10_000_000_000 - bounty);
    // Contract holds only the original stake again — bounty fully disbursed.
    assert_eq!(bal(&app, &registry), stake, "contract holds only stake");

    // Reputation bumped from 0 → 1.
    let agent: AgentResponse = app
        .wrap()
        .query_wasm_smart(
            &registry,
            &QueryMsg::Agent {
                operator: operator.clone(),
            },
        )
        .unwrap();
    let rec = agent.agent.expect("agent should exist");
    assert_eq!(rec.reputation, 1);
    assert_eq!(rec.stake.u128(), stake);
    assert!(!rec.jailed);

    // Task marked settled.
    let task: TaskResponse = app
        .wrap()
        .query_wasm_smart(&registry, &QueryMsg::Task { task_id: 1 })
        .unwrap();
    let t = task.task.expect("task should exist");
    assert!(t.settled);
    assert!(!t.slashed);

    // Burned total = 200.
    let burned: BurnedTotalResponse = app
        .wrap()
        .query_wasm_smart(&registry, &QueryMsg::BurnedTotal {})
        .unwrap();
    assert_eq!(burned.total.u128(), 200);
}

#[test]
fn dust_routes_to_burn() {
    // Bounty of 7 → 7 * 0.50 = 3 (floor), 7 * 0.30 = 2 (floor), burn = 2.
    // Conservation: 3 + 2 + 2 = 7. The dust accumulates into burn.
    let (mut app, operator, requester, treasury, burn_sink) = setup_app();
    let admin = Addr::unchecked("cosmos1admin");
    let registry = instantiate_registry(&mut app, &admin, &treasury, &burn_sink);

    app.execute_contract(
        operator.clone(),
        registry.clone(),
        &ExecuteMsg::RegisterAgent {
            moniker: "a".into(),
            endpoint: "x".into(),
        },
        &coins(100_000_000, DENOM),
    )
    .unwrap();

    app.execute_contract(
        requester.clone(),
        registry.clone(),
        &ExecuteMsg::CreateTask {
            agent: operator.clone(),
            spec: "dust test".into(),
        },
        &coins(7, DENOM),
    )
    .unwrap();
    app.execute_contract(
        operator.clone(),
        registry.clone(),
        &ExecuteMsg::SubmitResponse {
            task_id: 1,
            response_cid: "Qm123".into(),
        },
        &[],
    )
    .unwrap();
    app.execute_contract(
        requester.clone(),
        registry.clone(),
        &ExecuteMsg::SettleTask { task_id: 1 },
        &[],
    )
    .unwrap();

    let op_before_stake = 1_000_000_000_000u128 - 100_000_000u128;
    assert_eq!(bal(&app, &operator), op_before_stake + 3); // agent
    assert_eq!(bal(&app, &treasury), 2); // treasury (floor)
    assert_eq!(bal(&app, &burn_sink), 2); // burn (absorbs dust)
}

#[test]
fn fraud_proof_quorum_slashes_agent() {
    let (mut app, operator, requester, treasury, burn_sink) = setup_app();
    let admin = Addr::unchecked("cosmos1admin");
    let attestor_1 = Addr::unchecked("cosmos1att1");
    let attestor_2 = Addr::unchecked("cosmos1att2");
    let attestor_3 = Addr::unchecked("cosmos1att3");

    let registry = instantiate_registry(&mut app, &admin, &treasury, &burn_sink);

    // Operator registers with 500 GEN stake.
    let stake = 500_000_000u128;
    app.execute_contract(
        operator.clone(),
        registry.clone(),
        &ExecuteMsg::RegisterAgent {
            moniker: "bad-actor".into(),
            endpoint: "x".into(),
        },
        &coins(stake, DENOM),
    )
    .unwrap();

    // Requester creates a 1000 ugen task.
    let bounty = 1_000u128;
    app.execute_contract(
        requester.clone(),
        registry.clone(),
        &ExecuteMsg::CreateTask {
            agent: operator.clone(),
            spec: "this will be fraudulent".into(),
        },
        &coins(bounty, DENOM),
    )
    .unwrap();
    app.execute_contract(
        operator.clone(),
        registry.clone(),
        &ExecuteMsg::SubmitResponse {
            task_id: 1,
            response_cid: "QmFake".into(),
        },
        &[],
    )
    .unwrap();

    // 3 attestors submit fraud proofs (quorum threshold).
    for att in [&attestor_1, &attestor_2, &attestor_3] {
        app.execute_contract(
            att.clone(),
            registry.clone(),
            &ExecuteMsg::SubmitFraudProof {
                task_id: 1,
                evidence: "ipfs://evidence-cid".into(),
            },
            &[],
        )
        .unwrap();
    }

    // Verify: stake burned, bounty refunded, agent jailed.
    let agent: AgentResponse = app
        .wrap()
        .query_wasm_smart(
            &registry,
            &QueryMsg::Agent {
                operator: operator.clone(),
            },
        )
        .unwrap();
    let rec = agent.agent.unwrap();
    assert_eq!(rec.stake.u128(), 0, "stake zeroed");
    assert!(rec.jailed, "agent jailed");
    assert_eq!(rec.reputation, 0);

    // Requester got the bounty back.
    assert_eq!(bal(&app, &requester), 10_000_000_000);
    // Burn sink got the stake.
    assert_eq!(bal(&app, &burn_sink), stake);
    // Treasury untouched.
    assert_eq!(bal(&app, &treasury), 0);
}
