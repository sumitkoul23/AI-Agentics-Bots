package agenticrouter

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

	"github.com/sumitkoul23/agentic-chain/x/agenticrouter/keeper"
	"github.com/sumitkoul23/agentic-chain/x/agenticrouter/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string                                                { return types.ModuleName }
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino)                 {}
func (AppModuleBasic) RegisterInterfaces(cdctypes.InterfaceRegistry)               {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}
func (AppModuleBasic) GetTxCmd() *cobra.Command                                    { return nil }
func (AppModuleBasic) GetQueryCmd() *cobra.Command                                 { return nil }
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

type GenesisState struct {
	Params types.Params `json:"params"`
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
}

func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	p, _ := am.keeper.Params.Get(ctx)
	return mustJSON(&GenesisState{Params: p})
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
