package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Pool is the on-chain representation of a constant-product AMM pool.
//
// Reserves are kept canonically as `sdk.Coin` to preserve denom safety;
// `TotalShares` is stored as a stringified `math.Int` to avoid the proto
// codec's signed-bigint quirks.
type Pool struct {
	ID          uint64    `json:"id"`
	AssetA      sdk.Coin  `json:"asset_a"`       // reserve of asset A
	AssetB      sdk.Coin  `json:"asset_b"`       // reserve of asset B
	TotalShares math.Int  `json:"total_shares"`  // outstanding LP tokens
	SwapFee     math.LegacyDec `json:"swap_fee"` // applied to the input on every swap
	ExitFee     math.LegacyDec `json:"exit_fee"` // applied on MsgExitPool
}

// ShareDenom returns the bank denom under which this pool's LP shares
// circulate. Wallets and IBC use this denom directly.
func (p Pool) ShareDenom() string {
	return fmt.Sprintf("%s%d", PoolShareDenomPrefix, p.ID)
}

// SortedPair returns the pool's two reserves with the lower-lexicographic
// denom first. Callers that need stable ordering (e.g. price quotes) should
// use this rather than reading AssetA / AssetB directly.
func (p Pool) SortedPair() (sdk.Coin, sdk.Coin) {
	if p.AssetA.Denom < p.AssetB.Denom {
		return p.AssetA, p.AssetB
	}
	return p.AssetB, p.AssetA
}

// SpotPrice returns the marginal price of `denomOut` denominated in
// `denomIn`, i.e. how many `denomIn` units are needed to buy one
// `denomOut` at the current reserves (ignoring fees and slippage).
//
//   price = reserveIn / reserveOut
func (p Pool) SpotPrice(denomIn, denomOut string) (math.LegacyDec, error) {
	in, out, err := p.reservesForPair(denomIn, denomOut)
	if err != nil {
		return math.LegacyDec{}, err
	}
	if out.IsZero() {
		return math.LegacyDec{}, fmt.Errorf("empty out reserve")
	}
	return math.LegacyNewDecFromInt(in.Amount).Quo(math.LegacyNewDecFromInt(out.Amount)), nil
}

// reservesForPair returns (in, out) such that `in.Denom == denomIn` and
// `out.Denom == denomOut`. Errors if the denoms don't match the pool.
func (p Pool) reservesForPair(denomIn, denomOut string) (sdk.Coin, sdk.Coin, error) {
	switch {
	case denomIn == p.AssetA.Denom && denomOut == p.AssetB.Denom:
		return p.AssetA, p.AssetB, nil
	case denomIn == p.AssetB.Denom && denomOut == p.AssetA.Denom:
		return p.AssetB, p.AssetA, nil
	default:
		return sdk.Coin{}, sdk.Coin{}, fmt.Errorf("denoms (%s, %s) not in pool %d", denomIn, denomOut, p.ID)
	}
}

// CalcOutGivenIn applies the constant-product invariant after the swap
// fee, returning the amount of `denomOut` the swapper receives.
//
//   amountInAfterFee = amountIn * (1 - swapFee)
//   amountOut        = reserveOut * amountInAfterFee / (reserveIn + amountInAfterFee)
//
// Returns the amount-out and the fee (denominated in `denomIn`) so callers
// can route the protocol slice elsewhere.
func (p Pool) CalcOutGivenIn(denomIn, denomOut string, amountIn math.Int) (out math.Int, fee math.Int, err error) {
	in, outRes, err := p.reservesForPair(denomIn, denomOut)
	if err != nil {
		return math.Int{}, math.Int{}, err
	}
	if !amountIn.IsPositive() {
		return math.Int{}, math.Int{}, fmt.Errorf("amount_in must be positive")
	}

	feeAmt := math.LegacyNewDecFromInt(amountIn).Mul(p.SwapFee).TruncateInt()
	amountInNet := amountIn.Sub(feeAmt)

	numerator := outRes.Amount.Mul(amountInNet)
	denominator := in.Amount.Add(amountInNet)
	if denominator.IsZero() {
		return math.Int{}, math.Int{}, fmt.Errorf("zero denominator")
	}
	amountOut := numerator.Quo(denominator)
	if !amountOut.IsPositive() {
		return math.Int{}, math.Int{}, fmt.Errorf("output rounds to zero — increase amount_in")
	}
	return amountOut, feeAmt, nil
}
