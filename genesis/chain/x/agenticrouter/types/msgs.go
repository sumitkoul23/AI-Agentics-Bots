package types

import (
	"errors"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgRouteSwap is the single user-facing message. The off-chain quoter
// produces the Hops array; the keeper executes them atomically.
type MsgRouteSwap struct {
	User         string    `json:"user"`
	AmountIn     sdk.Coin  `json:"amount_in"`
	DenomOut     string    `json:"denom_out"`
	MinAmountOut math.Int  `json:"min_amount_out"` // applied to the FINAL hop's output
	Hops         []Hop     `json:"hops"`
}

func (m MsgRouteSwap) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.User); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}
	if !m.AmountIn.IsValid() || !m.AmountIn.IsPositive() {
		return errors.New("amount_in must be a valid positive coin")
	}
	if m.DenomOut == "" {
		return errors.New("denom_out required")
	}
	if len(m.Hops) == 0 || len(m.Hops) > 6 {
		return errors.New("hops must be 1..6")
	}
	// Hop chain continuity check.
	for i, h := range m.Hops {
		if i == 0 && h.AmountIn.Denom != m.AmountIn.Denom {
			return fmt.Errorf("hop[0].amount_in.denom %s != amount_in.denom %s", h.AmountIn.Denom, m.AmountIn.Denom)
		}
		if i > 0 && h.AmountIn.Denom != m.Hops[i-1].DenomOut {
			return fmt.Errorf("hop[%d] denom discontinuity", i)
		}
		if i == len(m.Hops)-1 && h.DenomOut != m.DenomOut {
			return fmt.Errorf("final hop must produce %s", m.DenomOut)
		}
	}
	return nil
}

type MsgRouteSwapResponse struct {
	RouteID    uint64    `json:"route_id"`     // non-zero only for async / IBC routes
	AmountOut  sdk.Coin  `json:"amount_out"`   // populated for synchronous routes
	Pending    bool      `json:"pending"`      // true when waiting on IBC acks
}

// Params govern router-wide settings.
type Params struct {
	// Router fee on the user's input, in basis points. Default 5 bps.
	// 100 % of this fee routes to the SKYMETRIC treasury.
	RouterFeeBps uint32 `json:"router_fee_bps"`

	// Whitelist of IBC channels the router is allowed to use. Prevents
	// griefing via untrusted / unconnected channels.
	AllowedChannels []string `json:"allowed_channels"`
}

func DefaultParams() Params {
	return Params{
		RouterFeeBps:    5,
		AllowedChannels: []string{}, // empty in genesis; governance whitelists post-IBC
	}
}

func (p Params) Validate() error {
	if p.RouterFeeBps > 100 { // hard cap 1 %
		return fmt.Errorf("router_fee_bps cannot exceed 100 (1%%)")
	}
	return nil
}
