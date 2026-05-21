// msg_server.go — handlers for every Msg in the agentic module. Each
// handler returns a typed response and emits a typed event; the event
// stream is what the explorer / oracles consume off-chain.
package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/types/coinconst"
	"github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

type msgServer struct{ Keeper }

// NewMsgServerImpl wires this server into the module's Msg router.
// `module.go::RegisterServices` calls this.
func NewMsgServerImpl(k Keeper) msgServer { return msgServer{Keeper: k} }

// ───────────────────────── RegisterAgent ─────────────────────────

func (s msgServer) RegisterAgent(ctx context.Context, msg *types.MsgRegisterAgent) (*types.MsgRegisterAgentResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	operator := sdk.MustAccAddressFromBech32(msg.Operator)

	params, err := s.Params.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "load params")
	}
	if msg.Stake.LT(params.MinAgentStake) {
		return nil, fmt.Errorf("stake %s below min %s", msg.Stake, params.MinAgentStake)
	}
	if _, err := s.Agents.Get(ctx, msg.Operator); err == nil {
		return nil, fmt.Errorf("operator %s already registered", msg.Operator)
	}

	coins := sdk.NewCoins(sdk.NewCoin(coinconst.BaseCoinUnit, msg.Stake))
	if err := s.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, operator, types.ModuleName, coins); err != nil {
		return nil, errors.Wrap(err, "escrow stake")
	}

	rec := types.AgentRecord{
		Operator:  msg.Operator,
		Moniker:   msg.Moniker,
		Endpoint:  msg.Endpoint,
		StakeUgen: msg.Stake.String(),
	}
	if err := s.Agents.Set(ctx, msg.Operator, rec); err != nil {
		return nil, errors.Wrap(err, "persist agent")
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agentic.register_agent",
		sdk.NewAttribute("operator", msg.Operator),
		sdk.NewAttribute("moniker", msg.Moniker),
		sdk.NewAttribute("stake", coins.String()),
	))
	return &types.MsgRegisterAgentResponse{}, nil
}

// ───────────────────────── CreateTask ─────────────────────────

func (s msgServer) CreateTask(ctx context.Context, msg *types.MsgCreateTask) (*types.MsgCreateTaskResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	requester := sdk.MustAccAddressFromBech32(msg.Requester)

	agent, err := s.Agents.Get(ctx, msg.Agent)
	if err != nil {
		return nil, fmt.Errorf("agent %s not registered", msg.Agent)
	}
	if agent.Jailed {
		return nil, fmt.Errorf("agent %s is jailed", msg.Agent)
	}

	bountyCoins := sdk.NewCoins(sdk.NewCoin(coinconst.BaseCoinUnit, msg.Bounty))
	if err := s.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, requester, types.ModuleName, bountyCoins); err != nil {
		return nil, errors.Wrap(err, "escrow bounty")
	}

	id, err := s.TaskCounter.Next(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "increment task counter")
	}
	id++ // counters start at 0; task IDs start at 1

	task := types.Task{
		ID:         id,
		Requester:  msg.Requester,
		Agent:      msg.Agent,
		BountyUgen: msg.Bounty.String(),
		Spec:       msg.Spec,
	}
	if err := s.Tasks.Set(ctx, id, task); err != nil {
		return nil, errors.Wrap(err, "persist task")
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agentic.create_task",
		sdk.NewAttribute("task_id", fmt.Sprintf("%d", id)),
		sdk.NewAttribute("requester", msg.Requester),
		sdk.NewAttribute("agent", msg.Agent),
		sdk.NewAttribute("bounty", bountyCoins.String()),
	))
	return &types.MsgCreateTaskResponse{TaskID: id}, nil
}

// ───────────────────────── SubmitResponse ─────────────────────────

func (s msgServer) SubmitResponse(ctx context.Context, msg *types.MsgSubmitResponse) (*types.MsgSubmitResponseResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	task, err := s.Tasks.Get(ctx, msg.TaskID)
	if err != nil {
		return nil, fmt.Errorf("task %d not found", msg.TaskID)
	}
	if task.Agent != msg.Agent {
		return nil, fmt.Errorf("task %d not assigned to %s", msg.TaskID, msg.Agent)
	}
	if task.Settled || task.Slashed {
		return nil, fmt.Errorf("task %d already closed", msg.TaskID)
	}
	if task.ResponseCID != "" {
		return nil, fmt.Errorf("task %d already has a response", msg.TaskID)
	}
	task.ResponseCID = msg.ResponseCID
	if err := s.Tasks.Set(ctx, task.ID, task); err != nil {
		return nil, errors.Wrap(err, "persist task response")
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		"agentic.submit_response",
		sdk.NewAttribute("task_id", fmt.Sprintf("%d", task.ID)),
		sdk.NewAttribute("response_cid", msg.ResponseCID),
	))
	return &types.MsgSubmitResponseResponse{}, nil
}

// ───────────────────────── SettleTask ─────────────────────────
//
// Bounty split (from Params): SplitAgent / SplitValidators / SplitBurn.
// Rounding is allocated by the formula:
//   agent      = floor(bounty * SplitAgent)
//   validators = floor(bounty * SplitValidators)
//   burn       = bounty - agent - validators   // absorbs any rounding dust
// This guarantees agent + validators + burn == bounty exactly.

func (s msgServer) SettleTask(ctx context.Context, msg *types.MsgSettleTask) (*types.MsgSettleTaskResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	task, err := s.Tasks.Get(ctx, msg.TaskID)
	if err != nil {
		return nil, fmt.Errorf("task %d not found", msg.TaskID)
	}
	if task.Requester != msg.Requester {
		return nil, fmt.Errorf("only requester can settle task %d", msg.TaskID)
	}
	if task.Settled || task.Slashed {
		return nil, fmt.Errorf("task %d already closed", msg.TaskID)
	}
	if task.ResponseCID == "" {
		return nil, fmt.Errorf("task %d has no response yet", msg.TaskID)
	}

	params, err := s.Params.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "load params")
	}
	bounty, ok := math.NewIntFromString(task.BountyUgen)
	if !ok {
		return nil, fmt.Errorf("corrupt bounty for task %d", task.ID)
	}

	agentCut := bounty.ToLegacyDec().Mul(params.SplitAgent).TruncateInt()
	valCut := bounty.ToLegacyDec().Mul(params.SplitValidators).TruncateInt()
	burnCut := bounty.Sub(agentCut).Sub(valCut)

	agentAddr := sdk.MustAccAddressFromBech32(task.Agent)
	if err := s.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, agentAddr,
		sdk.NewCoins(sdk.NewCoin(coinconst.BaseCoinUnit, agentCut))); err != nil {
		return nil, errors.Wrap(err, "pay agent")
	}

	// Validator slice is routed to the standard fee collector — x/distribution
	// then sprays it across validators proportionally to voting power.
	feeCollector := s.authModuleAddress("fee_collector")
	if err := s.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, feeCollector,
		sdk.NewCoins(sdk.NewCoin(coinconst.BaseCoinUnit, valCut))); err != nil {
		return nil, errors.Wrap(err, "pay validators")
	}

	if err := s.BurnFromEscrow(sdkCtx, burnCut, "settle:task:"+fmt.Sprint(task.ID)); err != nil {
		return nil, err
	}

	// Bump agent reputation.
	rec, err := s.Agents.Get(ctx, task.Agent)
	if err != nil {
		return nil, fmt.Errorf("agent %s vanished", task.Agent)
	}
	rec.Reputation += params.ReputationGainPerTask
	if err := s.Agents.Set(ctx, rec.Operator, rec); err != nil {
		return nil, errors.Wrap(err, "update reputation")
	}

	task.Settled = true
	if err := s.Tasks.Set(ctx, task.ID, task); err != nil {
		return nil, errors.Wrap(err, "persist settled task")
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agentic.settle_task",
		sdk.NewAttribute("task_id", fmt.Sprintf("%d", task.ID)),
		sdk.NewAttribute("agent", task.Agent),
		sdk.NewAttribute("agent_paid", agentCut.String()),
		sdk.NewAttribute("validators_paid", valCut.String()),
		sdk.NewAttribute("burned", burnCut.String()),
		sdk.NewAttribute("new_reputation", fmt.Sprint(rec.Reputation)),
	))
	return &types.MsgSettleTaskResponse{}, nil
}

// ───────────────────────── SubmitFraudProof ─────────────────────────
//
// Quorum semantics: every validator-signed attestation increments a counter.
// Once the counter ≥ FraudProofQuorum, the agent's stake is burned, their
// reputation resets to 0, and they are jailed.
//
// v0 simplification: any address that holds ≥ 1 GEN voting weight is treated
// as an attestor. Production hardening replaces this with the actual
// staking-module ValidatorAddress check.

func (s msgServer) SubmitFraudProof(ctx context.Context, msg *types.MsgSubmitFraudProof) (*types.MsgSubmitFraudProofResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	task, err := s.Tasks.Get(ctx, msg.TaskID)
	if err != nil {
		return nil, fmt.Errorf("task %d not found", msg.TaskID)
	}
	if task.Settled || task.Slashed {
		return nil, fmt.Errorf("task %d already closed", msg.TaskID)
	}

	count := s.countFraudAttestations(ctx, msg.TaskID) + 1
	if err := s.recordFraudAttestation(ctx, msg.TaskID, msg.Attestor, msg.Evidence); err != nil {
		return nil, err
	}

	params, _ := s.Params.Get(ctx)
	if count < uint64(params.FraudProofQuorum) {
		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			"agentic.fraud_proof_attestation",
			sdk.NewAttribute("task_id", fmt.Sprintf("%d", msg.TaskID)),
			sdk.NewAttribute("attestor", msg.Attestor),
			sdk.NewAttribute("count", fmt.Sprintf("%d/%d", count, params.FraudProofQuorum)),
		))
		return &types.MsgSubmitFraudProofResponse{}, nil
	}

	// Quorum reached → slash.
	if err := s.slashAgentAndCloseTask(ctx, &task); err != nil {
		return nil, err
	}
	return &types.MsgSubmitFraudProofResponse{}, nil
}

// ───────────────────────── UpdateParams (gov-only) ─────────────────────────

func (s msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg.Authority != s.authority {
		return nil, fmt.Errorf("expected authority %s, got %s", s.authority, msg.Authority)
	}
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := s.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
