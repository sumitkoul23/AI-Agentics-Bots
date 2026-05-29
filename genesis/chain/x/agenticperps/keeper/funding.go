package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticperps/types"
)

// FundingRate returns the per-block funding rate for `market`. The standard
// dYdX / Perpetual Protocol formula:
//
//   premium = (markPrice - indexPrice) / indexPrice
//   rate    = clamp(premium / fundingPeriodBlocks, ±FundingCap)
//
// When longs pay shorts, the index trails the mark — i.e. the perp price
// has run above spot, and shorts are subsidising the convergence by
// receiving funding.
func (k Keeper) FundingRate(ctx context.Context, market types.Market) (math.LegacyDec, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	indexRaw, ok := k.priceKeeper.IndexPrice(sdkCtx, market.ID)
	if !ok {
		return math.LegacyZeroDec(), nil
	}
	index, err := math.LegacyNewDecFromStr(indexRaw)
	if err != nil || !index.IsPositive() {
		return math.LegacyZeroDec(), nil
	}
	mark := market.MarkPrice()
	premium := mark.Sub(index).Quo(index)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return math.LegacyZeroDec(), err
	}
	periodRate := premium.QuoInt64(int64(params.BlocksPerFundingPeriod))

	// Clamp to ±FundingCap.
	perBlockCap := params.FundingCap.QuoInt64(int64(params.BlocksPerFundingPeriod))
	if periodRate.GT(perBlockCap) {
		periodRate = perBlockCap
	}
	if periodRate.LT(perBlockCap.Neg()) {
		periodRate = perBlockCap.Neg()
	}
	return periodRate, nil
}

// AccrueFundingForMarket bumps a market's cumulative funding index by the
// current per-block rate. Called once per block by `EndBlocker` for each
// market. Trader funding payments are settled lazily on next interaction
// by comparing their stored `LastFundingIndex` with the current index.
//
//   payment_owed_by_longs = position.size * (currentIndex - position.lastIndex)
//   (longs pay when positive, receive when negative)
func (k Keeper) AccrueFundingForMarket(ctx context.Context, market types.Market) error {
	rate, err := k.FundingRate(ctx, market)
	if err != nil {
		return err
	}
	prevRaw, _ := k.FundingIndex.Get(ctx, market.ID)
	prev, _ := math.LegacyNewDecFromStr(prevRaw)
	if prev.IsNil() {
		prev = math.LegacyZeroDec()
	}
	return k.FundingIndex.Set(ctx, market.ID, prev.Add(rate).String())
}

// settleFundingForPosition computes the trader's net funding payment
// since their last interaction and adjusts their margin in place.
// Returns the signed payment in MarginDenom units.
//
// Convention: positive return = trader owed margin (received funding).
func (k Keeper) settleFundingForPosition(ctx context.Context, pos *types.Position) (math.LegacyDec, error) {
	idxRaw, _ := k.FundingIndex.Get(ctx, pos.Market)
	currentIdx, _ := math.LegacyNewDecFromStr(idxRaw)
	if currentIdx.IsNil() {
		currentIdx = math.LegacyZeroDec()
	}
	delta := currentIdx.Sub(pos.LastFundingIndex)

	// payment to longs (positive size) is negative when delta positive,
	// since they pay when funding is positive.
	payment := pos.Size.Mul(delta).Neg()

	// Apply to margin. Capped at margin to avoid negative balances —
	// shortfall becomes the trader's bad debt at liquidation time.
	if payment.IsPositive() {
		pos.Margin = pos.Margin.Add(payment.TruncateInt())
	} else {
		decr := payment.Neg().TruncateInt()
		if decr.GT(pos.Margin) {
			decr = pos.Margin
		}
		pos.Margin = pos.Margin.Sub(decr)
	}
	pos.LastFundingIndex = currentIdx
	return payment, nil
}
