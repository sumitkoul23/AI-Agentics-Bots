// Hand-rolled Msg types for `x/agenticdex`. Same v0 rationale as
// `x/agentic/types/msgs.go` — these document the wire shape and let the
// package compile before `buf generate` is wired up.
package types

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ───────────────────────── MsgCreatePool ─────────────────────────

type MsgCreatePool struct {
	Creator       string   `json:"creator"`
	InitialAssetA sdk.Coin `json:"initial_asset_a"`
	InitialAssetB sdk.Coin `json:"initial_asset_b"`
	SwapFee       math.LegacyDec `json:"swap_fee"` // 0 → use Params.DefaultSwapFee
	ExitFee       math.LegacyDec `json:"exit_fee"` // 0 → use Params.DefaultExitFee
}

func (m MsgCreatePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return fmt.Errorf("invalid creator: %w", err)
	}
	if !m.InitialAssetA.IsValid() || !m.InitialAssetA.IsPositive() {
		return errors.New("initial_asset_a must be a valid positive coin")
	}
	if !m.InitialAssetB.IsValid() || !m.InitialAssetB.IsPositive() {
		return errors.New("initial_asset_b must be a valid positive coin")
	}
	if m.InitialAssetA.Denom == m.InitialAssetB.Denom {
		return errors.New("initial assets must be distinct denoms")
	}
	return nil
}

// ───────────────────────── MsgJoinPool ─────────────────────────

type MsgJoinPool struct {
	Joiner        string   `json:"joiner"`
	PoolID        uint64   `json:"pool_id"`
	ShareOutMin   math.Int `json:"share_out_min"`   // slippage guard
	MaxAmountsIn  sdk.Coins `json:"max_amounts_in"` // caller's ceiling per asset
}

func (m MsgJoinPool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Joiner); err != nil {
		return fmt.Errorf("invalid joiner: %w", err)
	}
	if m.PoolID == 0 {
		return errors.New("pool_id required")
	}
	if !m.ShareOutMin.IsPositive() {
		return errors.New("share_out_min must be positive")
	}
	if len(m.MaxAmountsIn) != 2 {
		return errors.New("max_amounts_in must specify both pool assets")
	}
	return nil
}

// ───────────────────────── MsgExitPool ─────────────────────────

type MsgExitPool struct {
	Exiter        string    `json:"exiter"`
	PoolID        uint64    `json:"pool_id"`
	ShareInAmount math.Int  `json:"share_in_amount"`
	MinAmountsOut sdk.Coins `json:"min_amounts_out"` // slippage guard per asset
}

func (m MsgExitPool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Exiter); err != nil {
		return fmt.Errorf("invalid exiter: %w", err)
	}
	if m.PoolID == 0 {
		return errors.New("pool_id required")
	}
	if !m.ShareInAmount.IsPositive() {
		return errors.New("share_in_amount must be positive")
	}
	return nil
}

// ───────────────────────── MsgSwap ─────────────────────────

type MsgSwap struct {
	Swapper       string   `json:"swapper"`
	PoolID        uint64   `json:"pool_id"`
	AmountIn      sdk.Coin `json:"amount_in"`
	DenomOut      string   `json:"denom_out"`
	MinAmountOut  math.Int `json:"min_amount_out"`  // slippage guard
}

func (m MsgSwap) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Swapper); err != nil {
		return fmt.Errorf("invalid swapper: %w", err)
	}
	if m.PoolID == 0 {
		return errors.New("pool_id required")
	}
	if !m.AmountIn.IsValid() || !m.AmountIn.IsPositive() {
		return errors.New("amount_in must be a valid positive coin")
	}
	if m.DenomOut == "" || m.DenomOut == m.AmountIn.Denom {
		return errors.New("denom_out must be set and distinct from amount_in.denom")
	}
	if !m.MinAmountOut.IsPositive() {
		return errors.New("min_amount_out must be positive (set to 1 for no slippage cap)")
	}
	return nil
}

// ───────────────────────── MsgUpdateParams (gov) ─────────────────────────

type MsgUpdateParams struct {
	Authority string `json:"authority"`
	Params    Params `json:"params"`
}

func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return fmt.Errorf("invalid authority: %w", err)
	}
	return m.Params.Validate()
}

// Responses

type MsgCreatePoolResponse struct {
	PoolID      uint64   `json:"pool_id"`
	ShareDenom  string   `json:"share_denom"`
	SharesOut   math.Int `json:"shares_out"`
}
type MsgJoinPoolResponse struct {
	SharesOut math.Int  `json:"shares_out"`
	UsedIn    sdk.Coins `json:"used_in"`
}
type MsgExitPoolResponse struct {
	AmountsOut sdk.Coins `json:"amounts_out"`
}
type MsgSwapResponse struct {
	AmountOut math.Int `json:"amount_out"`
	FeePaid   math.Int `json:"fee_paid"`
}
type MsgUpdateParamsResponse struct{}
