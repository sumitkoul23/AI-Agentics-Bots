package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticrouter/types"
)

// RouteSwap executes the multi-hop plan atomically.
//
//   1. Pull user funds + carve the router fee (5 bps → treasury).
//   2. Iterate hops:
//        a. NativeAMM   → call into x/agenticdex.Swap synchronously
//        b. IBCTransfer → SendTransfer + park as PendingRoute (v1)
//        c. RemoteAMM   → ICA call (v1)
//   3. If every hop is synchronous, pay the final output back to the user
//      in the same tx. Otherwise return a PendingRoute id and wait for IBC
//      acks (handled by `OnAcknowledgement` in the IBC middleware — v1).
//
// v0 supports only NativeAMM hops; the other branches are stubbed to
// return an explicit "not yet" error so misuse fails loudly.
func (k Keeper) RouteSwap(ctx context.Context, user sdk.AccAddress, in sdk.Coin, denomOut string, minOut math.Int, hops []types.Hop) (sdk.Coin, uint64, bool, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return sdk.Coin{}, 0, false, err
	}

	// Pull user funds.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, user, types.ModuleName, sdk.NewCoins(in)); err != nil {
		return sdk.Coin{}, 0, false, fmt.Errorf("escrow input: %w", err)
	}

	// Carve the router fee.
	feeAmt := in.Amount.MulRaw(int64(params.RouterFeeBps)).QuoRaw(10_000)
	netIn := in.Amount.Sub(feeAmt)
	if feeAmt.IsPositive() {
		feeCoin := sdk.NewCoin(in.Denom, feeAmt)
		if err := k.bankKeeper.SendCoinsFromModuleToModule(sdkCtx, types.ModuleName, "fee_collector", sdk.NewCoins(feeCoin)); err != nil {
			return sdk.Coin{}, 0, false, fmt.Errorf("forward router fee: %w", err)
		}
	}

	current := sdk.NewCoin(in.Denom, netIn)

	for i, h := range hops {
		switch h.Kind {
		case types.HopKindNativeAMM:
			out, _, err := k.dexKeeper.Swap(sdkCtx, k.moduleAddress(), h.PoolID, current, h.DenomOut, h.MinAmountOut)
			if err != nil {
				return sdk.Coin{}, 0, false, fmt.Errorf("hop[%d] native swap: %w", i, err)
			}
			current = sdk.NewCoin(h.DenomOut, out)

		case types.HopKindIBCTransfer, types.HopKindRemoteAMM:
			// v0: park as pending and bail out. The actual IBC plumbing
			// (channel verification, ack/timeout handlers, unwind logic)
			// is non-trivial; explicit error is safer than half-implemented.
			return sdk.Coin{}, 0, false, fmt.Errorf("hop[%d]: cross-chain hops not implemented in v0 — only native AMM hops supported", i)

		default:
			return sdk.Coin{}, 0, false, fmt.Errorf("hop[%d]: unknown kind", i)
		}
	}

	// Slippage check on final output.
	if current.Amount.LT(minOut) {
		return sdk.Coin{}, 0, false, fmt.Errorf("slippage: out %s < min %s", current.Amount, minOut)
	}

	// Pay the user.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, user, sdk.NewCoins(current)); err != nil {
		return sdk.Coin{}, 0, false, fmt.Errorf("pay user: %w", err)
	}

	// Update volume tally.
	prevRaw, _ := k.VolumeTotal.Get(ctx)
	prev, _ := math.NewIntFromString(prevRaw)
	if prev.IsNil() {
		prev = math.ZeroInt()
	}
	_ = k.VolumeTotal.Set(ctx, prev.Add(in.Amount).String())

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agenticrouter.route_swap",
		sdk.NewAttribute("user", user.String()),
		sdk.NewAttribute("hops", fmt.Sprint(len(hops))),
		sdk.NewAttribute("amount_in", in.String()),
		sdk.NewAttribute("amount_out", current.String()),
		sdk.NewAttribute("router_fee", feeAmt.String()),
	))
	return current, 0, false, nil
}

// moduleAddress is the bech32 address derived from the router's module
// name. The DEX keeper sees this address as the swapper when the router
// routes through native pools.
func (k Keeper) moduleAddress() sdk.AccAddress {
	return sdk.AccAddress([]byte(types.ModuleName))
}
