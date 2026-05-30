package keeper

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/sumitkoul23/agentic-chain/x/agenticperps/types"
)

// skyPerpMarket is a SKY-PERP seeded at $1.00 mark with symmetric
// 1 trillion / 1 trillion virtual reserves.
func skyPerpMarket() types.Market {
	return types.Market{
		ID:                  "SKY-PERP",
		BaseDenom:           "usky",
		MarginDenom:         "uusdc",
		VirtualBaseReserve:  math.LegacyNewDec(1_000_000_000_000),
		VirtualQuoteReserve: math.LegacyNewDec(1_000_000_000_000),
		MaxLeverage:         math.LegacyNewDec(10),
		MaintenanceMargin:   math.LegacyNewDecWithPrec(625, 4),
		OracleSource:        "dex_twap",
	}
}

// TestMarkPrice locks in the vAMM mark price formula: quote / base.
func TestMarkPrice(t *testing.T) {
	m := skyPerpMarket()

	// Symmetric reserves → price == 1.0
	if !m.MarkPrice().Equal(math.LegacyOneDec()) {
		t.Fatalf("symmetric market: expected mark=1.0, got %s", m.MarkPrice())
	}

	// Double the quote reserve → price doubles
	m.VirtualQuoteReserve = m.VirtualQuoteReserve.MulInt64(2)
	if !m.MarkPrice().Equal(math.LegacyNewDec(2)) {
		t.Fatalf("2× quote: expected mark=2.0, got %s", m.MarkPrice())
	}

	// Zero base reserve → mark is zero (no divide-by-zero panic)
	m.VirtualBaseReserve = math.LegacyZeroDec()
	if !m.MarkPrice().IsZero() {
		t.Fatalf("zero base: expected mark=0, got %s", m.MarkPrice())
	}
}

// TestSimulateOpen locks in the constant-product vAMM invariant.
//
// For every valid open the product k = base * quote must be conserved to
// within rounding. Long positions must receive an entry price at or above
// the pre-trade mark; short positions at or below.
func TestSimulateOpen(t *testing.T) {
	cases := []struct {
		name     string
		notional int64 // positive = long, negative = short
		wantErr  bool
	}{
		{"long 1k USDC", 1_000_000_000, false},
		{"long 100k USDC (large)", 100_000_000_000, false},
		{"short 1k USDC", -1_000_000_000, false},
		{"short 100k USDC (large)", -100_000_000_000, false},
		{"zero notional errors", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := skyPerpMarket()
			mark := m.MarkPrice()
			k := m.VirtualBaseReserve.Mul(m.VirtualQuoteReserve)

			n := math.LegacyNewDec(tc.notional)
			entry, delta, err := m.SimulateOpen(n)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got entry=%s delta=%s", entry, delta)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify constant-product: newBase * newQuote ≈ k.
			// In all cases: new reserves = (base - delta, quote + notional).
			newBase := m.VirtualBaseReserve.Sub(delta)
			newQuote := m.VirtualQuoteReserve.Add(n)
			kNew := newBase.Mul(newQuote)
			diff := k.Sub(kNew).Abs()
			// Allow up to 1 unit of rounding error.
			if diff.GT(math.LegacyNewDec(1)) {
				t.Errorf("constant-product violated: k=%s kNew=%s diff=%s", k, kNew, diff)
			}

			// Entry price must be positive.
			if !entry.IsPositive() {
				t.Errorf("entry price must be positive, got %s", entry)
			}

			// Long: slippage drives entry above pre-trade mark.
			// Short: slippage drives entry below pre-trade mark.
			if n.IsPositive() && entry.LT(mark) {
				t.Errorf("long entry %s below mark %s", entry, mark)
			}
			if n.IsNegative() && entry.GT(mark) {
				t.Errorf("short entry %s above mark %s", entry, mark)
			}
		})
	}
}

// TestDefaultPerpsParamsValidate asserts the documented defaults satisfy
// every module invariant enforced at upgrade time.
func TestDefaultPerpsParamsValidate(t *testing.T) {
	p := types.DefaultParams()
	if err := p.Validate(); err != nil {
		t.Fatalf("default params must validate: %v", err)
	}

	// Liquidator bounty + insurance fund share must sum to 1.0.
	sum := p.LiquidatorBounty.Add(p.InsuranceFundShare)
	if !sum.Equal(math.LegacyOneDec()) {
		t.Errorf("liquidator_bounty + insurance_fund_share = %s, want 1.0", sum)
	}

	// 8-hour funding period at 3-second blocks = 9 600 blocks.
	if p.BlocksPerFundingPeriod != 9600 {
		t.Errorf("expected 9600 blocks (8h @ 3s), got %d", p.BlocksPerFundingPeriod)
	}

	// Zero-notional opening fee is a regression guard: the fee is charged on
	// every open, so a negative or >= 100% value would brick the module.
	if p.OpenFee.IsNegative() || !p.OpenFee.LT(math.LegacyOneDec()) {
		t.Errorf("open_fee out of [0,1): %s", p.OpenFee)
	}
}
