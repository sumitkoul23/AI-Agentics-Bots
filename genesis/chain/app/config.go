package app

import (
	"os"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// AppName is the SDK app name used in version strings and home-dir defaults.
	AppName = "skymetricd"

	// HumanCoinUnit is the user-facing denom (`SKY`).
	HumanCoinUnit = "SKY"

	// BaseCoinUnit is the on-chain base denom (`usky` — micro-SKY).
	BaseCoinUnit = "usky"

	// SkyExponent is the conversion 1 SKY = 10^SkyExponent usky.
	SkyExponent = 6

	// Bech32MainPrefix prefixes every account / validator / consensus address.
	// Example: sky1xyz..., skyvaloper1xyz..., skyvalcons1xyz...
	Bech32MainPrefix = "sky"

	// ChainID is the genesis chain identifier for the public network.
	ChainID = "skymetric-1"
)

// DefaultNodeHome is the default home directory for `skymetricd`.
var DefaultNodeHome = func() string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return ".skymetricd"
	}
	return filepath.Join(userHome, "."+AppName)
}()

// SetAddressPrefixes wires the Skymetric-specific bech32 prefixes into the SDK
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
