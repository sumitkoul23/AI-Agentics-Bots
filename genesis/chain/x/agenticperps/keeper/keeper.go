package keeper

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sumitkoul23/agentic-chain/x/agenticperps/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService storetypes.KVStoreService
	authority    string

	bankKeeper  BankKeeper
	priceKeeper PriceKeeper

	Schema        collections.Schema
	Params        collections.Item[types.Params]
	Markets       collections.Map[string, types.Market]
	Positions     collections.Map[collections.Pair[string, string], types.Position] // (market, trader)
	FundingIndex  collections.Map[string, string]                                    // market_id → cumulative funding dec
	InsuranceFund collections.Item[string]                                           // running USDC balance
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, from sdk.AccAddress, name string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, name string, to sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// PriceKeeper abstracts the index-price source. v0 uses the
// `x/agenticdex` TWAP of the corresponding spot pool; later we can plug
// IBC oracles or external feeds.
type PriceKeeper interface {
	IndexPrice(ctx sdk.Context, marketID string) (price string, ok bool)
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	authority string,
	bk BankKeeper,
	pk PriceKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc: cdc, storeService: storeService, authority: authority,
		bankKeeper: bk, priceKeeper: pk,

		Params:        collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Markets:       collections.NewMap(sb, types.MarketsKey, "markets", collections.StringKey, codec.CollValue[types.Market](cdc)),
		Positions:     collections.NewMap(sb, types.PositionsKey, "positions", collections.PairKeyCodec(collections.StringKey, collections.StringKey), codec.CollValue[types.Position](cdc)),
		FundingIndex:  collections.NewMap(sb, types.FundingIndexKey, "funding_index", collections.StringKey, collections.StringValue),
		InsuranceFund: collections.NewItem(sb, types.InsuranceFundKey, "insurance_fund", collections.StringValue),
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
