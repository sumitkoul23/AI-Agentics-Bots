package cmd

import (
	"errors"
	"io"
	"os"

	cmtcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	"github.com/sumitkoul23/agentic-chain/app"
)

// NewRootCmd returns the root command for the `skymetricd` binary.
//
// Most of the wiring here is the standard Cosmos SDK root command; the
// SKYMETRIC-specific bits are:
//
//   - `Bech32MainPrefix = "agentic"` (set in `app/config.go`)
//   - genesis pre-seeding of the validator-pool, treasury, and airdrop buckets
//     via `genutilcli.InitCmd` which reads `app.NewDefaultGenesisState()`.
func NewRootCmd() *cobra.Command {
	encCfg := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		WithInterfaceRegistry(encCfg.InterfaceRegistry).
		WithTxConfig(encCfg.TxConfig).
		WithLegacyAmino(encCfg.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("AGENTICD")

	rootCmd := &cobra.Command{
		Use:   "skymetricd",
		Short: "Skymetric chain — the settlement layer for AI agents",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}
			return server.InterceptConfigsPreRunHandler(cmd, "", nil, cmtcfg.DefaultConfig())
		},
	}

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.GenTxCmd(app.ModuleBasics, encCfg.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome, nil, encCfg.TxConfig.SigningContext().ValidatorAddressCodec()),
		genutilcli.ValidateGenesisCmd(app.ModuleBasics),
		genutilcli.AddGenesisAccountCmd(app.DefaultNodeHome),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		keys.Commands(),
		rpc.StatusCommand(),
		queryCommand(),
		txCommand(),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, addModuleInitFlags)

	return rootCmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "query", Aliases: []string{"q"}, Short: "Querying subcommands"}
	cmd.AddCommand(
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)
	app.ModuleBasics.AddQueryCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "skymetric-1", "The network chain ID")
	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "tx", Short: "Transactions subcommands"}
	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)
	app.ModuleBasics.AddTxCommands(cmd)
	cmd.PersistentFlags().String(flags.FlagChainID, "skymetric-1", "The network chain ID")
	return cmd
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)
	return app.New(
		logger, db, traceStore, true,
		appOpts,
		baseappOptions...,
	)
}

func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}
	loadLatest := height == -1
	a := app.New(logger, db, traceStore, loadLatest, appOpts)
	if !loadLatest {
		if err := a.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}
	return a.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// guard against unused imports during the initial scaffold
var (
	_ = sdk.NewInt
	_ = simtestutil.NewAppOptionsWithFlagHome
	_ = viper.New
)
