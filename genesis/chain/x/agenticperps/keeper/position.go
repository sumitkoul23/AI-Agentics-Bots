package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticperps/types"
)

// OpenPosition is the keeper-level entry point for MsgOpenPosition. It
//   1. settles any unpaid funding on the trader's existing position
//   2. validates the trader's collateral is enough for the requested leverage
//   3. simulates the vAMM impact and aborts if slippage exceeds the trader's cap
//   4. updates virtual reserves
//   5. mutates the on-chain position (VWAP-merging entry price with prior fills)
func (k Keeper) OpenPosition(ctx context.Context, trader sdk.AccAddress, marketID string, notional math.LegacyDec, margin sdk.Coin, maxSlippage math.LegacyDec) (entry math.LegacyDec, newSize math.LegacyDec, newMargin math.Int, err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	market, err := k.Markets.Get(ctx, marketID)
	if err != nil {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("market %s not found", marketID)
	}
	if market.Paused {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("market %s paused", marketID)
	}
	if margin.Denom != market.MarginDenom {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("margin denom must be %s", market.MarginDenom)
	}

	// Pull collateral.
	if err := k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, trader, types.ModuleName, sdk.NewCoins(margin)); err != nil {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("escrow margin: %w", err)
	}

	// Load or initialise position.
	pos, err := k.Positions.Get(ctx, collections.Join(marketID, trader.String()))
	if err != nil {
		idxRaw, _ := k.FundingIndex.Get(ctx, marketID)
		idx, _ := math.LegacyNewDecFromStr(idxRaw)
		if idx.IsNil() {
			idx = math.LegacyZeroDec()
		}
		pos = types.Position{
			Market: marketID, Trader: trader.String(),
			Size: math.LegacyZeroDec(), Margin: math.ZeroInt(),
			EntryPrice: math.LegacyZeroDec(), LastFundingIndex: idx,
		}
	}

	// Settle funding to "now" before the position changes.
	if _, err := k.settleFundingForPosition(ctx, &pos); err != nil {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, err
	}
	pos.Margin = pos.Margin.Add(margin.Amount)

	// Leverage check at the pre-slippage mark.
	mark := market.MarkPrice()
	notionalAbs := notional.Abs()
	if !mark.IsZero() {
		leverage := notionalAbs.Quo(math.LegacyNewDecFromInt(pos.Margin))
		if leverage.GT(market.MaxLeverage) {
			return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("leverage %s exceeds max %s", leverage, market.MaxLeverage)
		}
	}

	entryPrice, baseDelta, err := market.SimulateOpen(notional)
	if err != nil {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, err
	}

	// Slippage = |entryPrice - markPrice| / markPrice
	if !mark.IsZero() && !maxSlippage.IsZero() {
		slip := entryPrice.Sub(mark).Abs().Quo(mark)
		if slip.GT(maxSlippage) {
			return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, fmt.Errorf("slippage %s exceeds cap %s", slip, maxSlippage)
		}
	}

	// Update virtual reserves consistently with SimulateOpen.
	if notional.IsPositive() {
		market.VirtualQuoteReserve = market.VirtualQuoteReserve.Add(notional)
		market.VirtualBaseReserve = market.VirtualBaseReserve.Sub(baseDelta) // baseDelta positive
	} else {
		market.VirtualQuoteReserve = market.VirtualQuoteReserve.Sub(notional.Neg())
		market.VirtualBaseReserve = market.VirtualBaseReserve.Add(baseDelta.Neg())
	}
	if err := k.Markets.Set(ctx, marketID, market); err != nil {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, err
	}

	// Merge VWAP entry price.
	if pos.Size.IsZero() {
		pos.EntryPrice = entryPrice
		pos.Size = baseDelta.Mul(signOf(notional))
	} else {
		// new VWAP = (|oldSize|*oldEntry + |delta|*newEntry) / |totalSize|
		oldSizeAbs := pos.Size.Abs()
		deltaAbs := baseDelta.Abs()
		num := oldSizeAbs.Mul(pos.EntryPrice).Add(deltaAbs.Mul(entryPrice))
		den := oldSizeAbs.Add(deltaAbs)
		pos.EntryPrice = num.Quo(den)
		pos.Size = pos.Size.Add(baseDelta.Mul(signOf(notional)))
	}

	if err := k.Positions.Set(ctx, collections.Join(marketID, trader.String()), pos); err != nil {
		return math.LegacyDec{}, math.LegacyDec{}, math.Int{}, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agenticperps.open_position",
		sdk.NewAttribute("market", marketID),
		sdk.NewAttribute("trader", trader.String()),
		sdk.NewAttribute("entry", entryPrice.String()),
		sdk.NewAttribute("size", pos.Size.String()),
		sdk.NewAttribute("margin", pos.Margin.String()),
	))
	return entryPrice, pos.Size, pos.Margin, nil
}

func signOf(d math.LegacyDec) math.LegacyDec {
	if d.IsNegative() {
		return math.LegacyNewDec(-1)
	}
	return math.LegacyOneDec()
}
