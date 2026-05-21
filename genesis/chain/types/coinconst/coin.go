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
	// HumanCoinUnit is the user-facing denom ("GEN").
	HumanCoinUnit = "GEN"

	// BaseCoinUnit is the on-chain base denom ("ugen" — micro-GEN).
	BaseCoinUnit = "ugen"

	// GenExponent is the conversion 1 GEN = 10^GenExponent ugen.
	GenExponent = 6

	// Bech32MainPrefix prefixes every account / validator / consensus
	// address (e.g. "agentic1...", "agenticvaloper1...").
	Bech32MainPrefix = "agentic"
)
