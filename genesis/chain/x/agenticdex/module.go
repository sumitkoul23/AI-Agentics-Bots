// Package agenticdex wires `x/agenticdex` into the AGENTIC app via the
// standard Cosmos SDK AppModule interfaces.
package agenticdex

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

	"github.com/sumitkoul23/agentic-chain/x/agenticdex/keeper"
	"github.com/sumitkoul23/agentic-chain/x/agenticdex/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

// ───────────────────────── AppModuleBasic ─────────────────────────

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string                                           { return types.ModuleName }
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino)          {}
func (AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry)        {}
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}
func (AppModuleBasic) GetTxCmd() *cobra.Command                               { return nil }
func (AppModuleBasic) GetQueryCmd() *cobra.Command                            { return nil }

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&types.GenesisState{Params: types.DefaultParams()})
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, raw json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(raw, &gs); err != nil {
		return fmt.Errorf("unmarshal %s genesis: %w", types.ModuleName, err)
	}
	return gs.Params.Validate()
}

// ───────────────────────── AppModule ─────────────────────────

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

func (AppModule) IsAppModule()        {}
func (AppModule) IsOnePerModuleType() {}

func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, raw json.RawMessage) {
	var gs types.GenesisState
	cdc.MustUnmarshalJSON(raw, &gs)
	if err := am.keeper.Params.Set(ctx, gs.Params); err != nil {
		panic(err)
	}
	for _, p := range gs.Pools {
		if err := am.keeper.Pools.Set(ctx, p.ID, p); err != nil {
			panic(err)
		}
	}
	if err := am.keeper.PoolCounter.Set(ctx, gs.PoolCounter); err != nil {
		panic(err)
	}
}

func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	params, _ := am.keeper.Params.Get(ctx)
	counter, _ := am.keeper.PoolCounter.Peek(ctx)
	gs := types.GenesisState{Params: params, PoolCounter: counter}
	_ = am.keeper.Pools.Walk(ctx, nil, func(_ uint64, p types.Pool) (bool, error) {
		gs.Pools = append(gs.Pools, p)
		return false, nil
	})
	return cdc.MustMarshalJSON(&gs)
}
