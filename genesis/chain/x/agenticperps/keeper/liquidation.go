package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticperps/types"
)

// Liquidate is the permissionless keeper-bounty entry point. Returns the
// liquidator's bounty, the insurance fund credit, and any bad-debt
// (margin shortfall that the insurance fund had to cover).
//
// Eligibility:
//
//   pos.MarginRatio(mark) < market.MaintenanceMargin
//
// Settlement:
//
//   notional         = |pos.size| * mark
//   liquidation_fee  = notional * params.LiquidationFee
//   bounty           = liquidation_fee * params.LiquidatorBounty
//   insurance_credit = liquidation_fee - bounty
//   remaining_margin = margin + unrealisedPnL - liquidation_fee
//
// If remaining_margin >= 0, the trader receives it back.
// If remaining_margin < 0, that's bad debt and the insurance fund absorbs it.
func (k Keeper) Liquidate(ctx context.Context, liquidator sdk.AccAddress, marketID, traderAddr string) (sdk.Coin, sdk.Coin, math.LegacyDec, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	market, err := k.Markets.Get(ctx, marketID)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, fmt.Errorf("market %s not found", marketID)
	}
	posKey := collections.Join(marketID, traderAddr)
	pos, err := k.Positions.Get(ctx, posKey)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, fmt.Errorf("no position for %s in %s", traderAddr, marketID)
	}

	// Settle funding to "now".
	if _, err := k.settleFundingForPosition(ctx, &pos); err != nil {
		return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, err
	}

	mark := market.MarkPrice()
	ratio := pos.MarginRatio(mark)
	if ratio.GTE(market.MaintenanceMargin) {
		return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, fmt.Errorf("position healthy (ratio %s >= maintenance %s)", ratio, market.MaintenanceMargin)
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, err
	}

	notional := pos.Notional(mark)
	liqFee := notional.Mul(params.LiquidationFee)
	bountyAmt := liqFee.Mul(params.LiquidatorBounty).TruncateInt()
	insAmt := liqFee.TruncateInt().Sub(bountyAmt)

	// Effective margin after PnL + fee.
	effMargin := math.LegacyNewDecFromInt(pos.Margin).Add(pos.UnrealisedPnL(mark)).Sub(liqFee)

	var badDebt math.LegacyDec
	if effMargin.IsNegative() {
		badDebt = effMargin.Neg()
		// Drain insurance fund.
		fundRaw, _ := k.InsuranceFund.Get(ctx)
		fund, _ := math.NewIntFromString(fundRaw)
		if fund.IsNil() {
			fund = math.ZeroInt()
		}
		if fund.LT(badDebt.TruncateInt()) {
			// Insurance underfunded — emit a critical event but settle anyway.
			sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
				"agenticperps.insurance_underfunded",
				sdk.NewAttribute("market", marketID),
				sdk.NewAttribute("shortfall", badDebt.String()),
			))
		}
		newFund := fund.Sub(badDebt.TruncateInt())
		if newFund.IsNegative() {
			newFund = math.ZeroInt()
		}
		_ = k.InsuranceFund.Set(ctx, newFund.String())
		effMargin = math.LegacyZeroDec()
	} else {
		// Refund the surplus to the trader.
		traderAddrBech, _ := sdk.AccAddressFromBech32(traderAddr)
		refund := sdk.NewCoin(market.MarginDenom, effMargin.TruncateInt())
		if refund.Amount.IsPositive() {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, traderAddrBech, sdk.NewCoins(refund)); err != nil {
				return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, fmt.Errorf("refund trader: %w", err)
			}
		}
		badDebt = math.LegacyZeroDec()
	}

	// Pay the liquidator their bounty + credit the insurance fund.
	bountyCoin := sdk.NewCoin(market.MarginDenom, bountyAmt)
	if bountyCoin.Amount.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, liquidator, sdk.NewCoins(bountyCoin)); err != nil {
			return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, fmt.Errorf("pay bounty: %w", err)
		}
	}
	if insAmt.IsPositive() {
		fundRaw, _ := k.InsuranceFund.Get(ctx)
		fund, _ := math.NewIntFromString(fundRaw)
		if fund.IsNil() {
			fund = math.ZeroInt()
		}
		_ = k.InsuranceFund.Set(ctx, fund.Add(insAmt).String())
	}

	// Close the vAMM exposure (mirror of OpenPosition logic with opposite sign).
	// In v0 we settle bookkeeping only; in v1 the vAMM unwind ships separately.
	if err := k.Positions.Remove(ctx, posKey); err != nil {
		return sdk.Coin{}, sdk.Coin{}, math.LegacyDec{}, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agenticperps.liquidate",
		sdk.NewAttribute("market", marketID),
		sdk.NewAttribute("trader", traderAddr),
		sdk.NewAttribute("liquidator", liquidator.String()),
		sdk.NewAttribute("bounty", bountyCoin.String()),
		sdk.NewAttribute("insurance_credit", sdk.NewCoin(market.MarginDenom, insAmt).String()),
		sdk.NewAttribute("bad_debt", badDebt.String()),
	))
	return bountyCoin, sdk.NewCoin(market.MarginDenom, insAmt), badDebt, nil
}
