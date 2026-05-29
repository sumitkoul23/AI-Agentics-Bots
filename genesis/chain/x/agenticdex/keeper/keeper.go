// Package keeper holds the state-transition logic for `x/agenticdex`.
package keeper

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticdex/types"
	"github.com/sumitkoul23/agentic-chain/types/jsonvalue"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService
	authority    string // gov module address

	bankKeeper BankKeeper

	Schema       collections.Schema
	Params       collections.Item[types.Params]
	Pools        collections.Map[uint64, types.Pool]
	PoolCounter  collections.Sequence
	TotalSwapFee collections.Item[string] // cumulative protocol fee collected, for telemetry
}

// BankKeeper is the slim x/bank surface used by the DEX. Defined as an
// interface so tests can mock it.
type BankKeeper interface {
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, from sdk.AccAddress, name string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, name string, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, from, to string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	authority string,
	bk BankKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		bankKeeper:   bk,

		Params:       collections.NewItem(sb, types.ParamsKey, "params", jsonvalue.Codec[types.Params]()),
		Pools:        collections.NewMap(sb, types.PoolsKey, "pools", collections.Uint64Key, jsonvalue.Codec[types.Pool]()),
		PoolCounter:  collections.NewSequence(sb, types.PoolCounterKey, "pool_counter"),
		TotalSwapFee: collections.NewItem(sb, types.TotalSwapFeeKey, "total_swap_fee", collections.StringValue),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) Authority() string { return k.authority }
