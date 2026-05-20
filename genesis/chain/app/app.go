// Package app wires together the modules that make up the AGENTIC chain.
//
// The Genesis System's v0 module set is deliberately conservative — the
// standard Cosmos SDK suite (auth, bank, staking, slashing, distribution,
// gov, mint, ibc) plus the bespoke `x/agentic` module that ships the
// agent-staking primitive described in `docs/01-architecture.md`.
//
// This file is intentionally a slim wiring layer; the heavy lifting lives in
// the Cosmos SDK's `runtime` package and the per-module keepers.
package app

import (
	"io"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sumitkoul23/agentic-chain/x/agentic"
	agentictypes "github.com/sumitkoul23/agentic-chain/x/agentic/types"
)

// ModuleBasics is the canonical list of modules exposed at the CLI / genesis
// level. New modules must be added here AND wired in `New(...)` below.
var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distribution.AppModuleBasic{},
	gov.AppModuleBasic{},
	slashing.AppModuleBasic{},
	genutil.AppModuleBasic{},
	agentic.AppModuleBasic{},
)

// maccPerms lists the module-account permissions. `x/agentic` gets `minter`
// + `burner` because agent slashing burns stake and reputation rewards mint.
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	distrtypes.ModuleName:          nil,
	minttypes.ModuleName:           {authtypes.Minter},
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:            {authtypes.Burner},
	slashingtypes.ModuleName:       nil,
	agentictypes.ModuleName:        {authtypes.Minter, authtypes.Burner},
}

// AgenticApp is the concrete `servertypes.Application` for this chain.
//
// In v0 we deliberately reuse `runtime.App` (the SDK's "app-wiring" composer)
// rather than rolling our own BaseApp wiring — it cuts ~400 LOC of boilerplate
// for free.
type AgenticApp struct {
	*runtime.App
}

// New constructs the AGENTIC application and registers every module.
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *AgenticApp {
	app := &AgenticApp{}
	app.App = runtime.NewAppBuilder(MakeEncodingConfig().Codec).Build(logger, db, traceStore, baseAppOptions...)

	if loadLatest {
		if err := app.Load(loadLatest); err != nil {
			panic(err)
		}
	}
	return app
}

// LoadHeight loads the app at a specific height (used by `appExport`).
func (a *AgenticApp) LoadHeight(height int64) error {
	return a.LoadHeightForStore(height, a.GetKey(storetypes.NewKVStoreKey(banktypes.StoreKey).Name()))
}
