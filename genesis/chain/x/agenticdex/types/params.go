package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Params govern protocol-wide DEX defaults. Pool creators can override
// SwapFee / ExitFee per-pool within the gov-set min/max range.
type Params struct {
	DefaultSwapFee  math.LegacyDec `json:"default_swap_fee"`  // applied to MsgSwap input
	DefaultExitFee  math.LegacyDec `json:"default_exit_fee"`  // applied to MsgExitPool

	// Split of the protocol's slice of the swap fee.
	// LP slice is (1 - ProtocolFeeShare) — paid implicitly to LPs by leaving
	// the fee inside the pool's reserves.
	ProtocolFeeShare    math.LegacyDec `json:"protocol_fee_share"`     // fraction of swap fee diverted away from LPs
	ProtocolFeeBurnShare math.LegacyDec `json:"protocol_fee_burn_share"` // fraction of *protocol* slice that burns

	// Pool creation guard.
	MinInitialDepositUsd math.Int `json:"min_initial_deposit_usd"` // in ucents (1/100 of a USD-cent)
}

// DefaultParams matches the values described in `genesis/docs/05-exchange-strategy.md`.
//   Total swap fee 0.30 %, split:
//     LP slice         0.25 %   (implicit, stays in pool)
//     Protocol slice   0.05 %   = ProtocolFeeShare 16.67% of 0.30%
//   Protocol slice further split 60/40 treasury/burn.
func DefaultParams() Params {
	return Params{
		DefaultSwapFee:       math.LegacyNewDecWithPrec(3, 3),    // 0.30 %
		DefaultExitFee:       math.LegacyNewDecWithPrec(0, 0),    // 0.00 %
		ProtocolFeeShare:     math.LegacyNewDecWithPrec(1667, 5), // 16.67 % of swap fee (≈ 0.05% of trade)
		ProtocolFeeBurnShare: math.LegacyNewDecWithPrec(4, 1),    // 40 % of protocol slice burns
		MinInitialDepositUsd: math.NewInt(1_000_00),              // $1,000 in ucents
	}
}

func (p Params) Validate() error {
	if p.DefaultSwapFee.IsNegative() || p.DefaultSwapFee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("default_swap_fee must be in [0, 1), got %s", p.DefaultSwapFee)
	}
	if p.DefaultExitFee.IsNegative() || p.DefaultExitFee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("default_exit_fee must be in [0, 1)")
	}
	if p.ProtocolFeeShare.IsNegative() || p.ProtocolFeeShare.GT(math.LegacyOneDec()) {
		return fmt.Errorf("protocol_fee_share must be in [0, 1]")
	}
	if p.ProtocolFeeBurnShare.IsNegative() || p.ProtocolFeeBurnShare.GT(math.LegacyOneDec()) {
		return fmt.Errorf("protocol_fee_burn_share must be in [0, 1]")
	}
	if !p.MinInitialDepositUsd.IsPositive() {
		return fmt.Errorf("min_initial_deposit_usd must be positive")
	}
	return nil
}
