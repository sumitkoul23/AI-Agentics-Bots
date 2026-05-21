// Package agentic plugs the `x/agentic` keeper into the AGENTIC app via the
// standard Cosmos SDK `AppModule` interface set.
package agentic

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

	"github.com/sumitkoul23/agentic-chain/x/agentic/keeper"
	"github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

// ───────────────────────── AppModuleBasic ─────────────────────────

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (AppModuleBasic) RegisterInterfaces(_ cdctypes.InterfaceRegistry) {}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return mustJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, raw json.RawMessage) error {
	var gs types.GenesisState
	if err := json.Unmarshal(raw, &gs); err != nil {
		return fmt.Errorf("unmarshal %s genesis: %w", types.ModuleName, err)
	}
	return gs.Validate()
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

func (AppModuleBasic) GetTxCmd() *cobra.Command    { return nil }
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

// ───────────────────────── AppModule ─────────────────────────

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

// IsAppModule + IsOnePerModuleType are marker methods required by the
// `appmodule.AppModule` interface set introduced in SDK v0.50.
func (AppModule) IsAppModule()        {}
func (AppModule) IsOnePerModuleType() {}

func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, raw json.RawMessage) {
	var gs types.GenesisState
	mustUnmarshalJSON(raw, &gs)
	if err := am.keeper.Params.Set(ctx, gs.Params); err != nil {
		panic(err)
	}
	for _, a := range gs.AgentRecords {
		if err := am.keeper.Agents.Set(ctx, a.Operator, a); err != nil {
			panic(err)
		}
	}
	for _, t := range gs.Tasks {
		if err := am.keeper.Tasks.Set(ctx, t.ID, t); err != nil {
			panic(err)
		}
	}
	if err := am.keeper.TaskCounter.Set(ctx, gs.TaskCounter); err != nil {
		panic(err)
	}
	if err := am.keeper.BurnedTotal.Set(ctx, gs.BurnedTotal); err != nil {
		panic(err)
	}
}

func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	params, _ := am.keeper.Params.Get(ctx)
	burned, _ := am.keeper.BurnedTotal.Get(ctx)
	counter, _ := am.keeper.TaskCounter.Peek(ctx)

	gs := types.GenesisState{
		Params:      params,
		BurnedTotal: burned,
		TaskCounter: counter,
	}
	_ = am.keeper.Agents.Walk(ctx, nil, func(_ string, a types.AgentRecord) (bool, error) {
		gs.AgentRecords = append(gs.AgentRecords, a)
		return false, nil
	})
	_ = am.keeper.Tasks.Walk(ctx, nil, func(_ uint64, t types.Task) (bool, error) {
		gs.Tasks = append(gs.Tasks, t)
		return false, nil
	})
	return mustJSON(&gs)
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
