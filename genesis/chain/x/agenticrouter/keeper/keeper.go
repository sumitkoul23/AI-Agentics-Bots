package keeper

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticrouter/types"
	"github.com/sumitkoul23/agentic-chain/types/jsonvalue"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService
	authority    string

	bankKeeper      BankKeeper
	dexKeeper       DEXKeeper
	transferKeeper  IBCTransferKeeper

	Schema         collections.Schema
	Params         collections.Item[types.Params]
	PendingRoutes  collections.Map[uint64, types.PendingRoute]
	RouteCounter   collections.Sequence
	VolumeTotal    collections.Item[string]
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, from sdk.AccAddress, name string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, name string, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, from, to string, amt sdk.Coins) error
}

// DEXKeeper is the slim subset of `x/agenticdex` we call into.
type DEXKeeper interface {
	Swap(ctx sdk.Context, swapper sdk.AccAddress, poolID uint64, amountIn sdk.Coin, denomOut string, minOut math.Int) (math.Int, math.Int, error)
}

// IBCTransferKeeper is the slim subset of `x/ibc/transfer` we use. In v0
// the aggregator only handles synchronous (native-AMM-only) routes; the
// IBC paths land in v1.
type IBCTransferKeeper interface {
	// SendTransfer initiates an ICS-20 transfer. Returns the sequence
	// number which the keeper stores against the PendingRoute.
	SendTransfer(ctx sdk.Context, sourcePort, sourceChannel string, token sdk.Coin, sender sdk.AccAddress, receiver string, timeoutHeight uint64) (uint64, error)
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	authority string,
	bk BankKeeper,
	dk DEXKeeper,
	tk IBCTransferKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc: cdc, storeService: storeService, authority: authority,
		bankKeeper: bk, dexKeeper: dk, transferKeeper: tk,

		Params:        collections.NewItem(sb, types.ParamsKey, "params", jsonvalue.Codec[types.Params]()),
		PendingRoutes: collections.NewMap(sb, types.PendingRoutesKey, "pending_routes", collections.Uint64Key, jsonvalue.Codec[types.PendingRoute]()),
		RouteCounter:  collections.NewSequence(sb, types.RouteCounterKey, "route_counter"),
		VolumeTotal:   collections.NewItem(sb, types.VolumeTotalKey, "volume_total", collections.StringValue),
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
