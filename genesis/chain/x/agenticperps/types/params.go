package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Params govern the module-wide perp economics. Per-market knobs live on
// the Market struct so different markets can run different leverage caps.
type Params struct {
	// Funding rate is recomputed every BlocksPerFundingPeriod blocks.
	// Default = 8h equivalent at 3s blocks = 9600 blocks.
	BlocksPerFundingPeriod uint64 `json:"blocks_per_funding_period"`

	// FundingCap is the maximum |funding rate| per period (e.g. 0.0075
	// → ±0.75 % per 8h). Tracks Binance / dYdX defaults.
	FundingCap math.LegacyDec `json:"funding_cap"`

	// Liquidation params.
	LiquidationFee     math.LegacyDec `json:"liquidation_fee"`     // % of notional paid by trader on liquidation
	LiquidatorBounty   math.LegacyDec `json:"liquidator_bounty"`   // fraction of fee paid to liquidator
	InsuranceFundShare math.LegacyDec `json:"insurance_fund_share"` // fraction of fee that funds insurance

	// Open-fee taken on every MsgOpenPosition / MsgClosePosition.
	OpenFee math.LegacyDec `json:"open_fee"`

	// Whitelist of denoms accepted as margin (typically just "uusdc" at v0).
	AllowedMarginDenoms []string `json:"allowed_margin_denoms"`
}

func DefaultParams() Params {
	return Params{
		BlocksPerFundingPeriod: 9600,
		FundingCap:             math.LegacyNewDecWithPrec(75, 4),  // 0.75 %
		LiquidationFee:         math.LegacyNewDecWithPrec(25, 4),  // 0.25 %
		LiquidatorBounty:       math.LegacyNewDecWithPrec(5, 1),   // 50 %
		InsuranceFundShare:     math.LegacyNewDecWithPrec(5, 1),   // 50 %
		OpenFee:                math.LegacyNewDecWithPrec(5, 4),   // 0.05 %
		AllowedMarginDenoms:    []string{"uusdc"},
	}
}

func (p Params) Validate() error {
	if p.BlocksPerFundingPeriod == 0 {
		return fmt.Errorf("blocks_per_funding_period must be positive")
	}
	if p.FundingCap.IsNegative() || p.FundingCap.GT(math.LegacyOneDec()) {
		return fmt.Errorf("funding_cap out of range")
	}
	if !p.LiquidatorBounty.Add(p.InsuranceFundShare).Equal(math.LegacyOneDec()) {
		return fmt.Errorf("liquidator_bounty + insurance_fund_share must equal 1.0")
	}
	if p.OpenFee.IsNegative() || p.OpenFee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("open_fee out of range")
	}
	if len(p.AllowedMarginDenoms) == 0 {
		return fmt.Errorf("at least one margin denom must be allowed")
	}
	return nil
}
