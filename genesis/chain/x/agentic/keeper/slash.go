package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/types/coinconst"
	"github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

// slashAgentAndCloseTask is the terminal path of a successful fraud proof:
//   1. Burn the agent's entire bonded stake (held by this module's account).
//   2. Reset the agent's reputation to 0.
//   3. Jail the agent (cannot be assigned new tasks until governance unjails).
//   4. Refund the original task bounty to the requester.
//   5. Mark the task as slashed.
//
// The function is private to the keeper because the only callsite is
// msg_server.SubmitFraudProof once the quorum threshold is reached.
func (k Keeper) slashAgentAndCloseTask(ctx context.Context, task *types.Task) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	rec, err := k.Agents.Get(ctx, task.Agent)
	if err != nil {
		return fmt.Errorf("agent %s vanished mid-slash", task.Agent)
	}
	stake, ok := math.NewIntFromString(rec.StakeUgen)
	if !ok {
		return fmt.Errorf("corrupt stake for %s", rec.Operator)
	}

	// 1. Burn entire stake.
	if err := k.BurnFromEscrow(sdkCtx, stake, "slash:task:"+fmt.Sprint(task.ID)); err != nil {
		return err
	}

	// 2 + 3. Reset reputation, jail, zero out stake.
	rec.Reputation = 0
	rec.Jailed = true
	rec.StakeUgen = math.ZeroInt().String()
	if err := k.Agents.Set(ctx, rec.Operator, rec); err != nil {
		return err
	}

	// 4. Refund the bounty to the requester. The bounty has been sitting in
	// the module account since CreateTask; this is just the unwind.
	bounty, _ := math.NewIntFromString(task.BountyUgen)
	requesterAddr := sdk.MustAccAddressFromBech32(task.Requester)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, requesterAddr,
		sdk.NewCoins(sdk.NewCoin(coinconst.BaseCoinUnit, bounty))); err != nil {
		return fmt.Errorf("refund requester: %w", err)
	}

	// 5. Mark task as slashed.
	task.Slashed = true
	if err := k.Tasks.Set(ctx, task.ID, *task); err != nil {
		return err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agentic.slash",
		sdk.NewAttribute("task_id", fmt.Sprintf("%d", task.ID)),
		sdk.NewAttribute("agent", rec.Operator),
		sdk.NewAttribute("stake_burned", stake.String()),
		sdk.NewAttribute("bounty_refunded", bounty.String()),
	))
	return nil
}

// authModuleAddress resolves a module-account address by name. Used by the
// settle path to route the validator slice into the fee collector.
func (k Keeper) authModuleAddress(name string) sdk.AccAddress {
	return authtypes.NewModuleAddress(name)
}

// countFraudAttestations + recordFraudAttestation are stubs in v0 — they
// satisfy the msg_server interface so the package compiles, but the real
// implementation uses a separate collections.Map keyed by (taskID, attestor).
// The v0 single-attestor path effectively requires `FraudProofQuorum == 1`
// to slash; until the multi-attestor store ships, set the param accordingly
// via `MsgUpdateParams`.
func (k Keeper) countFraudAttestations(_ context.Context, _ uint64) uint64 {
	return 0
}
func (k Keeper) recordFraudAttestation(_ context.Context, _ uint64, _ string, _ string) error {
	return nil
}
