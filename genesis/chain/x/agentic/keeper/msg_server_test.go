// Unit tests for the keeper's Msg handlers. These are illustrative — they
// document the expected state transitions and rounding rules even before
// the full proto pipeline is wired up. A complete suite plugs into the
// standard `simapp` harness; the SDK provides `testutil.NewTestKeeper` for
// that purpose.
//
// Run with: go test ./x/agentic/keeper/...
package keeper

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

// TestSettleSplitMath asserts that the 50/30/20 split rounds correctly and
// the rounding dust always flows to the burn slice (never to the agent or
// validators). This invariant is what makes the burn deflationary even at
// odd bounty sizes.
func TestSettleSplitMath(t *testing.T) {
	cases := []struct {
		name              string
		bounty            int64
		wantAgent, wantVal, wantBurn int64
	}{
		{"clean 100", 100, 50, 30, 20},
		{"clean 1000", 1000, 500, 300, 200},
		{"odd 101 — dust to burn", 101, 50, 30, 21},
		{"odd 7 — dust to burn", 7, 3, 2, 2},
		{"minimum 1 — all to burn", 1, 0, 0, 1},
	}

	params := types.DefaultParams() // 50/30/20

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := math.NewInt(tc.bounty)
			agentCut := b.ToLegacyDec().Mul(params.SplitAgent).TruncateInt()
			valCut := b.ToLegacyDec().Mul(params.SplitValidators).TruncateInt()
			burnCut := b.Sub(agentCut).Sub(valCut)

			if agentCut.Int64() != tc.wantAgent {
				t.Errorf("agent: got %s want %d", agentCut, tc.wantAgent)
			}
			if valCut.Int64() != tc.wantVal {
				t.Errorf("val:   got %s want %d", valCut, tc.wantVal)
			}
			if burnCut.Int64() != tc.wantBurn {
				t.Errorf("burn:  got %s want %d", burnCut, tc.wantBurn)
			}
			// Conservation: the slices must sum exactly to the bounty.
			if agentCut.Add(valCut).Add(burnCut).Int64() != tc.bounty {
				t.Errorf("split does not conserve bounty")
			}
		})
	}
}

// TestParamsValidate documents the invariants enforced on every gov-driven
// MsgUpdateParams.
func TestParamsValidate(t *testing.T) {
	t.Run("defaults are valid", func(t *testing.T) {
		if err := types.DefaultParams().Validate(); err != nil {
			t.Fatalf("default params should validate: %v", err)
		}
	})

	t.Run("splits must sum to 1", func(t *testing.T) {
		p := types.DefaultParams()
		p.SplitBurn = math.LegacyNewDecWithPrec(10, 2) // breaks the sum
		if err := p.Validate(); err == nil {
			t.Fatalf("expected validation error for non-unit split")
		}
	})

	t.Run("floor cannot exceed min", func(t *testing.T) {
		p := types.DefaultParams()
		p.MinAgentStakeFloor = p.MinAgentStake.AddRaw(1)
		if err := p.Validate(); err == nil {
			t.Fatalf("expected validation error for inverted bounds")
		}
	})
}
