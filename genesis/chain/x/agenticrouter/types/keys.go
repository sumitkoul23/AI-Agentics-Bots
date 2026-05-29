// Package types backs `x/agenticrouter`, the on-chain leg of the SKYMETRIC
// cross-chain aggregator (Tier 2 from genesis/docs/05-exchange-strategy.md).
//
// What lives on-chain vs off-chain:
//
//   - Route discovery & quoting → OFF-CHAIN. A Skip-Protocol-style API
//     (https://api-docs.skip.money/) computes the best route across
//     SKYMETRIC DEX, Osmosis, Astroport, etc. v0 calls the free Skip
//     endpoint; v1 ships our own quoter that respects agent-quote
//     priorities (see docs/06-financial-instruments.md §reputation
//     -weighted routing).
//
//   - Route execution → ON-CHAIN. The user signs a single MsgRouteSwap
//     containing the full N-hop plan. The keeper validates each hop,
//     executes them atomically within the tx, and reverts the whole
//     thing if any hop fails.
//
//   - Cross-chain hops → on-chain via IBC. The router submits standard
//     ICS-20 transfers + ICA messages to neighbouring chains. Asynchronous
//     legs flag a `PendingRoute` until the IBC acknowledgement comes back.
package types

import "cosmossdk.io/collections"

const (
	ModuleName   = "agenticrouter"
	StoreKey     = ModuleName
	RouterKey    = ModuleName
	QuerierRoute = ModuleName
)

var (
	ParamsKey         = collections.NewPrefix(0x00)
	PendingRoutesKey  = collections.NewPrefix(0x10) // key: route_id (uint64); awaiting IBC ack
	RouteCounterKey   = collections.NewPrefix(0x11)
	VolumeTotalKey    = collections.NewPrefix(0x20) // running notional, for fee-tier eligibility
)
