// Package coinconst holds the shared denom constants used by every module
// keeper. Lives in a leaf package (no internal imports) so it can be
// referenced by app/, x/agentic/, x/agenticdex/, x/agenticperps/, and
// x/agenticrouter/ without creating an import cycle.
//
// Mirrors the values declared in app/config.go — keep both in sync, or
// (better) make app/config.go re-export from here once we drop v0
// compatibility shims.
package coinconst

const (
	// HumanCoinUnit is the user-facing denom ("SKY").
	HumanCoinUnit = "SKY"

	// BaseCoinUnit is the on-chain base denom ("usky" — micro-SKY).
	BaseCoinUnit = "usky"

	// SkyExponent is the conversion 1 SKY = 10^SkyExponent usky.
	SkyExponent = 6

	// Bech32MainPrefix prefixes every account / validator / consensus
	// address on the Skymetric chain (e.g. "sky1...", "skyvaloper1...").
	Bech32MainPrefix = "sky"
)
