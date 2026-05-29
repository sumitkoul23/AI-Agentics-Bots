package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Params govern the `x/agentic` module's economic knobs. Changeable via a
// standard `MsgUpdateParams` gov proposal.
type Params struct {
	// MinAgentStake is the floor (in usky) an operator must bond to register
	// an agent. Halves with every +100 reputation points until it hits the
	// `MinAgentStakeFloor` below.
	MinAgentStake     math.Int `json:"min_agent_stake"`
	MinAgentStakeFloor math.Int `json:"min_agent_stake_floor"`

	// SplitAgent / SplitValidators / SplitBurn are the three slices of a
	// successful task escrow. They must sum to 1.0 exactly.
	SplitAgent      math.LegacyDec `json:"split_agent"`
	SplitValidators math.LegacyDec `json:"split_validators"`
	SplitBurn       math.LegacyDec `json:"split_burn"`

	// FraudProofQuorum is the number of validator-signed attestations
	// required to slash an agent. Defaults to ⌈2/3 * activeValidators⌉.
	FraudProofQuorum uint32 `json:"fraud_proof_quorum"`

	// ReputationGainPerTask is the integer reputation increment awarded on a
	// successfully-settled task.
	ReputationGainPerTask uint64 `json:"reputation_gain_per_task"`
}

// DefaultParams returns the genesis-level parameters. Values come straight
// from `genesis/docs/02-tokenomics.md`.
func DefaultParams() Params {
	return Params{
		MinAgentStake:         math.NewInt(100).MulRaw(1_000_000),  // 100 SKY
		MinAgentStakeFloor:    math.NewInt(10).MulRaw(1_000_000),   // 10 SKY
		SplitAgent:            math.LegacyNewDecWithPrec(50, 2),    // 50%
		SplitValidators:       math.LegacyNewDecWithPrec(30, 2),    // 30%
		SplitBurn:             math.LegacyNewDecWithPrec(20, 2),    // 20%
		FraudProofQuorum:      3,                                   // bootstrap value, raised by gov post-launch
		ReputationGainPerTask: 1,
	}
}

// Validate checks the params are internally consistent.
func (p Params) Validate() error {
	if !p.MinAgentStake.IsPositive() {
		return fmt.Errorf("min_agent_stake must be positive")
	}
	if p.MinAgentStakeFloor.GT(p.MinAgentStake) {
		return fmt.Errorf("min_agent_stake_floor cannot exceed min_agent_stake")
	}
	sum := p.SplitAgent.Add(p.SplitValidators).Add(p.SplitBurn)
	if !sum.Equal(math.LegacyOneDec()) {
		return fmt.Errorf("split shares must sum to 1.0, got %s", sum)
	}
	if p.FraudProofQuorum == 0 {
		return fmt.Errorf("fraud_proof_quorum must be at least 1")
	}
	return nil
}
