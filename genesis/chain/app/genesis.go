package app

import (
	"encoding/json"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GenesisState is the map[moduleName] -> rawJSON expected by every module's
// `InitGenesis`.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState builds the AGENTIC chain's default genesis state and
// overlays the tokenomics constants from `docs/02-tokenomics.md`.
//
//   - 1,000,000,000 GEN total supply at genesis
//   - bond denom = "ugen"
//   - mint inflation min 1 %, max 7 %, goal-bonded 67 %
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	gen := ModuleBasics.DefaultGenesis(cdc)

	gen[stakingtypes.ModuleName] = overrideStaking(cdc, gen[stakingtypes.ModuleName])
	gen[minttypes.ModuleName] = overrideMint(cdc, gen[minttypes.ModuleName])
	gen[banktypes.ModuleName] = overrideBankMetadata(cdc, gen[banktypes.ModuleName])

	return gen
}

func overrideStaking(cdc codec.JSONCodec, raw json.RawMessage) json.RawMessage {
	var s stakingtypes.GenesisState
	cdc.MustUnmarshalJSON(raw, &s)
	s.Params.BondDenom = BaseCoinUnit
	s.Params.MaxValidators = 100
	s.Params.UnbondingTime = stakingtypes.DefaultUnbondingTime / 2 // 14 days (default is 21)
	s.Params.MinCommissionRate = math.LegacyNewDecWithPrec(5, 2)    // 5%
	return cdc.MustMarshalJSON(&s)
}

func overrideMint(cdc codec.JSONCodec, raw json.RawMessage) json.RawMessage {
	var m minttypes.GenesisState
	cdc.MustUnmarshalJSON(raw, &m)
	m.Params.MintDenom = BaseCoinUnit
	m.Params.InflationMin = math.LegacyNewDecWithPrec(1, 2)        // 1 %
	m.Params.InflationMax = math.LegacyNewDecWithPrec(7, 2)        // 7 %
	m.Params.InflationRateChange = math.LegacyNewDecWithPrec(1, 2) // 1 %
	m.Params.GoalBonded = math.LegacyNewDecWithPrec(67, 2)         // 67 %
	m.Params.BlocksPerYear = 10_512_000                            // ~3 s blocks
	return cdc.MustMarshalJSON(&m)
}

func overrideBankMetadata(cdc codec.JSONCodec, raw json.RawMessage) json.RawMessage {
	var b banktypes.GenesisState
	cdc.MustUnmarshalJSON(raw, &b)
	b.DenomMetadata = append(b.DenomMetadata, banktypes.Metadata{
		Description: "The native staking + settlement coin of the AGENTIC chain.",
		Base:        BaseCoinUnit,
		Display:     HumanCoinUnit,
		Name:        "Agentic",
		Symbol:      HumanCoinUnit,
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: BaseCoinUnit, Exponent: 0, Aliases: []string{"micro-gen"}},
			{Denom: HumanCoinUnit, Exponent: GenExponent},
		},
	})
	return cdc.MustMarshalJSON(&b)
}

// GenesisSupply returns the canonical 1B GEN total-supply Coin used by the
// initial-balances pre-seeding in the genesis JSON.
func GenesisSupply() sdk.Coin {
	return sdk.NewCoin(BaseCoinUnit, math.NewInt(1_000_000_000).MulRaw(1_000_000))
}
