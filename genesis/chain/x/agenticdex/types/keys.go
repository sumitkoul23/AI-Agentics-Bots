// Package types defines the storage keys and shared types for the
// `x/agenticdex` module — the AGENTIC chain's native on-chain AMM.
//
// Design notes:
//   - Constant-product (x*y=k) curve only at v0. Concentrated liquidity is
//     a v1 follow-up.
//   - Pool shares are minted as a bespoke denom (`pool/<id>`) and held in
//     standard bank accounts — so wallets, IBC, and explorers display LP
//     positions for free with no extra module.
//   - Swap fees are split between LPs and the protocol; the protocol slice
//     hooks into the same burn machinery as `x/agentic` for deflationary
//     pressure.
package types

import "cosmossdk.io/collections"

const (
	ModuleName   = "agenticdex"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName

	// PoolShareDenomPrefix produces denoms like `pool/42` for the LP token
	// of pool ID 42. Reusing the bank module's denom space means LPs can
	// transfer / IBC-send their position without any extra primitives.
	PoolShareDenomPrefix = "pool/"
)

var (
	ParamsKey       = collections.NewPrefix(0x00)
	PoolsKey        = collections.NewPrefix(0x10) // key: pool_id (uint64)
	PoolCounterKey  = collections.NewPrefix(0x11)
	TotalSwapFeeKey = collections.NewPrefix(0x20) // running counter for telemetry
)
