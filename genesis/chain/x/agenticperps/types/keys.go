// Package types defines storage keys and shared types for `x/agenticperps`,
// the AGENTIC chain's native perpetual-futures module.
//
// Design — virtual-AMM (vAMM) à la dYdX v1 / Perpetual Protocol v1.
//
//   - Each market has a *virtual* x*y=k curve that quotes prices but
//     holds no real reserves. Real collateral lives in module accounts.
//   - Traders deposit USDC (or any whitelisted margin denom) and open
//     long/short positions. PnL is settled by the funding-rate mechanism
//     against the spot index price (sourced from the `x/agenticdex` TWAP
//     of the corresponding pool, or an external oracle once IBC opens).
//   - An insurance fund covers shortfalls from bad liquidations. Funded
//     by 50 % of liquidation penalties; the other 50 % goes to the
//     liquidator as a keeper bounty.
//
// Why vAMM not CLOB at v0:
//   - CLOB requires off-chain matching infra (Tier 3) — not yet built.
//   - vAMM ships with zero off-chain dependencies, matching the project's
//     $0-budget thesis.
//   - dYdX migrated *away* from on-chain CLOB to Cosmos for exactly this
//     reason; we start where they ended up.
package types

import "cosmossdk.io/collections"

const (
	ModuleName   = "agenticperps"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName

	// InsuranceFundName is the sub-account that holds the insurance fund.
	InsuranceFundName = "insurance_fund"
)

var (
	ParamsKey         = collections.NewPrefix(0x00)
	MarketsKey        = collections.NewPrefix(0x10) // key: market_id (string, e.g. "GEN-PERP")
	PositionsKey      = collections.NewPrefix(0x20) // key: (market_id, trader)
	PositionsByMarket = collections.NewPrefix(0x21) // secondary index
	FundingIndexKey   = collections.NewPrefix(0x30) // key: market_id → cumulative funding index
	InsuranceFundKey  = collections.NewPrefix(0x40) // balance counter
)
