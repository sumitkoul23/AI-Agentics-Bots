package agenticperps

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/sumitkoul23/agentic-chain/x/agenticperps/keeper"
	"github.com/sumitkoul23/agentic-chain/x/agenticperps/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string                                                    { return types.ModuleName }
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino)                     {}
func (AppModuleBasic) RegisterInterfaces(cdctypes.InterfaceRegistry)                   {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)     {}
func (AppModuleBasic) GetTxCmd() *cobra.Command                                        { return nil }
func (AppModuleBasic) GetQueryCmd() *cobra.Command                                     { return nil }
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return mustJSON(&GenesisState{Params: types.DefaultParams()})
}
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, raw json.RawMessage) error {
	var gs GenesisState
	if err := json.Unmarshal(raw, &gs); err != nil {
		return fmt.Errorf("%s: %w", types.ModuleName, err)
	}
	return gs.Params.Validate()
}

// GenesisState is the perp module's genesis JSON.
type GenesisState struct {
	Params       types.Params   `json:"params"`
	Markets      []types.Market `json:"markets"`
	InsuranceFund string        `json:"insurance_fund"`
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule { return AppModule{keeper: k} }

func (AppModule) IsAppModule()        {}
func (AppModule) IsOnePerModuleType() {}

func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, raw json.RawMessage) {
	var gs GenesisState
	mustUnmarshalJSON(raw, &gs)
	if err := am.keeper.Params.Set(ctx, gs.Params); err != nil {
		panic(err)
	}
	for _, m := range gs.Markets {
		if err := am.keeper.Markets.Set(ctx, m.ID, m); err != nil {
			panic(err)
		}
	}
	if gs.InsuranceFund == "" {
		gs.InsuranceFund = "0"
	}
	if err := am.keeper.InsuranceFund.Set(ctx, gs.InsuranceFund); err != nil {
		panic(err)
	}
}

func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	p, _ := am.keeper.Params.Get(ctx)
	fund, _ := am.keeper.InsuranceFund.Get(ctx)
	gs := GenesisState{Params: p, InsuranceFund: fund}
	_ = am.keeper.Markets.Walk(ctx, nil, func(_ string, m types.Market) (bool, error) {
		gs.Markets = append(gs.Markets, m)
		return false, nil
	})
	return mustJSON(&gs)
}

// EndBlock accrues funding for every active market once per block. This is
// the only place the cumulative funding index moves; trader-side settlement
// happens lazily on next interaction in the keeper.
func (am AppModule) EndBlock(ctx context.Context) error {
	return am.keeper.Markets.Walk(ctx, nil, func(_ string, m types.Market) (bool, error) {
		if m.Paused {
			return false, nil
		}
		if err := am.keeper.AccrueFundingForMarket(ctx, m); err != nil {
			return true, err
		}
		return false, nil
	})
}

// mustJSON marshals via encoding/json and panics on failure — the v0 stand-in
// for cdc.MustMarshalJSON until proto-gen produces typed messages.
func mustJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// mustUnmarshalJSON is the inverse of mustJSON.
func mustUnmarshalJSON(b []byte, v interface{}) {
	if err := json.Unmarshal(b, v); err != nil {
		panic(err)
	}
}
