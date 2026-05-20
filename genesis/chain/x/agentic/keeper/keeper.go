// Package keeper holds the state-transition logic for `x/agentic`. The keeper
// is the only path through which agent records, tasks, and burns are written
// — everything else goes through messages that ultimately call into this
// package.
package keeper

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

// Keeper is the AGENTIC module's state owner.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService
	authority    string // bech32 address allowed to update Params (usually `gov`)

	bankKeeper    BankKeeper
	stakingKeeper StakingKeeper

	Schema       collections.Schema
	Params       collections.Item[types.Params]
	Agents       collections.Map[string, types.AgentRecord]
	Tasks        collections.Map[uint64, types.Task]
	TaskCounter  collections.Sequence
	BurnedTotal  collections.Item[string]
}

// BankKeeper is the slim subset of `x/bank` this module uses. We accept an
// interface so tests can inject a mock.
type BankKeeper interface {
	MintCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, from sdk.AccAddress, name string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, name string, to sdk.AccAddress, amt sdk.Coins) error
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
}

// StakingKeeper is the subset of `x/staking` used for validator quorum lookup
// when validating fraud proofs.
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) sdk.Int
}

// NewKeeper wires up the module's state schema.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	authority string,
	bk BankKeeper,
	sk StakingKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:           cdc,
		storeService:  storeService,
		authority:     authority,
		bankKeeper:    bk,
		stakingKeeper: sk,

		Params:      collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Agents:      collections.NewMap(sb, types.AgentRecordsKey, "agents", collections.StringKey, codec.CollValue[types.AgentRecord](cdc)),
		Tasks:       collections.NewMap(sb, types.TasksKey, "tasks", collections.Uint64Key, codec.CollValue[types.Task](cdc)),
		TaskCounter: collections.NewSequence(sb, types.TaskCounterKey, "task_counter"),
		BurnedTotal: collections.NewItem(sb, types.BurnedTotalKey, "burned_total", collections.StringValue),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Logger returns a namespaced logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Authority returns the gov-owned address that can update Params.
func (k Keeper) Authority() string { return k.authority }
