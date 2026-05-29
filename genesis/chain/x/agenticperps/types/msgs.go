package types

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgOpenPosition opens (or adds to) a position. Direction is encoded by
// the sign of Notional — positive = long, negative = short.
type MsgOpenPosition struct {
	Trader        string         `json:"trader"`
	Market        string         `json:"market"`
	Notional      math.LegacyDec `json:"notional"`    // signed; in quote (USDC) units
	Margin        sdk.Coin       `json:"margin"`      // collateral to deposit
	MaxSlippage   math.LegacyDec `json:"max_slippage"` // e.g. 0.005 = 50 bps; abort if vAMM impact exceeds this
}

func (m MsgOpenPosition) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Trader); err != nil {
		return fmt.Errorf("invalid trader: %w", err)
	}
	if m.Market == "" {
		return errors.New("market required")
	}
	if m.Notional.IsZero() {
		return errors.New("notional must be non-zero")
	}
	if !m.Margin.IsValid() || !m.Margin.IsPositive() {
		return errors.New("margin must be a valid positive coin")
	}
	if m.MaxSlippage.IsNegative() {
		return errors.New("max_slippage must be non-negative")
	}
	return nil
}

// MsgClosePosition closes the trader's position in the given market.
// Partial closes are supported via SizeToClose (zero = close-all).
type MsgClosePosition struct {
	Trader      string         `json:"trader"`
	Market      string         `json:"market"`
	SizeToClose math.LegacyDec `json:"size_to_close"` // base units; 0 = entire position
	MaxSlippage math.LegacyDec `json:"max_slippage"`
}

func (m MsgClosePosition) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Trader); err != nil {
		return fmt.Errorf("invalid trader: %w", err)
	}
	if m.Market == "" {
		return errors.New("market required")
	}
	return nil
}

// MsgLiquidate is a permissionless keeper call. Anyone may try to
// liquidate any trader whose MarginRatio < MaintenanceMargin.
type MsgLiquidate struct {
	Liquidator string `json:"liquidator"`
	Market     string `json:"market"`
	Trader     string `json:"trader"`
}

func (m MsgLiquidate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Liquidator); err != nil {
		return fmt.Errorf("invalid liquidator: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.Trader); err != nil {
		return fmt.Errorf("invalid trader: %w", err)
	}
	if m.Market == "" {
		return errors.New("market required")
	}
	return nil
}

// MsgCreateMarket is gov-only. Used to bootstrap a new perp market.
type MsgCreateMarket struct {
	Authority string `json:"authority"`
	Market    Market `json:"market"`
}

func (m MsgCreateMarket) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	if m.Market.ID == "" || m.Market.BaseDenom == "" || m.Market.MarginDenom == "" {
		return errors.New("market id/base/margin required")
	}
	if !m.Market.MaxLeverage.IsPositive() {
		return errors.New("max_leverage must be positive")
	}
	return nil
}

// Responses

type MsgOpenPositionResponse struct {
	EntryPrice  math.LegacyDec `json:"entry_price"`
	NewSize     math.LegacyDec `json:"new_size"`
	NewMargin   math.Int       `json:"new_margin"`
}
type MsgClosePositionResponse struct {
	RealisedPnL math.LegacyDec `json:"realised_pnl"`
	PayoutCoins sdk.Coins      `json:"payout_coins"`
}
type MsgLiquidateResponse struct {
	Bounty          sdk.Coin       `json:"bounty"`
	InsuranceCredit sdk.Coin       `json:"insurance_credit"`
	BadDebt         math.LegacyDec `json:"bad_debt"`
}
type MsgCreateMarketResponse struct{}
