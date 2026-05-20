package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/app"
	"github.com/sumitkoul23/agentic-chain/x/agenticdex/types"
)

// Swap executes a single-hop swap on the given pool, respecting
// `minAmountOut` slippage protection.
//
// Fee routing:
//   1. The full swap fee is taken from the *input* amount.
//   2. The LP slice of the fee stays inside the pool reserves (raising the
//      constant-product k and accruing to all LPs proportionally).
//   3. The protocol slice (Params.ProtocolFeeShare) is moved out of the
//      pool's reserves into the module account.
//   4. The protocol slice is then further split: `ProtocolFeeBurnShare`
//      is burned, the rest is forwarded to the gov treasury via
//      `auth.fee_collector`.
//
// This is the single hook through which DEX activity feeds the chain's
// deflationary curve.
func (k Keeper) Swap(
	ctx context.Context,
	swapper sdk.AccAddress,
	poolID uint64,
	amountIn sdk.Coin,
	denomOut string,
	minAmountOut math.Int,
) (out math.Int, feePaid math.Int, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := k.Pools.Get(ctx, poolID)
	if err != nil {
		return math.Int{}, math.Int{}, fmt.Errorf("pool %d not found", poolID)
	}
	params, err := k.Params.Get(ctx)
	if err != nil {
		return math.Int{}, math.Int{}, err
	}

	// Validate denom membership.
	if amountIn.Denom != pool.AssetA.Denom && amountIn.Denom != pool.AssetB.Denom {
		return math.Int{}, math.Int{}, fmt.Errorf("denom %s not in pool %d", amountIn.Denom, poolID)
	}

	amountOut, swapFee, err := pool.CalcOutGivenIn(amountIn.Denom, denomOut, amountIn.Amount)
	if err != nil {
		return math.Int{}, math.Int{}, err
	}
	if amountOut.LT(minAmountOut) {
		return math.Int{}, math.Int{}, fmt.Errorf("slippage: out %s < min %s", amountOut, minAmountOut)
	}

	// Pull input from the swapper.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, swapper, types.ModuleName, sdk.NewCoins(amountIn)); err != nil {
		return math.Int{}, math.Int{}, fmt.Errorf("escrow input: %w", err)
	}

	// Update reserves: input added (full, including LP fee), output removed.
	if amountIn.Denom == pool.AssetA.Denom {
		pool.AssetA.Amount = pool.AssetA.Amount.Add(amountIn.Amount)
		pool.AssetB.Amount = pool.AssetB.Amount.Sub(amountOut)
	} else {
		pool.AssetB.Amount = pool.AssetB.Amount.Add(amountIn.Amount)
		pool.AssetA.Amount = pool.AssetA.Amount.Sub(amountOut)
	}

	// Carve the protocol slice out of the pool's input-side reserve and
	// route it. This is the only place where pool reserves can be reduced
	// outside of MsgExitPool / MsgSwap-out.
	protocolFee := math.LegacyNewDecFromInt(swapFee).Mul(params.ProtocolFeeShare).TruncateInt()
	if protocolFee.IsPositive() {
		if amountIn.Denom == pool.AssetA.Denom {
			pool.AssetA.Amount = pool.AssetA.Amount.Sub(protocolFee)
		} else {
			pool.AssetB.Amount = pool.AssetB.Amount.Sub(protocolFee)
		}

		burnAmt := math.LegacyNewDecFromInt(protocolFee).Mul(params.ProtocolFeeBurnShare).TruncateInt()
		treasuryAmt := protocolFee.Sub(burnAmt)

		feeCoin := sdk.NewCoin(amountIn.Denom, protocolFee)
		// Burn — but only if the fee denom matches the chain's burn-eligible
		// token. We can only burn coins we have authority over; for foreign
		// denoms we route the full slice to the treasury instead.
		if amountIn.Denom == app.BaseCoinUnit && burnAmt.IsPositive() {
			if err := k.bankKeeper.BurnCoins(sdkCtx, types.ModuleName,
				sdk.NewCoins(sdk.NewCoin(app.BaseCoinUnit, burnAmt))); err != nil {
				return math.Int{}, math.Int{}, fmt.Errorf("burn protocol fee: %w", err)
			}
		} else {
			treasuryAmt = protocolFee // foreign denom — full slice to treasury
			_ = burnAmt
		}

		if treasuryAmt.IsPositive() {
			if err := k.bankKeeper.SendCoinsFromModuleToModule(
				sdkCtx, types.ModuleName, "fee_collector",
				sdk.NewCoins(sdk.NewCoin(amountIn.Denom, treasuryAmt)),
			); err != nil {
				return math.Int{}, math.Int{}, fmt.Errorf("send treasury fee: %w", err)
			}
		}
		_ = feeCoin
	}

	// Persist the updated pool and pay out the swapper.
	if err := k.Pools.Set(ctx, poolID, pool); err != nil {
		return math.Int{}, math.Int{}, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		sdkCtx, types.ModuleName, swapper,
		sdk.NewCoins(sdk.NewCoin(denomOut, amountOut)),
	); err != nil {
		return math.Int{}, math.Int{}, fmt.Errorf("pay swapper: %w", err)
	}

	// Telemetry — keep a running tally of total fees collected.
	prev, _ := k.TotalSwapFee.Get(ctx)
	prevInt, _ := math.NewIntFromString(prev)
	_ = k.TotalSwapFee.Set(ctx, prevInt.Add(swapFee).String())

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agenticdex.swap",
		sdk.NewAttribute("pool_id", fmt.Sprint(poolID)),
		sdk.NewAttribute("swapper", swapper.String()),
		sdk.NewAttribute("amount_in", amountIn.String()),
		sdk.NewAttribute("amount_out", sdk.NewCoin(denomOut, amountOut).String()),
		sdk.NewAttribute("fee", swapFee.String()),
		sdk.NewAttribute("protocol_fee", protocolFee.String()),
	))
	return amountOut, swapFee, nil
}
