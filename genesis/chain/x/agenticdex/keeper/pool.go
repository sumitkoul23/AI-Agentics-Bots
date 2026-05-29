package keeper

import (
	"context"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticdex/types"
)

// CreatePool mints the genesis LP shares for a new constant-product pool
// and escrows the initial reserves in the module account.
//
// Genesis share supply is geometric-mean of the two reserves — the same
// convention Uniswap v2 and Osmosis use. This guarantees that the LP-token
// price scales smoothly with deposits and the very first deposit cannot
// be Sybil-amplified.
//
//   shares_minted = sqrt(amountA * amountB)
func (k Keeper) CreatePool(ctx context.Context, creator sdk.AccAddress, a, b sdk.Coin, swapFee, exitFee math.LegacyDec) (types.Pool, math.Int, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return types.Pool{}, math.Int{}, err
	}
	if swapFee.IsNil() || swapFee.IsZero() {
		swapFee = params.DefaultSwapFee
	}
	if exitFee.IsNil() || exitFee.IsZero() {
		exitFee = params.DefaultExitFee
	}

	// Escrow the initial reserves.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, creator, types.ModuleName, sdk.NewCoins(a, b)); err != nil {
		return types.Pool{}, math.Int{}, fmt.Errorf("escrow initial reserves: %w", err)
	}

	// Geometric-mean LP share minting.
	shares := geomMean(a.Amount, b.Amount)
	if !shares.IsPositive() {
		return types.Pool{}, math.Int{}, fmt.Errorf("initial deposit too small — shares round to zero")
	}

	// Mint the LP shares to the creator under denom `pool/<id>`.
	idSeq, err := k.PoolCounter.Next(ctx)
	if err != nil {
		return types.Pool{}, math.Int{}, err
	}
	id := idSeq + 1 // pool IDs are 1-indexed
	pool := types.Pool{
		ID:          id,
		AssetA:      a,
		AssetB:      b,
		TotalShares: shares,
		SwapFee:     swapFee,
		ExitFee:     exitFee,
	}
	shareCoins := sdk.NewCoins(sdk.NewCoin(pool.ShareDenom(), shares))
	if err := k.bankKeeper.MintCoins(sdkCtx, types.ModuleName, shareCoins); err != nil {
		return types.Pool{}, math.Int{}, fmt.Errorf("mint shares: %w", err)
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, creator, shareCoins); err != nil {
		return types.Pool{}, math.Int{}, fmt.Errorf("send shares: %w", err)
	}

	if err := k.Pools.Set(ctx, id, pool); err != nil {
		return types.Pool{}, math.Int{}, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agenticdex.create_pool",
		sdk.NewAttribute("pool_id", fmt.Sprint(id)),
		sdk.NewAttribute("creator", creator.String()),
		sdk.NewAttribute("asset_a", a.String()),
		sdk.NewAttribute("asset_b", b.String()),
		sdk.NewAttribute("shares_out", shares.String()),
	))
	return pool, shares, nil
}

// JoinPool deposits proportional amounts of both reserves and mints LP shares.
//
// To preserve the constant-product invariant on join, the caller must deposit
// in the *exact* current ratio of reserves. We compute the deposit from the
// requested `shareOut` amount:
//
//   amountA = shareOut * reserveA / totalShares
//   amountB = shareOut * reserveB / totalShares
//
// If either computed amount exceeds the caller's `max_amounts_in` budget,
// the tx aborts.
func (k Keeper) JoinPool(ctx context.Context, joiner sdk.AccAddress, poolID uint64, shareOut math.Int, maxIn sdk.Coins) (math.Int, sdk.Coins, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pool, err := k.Pools.Get(ctx, poolID)
	if err != nil {
		return math.Int{}, nil, fmt.Errorf("pool %d not found", poolID)
	}

	amtA := pool.AssetA.Amount.Mul(shareOut).Quo(pool.TotalShares)
	amtB := pool.AssetB.Amount.Mul(shareOut).Quo(pool.TotalShares)
	if !amtA.IsPositive() || !amtB.IsPositive() {
		return math.Int{}, nil, fmt.Errorf("computed deposit rounds to zero — request more shares")
	}
	depositA := sdk.NewCoin(pool.AssetA.Denom, amtA)
	depositB := sdk.NewCoin(pool.AssetB.Denom, amtB)

	if got := maxIn.AmountOf(pool.AssetA.Denom); got.LT(amtA) {
		return math.Int{}, nil, fmt.Errorf("max_amounts_in for %s too low: need %s, have %s", pool.AssetA.Denom, amtA, got)
	}
	if got := maxIn.AmountOf(pool.AssetB.Denom); got.LT(amtB) {
		return math.Int{}, nil, fmt.Errorf("max_amounts_in for %s too low: need %s, have %s", pool.AssetB.Denom, amtB, got)
	}

	deposit := sdk.NewCoins(depositA, depositB)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, joiner, types.ModuleName, deposit); err != nil {
		return math.Int{}, nil, fmt.Errorf("escrow deposit: %w", err)
	}

	// Mint shares.
	shareCoins := sdk.NewCoins(sdk.NewCoin(pool.ShareDenom(), shareOut))
	if err := k.bankKeeper.MintCoins(sdkCtx, types.ModuleName, shareCoins); err != nil {
		return math.Int{}, nil, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, joiner, shareCoins); err != nil {
		return math.Int{}, nil, err
	}

	pool.AssetA.Amount = pool.AssetA.Amount.Add(amtA)
	pool.AssetB.Amount = pool.AssetB.Amount.Add(amtB)
	pool.TotalShares = pool.TotalShares.Add(shareOut)
	if err := k.Pools.Set(ctx, poolID, pool); err != nil {
		return math.Int{}, nil, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agenticdex.join_pool",
		sdk.NewAttribute("pool_id", fmt.Sprint(poolID)),
		sdk.NewAttribute("joiner", joiner.String()),
		sdk.NewAttribute("shares_out", shareOut.String()),
		sdk.NewAttribute("deposited", deposit.String()),
	))
	return shareOut, deposit, nil
}

// ExitPool burns LP shares and pays out proportional reserves, less the
// pool's exit fee (which stays inside the pool, accruing to remaining LPs).
//
//   amountOutGross = shareIn * reserve / totalShares
//   amountOutNet   = amountOutGross * (1 - exitFee)
func (k Keeper) ExitPool(ctx context.Context, exiter sdk.AccAddress, poolID uint64, shareIn math.Int, minOut sdk.Coins) (sdk.Coins, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pool, err := k.Pools.Get(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("pool %d not found", poolID)
	}
	if shareIn.GT(pool.TotalShares) {
		return nil, fmt.Errorf("share_in exceeds total supply")
	}

	feeFactor := math.LegacyOneDec().Sub(pool.ExitFee)
	outA := pool.AssetA.Amount.Mul(shareIn).Quo(pool.TotalShares)
	outB := pool.AssetB.Amount.Mul(shareIn).Quo(pool.TotalShares)
	netA := math.LegacyNewDecFromInt(outA).Mul(feeFactor).TruncateInt()
	netB := math.LegacyNewDecFromInt(outB).Mul(feeFactor).TruncateInt()
	if !netA.IsPositive() && !netB.IsPositive() {
		return nil, fmt.Errorf("exit rounds to zero — burn more shares")
	}
	netOut := sdk.NewCoins(
		sdk.NewCoin(pool.AssetA.Denom, netA),
		sdk.NewCoin(pool.AssetB.Denom, netB),
	)

	if got := minOut.AmountOf(pool.AssetA.Denom); netA.LT(got) {
		return nil, fmt.Errorf("min_amounts_out for %s not met", pool.AssetA.Denom)
	}
	if got := minOut.AmountOf(pool.AssetB.Denom); netB.LT(got) {
		return nil, fmt.Errorf("min_amounts_out for %s not met", pool.AssetB.Denom)
	}

	// Burn the LP shares.
	shareCoins := sdk.NewCoins(sdk.NewCoin(pool.ShareDenom(), shareIn))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, exiter, types.ModuleName, shareCoins); err != nil {
		return nil, fmt.Errorf("collect shares: %w", err)
	}
	if err := k.bankKeeper.BurnCoins(sdkCtx, types.ModuleName, shareCoins); err != nil {
		return nil, fmt.Errorf("burn shares: %w", err)
	}

	// Pay out the reserves.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, exiter, netOut); err != nil {
		return nil, fmt.Errorf("pay reserves: %w", err)
	}

	// Update pool — exit fee remains in the pool, raising the share price
	// for everyone still in.
	pool.AssetA.Amount = pool.AssetA.Amount.Sub(netA)
	pool.AssetB.Amount = pool.AssetB.Amount.Sub(netB)
	pool.TotalShares = pool.TotalShares.Sub(shareIn)
	if err := k.Pools.Set(ctx, poolID, pool); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agenticdex.exit_pool",
		sdk.NewAttribute("pool_id", fmt.Sprint(poolID)),
		sdk.NewAttribute("exiter", exiter.String()),
		sdk.NewAttribute("shares_in", shareIn.String()),
		sdk.NewAttribute("paid_out", netOut.String()),
	))
	return netOut, nil
}

// ───────────────────────── helpers ─────────────────────────

// geomMean returns ⌊sqrt(a * b)⌋ using big.Int to avoid math.Int overflow
// for large initial deposits. Standard Uniswap v2 / Osmosis convention.
func geomMean(a, b math.Int) math.Int {
	prod := new(big.Int).Mul(a.BigInt(), b.BigInt())
	sqrt := new(big.Int).Sqrt(prod)
	return math.NewIntFromBigInt(sqrt)
}
