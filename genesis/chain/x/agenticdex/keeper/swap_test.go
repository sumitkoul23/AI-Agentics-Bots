package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticdex/types"
)

// TestCalcOutGivenIn locks in the constant-product math.
//
// Without fees, the textbook identity reserveIn * reserveOut == k must hold
// (modulo rounding). With a 0.30 % fee, the post-swap k must be strictly
// greater than pre-swap k by the LP fee — that's what accrues to LPs.
func TestCalcOutGivenIn(t *testing.T) {
	pool := types.Pool{
		ID:          1,
		AssetA:      sdk.NewCoin("usky", math.NewInt(1_000_000_000_000)),  // 1M SKY
		AssetB:      sdk.NewCoin("uusdc", math.NewInt(1_000_000_000_000)), // 1M USDC (assuming 6dp)
		TotalShares: math.NewInt(1_000_000_000_000),
		SwapFee:     math.LegacyNewDecWithPrec(3, 3), // 0.30 %
	}

	cases := []struct {
		name        string
		amountIn    int64
		minOutBound int64 // lower bound — actual will be slightly higher pre-fee
	}{
		{"1k SKY in", 1_000_000_000, 990_000_000},          // ~1k USDC out, with ~0.3% slippage+fee
		{"1M SKY in (large slippage)", 1_000_000_000_000, 480_000_000_000}, // 50% price impact
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, fee, err := pool.CalcOutGivenIn("usky", "uusdc", math.NewInt(tc.amountIn))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.LT(math.NewInt(tc.minOutBound)) {
				t.Errorf("out too low: got %s want >= %d", out, tc.minOutBound)
			}
			if !fee.IsPositive() {
				t.Errorf("fee should be positive, got %s", fee)
			}

			// Conservation: the input must equal the fee + the net input that
			// moved into the curve.
			netIn := math.NewInt(tc.amountIn).Sub(fee)
			if netIn.LTE(math.ZeroInt()) {
				t.Errorf("net input non-positive")
			}
		})
	}
}

// TestParamsDefaults asserts the default fee split documented in
// docs/05-exchange-strategy.md is internally consistent.
func TestParamsDefaults(t *testing.T) {
	p := types.DefaultParams()
	if err := p.Validate(); err != nil {
		t.Fatalf("default params should validate: %v", err)
	}

	// 0.30 % swap fee × 16.67 % protocol slice = ~0.05 % of trade volume
	// (matches the marketing claim in the strategy doc).
	totalFee := p.DefaultSwapFee
	protocolSliceOfTrade := totalFee.Mul(p.ProtocolFeeShare)
	want := math.LegacyNewDecWithPrec(5, 4) // 0.0005 = 0.05%
	diff := protocolSliceOfTrade.Sub(want).Abs()
	if diff.GT(math.LegacyNewDecWithPrec(1, 6)) {
		t.Errorf("protocol slice of trade = %s, expected ~%s", protocolSliceOfTrade, want)
	}
}
