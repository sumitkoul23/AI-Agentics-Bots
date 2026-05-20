package app

import (
	"os"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// AppName is the SDK app name used in version strings and home-dir defaults.
	AppName = "agenticd"

	// HumanCoinUnit is the user-facing denom (`GEN`).
	HumanCoinUnit = "GEN"

	// BaseCoinUnit is the on-chain base denom (`ugen` — micro-GEN).
	BaseCoinUnit = "ugen"

	// GenExponent is the conversion 1 GEN = 10^GenExponent ugen.
	GenExponent = 6

	// Bech32MainPrefix prefixes every account / validator / consensus address.
	// Example: agentic1xyz..., agenticvaloper1xyz..., agenticvalcons1xyz...
	Bech32MainPrefix = "agentic"

	// ChainID is the genesis chain identifier for the public network.
	ChainID = "agentic-1"
)

// DefaultNodeHome is the default home directory for `agenticd`.
var DefaultNodeHome = func() string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return ".agenticd"
	}
	return filepath.Join(userHome, "."+AppName)
}()

// SetAddressPrefixes wires the agentic-specific bech32 prefixes into the SDK
// global config. Called once at process start before any address is parsed.
func SetAddressPrefixes() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(Bech32MainPrefix, Bech32MainPrefix+"pub")
	cfg.SetBech32PrefixForValidator(Bech32MainPrefix+"valoper", Bech32MainPrefix+"valoperpub")
	cfg.SetBech32PrefixForConsensusNode(Bech32MainPrefix+"valcons", Bech32MainPrefix+"valconspub")
	cfg.SetCoinType(118) // shared with the Cosmos ecosystem for wallet interop
	cfg.Seal()
}

func init() {
	SetAddressPrefixes()
}
