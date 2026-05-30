package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticrouter/types"
)

// TestDefaultRouterParamsValidate asserts the documented defaults satisfy
// all governance invariants enforced at upgrade time.
func TestDefaultRouterParamsValidate(t *testing.T) {
	p := types.DefaultParams()
	if err := p.Validate(); err != nil {
		t.Fatalf("default params must validate: %v", err)
	}

	// Default fee is 5 bps.
	if p.RouterFeeBps != 5 {
		t.Errorf("expected RouterFeeBps=5, got %d", p.RouterFeeBps)
	}

	// 5 bps of 1_000_000 (1 USDC with 6dp) = 50 µUSDC.
	amount := math.NewInt(1_000_000)
	fee := amount.MulRaw(int64(p.RouterFeeBps)).QuoRaw(10_000)
	if !fee.Equal(math.NewInt(50)) {
		t.Errorf("5 bps of 1_000_000 = %s, want 50", fee)
	}
}

// TestRouterFeeCapValidation checks that fee params exceeding 1% are rejected
// and the boundary value (100 bps exactly) is accepted.
func TestRouterFeeCapValidation(t *testing.T) {
	cases := []struct {
		bps     uint32
		wantErr bool
	}{
		{0, false},
		{5, false},
		{100, false},
		{101, true},
		{1000, true},
	}
	for _, tc := range cases {
		p := types.Params{RouterFeeBps: tc.bps}
		err := p.Validate()
		if tc.wantErr && err == nil {
			t.Errorf("bps=%d: expected error, got nil", tc.bps)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("bps=%d: unexpected error: %v", tc.bps, err)
		}
	}
}

// TestValidateBasicHopChain locks in MsgRouteSwap.ValidateBasic's denom
// continuity, hop-count, and final-denom assertions.
func TestValidateBasicHopChain(t *testing.T) {
	// 20-byte zero address — valid bech32 format used only for unit tests.
	user := sdk.AccAddress(make([]byte, 20)).String()

	cases := []struct {
		name    string
		msg     types.MsgRouteSwap
		wantErr bool
	}{
		{
			name: "valid single-hop",
			msg: types.MsgRouteSwap{
				User:         user,
				AmountIn:     sdk.NewCoin("usky", math.NewInt(1_000_000)),
				DenomOut:     "uusdc",
				MinAmountOut: math.NewInt(990_000),
				Hops: []types.Hop{
					{Kind: types.HopKindNativeAMM, AmountIn: sdk.NewCoin("usky", math.NewInt(1_000_000)), DenomOut: "uusdc"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid two-hop",
			msg: types.MsgRouteSwap{
				User:         user,
				AmountIn:     sdk.NewCoin("usky", math.NewInt(1_000_000)),
				DenomOut:     "uatom",
				MinAmountOut: math.NewInt(1),
				Hops: []types.Hop{
					{Kind: types.HopKindNativeAMM, AmountIn: sdk.NewCoin("usky", math.NewInt(1_000_000)), DenomOut: "uusdc"},
					{Kind: types.HopKindNativeAMM, AmountIn: sdk.NewCoin("uusdc", math.NewInt(0)), DenomOut: "uatom"},
				},
			},
			wantErr: false,
		},
		{
			name: "denom discontinuity between hops",
			msg: types.MsgRouteSwap{
				User:         user,
				AmountIn:     sdk.NewCoin("usky", math.NewInt(1_000_000)),
				DenomOut:     "uusdc",
				MinAmountOut: math.NewInt(1),
				Hops: []types.Hop{
					{Kind: types.HopKindNativeAMM, AmountIn: sdk.NewCoin("usky", math.NewInt(1_000_000)), DenomOut: "uatom"},
					// hop[1].AmountIn.Denom should be "uatom" (output of hop[0]), not "uusdc"
					{Kind: types.HopKindNativeAMM, AmountIn: sdk.NewCoin("uusdc", math.NewInt(0)), DenomOut: "uusdc"},
				},
			},
			wantErr: true,
		},
		{
			name: "zero hops rejected",
			msg: types.MsgRouteSwap{
				User:         user,
				AmountIn:     sdk.NewCoin("usky", math.NewInt(1_000_000)),
				DenomOut:     "uusdc",
				MinAmountOut: math.NewInt(1),
				Hops:         []types.Hop{},
			},
			wantErr: true,
		},
		{
			name: "seven hops exceeds limit",
			msg: types.MsgRouteSwap{
				User:         user,
				AmountIn:     sdk.NewCoin("usky", math.NewInt(1_000_000)),
				DenomOut:     "uusdc",
				MinAmountOut: math.NewInt(1),
				Hops:         make([]types.Hop, 7),
			},
			wantErr: true,
		},
		{
			name: "first hop denom mismatch with amount_in",
			msg: types.MsgRouteSwap{
				User:         user,
				AmountIn:     sdk.NewCoin("usky", math.NewInt(1_000_000)),
				DenomOut:     "uusdc",
				MinAmountOut: math.NewInt(1),
				Hops: []types.Hop{
					{Kind: types.HopKindNativeAMM, AmountIn: sdk.NewCoin("uatom", math.NewInt(1_000_000)), DenomOut: "uusdc"},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
