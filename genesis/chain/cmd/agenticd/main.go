package main

import (
	"os"

	"cosmossdk.io/log"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/sumitkoul23/agentic-chain/app"
	"github.com/sumitkoul23/agentic-chain/cmd/agenticd/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "AGENTICD", app.DefaultNodeHome); err != nil {
		log.NewLogger(os.Stderr).Error("agenticd exited", "err", err)
		os.Exit(1)
	}
}
